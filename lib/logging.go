package lib

import (
	"fmt"
	"log"
	"os"
)

// Logger provides logging functionality for BubbleGum.
type Logger struct {
	debugEnabled bool
	logger       *log.Logger
}

var defaultLogger *Logger

func init() {
	defaultLogger = &Logger{
		debugEnabled: os.Getenv("BUBBLEGUM_DEBUG") != "",
		logger:       log.New(os.Stderr, "[BubbleGum] ", log.LstdFlags),
	}
}

// Debug logs a debug message if debug mode is enabled.
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debugEnabled {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

// Info logs an informational message.
func (l *Logger) Info(format string, args ...interface{}) {
	l.logger.Printf("[INFO] "+format, args...)
}

// Error logs an error message.
func (l *Logger) Error(format string, args ...interface{}) {
	l.logger.Printf("[ERROR] "+format, args...)
}

// Warn logs a warning message.
func (l *Logger) Warn(format string, args ...interface{}) {
	l.logger.Printf("[WARN] "+format, args...)
}

// SetDebug enables or disables debug logging.
func (l *Logger) SetDebug(enabled bool) {
	l.debugEnabled = enabled
}

// Global logging functions

// Debug logs a debug message if debug mode is enabled.
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info logs an informational message.
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Error logs an error message.
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Warn logs a warning message.
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// SetDebug enables or disables debug logging globally.
func SetDebug(enabled bool) {
	defaultLogger.SetDebug(enabled)
}

// ErrorMsg is a message type that wraps an error for delivery to Update.
type ErrorMsg struct {
	Err error
}

// Error implements the error interface.
func (e ErrorMsg) Error() string {
	return fmt.Sprintf("error: %v", e.Err)
}
