package tracer

import (
	"context"
	"fmt"
)

var tracer Tracer

// Tracer  tracing interface for gobrick
type Tracer interface {
	Trace(ctx context.Context, format string, args ...interface{})
}

// SetTracer initializes tracer with custom tracer
func SetTracer(customTracer Tracer) {
	tracer = customTracer
}

func init() {
	tracer = &DummyTracer{}
}

// TraceFuncCall performs tracing
func TraceFuncCall(ctx context.Context, funcName string) func() {
	tracer.Trace(ctx, "START: %s", funcName)
	return func() {
		Trace(ctx, "END: %s", funcName)
	}
}

// DummyTracer is used for testing purpose
type DummyTracer struct{}

// Trace contains dummy implementation of trace function
func (dl *DummyTracer) Trace(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

// Trace contains implementation of trace function
func Trace(ctx context.Context, format string, args ...interface{}) {
	tracer.Trace(ctx, format, args...)
}
