package auth_test

import (
	"os"
	"testing"

	"github.com/Okabe-Junya/golink-backend/auth"
	"github.com/stretchr/testify/assert"
)

func setupAuthEnvironment(t *testing.T) {
	os.Setenv("AUTH_DISABLED", "false")
	os.Setenv("SESSION_SECRET_KEY", "test-secret-key")
	os.Setenv("GOOGLE_CLIENT_ID", "test-client-id")
	os.Setenv("GOOGLE_CLIENT_SECRET", "test-client-secret")
}

func cleanupAuthEnvironment() {
	os.Unsetenv("AUTH_DISABLED")
	os.Unsetenv("SESSION_SECRET_KEY")
	os.Unsetenv("GOOGLE_CLIENT_ID")
	os.Unsetenv("GOOGLE_CLIENT_SECRET")
}

func TestInitSessionManager(t *testing.T) {
	tests := []struct {
		envVars map[string]string
		name    string
		wantErr bool
	}{
		{
			name: "Auth Disabled",
			envVars: map[string]string{
				"AUTH_DISABLED": "true",
			},
			wantErr: false,
		},
		{
			name: "With Secret Key",
			envVars: map[string]string{
				"AUTH_DISABLED":      "false",
				"SESSION_SECRET_KEY": "test-secret-key",
			},
			wantErr: false,
		},
		{
			name: "Without Secret Key",
			envVars: map[string]string{
				"AUTH_DISABLED":      "false",
				"SESSION_SECRET_KEY": "",
			},
			wantErr: false, // Should generate random key
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

			err := auth.InitSessionManager()
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSessionTokenCreationAndValidation(t *testing.T) {
	setupAuthEnvironment(t)
	defer cleanupAuthEnvironment()

	// Initialize session manager
	err := auth.InitSessionManager()
	assert.NoError(t, err)

	// Initialize auth
	err = auth.InitAuth()
	assert.NoError(t, err)

	// Test user
	testUser := &auth.User{
		ID:     "test-user-id",
		Email:  "test@example.com",
		Name:   "Test User",
		Domain: "example.com",
	}

	// Create session token
	token, err := auth.CreateSessionToken(testUser)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate session token
	validatedUser, err := auth.ValidateSessionToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, validatedUser)
	assert.Equal(t, testUser.ID, validatedUser.ID)
	assert.Equal(t, testUser.Email, validatedUser.Email)
	assert.Equal(t, testUser.Name, validatedUser.Name)
	assert.Equal(t, testUser.Domain, validatedUser.Domain)
}

func TestInvalidSessionToken(t *testing.T) {
	setupAuthEnvironment(t)
	defer cleanupAuthEnvironment()

	// Initialize session manager
	err := auth.InitSessionManager()
	assert.NoError(t, err)

	// Initialize auth
	err = auth.InitAuth()
	assert.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "Invalid Format",
			token:       "invalid-token",
			expectError: true,
		},
		{
			name:        "Invalid Signature",
			token:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.invalid-signature",
			expectError: true,
		},
		{
			name:        "Empty Token",
			token:       "",
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user, err := auth.ValidateSessionToken(tc.token)
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}
		})
	}
}

func TestSessionTokenExpiration(t *testing.T) {
	setupAuthEnvironment(t)
	defer cleanupAuthEnvironment()

	// Initialize session manager
	err := auth.InitSessionManager()
	assert.NoError(t, err)

	// Initialize auth
	err = auth.InitAuth()
	assert.NoError(t, err)

	// Test user
	testUser := &auth.User{
		ID:     "test-user-id",
		Email:  "test@example.com",
		Name:   "Test User",
		Domain: "example.com",
	}

	// Create session token
	token, err := auth.CreateSessionToken(testUser)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token immediately (should work)
	validatedUser, err := auth.ValidateSessionToken(token)
	assert.NoError(t, err)
	assert.NotNil(t, validatedUser)

	// Note: JWT expiration can't be easily tested without mocking time
	// This is a placeholder for future implementation
	t.Skip("Session token expiration test requires time manipulation")
}
