package api

import (
	"database/sql"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/samiulice/photostock/internal/models"
	"github.com/samiulice/photostock/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// --- Authentication & User Management ---
// Register handles the process of registering new users
func (app *application) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR: SignUp => unable to read json %w", err))
		return
	}
	//sanitize user input
	user.Email = strings.TrimSpace(user.Email)
	user.Username = strings.Split(user.Email, "@")[0] + "_" + app.GenerateRandomAlphanumericCode(4)
	user.Name = strings.TrimSpace(user.Name)
	user.Status = true
	user.Role = "user"
	user.Password = strings.TrimSpace(user.Password)
	user.AvatarID = strings.Split(user.Email, "@")[0] + "_" + uuid.NewString() + ".jpg"
	//hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		app.errorLog.Println("ERROR: SignUp => unable to hash password:", err)
		app.badRequest(w, fmt.Errorf("Internal Server Error: unable to hash password %w", err))
		return
	}
	user.Password = string(hashedPassword)

	err = app.DB.UserRepo.Create(r.Context(), &user)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			app.errorLog.Println("ERROR: SignUp => Email already associated with another account:", err)
			app.badRequest(w, errors.New("Email already associated with another account"))
			return
		} else {
			app.errorLog.Println("ERROR: SignUp => unable to create user:", err)
			app.badRequest(w, errors.New("Internal Server Error: Unable to create user"))
			return
		}
	}

	//sanitize input
	user.Password = ""
	//after adding user successfully, go to the login process
	//Generate signed token
	token, err := generateSignedToken(&user)
	if err != nil {
		app.errorLog.Println("ERROR: SignUp => Unable to generate token for user: ", user.Username)
		app.badRequest(w, errors.New("Internal server error"))
		return
	}
	// Remove sensitive data before sending response
	user.Password = ""

	// Prepare and send response
	response := struct {
		Error   bool         `json:"error"`
		Message string       `json:"message"`
		Token   string       `json:"token"`
		User    *models.User `json:"user"`
	}{
		Error:   false,
		Message: "Registration successful. Auto sign-in complete.",
		Token:   token,
		User:    &user,
	}

	app.writeJSON(w, http.StatusOK, response)
}

// generateSignedToken generate a token string for implementing JWT
func generateSignedToken(user *models.User) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"id":       user.ID,
		"name":     user.Name,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"iss":      app.config.jwt.issuer,
		"aud":      app.config.jwt.audience,
		"exp":      time.Now().Add(app.config.jwt.expiry).Unix(),
		"iat":      time.Now().Unix(),
	}

	// Sign the token with the secret key
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(app.config.jwt.secretKey))
	return tokenString, err
}

// Login authenticates the user and generates a JWT token for them.
// This function is used for the new authentication system using JWT.
func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	var user models.User

	// Decode JSON credentials
	if err := app.readJSON(w, r, &user); err != nil {
		app.errorLog.Println("ERROR: Unable to read JSON -", err)
		app.badRequest(w, errors.New("Failed to read username and password"))
		return
	}

	// Lookup user by username
	validUser, err := app.DB.UserRepo.GetByEmail(r.Context(), user.Email)
	if err != nil {
		app.errorLog.Println("ERROR: User lookup failed -", err)
		if errors.Is(err, sql.ErrNoRows) {
			app.badRequest(w, errors.New("Wrong username or password"))
		} else {
			app.badRequest(w, errors.New("Failed to retrieve user"+err.Error()))
		}
		return
	}

	// Check if user account is active
	if !validUser.Status {
		app.infoLog.Printf("Login denied for deactivated user: %s", user.Email)
		app.badRequest(w, errors.New("Account deactivated. Please contact support"))
		return
	}

	// Compare password hash
	if err := bcrypt.CompareHashAndPassword([]byte(validUser.Password), []byte(user.Password)); err != nil {
		app.errorLog.Printf("ERROR: Password mismatch for user: %s", user.Username)
		app.badRequest(w, errors.New("Wrong username or password"))
		return
	}

	//Generate signed token
	token, err := generateSignedToken(validUser)
	if err != nil {
		app.errorLog.Println("ERROR: Unable to generate token for user: ", user.Username)
		app.badRequest(w, errors.New("Internal server error"))
		return
	}
	// Remove sensitive data before sending response
	validUser.Password = ""

	// Prepare and send response
	response := struct {
		Error   bool         `json:"error"`
		Message string       `json:"message"`
		Token   string       `json:"token"`
		User    *models.User `json:"user"`
	}{
		Error:   false,
		Message: "Sign in successful",
		Token:   token,
		User:    validUser,
	}

	app.infoLog.Printf("User %s signed in successfully", user.Username)
	app.writeJSON(w, http.StatusOK, response)
}

// Profile return the profile info of a user by username from request context
func (app *application) Profile(w http.ResponseWriter, r *http.Request) {
	token, ok := app.GetUserTokenFromContext(r.Context())
	if !ok || token == nil {
		app.badRequest(w, errors.New("No user found in context"))
		return
	}

	//user profile
	user, err := app.DB.UserRepo.GetByID(r.Context(), token.ID)
	if err != nil {
		app.errorLog.Println("No user found")
		app.badRequest(w, errors.New("No user found"))
		return
	}

	var sbh []*models.Subscription
	var uph []*models.UploadHistory
	var dwh []*models.DownloadHistory

	var cat []*models.MediaCategory
	var sbp []*models.SubscriptionPlan

	if user.Role == "user" {
		//Subscriptions History
		sbh, err = app.DB.SubscriptionRepo.GetByUserID(r.Context(), user.ID)
		if err != nil {
			app.errorLog.Println("no subscription history available for user: ", token.Name)
		}
		//upload history
		uph, err = app.DB.UploadHistoryRepo.GetAllByUserID(r.Context(), token.ID)
		if err != nil {
			app.errorLog.Println("no upload history available for user: ", token.Name)
		}
		//download history
		dwh, err = app.DB.DownloadHistoryRepo.GetAllByUserID(r.Context(), token.ID)
		if err != nil {
			app.errorLog.Println("no upload history available for user: ", token.Name)
		}
	} else {
		//Subscriptions History
		sbh, err = app.DB.SubscriptionRepo.GetAll(r.Context())
		if err != nil {
			app.errorLog.Println("no subscription history available for admin: ", token.Name)
		}
		//upload history
		uph, err = app.DB.UploadHistoryRepo.GetAll(r.Context())
		if err != nil {
			app.errorLog.Println("no upload history available for admin: ", token.Name)
		}
		//download history
		dwh, err = app.DB.DownloadHistoryRepo.GetAll(r.Context())
		if err != nil {
			app.errorLog.Println("no upload history available for admin: ", token.Name)
		}
	}
	cat, err = app.DB.MediaCategoryRepo.GetAll(r.Context())
	if err != nil {
		app.errorLog.Println("no categories available")
	}
	sbp, err = app.DB.SubscriptionTypeRepo.GetAll(r.Context())
	if err != nil {
		app.errorLog.Println("no plans available")
	}
	// Prepare and send response
	response := struct {
		Error                bool                       `json:"error"`
		Message              string                     `json:"message"`
		User                 *models.User               `json:"user"`
		SubscriptionsHistory []*models.Subscription     `json:"subscription_history"`
		DownloadHistory      []*models.DownloadHistory  `json:"download_history"`
		UploadHistory        []*models.UploadHistory    `json:"upload_history"`
		CategoryList         []*models.MediaCategory    `json:"media_categories"`
		SubscriptionPlans    []*models.SubscriptionPlan `json:"plans"`
	}{
		Error:                false,
		Message:              "user data fetched successfully",
		User:                 user,
		SubscriptionsHistory: sbh,
		DownloadHistory:      dwh,
		UploadHistory:        uph,
		CategoryList:         cat,
		SubscriptionPlans:    sbp,
	}

	app.infoLog.Printf("User %s data retrieved successfully", user.Username)
	app.writeJSON(w, http.StatusOK, response)
}

// UpdateProfile updates user's info in the users registry
func (app *application) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to read json %w", err))
		return
	}
	fmt.Println("Received User data: ", user)
	//sanitize user input
	//name
	//mobile
	//address
	//email

	user.Name = strings.TrimSpace(user.Name)
	user.Mobile = strings.TrimSpace(user.Mobile)
	user.Email = strings.TrimSpace(user.Email)
	user.Address = strings.TrimSpace(user.Address)

	// Update user details in the database
	err = app.DB.UserRepo.UpdateBasicInfo(r.Context(), &user)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	Resp.Error = false
	Resp.Message = "Profile details updated successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

func (app *application) UpdateProfileImage(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error   bool         `json:"error"`
		Message string       `json:"message"`
		User    *models.User `json:"user"`
	}
	err := r.ParseMultipartForm(20 << 20) // 20MB max
	if err != nil {
		app.errorLog.Println("Could not parse multipart form")
		Resp.Error = true
		Resp.Message = "Could not parse multipart form"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		app.errorLog.Println("Image File Required")
		Resp.Error = true
		Resp.Message = "Image File Required"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}
	defer file.Close()

	//get uploader details
	token, ok := app.GetUserTokenFromContext(r.Context())
	if !ok {
		app.errorLog.Println("Unable to get user token from request context")
		Resp.Error = true
		Resp.Message = "Access Denied"
		app.writeJSON(w, http.StatusUnauthorized, Resp)
		return
	}
	user, err := app.DB.UserRepo.GetByID(r.Context(), token.ID)
	if err != nil {
		app.errorLog.Println("Could not get user details from database", err.Error())
		Resp.Error = true
		Resp.Message = "Could not get user details from database"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	uploadDir := filepath.Join(".", "assets", "images", "public", "profile")

	// Check if folder exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		// Folder doesn't exist, create it
		err := os.MkdirAll(uploadDir, os.ModePerm)
		if err != nil {
			app.errorLog.Println("Could not create upload directory:", err.Error())
			Resp.Error = true
			Resp.Message = "Could not create upload directory"
			app.writeJSON(w, http.StatusInternalServerError, Resp)
			return
		}
	}

	//delete file if exist on dstpath
	_ = os.Remove(filepath.Join(uploadDir, user.AvatarID))

	// Generate safe filename
	filename := app.GenerateSafeFilename("", handler)

	err = app.DB.UserRepo.UpdateProfileAvatarURL(r.Context(), user.ID, filename)
	if err != nil {
		app.errorLog.Println("Could not create upload image:", err.Error())
		Resp.Error = true
		Resp.Message = "Internal server error"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	dstPath := filepath.Join(uploadDir, filename)
	//recreate the filepath
	dst, err := os.Create(dstPath)
	if err != nil {
		app.errorLog.Println("Could not save image to filesystem", err.Error())
		Resp.Error = true
		Resp.Message = "Could not save image to filesystem"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		app.errorLog.Println("Error Saving file")
		Resp.Error = true
		Resp.Message = "Error saving file"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	Resp.Error = false
	Resp.Message = "Image uploaded successfully"

	//update avatar url
	baseURL, _ := url.Parse(models.APIEndPoint)
	baseURL.Path = path.Join(baseURL.Path, "public", "profile", filename)
	user.AvatarURL = baseURL.String()

	Resp.User = user
	app.writeJSON(w, http.StatusCreated, Resp)
}

func (app *application) DeactivateProfile(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	token, ok := app.GetUserTokenFromContext(r.Context())

	if !ok {
		app.errorLog.Println("User token not found")
		Resp.Error = true
		Resp.Message = "Invalid token: Access denied"
		app.writeJSON(w, http.StatusOK, Resp)
	}

	status, err := strconv.ParseBool(r.URL.Query().Get("status"))
	if err != nil {
		app.errorLog.Printf("invalid status: %v", status)
		Resp.Error = true
		Resp.Message = "Invalid account status"
		app.writeJSON(w, http.StatusOK, Resp)
	}
	err = app.DB.UserRepo.Deactivate(r.Context(), token.ID, status)
	if err != nil {
		app.errorLog.Println("User token not found")
		Resp.Error = true
		Resp.Message = "Invalid token: Access denied"
		app.writeJSON(w, http.StatusOK, Resp)
	}
	Resp.Error = false
	Resp.Message = "Profile details updated successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

// DeleteProfile removes user's profile from users registry
func (app *application) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to read json %w", err))
		return
	}
	err = app.DB.UserRepo.DeleteByID(r.Context(), user.ID)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	Resp.Error = false
	Resp.Message = "User updated successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

// GetAllUsers retrieves users list from the users registry
func (app *application) GetAllUsers(w http.ResponseWriter, r *http.Request) {

	users, err := app.DB.UserRepo.GetAll(r.Context())
	if err != nil {
		app.badRequest(w, errors.New("Internal Server Error: Unable to retrieve users"))
		return
	}
	var Resp struct {
		Error   bool           `json:"error"`
		Message string         `json:"message"`
		Users   []*models.User `json:"users"`
	}
	Resp.Error = false
	Resp.Users = users
	Resp.Message = "Data fetched successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

// GetMediaCategories retrieves all possible category list from the media_categories registry
func (app *application) GetMediaCategories(w http.ResponseWriter, r *http.Request) {
	var categories []*models.MediaCategory
	var err error

	categories, err = app.DB.MediaCategoryRepo.GetAll(r.Context())
	if err != nil {
		app.errorLog.Println("No category available")
		app.badRequest(w, errors.New("Internal Server Error: No category available"))
		return
	}
	var Resp struct {
		Error           bool                    `json:"error"`
		Message         string                  `json:"message"`
		MediaCategories []*models.MediaCategory `json:"media_categories"`
	}

	if r.URL.Query().Get("isnav") == "true" {
		Resp.MediaCategories = append(Resp.MediaCategories,
			&models.MediaCategory{
				ID:   0,
				Name: "All",
			})
	}

	for _, v := range categories {
		if v != nil {
			baseURL, _ := url.Parse(models.APIEndPoint)
			baseURL.Path = path.Join(baseURL.Path, "public", "categories", v.ThumbnailURL)
			v.ThumbnailURL = baseURL.String()
		}
	}
	Resp.Error = false
	Resp.MediaCategories = append(Resp.MediaCategories, categories...)
	Resp.Message = "Data fetched successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

// CreateMediaCategory creates a new category to the database
func (app *application) CreateMediaCategory(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error         bool                  `json:"error"`
		Message       string                `json:"message"`
		MediaCategory *models.MediaCategory `json:"media_category"`
	}
	err := r.ParseMultipartForm(20 << 20) // 20MB max
	if err != nil {
		app.errorLog.Println("Could not parse multipart form")
		Resp.Error = true
		Resp.Message = "Could not parse multipart form"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		app.errorLog.Println("Image File Required")
		Resp.Error = true
		Resp.Message = "Image File Required"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}
	defer file.Close()

	name := r.FormValue("name")
	// Validate fields
	if name == "" {
		app.errorLog.Println("Missing or invalid fields", "name:", name)
		Resp.Error = true
		Resp.Message = "Missing or invalid fields"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}
	// Generate safe filename
	filename := app.GenerateSafeFilename(name, handler)
	uploadDir := filepath.Join(".", "assets", "images", "public", "categories")

	// Check if folder exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		// Folder doesn't exist, create it
		err := os.MkdirAll(uploadDir, os.ModePerm)
		if err != nil {
			app.errorLog.Println("Could not create upload directory:", err.Error())
			Resp.Error = true
			Resp.Message = "Could not create upload directory"
			app.writeJSON(w, http.StatusInternalServerError, Resp)
			return
		}
	}
	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		app.errorLog.Println("Could not save image to filesystem", err.Error())
		Resp.Error = true
		Resp.Message = "Could not save image to filesystem"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		app.errorLog.Println("Error Saving file: ", err.Error())
		Resp.Error = true
		Resp.Message = "Error saving file"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	//resize the image to 540x540 px
	err = utils.ResizeImage(dstPath, dstPath, 540, 540, true)
	if err != nil {
		app.errorLog.Println("Error resizing file: ", err.Error())
		Resp.Error = true
		Resp.Message = "Error resizing file"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	//save metadata to the backend
	category := models.MediaCategory{
		Name:         name,
		ThumbnailURL: filename,
	}
	err = app.DB.MediaCategoryRepo.Create(r.Context(), &category)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	Resp.Error = false
	Resp.Message = "Media category added successfully"
	baseURL, _ := url.Parse(models.APIEndPoint)
	baseURL.Path = path.Join(baseURL.Path, "public", "categories", category.ThumbnailURL)
	category.ThumbnailURL = baseURL.String()
	Resp.MediaCategory = &category
	app.writeJSON(w, http.StatusOK, Resp)
}

func (app *application) UpdateMediaCategory(w http.ResponseWriter, r *http.Request) {
	var category models.MediaCategory
	err := app.readJSON(w, r, &category)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to read json %w", err))
		return
	}
	err = app.DB.MediaCategoryRepo.Update(r.Context(), &category)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	Resp.Error = false
	Resp.Message = "Media category updated successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

func (app *application) DeleteMediaCategory(w http.ResponseWriter, r *http.Request) {
	param := r.URL.Query().Get("cat_id")
	cat_id, err := strconv.Atoi(param)
	if err != nil {
		app.badRequest(w, fmt.Errorf("Invalid id: %w", err))
		return
	}
	err = app.DB.MediaCategoryRepo.Delete(r.Context(), cat_id)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	Resp.Error = false
	Resp.Message = "Media category removed"
	app.writeJSON(w, http.StatusOK, Resp)
}

// Media content management
func (app *application) ListMedia(w http.ResponseWriter, r *http.Request) {
	category := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("category")))
	var Resp struct {
		Error   bool            `json:"error"`
		Message string          `json:"message"`
		Medias  []*models.Media `json:"medias"`
	}
	categoryID, err := strconv.Atoi(category)
	if err != nil || category == "" {
		app.errorLog.Println("Please specify image category id")
		Resp.Error = true
		Resp.Message = "Please specify image category id"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	var list []*models.Media

	//get the images metadata from database
	if category == "0" {
		list, err = app.DB.MediaRepo.GetAll(r.Context())
	} else {
		list, err = app.DB.MediaRepo.GetAllByCategoryID(r.Context(), categoryID)
	}

	if err != nil && err != sql.ErrNoRows {
		app.errorLog.Println("Could not get image metadata: ", err)
		Resp.Error = true
		Resp.Message = "Image metadata can't be loaded"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	fileDir := filepath.Join(".", "assets", "images", "public", "thumbnails")
	for _, v := range list {
		_, err := os.Stat(filepath.Join(fileDir, "thumb_"+v.MediaUUID))
		if err == nil {
			//TODO:
			baseURL, _ := url.Parse(models.APIEndPoint)
			baseURL.Path = path.Join(baseURL.Path, "public", "thumbnails", "thumb_"+v.MediaUUID)
			v.MediaURL = baseURL.String()
			v.MediaUUID = ""
			Resp.Medias = append(Resp.Medias, v)
		} else {
			app.errorLog.Println(*v)
		}
	}

	Resp.Error = false
	Resp.Message = "Images retrieved successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

func (app *application) FetchMediaDetails(w http.ResponseWriter, r *http.Request) {
	id := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("id")))
	var Resp struct {
		Error   bool          `json:"error"`
		Message string        `json:"message"`
		Media   *models.Media `json:"media"`
	}
	mediaID, err := strconv.Atoi(id)
	if err != nil || mediaID == 0 {
		app.errorLog.Println("Please specify image id")
		Resp.Error = true
		Resp.Message = "Please specify image id"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	var media *models.Media

	//get the images metadata from database
	media, err = app.DB.MediaRepo.GetByID(r.Context(), mediaID)

	if err != nil && err != sql.ErrNoRows {
		app.errorLog.Println("Could not get image metadata: ", err)
		Resp.Error = true
		Resp.Message = "Image metadata can't be loaded"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	fileDir := filepath.Join(".", "assets", "images", "public", "thumbnails")
	_, err = os.Stat(filepath.Join(fileDir, "thumb_"+media.MediaUUID))
	if err == nil {
		baseURL, _ := url.Parse(models.APIEndPoint)
		baseURL.Path = path.Join(baseURL.Path, "public", "thumbnails", "thumb_"+media.MediaUUID)
		media.MediaURL = baseURL.String()
		media.MediaUUID = ""
		Resp.Media = media
	} else {
		Resp.Error = true
		Resp.Message = "Images data not found"
		app.writeJSON(w, http.StatusOK, Resp)
		return
	}

	Resp.Error = false
	Resp.Message = "Images retrieved successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

func (app *application) UploadMedia(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	err := r.ParseMultipartForm(20 << 20) // 20MB max
	if err != nil {
		app.errorLog.Println("Could not parse multipart form")
		Resp.Error = true
		Resp.Message = "Could not parse multipart form"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	file, handler, err := r.FormFile("image")
	if err != nil {
		app.errorLog.Println("Image File Required")
		Resp.Error = true
		Resp.Message = "Image File Required"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}
	defer file.Close()

	title := r.FormValue("media_title")
	description := r.FormValue("description")
	catId := r.FormValue("category_id")
	license_type := r.FormValue("license_type") // "free = 0" or "premium = 1"
	//validate categoryId
	categoryId, catErr := strconv.Atoi(catId)
	licenseType := 1
	imageType := "premium"
	if strings.ToLower(license_type) == "free" {
		licenseType = 0
		imageType = "free"
	}
	licOk := strings.ToLower(license_type) == "free" || strings.ToLower(license_type) == "premium"
	// Validate fields
	if catErr != nil {
		app.errorLog.Println("category id error fields: ", catErr)
		Resp.Error = true
		Resp.Message = "Missing or invalid fields"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}
	if !licOk {
		app.errorLog.Println("Invalid license type: ")
		Resp.Error = true
		Resp.Message = "Missing or invalid fields"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}
	if title == "" {
		app.errorLog.Println("Missing title field: ", title)
		Resp.Error = true
		Resp.Message = "Missing or invalid fields"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	// Generate safe filename
	filename := app.GenerateSafeFilename("", handler)

	uploadDir := filepath.Join(".", "assets", "images", imageType)

	// Check if folder exists
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		// Folder doesn't exist, create it
		err := os.MkdirAll(uploadDir, os.ModePerm)
		if err != nil {
			app.errorLog.Println("Could not create upload directory:", err.Error())
			Resp.Error = true
			Resp.Message = "Could not create upload directory"
			app.writeJSON(w, http.StatusInternalServerError, Resp)
			return
		}
	}

	dstPath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		app.errorLog.Println("Could not save image to filesystem", err.Error())
		Resp.Error = true
		Resp.Message = "Could not save image to filesystem"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		app.errorLog.Println("Error Saving file")
		Resp.Error = true
		Resp.Message = "Error saving file"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	//save watermarked image
	outputBaseDir := filepath.Join(".", "assets", "images", "public")
	//save thumbnail
	err = utils.GenerateImageVariants(dstPath, outputBaseDir, filename)
	if err != nil {
		app.errorLog.Println("Unable to save image variations: ", err.Error())
		Resp.Error = true
		Resp.Message = "Unable to save image variations"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	//get uploader details
	token, ok := app.GetUserTokenFromContext(r.Context())
	if !ok {
		app.errorLog.Println("Unable to get user token from request context")
		Resp.Error = true
		Resp.Message = "Access Denied"
		app.writeJSON(w, http.StatusUnauthorized, Resp)
		return
	}
	// Save metadata to DB
	imageMetadata := &models.Media{
		MediaTitle:   title,
		MediaUUID:    filename,
		Description:  description,
		CategoryID:   categoryId,
		LicenseType:  licenseType,
		UploaderID:   token.ID,
		UploaderName: token.Name,
		FileType:     utils.GetFileType(handler),
		FileExt:      filepath.Ext(filename),
		FileName:     title,
		FileSize:     utils.GetFormattedFileSize(handler),
		Resolution:   utils.GetImageResolutionString(handler),
	}
	err = app.DB.MediaRepo.Create(r.Context(), imageMetadata)
	if err != nil {
		app.errorLog.Println("Could not save image metadata", err.Error())
		Resp.Error = true
		Resp.Message = "Could not save image metadata"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	err = app.DB.MediaCategoryRepo.IncrementDownloads(r.Context(), int64(categoryId))
	if err != nil {
		app.errorLog.Println("Could not save image metadata", err.Error())
		Resp.Error = true
		Resp.Message = "Could not save image metadata"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	h := &models.UploadHistory{
		MediaUUID:  filename,
		UserID:     token.ID,
		FileType:   utils.GetFileType(handler),
		FileExt:    filepath.Ext(filename),
		FileName:   title,
		FileSize:   utils.GetFormattedFileSize(handler),
		Resolution: utils.GetImageResolutionString(handler),
		UploadedAt: time.Now(),
	}
	err = app.DB.UploadHistoryRepo.Create(r.Context(), h)
	if err != nil {
		app.errorLog.Println("Could not save upload history", err.Error())
		Resp.Error = true
		Resp.Message = "Could not save image metadata"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	Resp.Error = false
	Resp.Message = "Image uploaded successfully"
	app.writeJSON(w, http.StatusCreated, Resp)
}

func (app *application) ServeMedia(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	// 1. Validate media ID
	id, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("id")))
	if err != nil || id <= 0 {
		app.errorLog.Println("Invalid or missing media id")
		Resp.Error = true
		Resp.Message = "Invalid or missing media ID"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	// 2. Get media from DB
	media, err := app.DB.MediaRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			app.errorLog.Printf("Media not found for ID: %d", id)
			Resp.Error = true
			Resp.Message = "Media not found"
			app.writeJSON(w, http.StatusNotFound, Resp)
			return
		}
		app.errorLog.Println("Database error fetching media:", err)
		Resp.Error = true
		Resp.Message = "Could not retrieve media"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	// 3. Get user from context
	token, ok := app.GetUserTokenFromContext(r.Context())
	if !ok {
		app.errorLog.Println("User token not found in context")
		Resp.Error = true
		Resp.Message = "Access Denied: Please log in"
		app.writeJSON(w, http.StatusUnauthorized, Resp)
		return
	}

	// 4. Fetch user from DB
	user, err := app.DB.UserRepo.GetByID(r.Context(), token.ID)
	if err != nil {
		app.errorLog.Printf("Failed to load user ID %d: %v", token.ID, err)
		Resp.Error = true
		Resp.Message = "Could not load user"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	// 5. If media is premium, check subscription
	if media.LicenseType == 1 {
		plan := user.CurrentSubscription
		if plan == nil || plan.PlanDetails == nil || !plan.Status {
			app.errorLog.Printf("User %d has no active subscription", user.ID)
			Resp.Error = true
			Resp.Message = "You must have an active subscription"
			app.writeJSON(w, http.StatusForbidden, Resp)
			return
		}

		// 6. Check subscription expiry
		expiry := plan.PaymentTime.AddDate(0, 0, plan.PlanDetails.ExpiresAt)
		if time.Now().After(expiry) {
			app.errorLog.Printf("Subscription expired for user %d", user.ID)
			Resp.Error = true
			Resp.Message = "Your subscription has expired"
			app.writeJSON(w, http.StatusForbidden, Resp)
			return
		}

		// 7. Check download limit
		if plan.PlanDetails.DownloadLimit <= plan.TotalDownloads {
			app.errorLog.Printf("User %d reached download limit", user.ID)
			Resp.Error = true
			Resp.Message = "Download limit reached. Please upgrade your plan."
			app.writeJSON(w, http.StatusForbidden, Resp)
			return
		}
	}

	// 8. Decrement user download limit if media is premium
	if media.LicenseType == 1 {
		err := app.DB.UserRepo.IncrementDownloadCounts(r.Context(), user.ID)
		if err != nil {
			app.errorLog.Printf("Failed to decrement download limit for user %d: %v", user.ID, err)
		} else {
			app.infoLog.Printf("Decremented download limit for user %d", user.ID)
		}
	}

	// 9. Log download history
	download := models.DownloadHistory{
		MediaUUID:    media.MediaUUID,
		UserID:       user.ID,
		FileType:     media.FileType,
		FileExt:      media.FileExt,
		FileName:     media.MediaTitle,
		FileSize:     media.FileSize,
		Resolution:   media.Resolution,
		DownloadedAt: time.Now(),
	}
	if err := app.DB.DownloadHistoryRepo.Create(r.Context(), &download); err != nil {
		app.errorLog.Printf("Failed to log download for user %d: %v", user.ID, err)
		Resp.Error = true
		Resp.Message = "Failed to log download history"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	// 10. Increment media category download count
	if err := app.DB.MediaCategoryRepo.IncrementDownloads(r.Context(), int64(media.CategoryID)); err != nil {
		app.errorLog.Printf("Failed to increment category downloads for user %d: %v", user.ID, err)
		Resp.Error = true
		Resp.Message = "Failed to update category download count"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	// 11. Increment media download count
	if err := app.DB.MediaRepo.IncrementDownloadCountByID(r.Context(), media.ID); err != nil {
		app.errorLog.Printf("Failed to increment media downloads for user %d: %v", user.ID, err)
		Resp.Error = true
		Resp.Message = "Failed to update media download count"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	// 12. Locate the media file path
	imageType := "premium"
	if media.LicenseType == 0 {
		imageType = "free"
	}
	mediaPath := path.Join("assets", "images", imageType, media.MediaUUID)

	if _, err := os.Stat(mediaPath); err != nil {
		app.errorLog.Printf("File not found: %s", mediaPath)
		Resp.Error = true
		Resp.Message = "File not found"
		app.writeJSON(w, http.StatusNotFound, Resp)
		return
	}

	// 13. Serve the media file
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", media.FileName))

	http.ServeFile(w, r, mediaPath)
}

// --- Subscription Plan Management ---

// CreatePlan creates a new subscription Plan Type to the database
func (app *application) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error   bool                    `json:"error"`
		Message string                  `json:"message"`
		Plan    models.SubscriptionPlan `json:"plan"`
	}
	var p models.SubscriptionPlan
	err := app.readJSON(w, r, &p)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	err = app.DB.SubscriptionTypeRepo.Create(r.Context(), &p)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	Resp.Error = false
	Resp.Message = "New Plan added successfully"
	Resp.Plan = p
	app.writeJSON(w, http.StatusOK, Resp)
}

func (app *application) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	var plan models.SubscriptionPlan
	err := app.readJSON(w, r, &plan)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	err = app.DB.SubscriptionTypeRepo.Update(r.Context(), &plan)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	Resp.Error = false
	Resp.Message = "Plan details updated successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}
func (app *application) GetPlans(w http.ResponseWriter, r *http.Request) {
	var plans []*models.SubscriptionPlan
	var err error

	plans, err = app.DB.SubscriptionTypeRepo.GetAll(r.Context())
	if err != nil {
		app.errorLog.Println("No plan available")
		app.badRequest(w, errors.New("Internal Server Error: No plan available"))
		return
	}
	var Resp struct {
		Error   bool                       `json:"error"`
		Message string                     `json:"message"`
		Plans   []*models.SubscriptionPlan `json:"plans"`
	}

	Resp.Error = false
	Resp.Plans = plans
	Resp.Message = "Data fetched successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}
func (app *application) PurchasePlan(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error               bool                 `json:"error"`
		Message             string               `json:"message"`
		SubscriptionDetails *models.Subscription `json:"subscription"`
	}

	planID, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("plan_id")))
	if err != nil {
		app.errorLog.Println("Invalid plan id: ", err)
		Resp.Error = true
		Resp.Message = "Invalid plan id"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	token, ok := app.GetUserTokenFromContext(r.Context())
	if !ok {
		app.errorLog.Println("Invalid user token: ", err)
		Resp.Error = true
		Resp.Message = "Invalid user token"
		app.writeJSON(w, http.StatusForbidden, Resp)
		return
	}

	user, err := app.DB.UserRepo.GetByID(r.Context(), token.ID)
	if err != nil {
		app.errorLog.Println("Invalid user token: ", err)
		Resp.Error = true
		Resp.Message = "Invalid user token"
		app.writeJSON(w, http.StatusForbidden, Resp)
		return
	}

	if user.CurrentSubscription != nil {
		expiry := user.CurrentSubscription.PaymentTime.AddDate(0, 0, user.CurrentSubscription.PlanDetails.ExpiresAt)

		if !time.Now().After(expiry) {
			app.errorLog.Printf("Purchased Subscription plan exist for user %d", user.ID)
			Resp.Error = true
			Resp.Message = "You already have an active subscription plan. Please try again after your current plan expires."
			app.writeJSON(w, http.StatusForbidden, Resp)
			return
		}

		if user.CurrentSubscription.PlanDetails.DownloadLimit <= 0 {
			app.errorLog.Printf("Purchased Subscription plan exist for user %d", user.ID)
			Resp.Error = true
			Resp.Message = "You already have an active subscription plan. Please try again after your current plan expires."
			app.writeJSON(w, http.StatusForbidden, Resp)
			return
		}
	}
	//get subscription type info
	spt, err := app.DB.SubscriptionTypeRepo.GetByID(r.Context(), planID)
	if err != nil {
		app.errorLog.Println(err)
		Resp.Error = true
		Resp.Message = "Invalid plan id"
		app.writeJSON(w, http.StatusForbidden, Resp)
		return
	}
	//complete payment process(via stripe, MFS or Banking system)
	sub := &models.Subscription{
		UserID:             user.ID,
		SubscriptionPlanID: spt.ID,
		PaymentAmount:      float64(spt.Price),
		PaymentTime:        time.Now(),
		TotalDownloads:     0,
		Status:             true,
	}
	err = app.DB.SubscriptionRepo.Create(r.Context(), sub)
	if err != nil {
		app.errorLog.Println(err)
		Resp.Error = true
		Resp.Message = "Internal Server Error: Unable to complete purchase"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	err = app.DB.UserRepo.UpdateSubscriptionPlanByUserID(r.Context(), sub.ID, user.ID)
	if err != nil {
		app.errorLog.Println(err)
		Resp.Error = true
		Resp.Message = "Internal Server Error: Unable to complete purchase"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	Resp.Error = true
	Resp.Message = "Purchase completed"
	sub.PlanDetails = spt
	Resp.SubscriptionDetails = sub
	app.writeJSON(w, http.StatusOK, Resp)

}

// --- Report Management ---

func (app *application) GetDownloadHistory(w http.ResponseWriter, r *http.Request) {
	var resp struct {
		Error           bool                      `json:"error"`
		Message         string                    `json:"message"`
		DownloadHistory []*models.DownloadHistory `json:"history"`
	}

	// Step 1: Extract user token from context
	token, ok := app.GetUserTokenFromContext(r.Context())
	if !ok {
		app.errorLog.Println("User not found in context")
		resp.Error = true
		resp.Message = "Invalid token: Access Denied"
		app.writeJSON(w, http.StatusForbidden, resp)
		return
	}

	var (
		history []*models.DownloadHistory
		err     error
	)

	// Step 2: Handle based on user role
	if token.Role == "admin" {
		// Admins can optionally pass ?user_id to filter
		userIDStr := strings.TrimSpace(r.URL.Query().Get("user_id"))

		if userIDStr == "" {
			// No filter, return all download history
			history, err = app.DB.DownloadHistoryRepo.GetAll(r.Context())
		} else {
			// Filter by user_id
			var userID int
			userID, err = strconv.Atoi(userIDStr)
			if err != nil {
				app.errorLog.Printf("Invalid user_id query param: %v", err)
				resp.Error = true
				resp.Message = "Invalid user_id in query"
				app.writeJSON(w, http.StatusBadRequest, resp)
				return
			}
			history, err = app.DB.DownloadHistoryRepo.GetAllByUserID(r.Context(), userID)
		}
	} else {
		// Normal users only get their own download history
		history, err = app.DB.DownloadHistoryRepo.GetAllByUserID(r.Context(), token.ID)
	}

	// Step 3: Handle DB or logic error
	if err != nil {
		app.errorLog.Println("Error fetching download history:", err)
		resp.Error = true
		resp.Message = "Internal Server Error"
		app.writeJSON(w, http.StatusInternalServerError, resp)
		return
	}

	// Step 4: Successful response
	resp.Error = false
	resp.Message = "Download history retrieved successfully"
	resp.DownloadHistory = history
	app.writeJSON(w, http.StatusOK, resp)
}

func (app *application) GetUploadHistory(w http.ResponseWriter, r *http.Request) {
	var resp struct {
		Error         bool                    `json:"error"`
		Message       string                  `json:"message"`
		UploadHistory []*models.UploadHistory `json:"history"`
	}

	// Step 1: Extract user token from context
	token, ok := app.GetUserTokenFromContext(r.Context())
	if !ok {
		app.errorLog.Println("User not found in context")
		resp.Error = true
		resp.Message = "Invalid token: Access Denied"
		app.writeJSON(w, http.StatusForbidden, resp)
		return
	}

	var (
		history []*models.UploadHistory
		err     error
	)

	// Step 2: Handle based on user role
	if token.Role == "admin" {
		// Admins can optionally pass ?user_id to filter
		userIDStr := strings.TrimSpace(r.URL.Query().Get("user_id"))

		if userIDStr == "" {
			// No filter, return all upload history
			history, err = app.DB.UploadHistoryRepo.GetAll(r.Context())
		} else {
			// Filter by user_id
			var userID int
			userID, err = strconv.Atoi(userIDStr)
			if err != nil {
				app.errorLog.Printf("Invalid user_id query param: %v", err)
				resp.Error = true
				resp.Message = "Invalid user_id in query"
				app.writeJSON(w, http.StatusBadRequest, resp)
				return
			}
			history, err = app.DB.UploadHistoryRepo.GetAllByUserID(r.Context(), userID)
		}
	} else {
		// Normal users only get their own upload history
		history, err = app.DB.UploadHistoryRepo.GetAllByUserID(r.Context(), token.ID)
	}

	// Step 3: Handle DB or logic error
	if err != nil {
		app.errorLog.Println("Error fetching upload history:", err)
		resp.Error = true
		resp.Message = "Internal Server Error"
		app.writeJSON(w, http.StatusInternalServerError, resp)
		return
	}

	// Step 4: Successful response
	resp.Error = false
	resp.Message = "Upload history retrieved successfully"
	resp.UploadHistory = history
	app.writeJSON(w, http.StatusOK, resp)
}
