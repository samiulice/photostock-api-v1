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

	// Serve /images/ → ./assets/images/
	thumbnailDir := filepath.Join(".", "assets", "images", "public")
	fs := http.StripPrefix("/public/", http.FileServer(http.Dir(thumbnailDir)))
	mux.Handle("/public/*", fs)

	// --- Authentication & User Management ---
	mux.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/register", app.Register) // Register a new user
		r.Post("/login", app.Login)       // User login
		// r.Post("/logout", app.Logout)               // User logout
		r.Group(func(r chi.Router) {
			r.Use(app.AuthUser)
			r.Get("/profile", app.Profile)       // Get currently logged-in user's profile
			//TODO: separate update profile functionality
			r.Put("/profile", app.UpdateProfile) // Update user profile information
			r.Put("/profile/deactivate", app.DeactivateProfile)        // Deactivate user profile information
			r.Delete("/profile/delete", app.DeleteProfile) // Delete user profile information
			// r.Put("/password", app.ChangePassword)      // Change password for logged-in user
			// r.Post("/forgot-password", app.ForgotPassword) // Request password reset via email
			// r.Post("/reset-password", app.ResetPassword)   // Reset password using token
		})
	})

	// --- Media Management ---
	mux.Route("/api/v1/media", func(r chi.Router) {
		r.Get("/", app.ListMedia)                // List all media
		r.Get("/details", app.FetchMediaDetails) // List all media
		r.Group(func(r chi.Router) {
			r.Use(app.AuthUser)
			r.Post("/", app.UploadMedia) // Upload new media
			// Secure premium endpoint
			r.Group(func(r chi.Router) { // Regular auth check
				// r.Use(app.WithSubscriptionCheck) // Premium subscription check

				r.Get("/premium", app.ServeMedia)
			}) // Retrieve a single media item by ID
		})

		// 	r.Put("/{id}", app.UpdateMedia)                   // Update an existing media item
		// 	r.Delete("/{id}", app.DeleteMedia)                // Delete a media item

		// 	r.Get("/user/{userId}", app.UserMedia)            // List all media uploaded by a specific user
		// 	r.Get("/category/{slug}", app.CategoryMedia)      // List media by category slug
	})

	// --- Categories Management ---
	mux.Route("/api/v1/categories", func(r chi.Router) {
		r.Get("/", app.GetMediaCategories) // List all categories
		r.Group(func(r chi.Router) {
			// TODO:
			// r.Use(app.AuthAdmin)
			r.Post("/", app.CreateMediaCategory)   // Create a new category
			r.Put("/", app.UpdateMediaCategory)    // Update an existing category
			r.Delete("/", app.DeleteMediaCategory) // Delete a category
		})
	})

	mux.Route("/api/v1/plans", func(r chi.Router) {
		r.Get("/",  app.GetPlans)
		r.Post("/", app.CreatePlan)
		r.Put("/", app.UpdatePlan)

		r.Group(func(r chi.Router) {
			r.Post("/purchase", app.PurchasePlan)
		})
	})

	mux.Route("/api/v1/history", func(r chi.Router) {
		r.Use(app.AuthUser)
		r.Get("/download", app.GetDownloadHistory)
		r.Get("/upload", app.GetUploadHistory)
	})

	return mux
}
