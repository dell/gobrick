/*
Copyright © 2020 - 2022 Dell Inc. or its subsidiaries. All Rights Reserved.

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
