package routes

import (
	"net/http"
	"strings"

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

// SetupRoutes configures the HTTP routes
func (r *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/links", r.handleLinks)
	mux.HandleFunc("/api/links/", r.handleLinkByShort)

	// Health check endpoint
	mux.HandleFunc("/health", r.handleHealth)

	// Redirect route (catch-all)
	mux.HandleFunc("/", r.handleRedirect)

	logger.Info("Routes configured", logger.Fields{
		"endpoints": []string{"/api/links", "/api/links/{short}", "/health", "/{short}"},
	})

	// Add CORS middleware
	return r.corsMiddleware(mux)
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
	if strings.HasPrefix(req.URL.Path, "/api/") || req.URL.Path == "/health" {
		http.NotFound(w, req)
		return
	}

	r.linkHandler.RedirectLink(w, req)
}

// corsMiddleware adds CORS headers to responses
func (r *Router) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-User-ID")

		// Handle preflight requests
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, req)
	})
}
