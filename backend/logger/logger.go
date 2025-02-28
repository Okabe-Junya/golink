package logger

import (
	"flag"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

// For testing purposes
var output io.Writer = os.Stdout

// Fields type for structured logging
type Fields map[string]interface{}

func init() {
	log = logrus.New()
	log.SetOutput(output)
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Set log level based on environment
	level := os.Getenv("LOG_LEVEL")
	if flag.Lookup("test.v") != nil {
		level = "debug"
	} else if level == "" {
		level = "info"
	}

	switch level {
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	default:
		log.SetLevel(logrus.InfoLevel)
	}
	if flag.Lookup("test.v") != nil {
		log.ExitFunc = func(code int) {}
	}
}

// SetLevel sets the logging level
func SetLevel(level logrus.Level) {
	log.SetLevel(level)
}

// GetLevel returns the current logging level
func GetLevel() logrus.Level {
	return log.GetLevel()
}

// Info logs info level messages with structured fields
func Info(msg string, fields Fields) {
	if fields == nil {
		log.Info(msg)
	} else {
		log.WithFields(logrus.Fields(fields)).Info(msg)
	}
}

// Debug logs debug level messages with structured fields
func Debug(msg string, fields Fields) {
	// Force debug level for ensuring debug messages are output during testing.
	log.SetLevel(logrus.DebugLevel)
	if fields == nil {
		log.Debug(msg)
	} else {
		log.WithFields(logrus.Fields(fields)).Debug(msg)
	}
}

// Warn logs warning level messages with structured fields
func Warn(msg string, fields Fields) {
	if fields == nil {
		log.Warn(msg)
	} else {
		log.WithFields(logrus.Fields(fields)).Warn(msg)
	}
}

// Error logs error level messages with structured fields
func Error(msg string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	log.WithFields(logrus.Fields(fields)).Error(msg)
}

// Fatal logs fatal level messages with structured fields and exits
func Fatal(msg string, err error, fields Fields) {
	if fields == nil {
		fields = Fields{}
	}
	if err != nil {
		fields["error"] = err.Error()
	}
	log.WithFields(logrus.Fields(fields)).Fatal(msg)
}

// SetOutput changes the output writer - useful for testing
// Returns the previous output writer
func SetOutput(w io.Writer) io.Writer {
	oldOutput := output
	if w == nil {
		w = os.Stdout
	}
	output = w
	log.SetOutput(w)
	return oldOutput
}
