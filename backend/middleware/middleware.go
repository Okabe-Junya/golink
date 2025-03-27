package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/Okabe-Junya/golink-backend/auth"
	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/pkg/errors"
	"github.com/Okabe-Junya/golink-backend/pkg/response"
)

// Middleware represents an HTTP middleware
type Middleware func(http.Handler) http.Handler

// Chain applies multiple middlewares to a handler
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// Logging logs all requests with their path, method, and duration
func Logging() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture status code
			rw := NewResponseWriter(w)

			// Call the next handler
			next.ServeHTTP(rw, r)

			// Log the request
			duration := time.Since(start)
			logger.Info("Request completed", logger.Fields{
				"method":     r.Method,
				"path":       r.URL.Path,
				"query":      r.URL.RawQuery,
				"user_agent": r.UserAgent(),
				"status":     rw.Status(),
				"duration":   duration.String(),
			})
		})
	}
}

// Recover recovers from panics and returns a 500 error
func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					switch e := err.(type) {
					case error:
						logger.Error("Panic recovered", e, logger.Fields{
							"path":   r.URL.Path,
							"method": r.Method,
						})
					default:
						logger.Error("Panic recovered", errors.NewInternalError(nil), logger.Fields{
							"path":   r.URL.Path,
							"method": r.Method,
						})
					}
					response.Error(w, errors.NewInternalError(nil))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// Authenticate checks if the request is authenticated
func Authenticate() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from header for backward compatibility
			userID := r.Header.Get("X-User-ID")
			if userID != "" {
				// Create a minimal user with just the ID
				user := &auth.User{
					ID:    userID,
					Email: r.Header.Get("X-User-Email"),
					Name:  r.Header.Get("X-User-Name"),
				}

				// Add user to context
				ctx := r.Context()
				ctx = context.WithValue(ctx, "user", user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Continue without authentication for public endpoints
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth requires authentication for the handler
func RequireAuth() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from header
			userID := r.Header.Get("X-User-ID")
			if userID == "" {
				response.Error(w, errors.NewUnauthorized("認証が必要です"))
				return
			}

			// Create a minimal user with just the ID
			user := &auth.User{
				ID:    userID,
				Email: r.Header.Get("X-User-Email"),
				Name:  r.Header.Get("X-User-Name"),
			}

			// Add user to context
			ctx := r.Context()
			ctx = context.WithValue(ctx, "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// CORS adds CORS headers to the response
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if the origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			// Set CORS headers
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// ResponseWriter is a wrapper around http.ResponseWriter that captures the status code
type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Status returns the status code
func (rw *ResponseWriter) Status() int {
	return rw.statusCode
}
