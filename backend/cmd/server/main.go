package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/Okabe-Junya/golink-backend/handlers"
	"github.com/Okabe-Junya/golink-backend/repositories"
	"github.com/Okabe-Junya/golink-backend/routes"
	"google.golang.org/api/option"
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
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	defer client.Close()

	// Get domain from environment variable or use default
	domain := os.Getenv("APP_DOMAIN")
	if domain == "" {
		domain = "example.com"
		log.Printf("No APP_DOMAIN environment variable set, using default: %s", domain)
	}

	// Create repository
	linkRepo := repositories.NewLinkRepository(client)
	// Create handler
	linkHandler := handlers.NewLinkHandler(linkRepo)
	// Set up routes
	router := routes.NewRouter(linkHandler)
	handler := router.SetupRoutes()
	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Application domain set to: %s", domain)
	// Replace log.Fatal with explicit error handling to ensure client.Close() runs; though defer won't run on os.Exit, we close before exit
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Printf("Server error: %v", err)
		client.Close()
		return
	}
}
