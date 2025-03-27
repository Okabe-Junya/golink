package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Okabe-Junya/golink-backend/auth"
	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/pkg/errors"
	"github.com/Okabe-Junya/golink-backend/pkg/response"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestsTotal counts total HTTP requests by path and method
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "golink_requests_total",
			Help: "Total number of HTTP requests by path and method",
		},
		[]string{"path", "method", "status"},
	)

	// RequestDuration measures the duration of HTTP requests
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "golink_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	// ActiveRequests tracks currently active requests
	ActiveRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "golink_active_requests",
			Help: "Current number of active HTTP requests",
		},
	)

	// RedirectsTotal counts link redirects
	RedirectsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "golink_redirects_total",
			Help: "Total number of link redirects",
		},
	)

	// ErrorsTotal counts HTTP errors
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "golink_errors_total",
			Help: "Total number of HTTP errors",
		},
		[]string{"path", "method", "status"},
	)
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
					// Log the error and stack trace
					stackTrace := debug.Stack()
					logger.Error("Request handler panic", fmt.Errorf("%v", err), logger.Fields{
						"path":        r.URL.Path,
						"method":      r.Method,
						"remoteAddr":  r.RemoteAddr,
						"stackTrace":  string(stackTrace),
						"userAgent":   r.UserAgent(),
						"requestId":   r.Header.Get("X-Request-ID"),
						"error":       err,
						"recoverable": true,
					})

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
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID, X-User-Email, X-User-Name")
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

// Metrics collects metrics about HTTP requests
func Metrics() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start a timer
			start := time.Now()

			// Increment active requests counter
			ActiveRequests.Inc()
			defer ActiveRequests.Dec()

			// Create a response writer that tracks status code
			ww := newStatusResponseWriter(w)

			// Process the request
			next.ServeHTTP(ww, r)

			// Get normalized path for metrics (prevent cardinality explosion)
			path := normalizePath(r.URL.Path)

			// Record metrics
			duration := time.Since(start).Seconds()
			RequestDuration.WithLabelValues(path, r.Method).Observe(duration)
			RequestsTotal.WithLabelValues(path, r.Method, strconv.Itoa(ww.status)).Inc()

			// Record error metrics for 4xx and 5xx responses
			if ww.status >= 400 {
				ErrorsTotal.WithLabelValues(path, r.Method, strconv.Itoa(ww.status)).Inc()
			}

			// Track redirects
			if path == "/{short}" {
				RedirectsTotal.Inc()
			}
		})
	}
}

// RequestID adds a unique ID to each request
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = fmt.Sprintf("%d", time.Now().UnixNano())
				r.Header.Set("X-Request-ID", requestID)
			}
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self';")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit limits the rate of requests per client IP
func RateLimit() Middleware {
	// Simple in-memory rate limiter
	type client struct {
		lastSeen     time.Time
		blockedUntil time.Time
		count        int
	}
	clients := make(map[string]*client)
	var mu sync.Mutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := r.RemoteAddr
			if fwdIP := r.Header.Get("X-Forwarded-For"); fwdIP != "" {
				ip = strings.Split(fwdIP, ",")[0]
			}

			// Check if client is rate limited
			mu.Lock()
			defer mu.Unlock()

			now := time.Now()
			// Clean up old clients every 100 requests, but only when we have clients
			if len(clients) > 0 && len(clients)%100 == 0 {
				for clientIP, c := range clients {
					if now.Sub(c.lastSeen) > 5*time.Minute {
						delete(clients, clientIP)
					}
				}
			}

			// Get or create client
			c, exists := clients[ip]
			if !exists {
				c = &client{
					lastSeen:     now,
					count:        0,
					blockedUntil: time.Time{},
				}
				clients[ip] = c
			}

			// Check if client is blocked
			if !c.blockedUntil.IsZero() && now.Before(c.blockedUntil) {
				w.Header().Set("Retry-After", strconv.Itoa(int(c.blockedUntil.Sub(now).Seconds())))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			// Reset counter if last request was more than a minute ago
			if now.Sub(c.lastSeen) > time.Minute {
				c.count = 0
			}

			// Increment counter
			c.count++
			c.lastSeen = now

			// Block client if too many requests
			if c.count > 100 { // 100 requests per minute
				c.blockedUntil = now.Add(time.Minute)
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
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

// statusResponseWriter is a wrapper for http.ResponseWriter that tracks the status code for metrics
type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	return &statusResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// normalizePath returns a normalized path for metrics to prevent cardinality explosion
func normalizePath(path string) string {
	// Special case for redirects
	if !strings.HasPrefix(path, "/api/") && path != "/health" && path != "/" {
		return "/{short}"
	}

	// Replace dynamic parts in API paths
	if strings.HasPrefix(path, "/api/links/") && len(path) > len("/api/links/") {
		return "/api/links/{short}"
	}

	if strings.HasPrefix(path, "/api/analytics/links/") && len(path) > len("/api/analytics/links/") {
		return "/api/analytics/links/{short}"
	}

	return path
}
