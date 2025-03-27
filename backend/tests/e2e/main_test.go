package e2e

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Okabe-Junya/golink-backend/auth"
	"github.com/Okabe-Junya/golink-backend/handlers"
	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/Okabe-Junya/golink-backend/repositories/mocks"
)

var (
	// Global variables for tests
	apiBaseURL string
	testClient *http.Client
	testUserID string
	mockRepo   *mocks.MockLinkRepository
	testServer *httptest.Server
)

// TestMain runs before all tests
func TestMain(m *testing.M) {
	// Setup
	if err := setup(); err != nil {
		log.Fatalf("Failed to set up test environment: %v", err)
	}

	// Run tests
	code := m.Run()

	// Teardown
	if err := teardown(); err != nil {
		log.Printf("Error during test environment cleanup: %v", err)
	}

	os.Exit(code)
}

// mockAuthMiddleware is a simplified auth middleware for testing
func mockAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user *auth.User

		// Check for Authorization header
		authHeader := r.Header.Get("Authorization")

		switch {
		case authHeader == "":
			// No auth header, treat as anonymous
			user = &auth.User{
				ID:    "anonymous",
				Email: "anonymous@example.com",
				Name:  "Anonymous User",
			}
		case strings.HasPrefix(authHeader, "Bearer "):
			// Parse Bearer token
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 {
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}

			userID := tokenParts[1]
			user = &auth.User{
				ID:    userID,
				Email: userID + "@example.com",
				Name:  "Test User (" + userID + ")",
			}
		default:
			// Invalid auth header format
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Add user to context and request context key
		ctx := auth.ContextWithUser(r.Context(), user)
		ctx = context.WithValue(ctx, "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// setup prepares the test environment
func setup() error {
	// Initialize mock repository
	mockRepo = mocks.NewMockLinkRepository()

	// Initialize handlers
	linkHandler := handlers.NewLinkHandler(mockRepo)

	// Create a test server
	mux := http.NewServeMux()

	// Set up API routes
	// GET and DELETE /api/links/:short
	mux.HandleFunc("/api/links/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Apply auth middleware and serve the request
			mockAuthMiddleware(linkHandler.GetLink).ServeHTTP(w, r)
		case http.MethodDelete:
			// Apply auth middleware and serve the request
			mockAuthMiddleware(linkHandler.DeleteLink).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// POST /api/links
	mux.HandleFunc("/api/links", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Apply auth middleware and serve the request
			mockAuthMiddleware(linkHandler.CreateLink).ServeHTTP(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Create test server
	testServer = httptest.NewServer(mux)
	apiBaseURL = testServer.URL

	// Create HTTP client
	testClient = &http.Client{
		Timeout: 10 * time.Second,
	}

	// Set test user ID
	testUserID = "test-user-123"

	// Create test data
	if err := createTestData(); err != nil {
		return err
	}

	return nil
}

// createTestData creates test data in the mock repository
func createTestData() error {
	ctx := context.Background()

	// Create test links with different access levels
	testLinks := []struct {
		short       string
		url         string
		accessLevel string
	}{
		{"test-public", "https://example.com/public", "Public"},
		{"test-private", "https://example.com/private", "Private"},
		{"test-restricted", "https://example.com/restricted", "Restricted"},
	}

	for _, tl := range testLinks {
		// Create link in repository
		link := &models.Link{
			Short:        tl.short,
			URL:          tl.url,
			AccessLevel:  tl.accessLevel,
			CreatedBy:    testUserID,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			AllowedUsers: []string{},
		}

		if tl.accessLevel == "Restricted" {
			link.AllowedUsers = []string{"allowed-user-456"}
		}

		// Add to mock repository
		if err := mockRepo.Create(ctx, link); err != nil {
			return err
		}
	}

	return nil
}

// teardown cleans up the test environment
func teardown() error {
	if testServer != nil {
		testServer.Close()
	}
	return nil
}
