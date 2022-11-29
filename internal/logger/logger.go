/*
Copyright © 2020-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package logger is a generated GoMock package.
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
