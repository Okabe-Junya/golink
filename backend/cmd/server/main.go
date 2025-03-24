package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/Okabe-Junya/golink-backend/auth"
	"github.com/Okabe-Junya/golink-backend/handlers"
	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/repositories"
	"github.com/Okabe-Junya/golink-backend/routes"
	"github.com/rs/cors"
	"google.golang.org/api/option"
)

// Constants for timeouts
const (
	readTimeout     = 5 * time.Second
	writeTimeout    = 10 * time.Second
	idleTimeout     = 120 * time.Second
	shutdownTimeout = 30 * time.Second
)

// initFirebase initializes the Firebase app and Firestore client
func initFirebase() (*firestore.Client, error) {
	ctx := context.Background()
	var opt option.ClientOption
	credJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	credFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")

	// Rewritten using switch
	switch {
	case credJSON != "":
		opt = option.WithCredentialsJSON([]byte(credJSON))
	case credFile != "":
		opt = option.WithCredentialsFile(credFile)
	default:
		credFile = "path/to/serviceAccountKey.json"
		opt = option.WithCredentialsFile(credFile)
	}

	// Initialize Firebase app
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	// Initialize Firestore client
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firestore: %v", err)
	}

	return client, nil
}

func main() {
	// Initialize Firebase
	client, err := initFirebase()
	if err != nil {
		logger.Fatal("Failed to initialize Firebase", err, nil)
	}
	defer client.Close()

	// Initialize authentication system
	if err := auth.InitSessionManager(); err != nil {
		logger.Warn("Failed to initialize session manager", logger.Fields{"error": err.Error()})
	}
	if err := auth.InitAuth(); err != nil {
		logger.Warn("Failed to initialize authentication", logger.Fields{"error": err.Error()})
	}
	logger.Info("Authentication system initialized successfully", nil)

	// Get domain from environment variable or use default
	domain := os.Getenv("APP_DOMAIN")
	if domain == "" {
		domain = "localhost:8080"
		logger.Warn("APP_DOMAIN environment variable not set", logger.Fields{
			"default_domain": domain,
			"message":        "Please set APP_DOMAIN for production use",
		})
	}

	// Create repository
	linkRepo := repositories.NewLinkRepository(client)

	// Create handlers
	linkHandler := handlers.NewLinkHandler(linkRepo)
	healthHandler := handlers.NewHealthHandler(linkRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(linkRepo)

	// Set up routes
	router := routes.NewRouter(linkHandler, healthHandler, analyticsHandler)
	handler := router.SetupRoutes()

	// Setup CORS
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "http://localhost:3001"
		logger.Warn("CORS_ORIGIN environment variable not set", logger.Fields{
			"default_origin": corsOrigin,
		})
	}

	corsHandler := cors.New(cors.Options{
		AllowedOrigins:   []string{corsOrigin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}).Handler(handler)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsHandler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	// Create a channel to listen for shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		logger.Info("Server starting", logger.Fields{
			"port":        port,
			"domain":      domain,
			"cors_origin": corsOrigin,
			"version":     os.Getenv("APP_VERSION"),
		})

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server error", err, nil)
		}
	}()

	// Wait for shutdown signal
	<-stop
	logger.Info("Server is shutting down...", nil)

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Shutdown the server gracefully
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", err, nil)
	}

	logger.Info("Server exited gracefully", nil)
}
