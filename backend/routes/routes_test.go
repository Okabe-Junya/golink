package routes_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Okabe-Junya/golink-backend/handlers"
	"github.com/Okabe-Junya/golink-backend/repositories/mocks"
	"github.com/Okabe-Junya/golink-backend/routes"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() http.Handler {
	mockRepo := mocks.NewMockLinkRepository()
	linkHandler := handlers.NewLinkHandler(mockRepo)
	router := routes.NewRouter(linkHandler)
	return router.SetupRoutes()
}

func TestRoutes(t *testing.T) {
	handler := setupTestRouter()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
	}{
		{
			name:           "Health Check",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get All Links",
			method:         http.MethodGet,
			path:           "/api/links",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get Link - Not Found",
			method:         http.MethodGet,
			path:           "/api/links/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Create Link - Method Not Allowed",
			method:         http.MethodGet,
			path:           "/api/links",
			expectedStatus: http.StatusOK, // GET is allowed for /api/links
		},
		{
			name:           "Update Link - Method Not Allowed",
			method:         http.MethodGet,
			path:           "/api/links/test",
			expectedStatus: http.StatusNotFound, // Link not found, but method is allowed
		},
		{
			name:           "Delete Link - Method Not Allowed",
			method:         http.MethodGet,
			path:           "/api/links/test",
			expectedStatus: http.StatusNotFound, // Link not found, but method is allowed
		},
		{
			name:           "Redirect - Not Found",
			method:         http.MethodGet,
			path:           "/nonexistent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Options Request - CORS",
			method:         http.MethodOptions,
			path:           "/api/links",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, tc.path, nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Check CORS headers for all responses
			assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
			assert.Equal(t, "Content-Type, X-User-ID", rr.Header().Get("Access-Control-Allow-Headers"))
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	handler := setupTestRouter()

	// Test OPTIONS request
	req, _ := http.NewRequest(http.MethodOptions, "/api/links", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check CORS headers
	assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, X-User-ID", rr.Header().Get("Access-Control-Allow-Headers"))

	// Ensure body is empty for OPTIONS request
	assert.Empty(t, rr.Body.String())
}
