package config

import (
	"os"
	"strconv"
	"time"

	"github.com/Okabe-Junya/golink-backend/logger"
)

// Config holds all the configuration for the application
type Config struct {
	Auth     AuthConfig
	Firebase FirebaseConfig
	CORS     CORSConfig
	Server   ServerConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port            string
	Domain          string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// FirebaseConfig holds Firebase-specific configuration
type FirebaseConfig struct {
	CredentialsJSON string
	CredentialsFile string
}

// AuthConfig holds authentication-specific configuration
type AuthConfig struct {
	JWTSecret        string
	SessionDomain    string
	SessionSameSite  string
	SessionKey       string
	SessionSignKey   string
	SessionEncrypKey string
	TokenExpiry      time.Duration
	SessionMaxAge    int
	SessionSecure    bool
	SessionHttpOnly  bool
}

// CORSConfig holds CORS-specific configuration
type CORSConfig struct {
	Origin             string
	AllowedMethods     []string
	AllowedHeaders     []string
	AllowCredentials   bool
	OptionsPassthrough bool
	MaxAge             int
}

// New creates a new Config instance with values from environment variables
func New() *Config {
	// Default values for timeouts
	const (
		defaultReadTimeout     = 5 * time.Second
		defaultWriteTimeout    = 10 * time.Second
		defaultIdleTimeout     = 120 * time.Second
		defaultShutdownTimeout = 30 * time.Second
		defaultTokenExpiry     = 24 * time.Hour
		defaultSessionMaxAge   = 86400 // 1 day in seconds
		defaultCORSMaxAge      = 300   // 5 minutes
	)

	// Get server configuration
	port := getEnv("PORT", "8080")
	domain := getEnv("APP_DOMAIN", "localhost:8080")

	// Get Firebase configuration
	credJSON := getEnv("FIREBASE_CREDENTIALS_JSON", "")
	credFile := getEnv("FIREBASE_CREDENTIALS_FILE", "")

	// Get auth configuration
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key")
	tokenExpiry := getDurationEnv("TOKEN_EXPIRY", defaultTokenExpiry)
	sessionMaxAge := getIntEnv("SESSION_MAX_AGE", defaultSessionMaxAge)
	sessionSecure := getBoolEnv("SESSION_SECURE", false)
	sessionHttpOnly := getBoolEnv("SESSION_HTTP_ONLY", true)
	sessionDomain := getEnv("SESSION_DOMAIN", domain)
	sessionSameSite := getEnv("SESSION_SAME_SITE", "Lax")
	sessionKey := getEnv("SESSION_KEY", "session")
	sessionSignKey := getEnv("SESSION_SIGN_KEY", "sign-key")
	sessionEncrypKey := getEnv("SESSION_ENCRYPT_KEY", "encr-key")

	// Get CORS configuration
	corsOrigin := getEnv("CORS_ORIGIN", "http://localhost:3001")
	corsAllowedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsAllowedHeaders := []string{"Content-Type", "Authorization"}
	corsAllowCredentials := true
	corsOptionsPassthrough := false
	corsMaxAge := getIntEnv("CORS_MAX_AGE", defaultCORSMaxAge)

	return &Config{
		Server: ServerConfig{
			Port:            port,
			Domain:          domain,
			ReadTimeout:     defaultReadTimeout,
			WriteTimeout:    defaultWriteTimeout,
			IdleTimeout:     defaultIdleTimeout,
			ShutdownTimeout: defaultShutdownTimeout,
		},
		Firebase: FirebaseConfig{
			CredentialsJSON: credJSON,
			CredentialsFile: credFile,
		},
		Auth: AuthConfig{
			JWTSecret:        jwtSecret,
			TokenExpiry:      tokenExpiry,
			SessionMaxAge:    sessionMaxAge,
			SessionSecure:    sessionSecure,
			SessionHttpOnly:  sessionHttpOnly,
			SessionDomain:    sessionDomain,
			SessionSameSite:  sessionSameSite,
			SessionKey:       sessionKey,
			SessionSignKey:   sessionSignKey,
			SessionEncrypKey: sessionEncrypKey,
		},
		CORS: CORSConfig{
			Origin:             corsOrigin,
			AllowedMethods:     corsAllowedMethods,
			AllowedHeaders:     corsAllowedHeaders,
			AllowCredentials:   corsAllowCredentials,
			OptionsPassthrough: corsOptionsPassthrough,
			MaxAge:             corsMaxAge,
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue != "" {
			logger.Warn(key+" environment variable not set", logger.Fields{
				"default_value": defaultValue,
			})
		}
		return defaultValue
	}
	return value
}

// getIntEnv gets an integer environment variable or returns a default value
func getIntEnv(key string, defaultValue int) int {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		logger.Warn("Invalid integer value for "+key, logger.Fields{
			"value":         strValue,
			"default_value": defaultValue,
			"error":         err.Error(),
		})
		return defaultValue
	}
	return value
}

// getBoolEnv gets a boolean environment variable or returns a default value
func getBoolEnv(key string, defaultValue bool) bool {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(strValue)
	if err != nil {
		logger.Warn("Invalid boolean value for "+key, logger.Fields{
			"value":         strValue,
			"default_value": defaultValue,
			"error":         err.Error(),
		})
		return defaultValue
	}
	return value
}

// getDurationEnv gets a duration environment variable or returns a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(strValue)
	if err != nil {
		logger.Warn("Invalid duration value for "+key, logger.Fields{
			"value":         strValue,
			"default_value": defaultValue,
			"error":         err.Error(),
		})
		return defaultValue
	}
	return value
}
