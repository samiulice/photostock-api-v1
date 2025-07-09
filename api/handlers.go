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
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	app.infoLog.Println("Registration successful. Auto sign-in complete")
	app.writeJSON(w, http.StatusOK, response)
}

// generateSignedToken generate a token string for implementing JWT
func generateSignedToken(user *models.User) (string, error) {
	app.infoLog.Println("signed in: ", user)
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

	//TODO::
	//upload history
	uph, err := app.DB.UploadHistoryRepo.GetAllByUserID(r.Context(), token.ID)
	if err != nil {
		app.errorLog.Println("no upload history available for user: ", token.Name)
	}
	//download history
	dwh, err := app.DB.DownloadHistoryRepo.GetAllByUserID(r.Context(), token.ID)
	if err != nil {
		app.errorLog.Println("no upload history available for user: ", token.Name)
	}
	// Prepare and send response
	response := struct {
		Error           bool                      `json:"error"`
		Message         string                    `json:"message"`
		User            *models.User              `json:"user"`
		DownloadHistory []*models.DownloadHistory `json:"download_history"`
		UploadHistory   []*models.UploadHistory   `json:"upload_history"`
	}{
		Error:           false,
		Message:         "user data fetched successfully",
		User:            user,
		DownloadHistory: dwh,
		UploadHistory:   uph,
	}

	app.infoLog.Printf("User %s signed in successfully", user.Username)
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
	user.Username = strings.Split(user.Email, "@")[0] + "_" + app.GenerateRandomAlphanumericCode(4)
	user.Name = strings.TrimSpace(user.Name)
	user.Password = strings.TrimSpace(user.Password)
	//hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		app.badRequest(w, errors.New("Internal Server Error: Try again"))
		return
	}
	user.Password = string(hashedPassword)

	// Update user details in the database
	err = app.DB.UserRepo.Update(r.Context(), &user)
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
	allCategory := models.MediaCategory{
		ID:   0,
		Name: "All",
	}
	if r.URL.Query().Get("isnav") == "true" {
		Resp.MediaCategories = append(Resp.MediaCategories, &allCategory)
	}
	Resp.Error = false
	Resp.MediaCategories = append(Resp.MediaCategories, categories...)
	Resp.Message = "Data fetched successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

// CreateMediaCategory creates a new category to the database
func (app *application) CreateMediaCategory(w http.ResponseWriter, r *http.Request) {
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
	filename := app.GenerateSafeFilename(handler)
	uploadDir := filepath.Join(".", "assets", "images", "categories")

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
	err = utils.ResizeImageInPlace(dstPath, 540, 540)
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
		ThumbnailURL: models.APIEndPoint + path.Join("images", "categories", filename),
	}
	err = app.DB.MediaCategoryRepo.Create(r.Context(), &category)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	Resp.Error = false
	Resp.Message = "Media category added successfully"
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
		app.badRequest(w, fmt.Errorf("Invalid id", err))
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

	app.infoLog.Println("Fetching ", category)
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

	fileDir := filepath.Join(".", "assets", "images", "thumbnails")
	for _, v := range list {
		_, err := os.Stat(filepath.Join(fileDir, "thumb_"+v.MediaUUID))
		if err == nil {
			//TODO:
			v.MediaURL = models.APIEndPoint + path.Join("images", "thumbnails", "thumb_"+v.MediaUUID)
			v.MediaUUID = ""
			Resp.Medias = append(Resp.Medias, v)
			app.infoLog.Println(*v)
		} else {
			app.errorLog.Println(*v)
		}
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
	licenseType := 0
	if license_type == "paid" {
		licenseType = 1
	}
	LicErr := license_type == "free" || license_type == "paid"
	// Validate fields
	if catErr != nil || LicErr || title == "" {
		app.errorLog.Println("Missing or invalid fields", "title:", title, "Description: ", description, "catid: ", catId, "lic_type", license_type)
		Resp.Error = true
		Resp.Message = "Missing or invalid fields"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	// Generate safe filename
	filename := app.GenerateSafeFilename(handler)

	uploadDir := filepath.Join(".", "assets", "images", "original")

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
	outputBaseDir := filepath.Join(".", "assets", "images")
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
	image := &models.Media{
		MediaTitle:   title,
		MediaUUID:    filename,
		Description:  description,
		CategoryID:   categoryId,
		LicenseType:  licenseType,
		UploaderID:   token.ID,
		UploaderName: token.Name,
	}
	app.infoLog.Println(image.UploaderID, token)
	err = app.DB.MediaRepo.Create(r.Context(), image)
	if err != nil {
		app.errorLog.Println("Could not save image metadata", err.Error())
		Resp.Error = true
		Resp.Message = "Could not save image metadata"
		app.writeJSON(w, http.StatusInternalServerError, Resp)
		return
	}

	Resp.Error = false
	Resp.Message = "Image uploaded successfully"
	app.writeJSON(w, http.StatusCreated, Resp)
}

func (app *application) DownloadPremiumMedia(w http.ResponseWriter, r *http.Request) {
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	// get media info by media_uuid
	media_uuid := r.URL.Query().Get("id")
	if strings.TrimSpace(media_uuid) == "" {
		app.errorLog.Println("No media id provided")
		Resp.Error = true
		Resp.Message = "Invalid or missing media id"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	m, err := app.DB.MediaRepo.GetByMediaUUID(r.Context(), media_uuid)
	if err != nil {
		if err == sql.ErrNoRows {
			app.errorLog.Println("Invalid media_uuid")
			Resp.Error = true
			Resp.Message = "wrong media id"
			app.writeJSON(w, http.StatusBadRequest, Resp)
			return
		}
		app.errorLog.Println("Unable to get media info")
		Resp.Error = true
		Resp.Message = "Internal Server Error: Unable to media info! Try again"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	if m.LicenseType == 1 {
		app.errorLog.Println("Free image! Redirect to image download page")
		Resp.Error = false
		Resp.Message = models.APIEndPoint + path.Join("images", "free", media_uuid)
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}
	token, ok := app.GetUserTokenFromContext(r.Context())
	if !ok {
		app.errorLog.Println("Could not get user token from context")
		Resp.Error = true
		Resp.Message = "Access Denied: Could not get user token from context"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	// get user by id
	user, err := app.DB.UserRepo.GetByID(r.Context(), token.ID)
	if err != nil {
		app.errorLog.Println("Could not get user data from database")
		Resp.Error = true
		Resp.Message = "Access Denied: Could not get user data from database"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	// check users subscription
	if user.SubscriptionPlan.Status {
		app.errorLog.Println("No subscription plan for this user")
		Resp.Error = true
		Resp.Message = "no plan"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	// check users download limit
	expiredDate, err := time.Parse(user.SubscriptionPlan.TimeLimit, "02-01-2006")
	if err != nil {
		app.errorLog.Println("Could not parse the expiry date")
		Resp.Error = true
		Resp.Message = "Internal Server Error: Could not parse the expiry date"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}
	notExpired := time.Now().Compare(expiredDate) <= 0
	if user.SubscriptionPlan.DownloadLimit > 0 && notExpired {
		app.errorLog.Println("No subscription plan for this user")
		Resp.Error = true
		Resp.Message = "no plan"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	http.ServeFile(w, r, path.Join("secure", "images", "premium", media_uuid))
}
