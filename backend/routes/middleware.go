package routes

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Okabe-Junya/golink-backend/logger"
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

// MetricsMiddleware collects metrics about HTTP requests
func MetricsMiddleware(next http.Handler) http.Handler {
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
		if path == "/" && !strings.HasPrefix(r.URL.Path, "/api/") && r.URL.Path != "/health" && r.URL.Path != "/" {
			RedirectsTotal.Inc()
		}
	})
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
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

				// Return a 500 error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security headers to responses
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
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

// RateLimitMiddleware limits the rate of requests per client IP
func RateLimitMiddleware(next http.Handler) http.Handler {
	// Simple in-memory rate limiter
	type client struct {
		lastSeen     time.Time
		blockedUntil time.Time

		count int
	}
	clients := make(map[string]*client)
	var mu sync.Mutex

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

// RequestIDMiddleware adds a unique ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
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

// Helper functions and types

// statusResponseWriter is a wrapper for http.ResponseWriter that tracks the status code
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

	return path
}
