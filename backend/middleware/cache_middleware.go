package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Okabe-Junya/golink-backend/logger"
)

// CacheItem represents a cached HTTP response
type CacheItem struct {
	CreatedAt   time.Time
	ContentType string
	Content     []byte
	Expiry      time.Duration
	StatusCode  int
}

// Cache is a simple in-memory cache for HTTP responses
type Cache struct {
	items map[string]CacheItem
	mutex sync.RWMutex
}

// Global cache instance
var (
	responseCache = NewCache()
)

// NewCache creates a new cache
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	// Start background cleanup
	go cache.periodicCleanup()

	return cache
}

// periodicCleanup removes expired items from the cache every minute
func (c *Cache) periodicCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		c.mutex.Lock()

		for key, item := range c.items {
			if now.Sub(item.CreatedAt) > item.Expiry {
				delete(c.items, key)
				logger.Info("Cache item expired and removed", logger.Fields{
					"key": key,
				})
			}
		}

		c.mutex.Unlock()
	}
}

// Set adds an item to the cache
func (c *Cache) Set(key string, content []byte, contentType string, statusCode int, expiry time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = CacheItem{
		Content:     content,
		ContentType: contentType,
		StatusCode:  statusCode,
		CreatedAt:   time.Now(),
		Expiry:      expiry,
	}

	logger.Info("Added item to cache", logger.Fields{
		"key":    key,
		"expiry": expiry.String(),
	})
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) (CacheItem, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found {
		return CacheItem{}, false
	}

	// Check if the item is expired
	if time.Since(item.CreatedAt) > item.Expiry {
		return CacheItem{}, false
	}

	return item, true
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, found := c.items[key]; found {
		delete(c.items, key)
		logger.Info("Removed item from cache", logger.Fields{"key": key})
	}
}

// createCacheKey generates a unique key for the request
func createCacheKey(r *http.Request) string {
	// For simplicity, we'll use the request path and query as the cache key
	// In a real-world scenario, you might want to include other things like authorization headers
	path := r.URL.Path
	query := r.URL.Query().Encode()

	// Combine path and query and create a hash
	keyStr := fmt.Sprintf("%s?%s", path, query)
	hasher := sha256.New()
	hasher.Write([]byte(keyStr))

	return hex.EncodeToString(hasher.Sum(nil))
}

// CacheMiddleware is a middleware that caches responses
func CacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Invalidate cache for non-GET requests
		if r.Method != http.MethodGet {
			// Generate cache key for the path
			key := createCacheKey(&http.Request{
				URL:    r.URL,
				Method: http.MethodGet,
			})
			responseCache.Delete(key)
			next.ServeHTTP(w, r)
			return
		}

		// Skip caching for certain paths
		if strings.HasPrefix(r.URL.Path, "/api/auth") ||
			r.URL.Path == "/health" ||
			r.URL.Path == "/health/detailed" ||
			r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		// Generate cache key
		key := createCacheKey(r)

		// Check if we have a cached response
		if item, found := responseCache.Get(key); found {
			// Set the content type and status code from the cached response
			w.Header().Set("Content-Type", item.ContentType)
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(item.StatusCode)

			// Write the cached content
			_, err := w.Write(item.Content)
			if err != nil {
				logger.Error("Failed to write cached response", err, logger.Fields{
					"key": key,
				})
			}

			logger.Info("Cache hit", logger.Fields{
				"path": r.URL.Path,
				"key":  key,
			})

			return
		}

		// Create a custom response writer to capture the response
		crw := &cachingResponseWriter{
			ResponseWriter: w,
			key:            key,
			statusCode:     http.StatusOK,
			path:           r.URL.Path,
		}

		// Set header to indicate cache miss
		w.Header().Set("X-Cache", "MISS")

		// Call the next handler with our custom response writer
		next.ServeHTTP(crw, r)

		// Finalize the caching process after the response is written
		crw.Close()
	})
}

// cachingResponseWriter is a custom response writer that captures the response for caching
type cachingResponseWriter struct {
	http.ResponseWriter
	key        string
	path       string
	content    bytes.Buffer
	statusCode int
	written    bool
}

// WriteHeader captures the status code and calls the underlying ResponseWriter.WriteHeader
func (crw *cachingResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
	crw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the content and calls the underlying ResponseWriter.Write
func (crw *cachingResponseWriter) Write(b []byte) (int, error) {
	// Only buffer the response if it's a success response
	if crw.statusCode < 400 {
		crw.content.Write(b)
	}

	crw.written = true
	// Write the actual response
	return crw.ResponseWriter.Write(b)
}

// Close is called after the response is written
// It adds the response to the cache if it was successful
func (crw *cachingResponseWriter) Close() {
	// Only cache successful responses that have actually been written
	if crw.statusCode < 400 && crw.written {
		// Determine expiry time based on the path
		var expiry time.Duration

		switch {
		case strings.HasPrefix(crw.path, "/api/analytics"):
			expiry = 5 * time.Minute // Cache analytics for a shorter time
		case strings.HasPrefix(crw.path, "/api/links"):
			expiry = 15 * time.Minute // Cache link data for a moderate time
		default:
			expiry = 30 * time.Minute // Cache redirects for longer
		}

		// Add the response to the cache
		contentType := crw.ResponseWriter.Header().Get("Content-Type")
		if contentType == "" {
			contentType = "application/json" // Default content type
		}

		responseCache.Set(crw.key, crw.content.Bytes(), contentType, crw.statusCode, expiry)

		logger.Info("Cached response", logger.Fields{
			"path":      crw.path,
			"key":       crw.key,
			"expiry":    expiry.String(),
			"size":      crw.content.Len(),
			"mediaType": contentType,
		})
	}
}
