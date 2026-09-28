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
// Source: interface.go

// Package powerpath is a generated GoMock package.
package powerpath

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockPowerpath is a mock of Powerpath interface.
type MockPowerpath struct {
	ctrl     *gomock.Controller
	recorder *MockPowerpathMockRecorder
}

// MockPowerpathMockRecorder is the mock recorder for MockPowerpath.
type MockPowerpathMockRecorder struct {
	mock *MockPowerpath
}

// NewMockPowerpath creates a new mock instance.
func NewMockPowerpath(ctrl *gomock.Controller) *MockPowerpath {
	mock := &MockPowerpath{ctrl: ctrl}
	mock.recorder = &MockPowerpathMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPowerpath) EXPECT() *MockPowerpathMockRecorder {
	return m.recorder
}

// FlushDevice mocks base method.
func (m *MockPowerpath) FlushDevice(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FlushDevice", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// FlushDevice indicates an expected call of FlushDevice.
func (mr *MockPowerpathMockRecorder) FlushDevice(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FlushDevice", reflect.TypeOf((*MockPowerpath)(nil).FlushDevice), ctx)
}

// GetPowerPathDevices mocks base method.
func (m *MockPowerpath) GetPowerPathDevices(ctx context.Context, devices []string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPowerPathDevices", ctx, devices)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPowerPathDevices indicates an expected call of GetPowerPathDevices.
func (mr *MockPowerpathMockRecorder) GetPowerPathDevices(ctx, devices interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPowerPathDevices", reflect.TypeOf((*MockPowerpath)(nil).GetPowerPathDevices), ctx, devices)
}

// IsDaemonRunning mocks base method.
func (m *MockPowerpath) IsDaemonRunning(ctx context.Context) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsDaemonRunning", ctx)
	ret0, _ := ret[0].(bool)
	return ret0
}

// IsDaemonRunning indicates an expected call of IsDaemonRunning.
func (mr *MockPowerpathMockRecorder) IsDaemonRunning(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsDaemonRunning", reflect.TypeOf((*MockPowerpath)(nil).IsDaemonRunning), ctx)
}
