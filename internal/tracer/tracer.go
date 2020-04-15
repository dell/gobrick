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

func SetTracer(customTracer Tracer) {
	tracer = customTracer
}

func init() {
	tracer = &DummyTracer{}
}

func TraceFuncCall(ctx context.Context, funcName string) func() {
	tracer.Trace(ctx, "START: %s", funcName)
	return func() {
		Trace(ctx, "END: %s", funcName)
	}
}

type DummyTracer struct{}

func (dl *DummyTracer) Trace(ctx context.Context, format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func Trace(ctx context.Context, format string, args ...interface{}) {
	tracer.Trace(ctx, format, args...)
}
