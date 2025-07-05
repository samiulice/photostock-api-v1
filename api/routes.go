package api

import (
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Set up a simple logger middleware
	mux.Use(app.Logger)

	//setup image server
	// Check if image directory exists
	imageDir := filepath.Join("static", "images")
	//Image for serving image files
	fileServer := http.FileServer(http.Dir(imageDir))
	mux.Handle("/images/*", http.StripPrefix("/images", fileServer))

	// --- Authentication & User Management ---
	mux.Route("/api/v1/auth", func(r chi.Router) {
		mux.Post("/register", app.Register) // Register a new user
		mux.Post("/login", app.Login)       // User login
		// mux.Post("/logout", app.Logout)               // User logout
		mux.Get("/profile", app.Profile)       // Get currently logged-in user's profile
		mux.Put("/profile", app.UpdateProfile) // Update user profile information
		// mux.Put("/profile/deactivate", app.DeactivateProfile)        // Deactivate user profile information
		mux.Delete("/profile/delete", app.DeleteProfile) // Delete user profile information
		// mux.Put("/password", app.ChangePassword)      // Change password for logged-in user
		// mux.Post("/forgot-password", app.ForgotPassword) // Request password reset via email
		// mux.Post("/reset-password", app.ResetPassword)   // Reset password using token
	})

	// --- Media Management ---
	// mux.Route("/api/v1/media", func(r chi.Router) {
	// 	mux.Get("/", app.ListMedia)                         // List all media
	// 	mux.Post("/", app.UploadMedia)                      // Upload new media
	// 	mux.Get("/{id}", app.GetMedia)                      // Retrieve a single media item by ID
	// 	mux.Put("/{id}", app.UpdateMedia)                   // Update an existing media item
	// 	mux.Delete("/{id}", app.DeleteMedia)                // Delete a media item

	// 	mux.Get("/user/{userId}", app.UserMedia)            // List all media uploaded by a specific user
	// 	mux.Get("/category/{slug}", app.CategoryMedia)      // List media by category slug
	// })

	// --- Categories Management ---
	mux.Route("/api/v1/categories", func(r chi.Router) {
		mux.Get("/", app.GetMediaCategories)     // List all categories
		mux.Post("/", app.CreateMediaCategory)   // Create a new category
		mux.Put("/", app.UpdateMediaCategory)    // Update an existing category
		mux.Delete("/", app.DeleteMediaCategory) // Delete a category
	})

	return mux
}
