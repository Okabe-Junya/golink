package logger_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/Okabe-Junya/golink-backend/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// captureOutput captures logger output during test execution
func captureOutput(f func()) string {
	// Create a buffer to store output
	var buf bytes.Buffer

	// Redirect logger output to buffer
	logger.SetOutput(&buf)

	// Execute the function that produces output
	f()

	// Capture the output
	output := buf.String()

	// Restore original output (to stdout)
	logger.SetOutput(nil)

	// Return captured output
	return output
}

func TestInfo(t *testing.T) {
	output := captureOutput(func() {
		logger.Info("test info message", logger.Fields{"key": "value"})
	})

	// Check if output contains expected data
	assert.Contains(t, output, "test info message")
	assert.Contains(t, output, "info")
	assert.Contains(t, output, "key")
	assert.Contains(t, output, "value")

	// Parse JSON to verify structure
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "test info message", logEntry["msg"])
	assert.Equal(t, "info", logEntry["level"])
	assert.Equal(t, "value", logEntry["key"])
}

func TestDebug(t *testing.T) {
	// Set log level to debug for this test
	originalLevel := logrus.GetLevel()
	logrus.SetLevel(logrus.DebugLevel)
	defer logrus.SetLevel(originalLevel)

	output := captureOutput(func() {
		logger.Debug("test debug message", logger.Fields{"key": "value"})
	})

	// Check if output contains expected data
	assert.Contains(t, output, "test debug message")
	assert.Contains(t, output, "debug")

	// Parse JSON to verify structure
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "test debug message", logEntry["msg"])
	assert.Equal(t, "debug", logEntry["level"])
	assert.Equal(t, "value", logEntry["key"])
}

func TestWarn(t *testing.T) {
	output := captureOutput(func() {
		logger.Warn("test warning message", logger.Fields{"key": "value"})
	})

	// Check if output contains expected data
	assert.Contains(t, output, "test warning message")
	assert.Contains(t, output, "warning")

	// Parse JSON to verify structure
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "test warning message", logEntry["msg"])
	assert.Equal(t, "warning", logEntry["level"])
	assert.Equal(t, "value", logEntry["key"])
}

func TestError(t *testing.T) {
	testErr := errors.New("test error")
	output := captureOutput(func() {
		logger.Error("test error message", testErr, logger.Fields{"key": "value"})
	})

	// Check if output contains expected data
	assert.Contains(t, output, "test error message")
	assert.Contains(t, output, "error")
	assert.Contains(t, output, "test error")

	// Parse JSON to verify structure
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "test error message", logEntry["msg"])
	assert.Equal(t, "error", logEntry["level"])
	assert.Equal(t, "value", logEntry["key"])
	assert.Equal(t, "test error", logEntry["error"])
}

func TestNilFields(t *testing.T) {
	// Test with nil fields
	output := captureOutput(func() {
		logger.Info("test nil fields", nil)
	})

	// Check if output contains expected data
	assert.Contains(t, output, "test nil fields")

	// Parse JSON to verify structure
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "test nil fields", logEntry["msg"])
	assert.Equal(t, "info", logEntry["level"])
}
