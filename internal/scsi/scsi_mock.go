// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go

// Package scsi is a generated GoMock package.
package scsi

import (
	context "context"
	reflect "reflect"

	scsi "github.com/dell/gobrick/pkg/scsi"
	gomock "github.com/golang/mock/gomock"
)

// MockSCSI is a mock of SCSI interface.
type MockSCSI struct {
	ctrl     *gomock.Controller
	recorder *MockSCSIMockRecorder
}

// MockSCSIMockRecorder is the mock recorder for MockSCSI.
type MockSCSIMockRecorder struct {
	mock *MockSCSI
}

// NewMockSCSI creates a new mock instance.
func NewMockSCSI(ctrl *gomock.Controller) *MockSCSI {
	mock := &MockSCSI{ctrl: ctrl}
	mock.recorder = &MockSCSIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSCSI) EXPECT() *MockSCSIMockRecorder {
	return m.recorder
}

// CheckDeviceIsValid mocks base method.
func (m *MockSCSI) CheckDeviceIsValid(ctx context.Context, device string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckDeviceIsValid", ctx, device)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckDeviceIsValid indicates an expected call of CheckDeviceIsValid.
func (mr *MockSCSIMockRecorder) CheckDeviceIsValid(ctx, device interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckDeviceIsValid", reflect.TypeOf((*MockSCSI)(nil).CheckDeviceIsValid), ctx, device)
}

// DeleteSCSIDeviceByHCTL mocks base method.
func (m *MockSCSI) DeleteSCSIDeviceByHCTL(ctx context.Context, h scsi.HCTL) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSCSIDeviceByHCTL", ctx, h)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSCSIDeviceByHCTL indicates an expected call of DeleteSCSIDeviceByHCTL.
func (mr *MockSCSIMockRecorder) DeleteSCSIDeviceByHCTL(ctx, h interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSCSIDeviceByHCTL", reflect.TypeOf((*MockSCSI)(nil).DeleteSCSIDeviceByHCTL), ctx, h)
}

// DeleteSCSIDeviceByName mocks base method.
func (m *MockSCSI) DeleteSCSIDeviceByName(ctx context.Context, name string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSCSIDeviceByName", ctx, name)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSCSIDeviceByName indicates an expected call of DeleteSCSIDeviceByName.
func (mr *MockSCSIMockRecorder) DeleteSCSIDeviceByName(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSCSIDeviceByName", reflect.TypeOf((*MockSCSI)(nil).DeleteSCSIDeviceByName), ctx, name)
}

// DeleteSCSIDeviceByPath mocks base method.
func (m *MockSCSI) DeleteSCSIDeviceByPath(ctx context.Context, devPath string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSCSIDeviceByPath", ctx, devPath)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSCSIDeviceByPath indicates an expected call of DeleteSCSIDeviceByPath.
func (mr *MockSCSIMockRecorder) DeleteSCSIDeviceByPath(ctx, devPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSCSIDeviceByPath", reflect.TypeOf((*MockSCSI)(nil).DeleteSCSIDeviceByPath), ctx, devPath)
}

// GetDMChildren mocks base method.
func (m *MockSCSI) GetDMChildren(ctx context.Context, dmPath string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDMChildren", ctx, dmPath)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDMChildren indicates an expected call of GetDMChildren.
func (mr *MockSCSIMockRecorder) GetDMChildren(ctx, dmPath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDMChildren", reflect.TypeOf((*MockSCSI)(nil).GetDMChildren), ctx, dmPath)
}

// GetDMDeviceByChildren mocks base method.
func (m *MockSCSI) GetDMDeviceByChildren(ctx context.Context, devices []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDMDeviceByChildren", ctx, devices)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDMDeviceByChildren indicates an expected call of GetDMDeviceByChildren.
func (mr *MockSCSIMockRecorder) GetDMDeviceByChildren(ctx, devices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDMDeviceByChildren", reflect.TypeOf((*MockSCSI)(nil).GetDMDeviceByChildren), ctx, devices)
}

// GetDeviceNameByHCTL mocks base method.
func (m *MockSCSI) GetDeviceNameByHCTL(ctx context.Context, h scsi.HCTL) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceNameByHCTL", ctx, h)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceNameByHCTL indicates an expected call of GetDeviceNameByHCTL.
func (mr *MockSCSIMockRecorder) GetDeviceNameByHCTL(ctx, h interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceNameByHCTL", reflect.TypeOf((*MockSCSI)(nil).GetDeviceNameByHCTL), ctx, h)
}

// GetDeviceWWN mocks base method.
func (m *MockSCSI) GetDeviceWWN(ctx context.Context, devices []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDeviceWWN", ctx, devices)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDeviceWWN indicates an expected call of GetDeviceWWN.
func (mr *MockSCSIMockRecorder) GetDeviceWWN(ctx, devices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDeviceWWN", reflect.TypeOf((*MockSCSI)(nil).GetDeviceWWN), ctx, devices)
}

// GetDevicesByWWN mocks base method.
func (m *MockSCSI) GetDevicesByWWN(ctx context.Context, wwn string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDevicesByWWN", ctx, wwn)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDevicesByWWN indicates an expected call of GetDevicesByWWN.
func (mr *MockSCSIMockRecorder) GetDevicesByWWN(ctx, wwn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDevicesByWWN", reflect.TypeOf((*MockSCSI)(nil).GetDevicesByWWN), ctx, wwn)
}

// GetNVMEDMDeviceByChildren mocks base method.
func (m *MockSCSI) GetNVMEDMDeviceByChildren(ctx context.Context, devices []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNVMEDMDeviceByChildren", ctx, devices)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNVMEDMDeviceByChildren indicates an expected call of GetNVMEDMDeviceByChildren.
func (mr *MockSCSIMockRecorder) GetNVMEDMDeviceByChildren(ctx, devices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNVMEDMDeviceByChildren", reflect.TypeOf((*MockSCSI)(nil).GetNVMEDMDeviceByChildren), ctx, devices)
}

// GetNVMEDeviceWWN mocks base method.
func (m *MockSCSI) GetNVMEDeviceWWN(ctx context.Context, devices []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNVMEDeviceWWN", ctx, devices)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNVMEDeviceWWN indicates an expected call of GetNVMEDeviceWWN.
func (mr *MockSCSIMockRecorder) GetNVMEDeviceWWN(ctx, devices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNVMEDeviceWWN", reflect.TypeOf((*MockSCSI)(nil).GetNVMEDeviceWWN), ctx, devices)
}

// GetNVMEMultipathDMName mocks base method.
func (m *MockSCSI) GetNVMEMultipathDMName(device, pattern string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNVMEMultipathDMName", device, pattern)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNVMEMultipathDMName indicates an expected call of GetNVMEMultipathDMName.
func (mr *MockSCSIMockRecorder) GetNVMEMultipathDMName(device, pattern interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNVMEMultipathDMName", reflect.TypeOf((*MockSCSI)(nil).GetNVMEMultipathDMName), device, pattern)
}

// IsDeviceExist mocks base method.
func (m *MockSCSI) IsDeviceExist(ctx context.Context, device string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDeviceExist", ctx, device)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDeviceExist indicates an expected call of IsDeviceExist.
func (mr *MockSCSIMockRecorder) IsDeviceExist(ctx, device interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDeviceExist", reflect.TypeOf((*MockSCSI)(nil).IsDeviceExist), ctx, device)
}

// RescanSCSIDeviceByHCTL mocks base method.
func (m *MockSCSI) RescanSCSIDeviceByHCTL(ctx context.Context, h scsi.HCTL) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RescanSCSIDeviceByHCTL", ctx, h)
	ret0, _ := ret[0].(error)
	return ret0
}

// RescanSCSIDeviceByHCTL indicates an expected call of RescanSCSIDeviceByHCTL.
func (mr *MockSCSIMockRecorder) RescanSCSIDeviceByHCTL(ctx, h interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RescanSCSIDeviceByHCTL", reflect.TypeOf((*MockSCSI)(nil).RescanSCSIDeviceByHCTL), ctx, h)
}

// RescanSCSIHostByHCTL mocks base method.
func (m *MockSCSI) RescanSCSIHostByHCTL(ctx context.Context, h scsi.HCTL) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RescanSCSIHostByHCTL", ctx, h)
	ret0, _ := ret[0].(error)
	return ret0
}

// RescanSCSIHostByHCTL indicates an expected call of RescanSCSIHostByHCTL.
func (mr *MockSCSIMockRecorder) RescanSCSIHostByHCTL(ctx, h interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RescanSCSIHostByHCTL", reflect.TypeOf((*MockSCSI)(nil).RescanSCSIHostByHCTL), ctx, h)
}

// WaitUdevSymlink mocks base method.
func (m *MockSCSI) WaitUdevSymlink(ctx context.Context, deviceName, wwn string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUdevSymlink", ctx, deviceName, wwn)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitUdevSymlink indicates an expected call of WaitUdevSymlink.
func (mr *MockSCSIMockRecorder) WaitUdevSymlink(ctx, deviceName, wwn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUdevSymlink", reflect.TypeOf((*MockSCSI)(nil).WaitUdevSymlink), ctx, deviceName, wwn)
}

// WaitUdevSymlinkNVMe mocks base method.
func (m *MockSCSI) WaitUdevSymlinkNVMe(ctx context.Context, deviceName, wwn string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUdevSymlinkNVMe", ctx, deviceName, wwn)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitUdevSymlinkNVMe indicates an expected call of WaitUdevSymlinkNVMe.
func (mr *MockSCSIMockRecorder) WaitUdevSymlinkNVMe(ctx, deviceName, wwn interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUdevSymlinkNVMe", reflect.TypeOf((*MockSCSI)(nil).WaitUdevSymlinkNVMe), ctx, deviceName, wwn)
}
