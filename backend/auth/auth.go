package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Okabe-Junya/golink-backend/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	// OAuth config
	oauthConfig *oauth2.Config
	// Domain constraint for Google Workspace
	allowedDomain string
	// Is authentication enabled
	authEnabled = true
)

// generateStateToken creates a random state token
func generateStateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// stateCookieName is the name of the cookie that stores the OAuth state
const stateCookieName = "oauth_state"

// User represents an authenticated user
type User struct {
	Email   string `json:"email"`
	ID      string `json:"id"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
	Domain  string `json:"-"` // Domain extracted from email

	VerifiedEmail bool `json:"verified_email"`
}

// InitAuth initializes the authentication system
func InitAuth() error {
	// Enable authentication by default
	authEnabled = true

	// Check if authentication is explicitly disabled
	authDisabled := os.Getenv("AUTH_DISABLED")
	if strings.ToLower(authDisabled) == "true" {
		authEnabled = false
		logger.Info("Authentication is disabled", nil)
		return nil
	}

	// Get client ID and secret from environment variables
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		authEnabled = false
		logger.Warn("Missing GOOGLE_CLIENT_ID or GOOGLE_CLIENT_SECRET environment variable, authentication will be disabled", nil)
		return nil
	}

	// Get allowed domain from environment variable
	allowedDomain = os.Getenv("GOOGLE_ALLOWED_DOMAIN")
	if allowedDomain == "" {
		logger.Warn("No GOOGLE_ALLOWED_DOMAIN set, all Google accounts will be allowed", nil)
	}

	// Get redirect URL from environment variable or use default
	redirectURL := os.Getenv("OAUTH_REDIRECT_URL")
	if redirectURL == "" {
		appDomain := os.Getenv("APP_DOMAIN")
		if appDomain == "" {
			appDomain = "localhost:8080"
		}
		redirectURL = fmt.Sprintf("http://%s/api/auth/callback", appDomain)
	}

	// Initialize OAuth config
	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	logger.Info("Authentication system initialized successfully", logger.Fields{
		"allowedDomain": allowedDomain,
		"redirectURL":   redirectURL,
	})

	return nil
}

// IsAuthEnabled returns whether authentication is enabled
func IsAuthEnabled() bool {
	return authEnabled
}

// GetLoginURL returns the URL to redirect users to for login
func GetLoginURL() (string, string, error) {
	if !authEnabled || oauthConfig == nil {
		return "", "", errors.New("authentication is not enabled")
	}

	state, err := generateStateToken()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate state token: %w", err)
	}

	return oauthConfig.AuthCodeURL(state), state, nil
}

// HandleLogin redirects the user to Google's OAuth login page
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if !authEnabled {
		http.Error(w, "Authentication is disabled", http.StatusNotImplemented)
		return
	}

	url, state, err := GetLoginURL()
	if err != nil {
		http.Error(w, "Failed to generate login URL", http.StatusInternalServerError)
		logger.Error("Failed to generate login URL", err, nil)
		return
	}

	// Set the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(5 * time.Minute.Seconds()), // State cookie expires in 5 minutes
	})

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback processes the OAuth callback from Google
func HandleCallback(w http.ResponseWriter, r *http.Request) {
	if !authEnabled {
		http.Error(w, "Authentication is disabled", http.StatusNotImplemented)
		return
	}

	// Get state from cookie
	stateCookie, err := r.Cookie(stateCookieName)
	if err != nil {
		http.Error(w, "State cookie not found", http.StatusBadRequest)
		logger.Error("State cookie not found", err, nil)
		return
	}

	// Clear the state cookie immediately
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		MaxAge:   -1,
	})

	// Verify state parameter
	state := r.FormValue("state")
	if state == "" || state != stateCookie.Value {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		logger.Error("Invalid OAuth state", nil, logger.Fields{
			"expected": stateCookie.Value,
			"received": state,
		})
		return
	}

	// Exchange authorization code for token
	code := r.FormValue("code")
	token, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		logger.Error("Failed to exchange token", err, nil)
		return
	}

	// Get user info
	user, err := getUserInfo(r.Context(), token)
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		logger.Error("Failed to get user info", err, nil)
		return
	}

	// Check if user's email domain is allowed
	if allowedDomain != "" && user.Domain != allowedDomain {
		http.Error(w, "Unauthorized domain", http.StatusUnauthorized)
		logger.Warn("Login attempt from unauthorized domain", logger.Fields{
			"email":         user.Email,
			"domain":        user.Domain,
			"allowedDomain": allowedDomain,
		})
		return
	}

	// Create a session token
	sessionToken, err := CreateSessionToken(user)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		logger.Error("Failed to create session token", err, nil)
		return
	}

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Hour * 24 * 7 / time.Second), // 7 days
	})

	// Redirect to frontend
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "/"
	}
	http.Redirect(w, r, frontendURL, http.StatusTemporaryRedirect)
}

// getUserInfo gets the user information from Google API
func getUserInfo(ctx context.Context, token *oauth2.Token) (*User, error) {
	if !authEnabled || oauthConfig == nil {
		return nil, errors.New("authentication is not enabled")
	}

	client := oauthConfig.Client(ctx, token)

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	// Extract domain from email
	parts := strings.Split(user.Email, "@")
	if len(parts) == 2 {
		user.Domain = parts[1]
	}

	return &user, nil
}

// GetCurrentUser gets the current user from the request
func GetCurrentUser(r *http.Request) (*User, error) {
	if !authEnabled {
		return &User{
			ID:    "anonymous",
			Email: "anonymous@example.com",
			Name:  "Anonymous User",
		}, nil
	}

	// Get the session token from the cookie
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, err
	}

	// Validate the session token
	user, err := ValidateSessionToken(cookie.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// AuthMiddleware is a middleware that checks if the user is authenticated
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth check if authentication is disabled
		if !authEnabled {
			// 認証が無効の場合は匿名ユーザーをコンテキストに追加
			anonymousUser := &User{
				ID:    "anonymous",
				Email: "anonymous@example.com",
				Name:  "Anonymous User",
			}
			ctx := context.WithValue(r.Context(), "user", anonymousUser)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Skip auth for auth-related paths
		if strings.HasPrefix(r.URL.Path, "/api/auth/") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for redirect paths
		if r.URL.Path == "/" || r.URL.Path == "/favicon.ico" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Get the user from the request
		user, err := GetCurrentUser(r)
		if err != nil {
			// Return 401 for API requests
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Redirect to login for other requests
			http.Redirect(w, r, "/api/auth/login", http.StatusTemporaryRedirect)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
