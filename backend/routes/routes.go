package routes

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/Okabe-Junya/golink-backend/auth"
	"github.com/Okabe-Junya/golink-backend/handlers"
	"github.com/Okabe-Junya/golink-backend/logger"
)

// Router handles HTTP routing
type Router struct {
	linkHandler *handlers.LinkHandler
}

// NewRouter creates a new Router
func NewRouter(linkHandler *handlers.LinkHandler) *Router {
	return &Router{
		linkHandler: linkHandler,
	}
}

// CORSMiddleware adds CORS headers to all responses
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3001")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass to the next middleware
		next.ServeHTTP(w, r)
	})
}

// SetupRoutes configures the HTTP routes
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/links", r.handleLinks)
	mux.HandleFunc("/api/links/", r.handleLinkByShort)

	// Auth routes
	mux.HandleFunc("/api/auth/login", auth.HandleLogin)
	mux.HandleFunc("/api/auth/callback", auth.HandleCallback)
	mux.HandleFunc("/api/auth/user", r.handleCurrentUser)

	// Health check endpoint
	mux.HandleFunc("/health", r.handleHealth)

	// Redirect route (catch-all)
	mux.HandleFunc("/", r.handleRedirect)

	logger.Info("Routes configured", logger.Fields{
		"endpoints": []string{
			"/api/links",
			"/api/links/{short}",
			"/api/auth/login",
			"/api/auth/callback",
			"/api/auth/user",
			"/health",
			"/{short}",
		},
	})

	// Apply middlewares in the correct order:
	// 1. CORS middleware first so headers are always set
	// 2. Authentication middleware second
	handler := CORSMiddleware(mux)

	// Only apply auth middleware if not in test mode
	if os.Getenv("TEST_MODE") != "true" {
		handler = auth.AuthMiddleware(handler)
	}

	return handler
}

// handleCurrentUser handles /api/auth/user requests
func (r *Router) handleCurrentUser(w http.ResponseWriter, req *http.Request) {
	// Only GET requests are allowed
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user
	user, err := auth.GetCurrentUser(req)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Return user info as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// We can just use the json encoding from the auth package since it already has json tags
	if err := json.NewEncoder(w).Encode(user); err != nil {
		logger.Error("Failed to encode user", err, nil)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// handleLinks handles /api/links requests
func (r *Router) handleLinks(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		r.linkHandler.GetLinks(w, req)
	case http.MethodPost:
		r.linkHandler.CreateLink(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleLinkByShort handles /api/links/{short} requests
func (r *Router) handleLinkByShort(w http.ResponseWriter, req *http.Request) {
	// Extract the short code from the URL
	path := req.URL.Path
	if !strings.HasPrefix(path, "/api/links/") {
		http.NotFound(w, req)
		return
	}

	switch req.Method {
	case http.MethodGet:
		r.linkHandler.GetLink(w, req)
	case http.MethodPut:
		r.linkHandler.UpdateLink(w, req)
	case http.MethodDelete:
		r.linkHandler.DeleteLink(w, req)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleHealth handles /health requests for health check
func (r *Router) handleHealth(w http.ResponseWriter, req *http.Request) {
	r.linkHandler.HealthCheck(w, req)
}

// handleRedirect handles /{short} requests for redirecting
func (r *Router) handleRedirect(w http.ResponseWriter, req *http.Request) {
	// Skip API routes and health check
	if strings.HasPrefix(req.URL.Path, "/api/") || req.URL.Path == "/health" || req.URL.Path == "/" {
		http.NotFound(w, req)
		return
	}

	// Remove leading slash
	shortCode := strings.TrimPrefix(req.URL.Path, "/")
	if shortCode == "" {
		http.NotFound(w, req)
		return
	}

	r.linkHandler.RedirectLink(w, req)
}
