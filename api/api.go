package api

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	db "github.com/samiulice/photostock/internal/database"
	"github.com/samiulice/photostock/internal/models"
	"github.com/samiulice/photostock/internal/repositories"
)

const version = models.APPVersion //app version

// config holds app configuration
type config struct {
	port int
	env  string //mediaion or development mode
	jwt  struct {
		secretKey string        //JWT secret key for signing tokens
		issuer    string        //Issuer of the JWT token
		expiry    time.Duration //Duration for which the JWT token is valid
		refresh   time.Duration //Duration for which the refresh token is valid
		audience  string        //Audience of the JWT token
		algorithm string        //Algorithm used for signing the JWT token
	}
	db struct {
		dsn string //Data source name : database connection name
	}
}

// application is the receiver for the various parts of the application
type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
	DB       *repositories.DBRepository
	Server   *http.Server
	ctx      context.Context
}

var app *application

// serve starts the server and listens for requests
func (app *application) serve() error {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Second,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Minute,
		WriteTimeout:      5 * time.Minute,
	}

	app.Server = srv
	app.infoLog.Printf("Starting HTTP Back end server in %s mode on port %d", app.config.env, app.config.port)
	app.infoLog.Println(".....................................")
	return srv.ListenAndServe()
}

// ShutdownServer gracefully shuts down the server
func (app *application) ShutdownServer() error {
	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	app.infoLog.Println("Shutting down the server gracefully...")
	// Shutdown the server with the context
	if err := app.Server.Shutdown(ctx); err != nil {
		app.errorLog.Printf("Server forced to shutdown: %s", err)
		return err
	}

	app.infoLog.Println("Server exited gracefully")
	return nil
}

// RunServer is the application entry point
func RunServer(ctx context.Context) error {
	var cfg config

	// Getting command line arguments
	flag.IntVar(&cfg.port, "port", 8080, "API Server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Application Environment{development|mediaion}")
	flag.Parse()

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	//setup JWT configuration
	cfg.jwt.secretKey = "photostock_app_v2_2024_Secure_JWT_Key_!@#$%^&*()_+" // Replace with your actual secret key
	cfg.jwt.issuer = "photostock_app_v2"
	cfg.jwt.expiry = 24 * time.Hour      // Token expiry duration
	cfg.jwt.refresh = 7 * 24 * time.Hour // Refresh token expiry duration
	cfg.jwt.audience = "photostock_app_v2"
	cfg.jwt.algorithm = "HS256" // JWT signing algorithm

	//for testing purpose
	cfg.db.dsn = "postgresql://photostock_db_kms3_user:0YqzS7ziqQjLx2nyfU9WGPdYRo7qNkd9@dpg-d1drniumcj7s73be3hqg-a.oregon-postgres.render.com/photostock_db_v1"
	// Connection to database
	dbConn, err := db.NewPgxPool(cfg.db.dsn)
	if err != nil {
		app.errorLog.Fatal(err)
		return err
	}
	defer dbConn.Close()
	db := repositories.NewDBRepository(dbConn)
	infoLog.Println("Connected to database")

	app = &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
		DB:       db,
		ctx:      ctx,
	}

	// Run the server in a separate goroutine so we can wait for shutdown signals
	go func() {
		if err := app.serve(); err != nil {
			errorLog.Printf("Error starting server: %s", err)
		}
	}()

	// Channel to listen for OS interrupt signals (e.g., from Ctrl+C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Wait for shutdown signal
	<-stop

	// Call ShutdownServer to gracefully shut down the server
	return app.ShutdownServer()
}

// Stop server from outer module
func StopServer() error {
	return app.ShutdownServer()
}
