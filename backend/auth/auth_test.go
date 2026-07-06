package auth_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Okabe-Junya/golink-backend/auth"
	"github.com/stretchr/testify/assert"
)

func TestInitAuth(t *testing.T) {
	// Clear all relevant environment variables before testing
	os.Unsetenv("AUTH_DISABLED")
	os.Unsetenv("GOOGLE_CLIENT_ID")
	os.Unsetenv("GOOGLE_CLIENT_SECRET")
	os.Unsetenv("GOOGLE_ALLOWED_DOMAIN")

	tests := []struct {
		envVars     map[string]string
		name        string
		wantEnabled bool
		wantErr     bool
	}{
		{
			name: "Auth Disabled",
			envVars: map[string]string{
				"AUTH_DISABLED": "true",
			},
			wantEnabled: false,
			wantErr:     false,
		},
		{
			name: "Missing Credentials",
			envVars: map[string]string{
				"AUTH_DISABLED":        "false",
				"GOOGLE_CLIENT_ID":     "",
				"GOOGLE_CLIENT_SECRET": "",
			},
			wantEnabled: false,
			wantErr:     false,
		},
		{
			name: "Valid Configuration",
			envVars: map[string]string{
				"AUTH_DISABLED":         "false",
				"GOOGLE_CLIENT_ID":      "test-client-id",
				"GOOGLE_CLIENT_SECRET":  "test-client-secret",
				"GOOGLE_ALLOWED_DOMAIN": "example.com",
			},
			wantEnabled: true,
			wantErr:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			for k, v := range tc.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tc.envVars {
					os.Unsetenv(k)
				}
			}()

			err := auth.InitAuth()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.wantEnabled, auth.IsAuthEnabled())
		})
	}
}

func TestAuthMiddleware(t *testing.T) {
	// Create a mock handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		path           string
		authEnabled    bool
		withCookie     bool
		expectedStatus int
	}{
		{
			name:           "Auth Disabled",
			path:           "/api/links",
			authEnabled:    false,
			withCookie:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Auth Path Bypass",
			path:           "/api/auth/login",
			authEnabled:    true,
			withCookie:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Health Check Bypass",
			path:           "/health",
			authEnabled:    true,
			withCookie:     false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API Path Without Auth",
			path:           "/api/links",
			authEnabled:    true,
			withCookie:     false,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			os.Unsetenv("AUTH_DISABLED")
			os.Unsetenv("GOOGLE_CLIENT_ID")
			os.Unsetenv("GOOGLE_CLIENT_SECRET")

			if !tc.authEnabled {
				os.Setenv("AUTH_DISABLED", "true")
			} else {
				os.Setenv("AUTH_DISABLED", "false")
				os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
				os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
			}
			auth.InitAuth()

			// Create request
			req := httptest.NewRequest("GET", tc.path, nil)
			if tc.withCookie {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: "test-token",
				})
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			// Create middleware
			handler := auth.AuthMiddleware(nextHandler)

			// Handle request
			handler.ServeHTTP(rr, req)

			// Check response
			assert.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	tests := []struct {
		name           string
		expectedUserID string
		authEnabled    bool
		withCookie     bool
		expectedError  bool
	}{
		{
			name:           "Auth Disabled",
			authEnabled:    false,
			withCookie:     false,
			expectedError:  false,
			expectedUserID: "anonymous",
		},
		{
			name:           "No Cookie",
			authEnabled:    true,
			withCookie:     false,
			expectedError:  true,
			expectedUserID: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Setup environment
			if !tc.authEnabled {
				os.Setenv("AUTH_DISABLED", "true")
			} else {
				os.Setenv("AUTH_DISABLED", "false")
				os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
				os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
			}
			auth.InitAuth()

			// Create request
			req := httptest.NewRequest("GET", "/api/links", nil)
			if tc.withCookie {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: "test-token",
				})
			}

			// Get current user
			user, err := auth.GetCurrentUser(req)
			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUserID, user.ID)
			}
		})
	}
}

// TestGetCurrentUser_HeaderNotTrustedInProduction pins the fix for the auth-bypass
// where GetCurrentUser trusted an attacker-controlled X-User-ID header. With auth
// enabled and no session cookie, the header must NOT authenticate a user unless
// TEST_MODE is explicitly set.
func TestGetCurrentUser_HeaderNotTrustedInProduction(t *testing.T) {
	t.Setenv("AUTH_DISABLED", "false")
	t.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
	if err := auth.InitAuth(); err != nil {
		t.Fatalf("InitAuth failed: %v", err)
	}

	newReq := func() *http.Request {
		req := httptest.NewRequest("GET", "/api/links", nil)
		req.Header.Set("X-User-ID", "victim@example.com")
		return req
	}

	t.Run("production rejects the header", func(t *testing.T) {
		t.Setenv("TEST_MODE", "")
		user, err := auth.GetCurrentUser(newReq())
		assert.Error(t, err)
		assert.Nil(t, user)
	})

	t.Run("test mode still honors the header", func(t *testing.T) {
		t.Setenv("TEST_MODE", "true")
		user, err := auth.GetCurrentUser(newReq())
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "victim@example.com", user.ID)
	})
}

func TestHandleLogout(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectsClear   bool
	}{
		{
			name:           "POST clears cookie",
			method:         http.MethodPost,
			expectedStatus: http.StatusNoContent,
			expectsClear:   true,
		},
		{
			name:           "GET is rejected",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			expectsClear:   false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/api/auth/logout", nil)
			req.AddCookie(&http.Cookie{Name: "session_token", Value: "some-token"})
			rr := httptest.NewRecorder()

			auth.HandleLogout(rr, req)

			assert.Equal(t, tc.expectedStatus, rr.Code)

			if !tc.expectsClear {
				return
			}

			cookies := rr.Result().Cookies()
			var sessionCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "session_token" {
					sessionCookie = c
					break
				}
			}
			assert.NotNil(t, sessionCookie, "session_token cookie should be set in response")
			assert.Empty(t, sessionCookie.Value)
			assert.Less(t, sessionCookie.MaxAge, 0, "MaxAge should be negative to delete cookie")
		})
	}
}
