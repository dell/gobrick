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

func SetLogger(customLogger Logger) {
	logger = customLogger
}

// Logger logging interface for gobrick
type Logger interface {
	Info(ctx context.Context, format string, args ...interface{})
	Debug(ctx context.Context, format string, args ...interface{})
	Error(ctx context.Context, format string, args ...interface{})
}

type DummyLogger struct{}

func (dl *DummyLogger) Info(ctx context.Context, format string, args ...interface{}) {
	log.Print("INFO: " + fmt.Sprintf(format, args...))
}
func (dl *DummyLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	log.Print("DEBUG: " + fmt.Sprintf(format, args...))
}
func (dl *DummyLogger) Error(ctx context.Context, format string, args ...interface{}) {
	log.Print("ERROR: " + fmt.Sprintf(format, args...))
}

func Info(ctx context.Context, format string, args ...interface{}) {
	logger.Info(ctx, format, args...)
}

func Debug(ctx context.Context, format string, args ...interface{}) {
	logger.Debug(ctx, format, args...)
}

func Error(ctx context.Context, format string, args ...interface{}) {
	logger.Error(ctx, format, args...)
}
