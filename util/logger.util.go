package util

import (
	"log/slog"
	"os"
)

/*
	Singleton logger instance for the application.
	Provides structured logging with JSON output.
	Supports different log levels: Info, Warn, Error, Debug.

	Initialize logger with sensible defaults
	Uses JSON output for structured logging
*/

var logger *slog.Logger

func init() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger = slog.New(handler)
}

// Get returns the shared application logger
func Get() *slog.Logger {
	return logger
}

// Info logs an info-level message with optional attributes
func Info(msg string, attrs ...any) {
	logger.Info(msg, attrs...)
}

// Warn logs a warning-level message with optional attributes
func Warn(msg string, attrs ...any) {
	logger.Warn(msg, attrs...)
}

// Error logs an error-level message with optional attributes
func Error(msg string, attrs ...any) {
	logger.Error(msg, attrs...)
}

// Debug logs a debug-level message with optional attributes
func Debug(msg string, attrs ...any) {
	logger.Debug(msg, attrs...)
}
