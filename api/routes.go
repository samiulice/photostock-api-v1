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

	//login routes
	mux.Post("/api/v1/auth/signin", app.SignIn)
	mux.Post("/api/v1/user", app.AddUser)
	//Secure routes
	mux.Route("/api/v1", func(mux chi.Router) {
		mux.Use(app.AuthUser)
	})

	// Media routes
	mux.Route("/media", func(r chi.Router) {
		// mux.Get("/", app.ListMedia)
		// mux.Post("/", app.UploadMedia)
		// mux.Get("/{id}", app.GetMedia)
		// mux.Put("/{id}", app.UpdateMedia)
		// mux.Delete("/{id}", app.DeleteMedia)

		// mux.Get("/user/{userId}", app.UserMedia)
		// mux.Get("/category/{slug}", app.CategoryMedia)
	})
	return mux
}
