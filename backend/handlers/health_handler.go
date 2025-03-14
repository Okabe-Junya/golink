package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// LastHealthCheckStatus stores the last health check status
	LastHealthCheckStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "golink_health_check_status",
			Help: "Status of the last health check (1 = healthy, 0 = unhealthy)",
		},
	)

	// LastHealthCheckTimestamp stores the timestamp of the last health check
	LastHealthCheckTimestamp = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "golink_health_check_timestamp",
			Help: "Timestamp of the last health check",
		},
	)
)

// HealthHandler handles health check requests
type HealthHandler struct {
	startTime time.Time
	repo      interface {
		GetAll(ctx context.Context) ([]*models.Link, error)
	}
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(repo interface {
	GetAll(ctx context.Context) ([]*models.Link, error)
}) *HealthHandler {
	return &HealthHandler{
		startTime: time.Now(),
		repo:      repo,
	}
}

// HealthInfo contains information about the health of the application
type HealthInfo struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Uptime    string            `json:"uptime"`
	Build     string            `json:"build,omitempty"`
	Database  map[string]string `json:"database"`
	System    SystemInfo        `json:"system"`
}

// SystemInfo contains information about the system
type SystemInfo struct {
	GoVersion  string `json:"go_version"`
	GOOS       string `json:"os"`
	GOARCH     string `json:"arch"`
	NumCPU     int    `json:"num_cpu"`
	Goroutines int    `json:"goroutines"`
	MemStats   struct {
		Alloc      uint64 `json:"alloc"`       // bytes allocated and not yet freed
		TotalAlloc uint64 `json:"total_alloc"` // total bytes allocated (even if freed)
		Sys        uint64 `json:"sys"`         // bytes obtained from system
		NumGC      uint32 `json:"num_gc"`      // number of completed GC cycles
	} `json:"mem_stats"`
}

// HealthCheck handles GET /health requests for detailed health check
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current time
	now := time.Now()

	// Check if we can connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dbStatus := map[string]string{
		"status": "connected",
	}

	_, err := h.repo.GetAll(ctx)
	if err != nil {
		dbStatus["status"] = "disconnected"
		dbStatus["error"] = err.Error()
		LastHealthCheckStatus.Set(0)
	} else {
		LastHealthCheckStatus.Set(1)
	}

	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Calculate uptime
	uptime := now.Sub(h.startTime).String()

	// Build the response
	response := HealthInfo{
		Status:    "healthy",
		Timestamp: now.Format(time.RFC3339),
		Version:   os.Getenv("APP_VERSION"),
		Build:     os.Getenv("BUILD_ID"),
		Uptime:    uptime,
		Database:  dbStatus,
		System: SystemInfo{
			GoVersion:  runtime.Version(),
			GOOS:       runtime.GOOS,
			GOARCH:     runtime.GOARCH,
			NumCPU:     runtime.NumCPU(),
			Goroutines: runtime.NumGoroutine(),
		},
	}

	// Set memory stats
	response.System.MemStats.Alloc = memStats.Alloc
	response.System.MemStats.TotalAlloc = memStats.TotalAlloc
	response.System.MemStats.Sys = memStats.Sys
	response.System.MemStats.NumGC = memStats.NumGC

	// Set status based on database connection
	if dbStatus["status"] != "connected" {
		response.Status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
		logger.Error("Health check failed", err, nil)
	} else {
		logger.Info("Health check passed", nil)
	}

	// Store the timestamp of the last health check
	LastHealthCheckTimestamp.Set(float64(now.Unix()))

	// Return the health check response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode health check response", http.StatusInternalServerError)
	}
}

// SimpleHealthCheck handles GET /health/simple requests for a simple health check
func (h *HealthHandler) SimpleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if we can connect to the database
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := h.repo.GetAll(ctx)

	response := map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err != nil {
		response["status"] = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
		LastHealthCheckStatus.Set(0)
	} else {
		LastHealthCheckStatus.Set(1)
	}

	// Store the timestamp of the last health check
	LastHealthCheckTimestamp.Set(float64(time.Now().Unix()))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode health check response", http.StatusInternalServerError)
	}
}
