package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/Okabe-Junya/golink-backend/logger"
)

// APIError represents a standardized API error response
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

// Common API error codes
const (
	ErrBadRequest          = "BAD_REQUEST"
	ErrUnauthorized        = "UNAUTHORIZED"
	ErrForbidden           = "FORBIDDEN"
	ErrNotFound            = "NOT_FOUND"
	ErrConflict            = "CONFLICT"
	ErrInternalServerError = "INTERNAL_SERVER_ERROR"
)

// ErrorHandler wraps an HTTP handler with standardized error handling
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a custom response writer to capture the status code
		crw := &customResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next.ServeHTTP(crw, r)

		// Only handle errors (status >= 400)
		if crw.statusCode >= 400 {
			// Don't override if the response has already been written
			if crw.written {
				return
			}

			// Get the appropriate error message based on status code
			var apiErr APIError
			switch crw.statusCode {
			case http.StatusBadRequest:
				apiErr = APIError{Status: crw.statusCode, Code: ErrBadRequest, Message: "Bad request"}
			case http.StatusUnauthorized:
				apiErr = APIError{Status: crw.statusCode, Code: ErrUnauthorized, Message: "Unauthorized"}
			case http.StatusForbidden:
				apiErr = APIError{Status: crw.statusCode, Code: ErrForbidden, Message: "Forbidden"}
			case http.StatusNotFound:
				apiErr = APIError{Status: crw.statusCode, Code: ErrNotFound, Message: "Not found"}
			case http.StatusConflict:
				apiErr = APIError{Status: crw.statusCode, Code: ErrConflict, Message: "Resource conflict"}
			default:
				apiErr = APIError{Status: crw.statusCode, Code: ErrInternalServerError, Message: "Internal server error"}
			}

			// Set content type
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(apiErr.Status)

			// Write the standardized error response
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"error": apiErr,
			}); err != nil {
				logger.Error("Failed to encode error response", err, nil)
			}
		}
	})
}

// customResponseWriter is a wrapper around http.ResponseWriter that captures the status code
type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader captures the status code and then calls the underlying ResponseWriter.WriteHeader
func (crw *customResponseWriter) WriteHeader(statusCode int) {
	crw.statusCode = statusCode
	crw.written = true
	crw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures that the response has been written and then calls the underlying ResponseWriter.Write
func (crw *customResponseWriter) Write(b []byte) (int, error) {
	crw.written = true
	return crw.ResponseWriter.Write(b)
}

// RespondWithError is a helper function to respond with a standardized error
func RespondWithError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	apiErr := APIError{
		Status:  status,
		Code:    code,
		Message: message,
	}

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error": apiErr,
	}); err != nil {
		logger.Error("Failed to encode error response", err, nil)
	}
}
