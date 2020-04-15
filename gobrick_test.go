package gobrick

import (
	"context"
	"github.com/dell/gobrick/internal/logger"
	"github.com/dell/gobrick/internal/mockhelper"
	"github.com/dell/gobrick/internal/tracer"
	"github.com/dell/gobrick/pkg/scsi"
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
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
		Lun:     strconv.FormatInt(int64(validLunNumber), 10)}
	validHCTL1Target1 = scsi.HCTL{
		Host:    validSCSIHost1,
		Channel: "0",
		Target:  "1",
		Lun:     strconv.FormatInt(int64(validLunNumber), 10)}
	validHCTL2 = scsi.HCTL{
		Host:    validSCSIHost2,
		Channel: "0",
		Target:  "0",
		Lun:     strconv.FormatInt(int64(validLunNumber), 10)}
	validDevice = Device{
		WWN:  mockhelper.ValidWWID,
		Name: mockhelper.ValidDeviceName,
	}
	validDeviceMultipath = Device{
		Name:        mockhelper.ValidDMName,
		WWN:         mockhelper.ValidWWID,
		MultipathID: mockhelper.ValidWWID}
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

func addToLogData(ctx context.Context, format string, args ...interface{}) {
	logData = append(logData, fmt.Sprintf(format, args...))
}

func (tl *testLogger) Info(ctx context.Context, format string, args ...interface{}) {
	addToLogData(ctx, format, args...)
}
func (tl *testLogger) Debug(ctx context.Context, format string, args ...interface{}) {
	addToLogData(ctx, format, args...)
}
func (tl *testLogger) Error(ctx context.Context, format string, args ...interface{}) {
	addToLogData(ctx, format, args...)
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

func (tt *testTracer) Trace(ctx context.Context, format string, args ...interface{}) {
	traceData = append(traceData, fmt.Sprintf(format, args...))
}
