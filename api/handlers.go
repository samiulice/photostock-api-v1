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
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/samiulice/photostock/internal/models"

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
	//download history
	// Prepare and send response
	response := struct {
		Error           bool                    `json:"error"`
		Message         string                  `json:"message"`
		User            *models.User            `json:"user"`
		DownloadHistory *models.DownloadHistory `json:"download_history"`
		UploadHistory   *models.UploadHistory   `json:"upload_history"`
	}{
		Error:   false,
		Message: "user data fetched successfully",
		User:    user,
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
	Resp.Error = false
	Resp.MediaCategories = categories
	Resp.Message = "Data fetched successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

// CreateMediaCategory creates a new category to the database
func (app *application) CreateMediaCategory(w http.ResponseWriter, r *http.Request) {
	var category models.MediaCategory
	err := app.readJSON(w, r, &category)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to read json %w", err))
		return
	}
	err = app.DB.MediaCategoryRepo.Create(r.Context(), &category)
	if err != nil {
		app.badRequest(w, err)
		return
	}
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
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

	title := r.FormValue("title")
	description := r.FormValue("description")
	catId := r.FormValue("category_id")
	license_type := r.FormValue("license_type") // "free" or "premium"
	//validate categoryId
	categoryId, catErr := strconv.Atoi(catId)
	licenseType, LicErr := strconv.Atoi(license_type)
	// Validate fields
	if catErr != nil || LicErr != nil || title == "" {
		app.errorLog.Println("Missing or invalid fields")
		Resp.Error = true
		Resp.Message = "Missing or invalid fields"
		app.writeJSON(w, http.StatusBadRequest, Resp)
		return
	}

	// Generate safe filename
	filename := app.GenerateSafeFilename(handler)

	uploadDir := filepath.Join(".", "assets", "images", "original")
	saveImage := func() (bool, string, int) {
		// Check if folder exists
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			// Folder doesn't exist, create it
			err := os.MkdirAll(uploadDir, os.ModePerm)
			if err != nil {
				app.errorLog.Println("Could not create upload directory:", err.Error())
				return true, "Could not create upload directory", http.StatusInternalServerError

			}
		}

		dstPath := filepath.Join(uploadDir, filename)
		dst, err := os.Create(dstPath)
		if err != nil {
			app.errorLog.Println("Could not save image to filesystem", err.Error())
			return true, "Could not save image to filesystem", http.StatusInternalServerError
		}
		defer dst.Close()

		_, err = io.Copy(dst, file)
		if err != nil {
			app.errorLog.Println("Error Saving file")
			return true, "Error Saving file", http.StatusInternalServerError
		}
		return false, "", 200
	}

	//save original image
	ok, msg, statusCode := saveImage()
	if ok {
		app.errorLog.Println("Could not save image to filesystem", err.Error())
		Resp.Error = true
		Resp.Message = msg
		app.writeJSON(w, statusCode, Resp)
		return
	}
	// TODO:
	//save watermarked image
	//save thumbnail

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
		MediaTitle:  title,
		MediaUUID:   filename,
		Description: description,
		CategoryID:  categoryId,
		LicenseType: licenseType,
		UploaderID:  token.ID,
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
