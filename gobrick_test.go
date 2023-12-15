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
package gobrick

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/dell/gobrick/internal/logger"
	"github.com/dell/gobrick/internal/mockhelper"
	"github.com/dell/gobrick/internal/tracer"
	"github.com/dell/gobrick/pkg/scsi"
	"github.com/stretchr/testify/assert"
)

// common vars
var (
	validLunNumber     = 199
	validSCSIHost1     = "34"
	validSCSIHost2     = "35"
	validHostOnlyHCTL1 = scsi.HCTL{Host: validSCSIHost1, Channel: "-", Target: "-", Lun: "-"}
	validHostOnlyHCTL2 = scsi.HCTL{Host: validSCSIHost2, Channel: "-", Target: "-", Lun: "-"}
	validHCTL1         = scsi.HCTL{
		Host:    validSCSIHost1,
		Channel: "0",
		Target:  "0",
		Lun:     strconv.FormatInt(int64(validLunNumber), 10),
	}
	validHCTL1Target1 = scsi.HCTL{
		Host:    validSCSIHost1,
		Channel: "0",
		Target:  "1",
		Lun:     strconv.FormatInt(int64(validLunNumber), 10),
	}
	validHCTL2 = scsi.HCTL{
		Host:    validSCSIHost2,
		Channel: "0",
		Target:  "0",
		Lun:     strconv.FormatInt(int64(validLunNumber), 10),
	}
	validDevice = Device{
		WWN:  mockhelper.ValidWWID,
		Name: mockhelper.ValidDeviceName,
	}
	validDeviceMultipath = Device{
		Name:        mockhelper.ValidDMName,
		WWN:         mockhelper.ValidWWID,
		MultipathID: mockhelper.ValidWWID,
	}
)

const (
	msg1 = "foo:%s"
	arg1 = "bar"
	msg2 = "spam"
)

var (
	logData   []string
	traceData []string
)

type testLogger struct{}

func addToLogData(format string, args ...interface{}) {
	logData = append(logData, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Info(_ context.Context, format string, args ...interface{}) {
	addToLogData(format, args...)
}

func (tl *testLogger) Debug(_ context.Context, format string, args ...interface{}) {
	addToLogData(format, args...)
}

func (tl *testLogger) Error(_ context.Context, format string, args ...interface{}) {
	addToLogData(format, args...)
}

func TestSetLogger(t *testing.T) {
	SetLogger(&testLogger{})
	defer SetLogger(&logger.DummyLogger{})
	ctx := context.Background()
	logger.Info(ctx, msg1, arg1)
	logger.Error(ctx, msg2)
	assert.Contains(t, logData, fmt.Sprintf(msg1, arg1))
	assert.Contains(t, logData, msg2)
}

func TestSetTracer(t *testing.T) {
	SetTracer(&testTracer{})
	defer SetTracer(&tracer.DummyTracer{})
	ctx := context.Background()
	tracer.Trace(ctx, msg1, arg1)
	tracer.Trace(ctx, msg2)
	assert.Contains(t, traceData, fmt.Sprintf(msg1, arg1))
	assert.Contains(t, traceData, msg2)
}

type testTracer struct{}

func (tt *testTracer) Trace(_ context.Context, format string, args ...interface{}) {
	traceData = append(traceData, fmt.Sprintf(format, args...))
}
