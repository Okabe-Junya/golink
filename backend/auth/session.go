package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Okabe-Junya/golink-backend/logger"
)

var (
	// Secret key for signing tokens
	secretKey []byte
)

// SessionClaims represents the data stored in a session token
type SessionClaims struct {
	ExpiresAt time.Time `json:"expires_at"`

	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// InitSessionManager initializes the session management system
func InitSessionManager() error {
	// Skip if auth is disabled
	authDisabled := os.Getenv("AUTH_DISABLED")
	if strings.ToLower(authDisabled) == "true" {
		logger.Info("Session management is disabled as authentication is disabled", nil)
		return nil
	}

	// Get or generate secret key
	keyString := os.Getenv("SESSION_SECRET_KEY")
	if keyString == "" {
		// Generate a random secret key if not provided
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("failed to generate secret key: %w", err)
		}
		secretKey = key
		logger.Warn("Generated random SESSION_SECRET_KEY; sessions will be invalidated on restart", nil)
	} else {
		// Use the provided secret key
		secretKey = []byte(keyString)
	}

	return nil
}

// CreateSessionToken creates a new session token for a user
func CreateSessionToken(user *User) (string, error) {
	// Check if auth is disabled
	if !IsAuthEnabled() {
		return "", errors.New("cannot create session token when authentication is disabled")
	}

	// Check if session management is initialized
	if len(secretKey) == 0 {
		return "", errors.New("session manager not initialized")
	}

	// Create claims
	claims := SessionClaims{
		UserID:    user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Domain:    user.Domain,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7), // 7 days
	}

	// Serialize claims
	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Base64 encode claims
	encodedClaims := base64.URLEncoding.EncodeToString(claimsJSON)

	// Create signature
	signature, err := createSignature(encodedClaims)
	if err != nil {
		return "", err
	}

	// Create token
	token := fmt.Sprintf("%s.%s", encodedClaims, signature)
	return token, nil
}

// ValidateSessionToken validates a session token and returns the user
func ValidateSessionToken(token string) (*User, error) {
	// Check if auth is disabled
	if !IsAuthEnabled() {
		return nil, errors.New("authentication is disabled")
	}

	// Check if session manager is initialized
	if len(secretKey) == 0 {
		return nil, errors.New("session manager not initialized")
	}

	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errors.New("invalid token format")
	}

	encodedClaims, signature := parts[0], parts[1]

	// Verify signature
	expectedSignature, err := createSignature(encodedClaims)
	if err != nil {
		return nil, err
	}
	if signature != expectedSignature {
		return nil, errors.New("invalid token signature")
	}

	// Decode claims
	claimsJSON, err := base64.URLEncoding.DecodeString(encodedClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to decode claims: %w", err)
	}

	// Unmarshal claims
	var claims SessionClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Check if token is expired
	if time.Now().After(claims.ExpiresAt) {
		return nil, errors.New("token expired")
	}

	// Create user
	user := &User{
		ID:     claims.UserID,
		Email:  claims.Email,
		Name:   claims.Name,
		Domain: claims.Domain,
	}

	return user, nil
}

// createSignature creates a signature for the given data
func createSignature(data string) (string, error) {
	if len(secretKey) == 0 {
		return "", errors.New("session manager not initialized")
	}

	// Create HMAC-SHA256 signature
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

// IsSessionEnabled returns whether session management is enabled
func IsSessionEnabled() bool {
	return IsAuthEnabled() && len(secretKey) > 0
}
