package logger

import (
	"context"
	"fmt"
	"log"
)

var logger Logger

func init() {
	logger = &DummyLogger{}
}

// SetLogger initializes custom logger for gobrick
func SetLogger(customLogger Logger) {
	logger = customLogger
}

// Logger logging interface for gobrick
type Logger interface {
	Info(ctx context.Context, format string, args ...interface{})
	Debug(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, format string, args ...interface{})
}

// DummyLogger for testing purposes
type DummyLogger struct{}

// Info is a dummy implementation of logger Info method
func (dl *DummyLogger) Info(ctx context.Context, format string, args ...interface{}) {
	log.Print("INFO: " + fmt.Sprintf(format, args...))
}

// Debug is a dummy implementation of logger Debug method
func (dl *DummyLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	log.Print("DEBUG: " + fmt.Sprintf(format, args...))
}

// Error is a dummy implementation of logger Error method
func (dl *DummyLogger) Error(ctx context.Context, format string, args ...interface{}) {
	log.Print("ERROR: " + fmt.Sprintf(format, args...))
}

// Info is a wrapper of logger Info method
func Info(ctx context.Context, format string, args ...interface{}) {
	logger.Info(ctx, format, args...)
}

// Debug is a wrapper of logger Debug method
func Debug(ctx context.Context, format string, args ...interface{}) {
	logger.Debug(ctx, format, args...)
}

// Error is a wrapper of logger Error method
func Error(ctx context.Context, format string, args ...interface{}) {
	logger.Error(ctx, format, args...)
}
