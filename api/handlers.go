package api

import (
	"database/sql"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"time"

	"github.com/samiulice/photostock/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// .......................Media MANAGEMENT.......................

// func (app *application) UploadMedia(w http.ResponseWriter, r *http.Request) {
// 	var resp models.Response
// 	// Limit request body size (optional)
// 	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10 MB

// 	// Parse multipart form
// 	err := r.ParseMultipartForm(10 << 20) // 10 MB
// 	if err != nil {
// 		app.errorLog.Println("Unable to parse form")
// 		resp.Error = true
// 		resp.Message = "Internal Server Error"
// 		app.writeJSON(w, http.StatusBadRequest, resp)
// 		return
// 	}

// 	//get uploader info from the contaxt
// 	user, ok := app.GetUserFromContext(r.Context())
// 	if !ok {
// 		app.writeJSON(w, http.StatusUnauthorized, nil)
// 		return
// 	}
// 	// Read form fields
// 	title := r.FormValue("title")
// 	description := r.FormValue("description")
// 	categoryID, err := strconv.Atoi(r.FormValue("category_id"))
// 	if err != nil {
// 		app.errorLog.Println("Inalid category id: Cannot convert to int")
// 		resp.Error = true
// 		resp.Message = "Invalid category id"
// 		app.writeJSON(w, http.StatusUnprocessableEntity, resp)
// 		return
// 	}
// 	mediaUUID := uuid.NewString()
// 	mediaURL := filepath.Join("fileserver", "media", "images", user.Username, mediaUUID)

// 	// Read image file
// 	file, _, err := r.FormFile("image")
// 	if err != nil {
// 		http.Error(w, "Error retrieving image", http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()

// 	// Save image to disk (or handle as needed)
// 	dst, err := os.Create(mediaURL)
// 	if err != nil {
// 		http.Error(w, "Error saving file", http.StatusInternalServerError)
// 		return
// 	}
// 	defer dst.Close()

// 	_, err = io.Copy(dst, file)
// 	if err != nil {
// 		http.Error(w, "Error writing file", http.StatusInternalServerError)
// 		return
// 	}

// 	//save media info to database
// 	var media models.Media
// 	media.MediaUUID = uuid.NewString()
// 	media.MediaTitle = title
// 	media.MediaURL = mediaURL
// 	media.Description = description
// 	media.CategoryID = &categoryID

// 	//TODO: Save media info to database
// 	media.Id, err = app.DB
// 	if err != nil {
// 		http.Error(w, "Error writing file", http.StatusInternalServerError)
// 		return
// 	}

// 	// Respond
// 	resp.Error = false
// 	resp.Message = "Upload Successfull"
// 	resp.Data = any(media)
// 	app.writeJSON(w, http.StatusOK, resp)
// }

// .......................APP USER MANAGEMENT.......................
// AddUser adds new user to the users registry
func (app *application) AddUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to read json %w", err))
		return
	}
	fmt.Println("Received User data: ", user)
	//sanitize user input
	user.Username = strings.TrimSpace(user.Username)
	user.Name = strings.TrimSpace(user.Name)
	user.Status = true
	user.Role = "user"
	user.Password = strings.TrimSpace(user.Password)
	//hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to hash password %w", err))
		return
	}
	user.Password = string(hashedPassword)

	err = app.DB.UserRepo.Create(r.Context(), &user)
	if err == sql.ErrNoRows {
		app.errorLog.Println("ERROR: AddUser => username already exists:", err)
		app.badRequest(w, errors.New("username already exists"))
		return
	} else if err != nil {
		app.errorLog.Println("ERROR: AddUser => Unable to create user:", err)
		app.badRequest(w, errors.New("Internal Server Error: Unable to create user"))
		return
	}
	var Resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	Resp.Error = false
	Resp.Message = "User added successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

// UpdateUser updates user's info in the users registry
func (app *application) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to read json %w", err))
		return
	}
	fmt.Println("Received User data: ", user)
	//sanitize user input
	user.Username = strings.TrimSpace(user.Username)
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
	Resp.Message = "User details updated successfully"
	app.writeJSON(w, http.StatusOK, Resp)
}

// DeleteUser removes user from users registry
func (app *application) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to read json %w", err))
		return
	}
	err = app.DB.UserRepo.Delete(r.Context(), user.ID)
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

// .......................APP USER SESSION MANAGEMENT.......................
// SignIn authenticates the user and generates a JWT token for them.
// This function is used for the new authentication system using JWT.

func (app *application) SignUp(w http.ResponseWriter, r *http.Request) {
	var user models.User
	err := app.readJSON(w, r, &user)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to read json %w", err))
		return
	}
	//sanitize user input
	user.Username = strings.Split(user.Email, "@")[0] + app.GenerateRandomAlphanumericCode(4)
	user.Name = strings.TrimSpace(user.Name)
	user.Email = strings.TrimSpace(user.Email)
	user.Status = true
	user.Role = "user"
	user.Password = strings.TrimSpace(user.Password)
	//hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		app.badRequest(w, fmt.Errorf("ERROR:unable to hash password %w", err))
		return
	}
	user.Password = string(hashedPassword)

	err = app.DB.UserRepo.Create(r.Context(), &user)
	if err == sql.ErrNoRows {
		app.errorLog.Println("ERROR: AddUser => username already exists:", err)
		app.badRequest(w, errors.New("username already exists"))
		return
	} else if err != nil {
		app.errorLog.Println("ERROR: AddUser => Unable to create user:", err)
		app.badRequest(w, errors.New("Internal Server Error: Unable to create user"))
		return
	}

	//after adding user successfully, go to the login process
	//Generate signed token
	token, err := generateSignedToken(&user)
	if err != nil{
		app.errorLog.Printf("ERROR: Unable to generate token for user: ", user.Username)
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
		Message: "Signup successfully and autoSign in",
		Token:   token,
		User:    &user,
	}

	app.infoLog.Printf("User %s signed in successfully", user.Username)
	app.writeJSON(w, http.StatusOK, response)
}

func (app *application) SignIn(w http.ResponseWriter, r *http.Request) {
	var user models.User

	// Decode JSON credentials
	if err := app.readJSON(w, r, &user); err != nil {
		app.errorLog.Println("ERROR: Unable to read JSON -", err)
		app.badRequest(w, errors.New("Failed to read username and password"))
		return
	}

	// Lookup user by username
	validUser, err := app.DB.UserRepo.GetByUsername(r.Context(), user.Username)
	if err != nil {
		app.errorLog.Println("ERROR: User lookup failed -", err)
		if errors.Is(err, sql.ErrNoRows) {
			app.badRequest(w, errors.New("Wrong username or password"))
		} else {
			app.badRequest(w, errors.New("Failed to retrieve user"))
		}
		return
	}

	// Check if user account is active
	if !validUser.Status {
		app.infoLog.Printf("SignIn denied for deactivated user: %s", user.Username)
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
	if err != nil{
		app.errorLog.Printf("ERROR: Unable to generate token for user: ", user.Username)
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


func generateSignedToken(user *models.User)(string, error){
	// Create JWT claims
	claims := jwt.MapClaims{
		"id":       user.ID,
		"name":     user.Name,
		"username": user.Username,
		"email": user.Email,
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