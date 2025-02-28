package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Okabe-Junya/golink-backend/models"
	"github.com/Okabe-Junya/golink-backend/repositories/mocks"
	"github.com/stretchr/testify/assert"
)

// setupTestHandler creates a new LinkHandler with a mock repository for testing
func setupTestHandler() (*LinkHandler, *mocks.MockLinkRepository) {
	mockRepo := mocks.NewMockLinkRepository()
	handler := NewLinkHandler(mockRepo)
	return handler, mockRepo
}

// createTestLink creates a test link
func createTestLink(short, url, userID string) *models.Link {
	link := models.NewLink(short, url, userID)
	link.AccessLevel = models.AccessLevels.Public
	return link
}

func TestCreateLink(t *testing.T) {
	// Setup
	handler, mockRepo := setupTestHandler()

	// Create a test link
	testLink := createTestLink("test", "https://example.com", "user1")

	// Add the link to the mock repository to test conflict case
	mockRepo.Create(context.Background(), testLink)

	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		userID         string
	}{
		{
			name: "Valid Link Creation",
			requestBody: map[string]string{
				"short": "newlink",
				"url":   "https://example.com",
			},
			expectedStatus: http.StatusCreated,
			userID:         "user1",
		},
		{
			name: "Missing Short Code",
			requestBody: map[string]string{
				"url": "https://example.com",
			},
			expectedStatus: http.StatusBadRequest,
			userID:         "user1",
		},
		{
			name: "Invalid Short Code Format",
			requestBody: map[string]string{
				"short": "invalid@code",
				"url":   "https://example.com",
			},
			expectedStatus: http.StatusBadRequest,
			userID:         "user1",
		},
		{
			name: "Conflict - Short Code Already Exists",
			requestBody: map[string]string{
				"short": "test", // This already exists in the mock repo
				"url":   "https://example.com",
			},
			expectedStatus: http.StatusConflict,
			userID:         "user1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/api/links", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if tc.userID != "" {
				req.Header.Set("X-User-ID", tc.userID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.CreateLink(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// If successful, check response body
			if tc.expectedStatus == http.StatusCreated {
				var response models.Link
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tc.requestBody["short"], response.Short)
				assert.Equal(t, tc.requestBody["url"], response.URL)
				assert.Equal(t, tc.userID, response.CreatedBy)
			}
		})
	}
}

func TestGetLinks(t *testing.T) {
	// Setup
	handler, mockRepo := setupTestHandler()

	// Add some test links to the mock repository
	ctx := context.Background()
	mockRepo.Create(ctx, createTestLink("test1", "https://example1.com", "user1"))
	mockRepo.Create(ctx, createTestLink("test2", "https://example2.com", "user1"))
	mockRepo.Create(ctx, createTestLink("test3", "https://example3.com", "user2"))

	// Create a private link
	privateLink := createTestLink("private", "https://private.com", "user1")
	privateLink.AccessLevel = models.AccessLevels.Private
	mockRepo.Create(ctx, privateLink)

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
		userID         string
		expectedCount  int
	}{
		{
			name:           "Get All Links - No User ID",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
			userID:         "",
			expectedCount:  4, // All links are accessible without user ID filtering
		},
		{
			name:           "Get All Links - With User ID",
			queryParams:    map[string]string{},
			expectedStatus: http.StatusOK,
			userID:         "user1",
			expectedCount:  4, // All links are accessible to user1 (3 public + 1 private owned by user1)
		},
		{
			name: "Filter By Access Level - Public",
			queryParams: map[string]string{
				"access_level": models.AccessLevels.Public,
			},
			expectedStatus: http.StatusOK,
			userID:         "user1",
			expectedCount:  3, // 3 public links
		},
		{
			name: "Filter By Creator",
			queryParams: map[string]string{
				"created_by": "user1",
			},
			expectedStatus: http.StatusOK,
			userID:         "user1",
			expectedCount:  3, // 2 public + 1 private owned by user1
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request with query parameters
			req, _ := http.NewRequest(http.MethodGet, "/api/links", nil)
			q := req.URL.Query()
			for key, value := range tc.queryParams {
				q.Add(key, value)
			}
			req.URL.RawQuery = q.Encode()

			// Set user ID if provided
			if tc.userID != "" {
				req.Header.Set("X-User-ID", tc.userID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.GetLinks(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// Check response body
			if tc.expectedStatus == http.StatusOK {
				var links []*models.Link
				err := json.Unmarshal(rr.Body.Bytes(), &links)
				assert.NoError(t, err)
				assert.Len(t, links, tc.expectedCount)
			}
		})
	}
}

func TestGetLink(t *testing.T) {
	// Setup
	handler, mockRepo := setupTestHandler()

	// Add a test link to the mock repository
	ctx := context.Background()
	publicLink := createTestLink("public", "https://public.com", "user1")
	mockRepo.Create(ctx, publicLink)

	// Add a private link
	privateLink := createTestLink("private", "https://private.com", "user1")
	privateLink.AccessLevel = models.AccessLevels.Private
	mockRepo.Create(ctx, privateLink)

	tests := []struct {
		name           string
		shortCode      string
		expectedStatus int
		userID         string
	}{
		{
			name:           "Get Public Link - No User ID",
			shortCode:      "public",
			expectedStatus: http.StatusOK,
			userID:         "",
		},
		{
			name:           "Get Private Link - Owner",
			shortCode:      "private",
			expectedStatus: http.StatusOK,
			userID:         "user1",
		},
		{
			name:           "Get Private Link - Not Owner",
			shortCode:      "private",
			expectedStatus: http.StatusForbidden,
			userID:         "user2",
		},
		{
			name:           "Link Not Found",
			shortCode:      "nonexistent",
			expectedStatus: http.StatusNotFound,
			userID:         "user1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/api/links/"+tc.shortCode, nil)

			// Set user ID if provided
			if tc.userID != "" {
				req.Header.Set("X-User-ID", tc.userID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.GetLink(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// If successful, check response body
			if tc.expectedStatus == http.StatusOK {
				var link models.Link
				err := json.Unmarshal(rr.Body.Bytes(), &link)
				assert.NoError(t, err)
				assert.Equal(t, tc.shortCode, link.Short)
			}
		})
	}
}

func TestUpdateLink(t *testing.T) {
	// Setup
	handler, mockRepo := setupTestHandler()

	// Add a test link to the mock repository
	ctx := context.Background()
	publicLink := createTestLink("public", "https://public.com", "user1")
	mockRepo.Create(ctx, publicLink)

	tests := []struct {
		name           string
		shortCode      string
		requestBody    map[string]string
		expectedStatus int
		userID         string
	}{
		{
			name:      "Update Link URL",
			shortCode: "public",
			requestBody: map[string]string{
				"url": "https://updated.com",
			},
			expectedStatus: http.StatusOK,
			userID:         "user1",
		},
		{
			name:      "Link Not Found",
			shortCode: "nonexistent",
			requestBody: map[string]string{
				"url": "https://updated.com",
			},
			expectedStatus: http.StatusNotFound,
			userID:         "user1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			body, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest(http.MethodPut, "/api/links/"+tc.shortCode, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Set user ID if provided
			if tc.userID != "" {
				req.Header.Set("X-User-ID", tc.userID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.UpdateLink(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// If successful, check response body
			if tc.expectedStatus == http.StatusOK {
				var link models.Link
				err := json.Unmarshal(rr.Body.Bytes(), &link)
				assert.NoError(t, err)
				assert.Equal(t, tc.shortCode, link.Short)
				assert.Equal(t, tc.requestBody["url"], link.URL)
			}
		})
	}
}

func TestDeleteLink(t *testing.T) {
	// Setup
	handler, mockRepo := setupTestHandler()

	// Add a test link to the mock repository
	ctx := context.Background()
	publicLink := createTestLink("public", "https://public.com", "user1")
	mockRepo.Create(ctx, publicLink)

	tests := []struct {
		name           string
		shortCode      string
		expectedStatus int
		userID         string
	}{
		{
			name:           "Delete Link",
			shortCode:      "public",
			expectedStatus: http.StatusNoContent,
			userID:         "user1",
		},
		{
			name:           "Link Not Found",
			shortCode:      "nonexistent",
			expectedStatus: http.StatusNotFound,
			userID:         "user1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest(http.MethodDelete, "/api/links/"+tc.shortCode, nil)

			// Set user ID if provided
			if tc.userID != "" {
				req.Header.Set("X-User-ID", tc.userID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.DeleteLink(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestRedirectLink(t *testing.T) {
	// Setup
	handler, mockRepo := setupTestHandler()

	// Add test links to the mock repository
	ctx := context.Background()
	publicLink := createTestLink("public", "https://public.com", "user1")
	mockRepo.Create(ctx, publicLink)

	// Add a private link
	privateLink := createTestLink("private", "https://private.com", "user1")
	privateLink.AccessLevel = models.AccessLevels.Private
	mockRepo.Create(ctx, privateLink)

	tests := []struct {
		name           string
		shortCode      string
		expectedStatus int
		userID         string
		expectedURL    string
	}{
		{
			name:           "Redirect Public Link",
			shortCode:      "public",
			expectedStatus: http.StatusFound,
			userID:         "",
			expectedURL:    "https://public.com",
		},
		{
			name:           "Redirect Private Link - Owner",
			shortCode:      "private",
			expectedStatus: http.StatusFound,
			userID:         "user1",
			expectedURL:    "https://private.com",
		},
		{
			name:           "Redirect Private Link - Not Owner",
			shortCode:      "private",
			expectedStatus: http.StatusForbidden,
			userID:         "user2",
			expectedURL:    "",
		},
		{
			name:           "Link Not Found",
			shortCode:      "nonexistent",
			expectedStatus: http.StatusNotFound,
			userID:         "",
			expectedURL:    "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create request
			req, _ := http.NewRequest(http.MethodGet, "/"+tc.shortCode, nil)

			// Set user ID if provided
			if tc.userID != "" {
				req.Header.Set("X-User-ID", tc.userID)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.RedirectLink(rr, req)

			// Check status code
			assert.Equal(t, tc.expectedStatus, rr.Code)

			// If redirect, check the Location header
			if tc.expectedStatus == http.StatusFound {
				assert.Equal(t, tc.expectedURL, rr.Header().Get("Location"))
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	// Setup
	handler, _ := setupTestHandler()

	// Create request
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler.HealthCheck(rr, req)

	// Check status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check response body
	var response map[string]string
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}
