package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetLink tests the GET /api/links/{short} endpoint
func TestGetLink(t *testing.T) {
	tests := []struct {
		name           string
		shortCode      string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Can retrieve public link",
			shortCode:      "test-public",
			authHeader:     "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Cannot retrieve private link without authentication",
			shortCode:      "test-private",
			authHeader:     "",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Owner can retrieve private link",
			shortCode:      "test-private",
			authHeader:     fmt.Sprintf("Bearer %s", testUserID),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "404 for non-existent link",
			shortCode:      "non-existent",
			authHeader:     fmt.Sprintf("Bearer %s", testUserID),
			expectedStatus: http.StatusNotFound,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			url := fmt.Sprintf("%s/api/links/%s", apiBaseURL, tt.shortCode)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Send the request
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Check the status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Status code should match expected")
		})
	}
}

// TestCreateLink tests the POST /api/links endpoint
func TestCreateLink(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "Can create a new link",
			requestBody:    `{"short":"test-create-1","url":"https://example.com/create1","access_level":"Public"}`,
			authHeader:     fmt.Sprintf("Bearer %s", testUserID),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Can create link as anonymous user",
			requestBody:    `{"short":"test-create-2","url":"https://example.com/create2","access_level":"Public"}`,
			authHeader:     "",
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Cannot create link with duplicate short code",
			requestBody:    `{"short":"test-public","url":"https://example.com/duplicate","access_level":"Public"}`,
			authHeader:     fmt.Sprintf("Bearer %s", testUserID),
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "400 for invalid request (missing short)",
			requestBody:    `{"url":"https://example.com/invalid","access_level":"Public"}`,
			authHeader:     fmt.Sprintf("Bearer %s", testUserID),
			expectedStatus: http.StatusBadRequest,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			url := fmt.Sprintf("%s/api/links", apiBaseURL)
			req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(tt.requestBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Send the request
			resp, err := testClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Check the status code
			assert.Equal(t, tt.expectedStatus, resp.StatusCode, "Status code should match expected")
		})
	}
}

// TestDeleteLink tests the DELETE /api/links/{short} endpoint
func TestDeleteLink(t *testing.T) {
	// First, create a link that we can delete
	shortToDelete := "test-delete"
	createBody := fmt.Sprintf(`{"short":"%s","url":"https://example.com/delete","access_level":"Public"}`, shortToDelete)

	createURL := fmt.Sprintf("%s/api/links", apiBaseURL)
	createReq, err := http.NewRequest(http.MethodPost, createURL, strings.NewReader(createBody))
	require.NoError(t, err)

	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testUserID))

	createResp, err := testClient.Do(createReq)
	require.NoError(t, err)
	createResp.Body.Close()

	// Test delete functionality
	t.Run("Owner can delete link", func(t *testing.T) {
		// Create a test request
		url := fmt.Sprintf("%s/api/links/%s", apiBaseURL, shortToDelete)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testUserID))

		// Send the request
		resp, err := testClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Check the status code should be 204 No Content
		assert.Equal(t, http.StatusNoContent, resp.StatusCode, "Status code should be 204 No Content")
	})
}
