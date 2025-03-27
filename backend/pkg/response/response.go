package response

import (
	"encoding/json"
	"net/http"

	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/Okabe-Junya/golink-backend/pkg/errors"
)

// Response represents a standard API response
type Response struct {
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Success bool        `json:"success"`
}

// ErrorInfo represents error information in a response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON sends a JSON response with the given status code and data
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	response := Response{
		Success: statusCode >= 200 && statusCode < 300,
		Data:    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode JSON response", err, nil)
	}
}

// Error sends an error response with the given error
func Error(w http.ResponseWriter, err error) {
	var statusCode int
	var code string
	var message string

	if appErr, ok := err.(*errors.Error); ok {
		statusCode = appErr.Code
		code = "ERROR"
		message = appErr.Message
	} else {
		statusCode = http.StatusInternalServerError
		code = "INTERNAL_SERVER_ERROR"
		message = "An internal server error occurred"
	}

	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}

	logger.Error("API error response", err, logger.Fields{
		"status_code": statusCode,
		"error_code":  code,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Failed to encode error response", err, nil)
	}
}
