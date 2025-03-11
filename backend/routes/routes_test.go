package routes_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Okabe-Junya/golink-backend/handlers"
	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/Okabe-Junya/golink-backend/repositories/mocks"
	"github.com/Okabe-Junya/golink-backend/routes"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() http.Handler {
	// Skip authentication for testing
	os.Setenv("TEST_MODE", "true")
	defer os.Unsetenv("TEST_MODE")

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
		body           interface{}
		userID         string
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name:           "Health Check",
			method:         http.MethodGet,
			path:           "/health",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]string{
				"status": "healthy",
			},
		},
		{
			name:   "Create Link",
			method: http.MethodPost,
			path:   "/api/links",
			body: map[string]string{
				"short": "test",
				"url":   "https://example.com",
			},
			userID:         "test-user",
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "Create Link - Invalid Body",
			method: http.MethodPost,
			path:   "/api/links",
			body: map[string]string{
				"invalid": "data",
			},
			userID:         "test-user",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get Links",
			method:         http.MethodGet,
			path:           "/api/links",
			userID:         "test-user",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get Link - Not Found",
			method:         http.MethodGet,
			path:           "/api/links/nonexistent",
			userID:         "test-user",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "Update Link - Non-existent",
			method: http.MethodPut,
			path:   "/api/links/nonexistent",
			body: map[string]string{
				"url": "https://updated.com",
			},
			userID:         "test-user",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Delete Link - Non-existent",
			method:         http.MethodDelete,
			path:           "/api/links/nonexistent",
			userID:         "test-user",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Options Request",
			method:         http.MethodOptions,
			path:           "/api/links",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			var err error
			if tc.body != nil {
				bodyBytes, err = json.Marshal(tc.body)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest(tc.method, tc.path, bytes.NewBuffer(bodyBytes))
			assert.NoError(t, err)

			if tc.userID != "" {
				req.Header.Set("X-User-ID", tc.userID)
			}
			if bodyBytes != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)
			// Check CORS headers
			corsOrigin := os.Getenv("CORS_ORIGIN")
			if corsOrigin == "" {
				corsOrigin = "http://localhost:3001"
			}
			assert.Equal(t, corsOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
			assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
			assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "X-User-ID")
			// Check response body for specific cases
			if tc.expectedBody != nil {
				var responseBody map[string]string
				err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				expectedBody := tc.expectedBody.(map[string]string)
				for k, v := range expectedBody {
					assert.Equal(t, v, responseBody[k])
				}
			}
		})
	}
}

func TestEndToEndLinkOperations(t *testing.T) {
	handler := setupTestRouter()

	// Test user identifier
	userID := "test-user"

	// 1. Create a new link
	createBody := map[string]string{
		"short": "test-link",
		"url":   "https://example.com",
	}
	bodyBytes, err := json.Marshal(createBody)
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// 2. Retrieve the created link
	req, _ = http.NewRequest(http.MethodGet, "/api/links/test-link", nil)
	req.Header.Set("X-User-ID", userID)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	var link models.Link
	err = json.Unmarshal(rr.Body.Bytes(), &link)
	assert.NoError(t, err)
	assert.Equal(t, "test-link", link.Short)
	assert.Equal(t, "https://example.com", link.URL)

	// 3. Update the link
	updateBody := map[string]string{
		"url": "https://updated-example.com",
	}
	bodyBytes, _ = json.Marshal(updateBody)

	req, _ = http.NewRequest(http.MethodPut, "/api/links/test-link", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// 4. Verify the update
	req, _ = http.NewRequest(http.MethodGet, "/api/links/test-link", nil)
	req.Header.Set("X-User-ID", userID)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	err = json.Unmarshal(rr.Body.Bytes(), &link)
	assert.NoError(t, err)
	assert.Equal(t, "https://updated-example.com", link.URL)

	// 5. Delete the link
	req, _ = http.NewRequest(http.MethodDelete, "/api/links/test-link", nil)
	req.Header.Set("X-User-ID", userID)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code) // Changed from StatusOK to StatusNoContent

	// 6. Verify the deletion
	req, _ = http.NewRequest(http.MethodGet, "/api/links/test-link", nil)
	req.Header.Set("X-User-ID", userID)

	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestCORSMiddleware(t *testing.T) {
	originalCorsOrigin := os.Getenv("CORS_ORIGIN")
	defer os.Setenv("CORS_ORIGIN", originalCorsOrigin)

	testCases := []struct {
		name           string
		corsOrigin     string
		expectedOrigin string
	}{
		{
			name:           "With CORS_ORIGIN set",
			corsOrigin:     "http://test.example.com",
			expectedOrigin: "http://test.example.com",
		},
		{
			name:           "Without CORS_ORIGIN",
			corsOrigin:     "",
			expectedOrigin: "http://localhost:3001",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("CORS_ORIGIN", tc.corsOrigin)
			handler := setupTestRouter()

			req, _ := http.NewRequest(http.MethodOptions, "/api/links", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			// Check status code
			assert.Equal(t, http.StatusOK, rr.Code)

			// Check CORS headers
			assert.Equal(t, tc.expectedOrigin, rr.Header().Get("Access-Control-Allow-Origin"))
			assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
			assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
			assert.Contains(t, rr.Header().Get("Access-Control-Allow-Headers"), "Authorization")
			assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))

			// Ensure body is empty for OPTIONS request
			assert.Empty(t, rr.Body.String())
		})
	}
}
