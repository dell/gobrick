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

// Package mockhelper is a generated GoMock package.
package mockhelper

import (
	"errors"
	"os"

	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/goiscsi"
	"github.com/golang/mock/gomock"
)

var (
	// TestErrMsg defines test error message
	TestErrMsg = "test error"
	// ErrTest defines test error
	ErrTest = errors.New(TestErrMsg)
	// ValidWWID defines wwid
	ValidWWID = "368ccf09800533dde4fa5e9ac6e9034ef"
	// ValidSYSFCWWID defines FC wwid
	ValidSYSFCWWID = "naa.68ccf09800533dde4fa5e9ac6e9034ef"
	// ValidDMName defines device mapper name
	ValidDMName = "dm-1"
	// ValidDMPath defines device mapper path
	ValidDMPath = "/dev/dm-1"
	// ValidDevicePath defines device path
	ValidDevicePath = "/dev/sde"
	// ValidDeviceName defines device name
	ValidDeviceName = "sde"
	// ValidDevicePath2 defines device path2
	ValidDevicePath2 = "/dev/sdf"
	// ValidDeviceName2 defines device name2
	ValidDeviceName2 = "sdf"
	// ValidDevices defines list of devices
	ValidDevices = []string{ValidDeviceName, ValidDeviceName2}
)

// MockHelper defines necessary fields used by mock methods
type MockHelper struct {
	Ctrl                                  *gomock.Controller
	OSOpenFileCallPath                    string
	OSStatCallPath                        string
	OSStatFileInfoIsDirOKReturn           bool
	OSEXECCommandContextName              string
	OSEXECCommandContextArgs              []string
	OSEXECCmdOKReturn                     string
	OSIsNotExistOKReturn                  bool
	IOUTILReadFileCallPath                string
	IOUTILReadFileOKReturn                string
	FileWriteStringCallData               string
	FilePathGlobCallPattern               string
	FilePathGlobOKReturn                  []string
	FilePathEvalSymlinksCallPath          string
	FilePathEvalSymlinksOKReturn          string
	ISCSILibGetInitiatorsCallFilename     string
	ISCSILibGetInitiatorsOKReturn         []string
	ISCSILibGetSessionsOKReturn           []goiscsi.ISCSISession
	ISCSILibPerformLoginCallTarget        goiscsi.ISCSITarget
	ISCSILibCreateOrUpdateNodeCallTarget  goiscsi.ISCSITarget
	ISCSILibCreateOrUpdateNodeCallOptions map[string]string
}

// OSStatCall contains mock implementation of Stat method in LimitedOS interface
func (mh *MockHelper) OSStatCall(m *wrp.MockLimitedOS) *gomock.Call {
	return m.EXPECT().Stat(mh.OSStatCallPath)
}

// OSStatCallErr mocks returning error for OSStatCall
func (mh *MockHelper) OSStatCallErr(m *wrp.MockLimitedOS) *gomock.Call {
	return mh.OSStatCall(m).Return(wrp.NewMockLimitedFileInfo(mh.Ctrl), ErrTest)
}

// OSStatCallOK mocks returning success for OSStatCall
func (mh *MockHelper) OSStatCallOK(m *wrp.MockLimitedOS) (*gomock.Call, *wrp.MockLimitedFileInfo) {
	statMock := wrp.NewMockLimitedFileInfo(mh.Ctrl)
	return mh.OSStatCall(m).Return(statMock, nil), statMock
}

// OSStatFileInfoIsDirCall contains mock implementation of IsDir method in LimitedFileInfo interface
func (mh *MockHelper) OSStatFileInfoIsDirCall(m *wrp.MockLimitedFileInfo) *gomock.Call {
	return m.EXPECT().IsDir()
}

// OSStatFileInfoIsDirOK mocks returning success from OSStatFileInfoIsDirCall
func (mh *MockHelper) OSStatFileInfoIsDirOK(m *wrp.MockLimitedFileInfo) *gomock.Call {
	return mh.OSStatFileInfoIsDirCall(m).Return(mh.OSStatFileInfoIsDirOKReturn)
}

// OSOpenFileCall mocks implementation of OpenFile method in LimitedOS interface
func (mh *MockHelper) OSOpenFileCall(m *wrp.MockLimitedOS) *gomock.Call {
	return m.EXPECT().OpenFile(
		mh.OSOpenFileCallPath,
		os.O_APPEND|os.O_WRONLY, os.FileMode(0200))
}

// OSOpenFileCallErr mocks returning error from OSOpenFileCall
func (mh *MockHelper) OSOpenFileCallErr(m *wrp.MockLimitedOS) *gomock.Call {
	return mh.OSOpenFileCall(m).Return(wrp.NewMockLimitedFile(mh.Ctrl), ErrTest)
}

// OSOpenFileCallOK mocks returning success from OSOpenFileCall
func (mh *MockHelper) OSOpenFileCallOK(m *wrp.MockLimitedOS) (*gomock.Call, *wrp.MockLimitedFile) {
	fileMock := wrp.NewMockLimitedFile(mh.Ctrl)
	call := mh.OSOpenFileCall(m).Return(fileMock, nil)
	return call, fileMock
}

// OSIsNotExistCall mocks implementation of IsNotExist method in LimitedOS interface
func (mh *MockHelper) OSIsNotExistCall(m *wrp.MockLimitedOS) *gomock.Call {
	return m.EXPECT().IsNotExist(gomock.Any())
}

// OSIsNotExistOK mocks returning success from OSIsNotExistCall
func (mh *MockHelper) OSIsNotExistOK(m *wrp.MockLimitedOS) *gomock.Call {
	return mh.OSIsNotExistCall(m).Return(mh.OSIsNotExistOKReturn)
}

// FileWriteStringCall mocks implementation of WriteString method in LimitedFile interface
func (mh *MockHelper) FileWriteStringCall(m *wrp.MockLimitedFile) *gomock.Call {
	return m.EXPECT().WriteString(mh.FileWriteStringCallData)
}

// FileWriteStringErr mocks returning error from FileWriteStringCall
func (mh *MockHelper) FileWriteStringErr(m *wrp.MockLimitedFile) *gomock.Call {
	return mh.FileWriteStringCall(m).Return(0, ErrTest)
}

// FileWriteStringOK mocks returning success from FileWriteStringCall
func (mh *MockHelper) FileWriteStringOK(m *wrp.MockLimitedFile) *gomock.Call {
	return mh.FileWriteStringCall(m).Return(5, nil)
}

// FileCloseCall mocks implementation of Close method in LimitedFile interface
func (mh *MockHelper) FileCloseCall(m *wrp.MockLimitedFile) *gomock.Call {
	return m.EXPECT().Close()
}

// FileCloseOK mocks returning success from FileCloseCall
func (mh *MockHelper) FileCloseOK(m *wrp.MockLimitedFile) *gomock.Call {
	return mh.FileCloseCall(m).Return(nil)
}

// FileCloseErr mocks returning error from FileCloseCall
func (mh *MockHelper) FileCloseErr(m *wrp.MockLimitedFile) *gomock.Call {
	return mh.FileCloseCall(m).Return(ErrTest)
}

// IOUTILReadFileCall mocks implementation of ReadFile method in LimitedIOUtil interface
func (mh *MockHelper) IOUTILReadFileCall(m *wrp.MockLimitedIOUtil) *gomock.Call {
	return m.EXPECT().ReadFile(mh.IOUTILReadFileCallPath)
}

// IOUTILReadFileErr mocks returning error from IOUTILReadFileCall
func (mh *MockHelper) IOUTILReadFileErr(m *wrp.MockLimitedIOUtil) *gomock.Call {
	return mh.IOUTILReadFileCall(m).Return(nil, ErrTest)
}

// IOUTILReadFileOK mocks returning success from IOUTILReadFileCall
func (mh *MockHelper) IOUTILReadFileOK(m *wrp.MockLimitedIOUtil) *gomock.Call {
	return mh.IOUTILReadFileCall(m).Return([]byte(mh.IOUTILReadFileOKReturn), nil)
}

// FilePathGlobCall mocks implementation of Glob method in LimitedFilepath interface
func (mh *MockHelper) FilePathGlobCall(m *wrp.MockLimitedFilepath) *gomock.Call {
	return m.EXPECT().Glob(mh.FilePathGlobCallPattern)
}

// FilePathGlobErr mocks returning error from FilePathGlobCall
func (mh *MockHelper) FilePathGlobErr(m *wrp.MockLimitedFilepath) *gomock.Call {
	return mh.FilePathGlobCall(m).Return(nil, ErrTest)
}

// FilePathGlobOK mocks returning success from FilePathGlobCall
func (mh *MockHelper) FilePathGlobOK(m *wrp.MockLimitedFilepath) *gomock.Call {
	return mh.FilePathGlobCall(m).Return(mh.FilePathGlobOKReturn, nil)
}

// OSExecCommandContextCall mocks implementation of CommandContext method from LimitedOSExec interface
func (mh *MockHelper) OSExecCommandContextCall(m *wrp.MockLimitedOSExec) *gomock.Call {
	return m.EXPECT().CommandContext(gomock.Any(),
		mh.OSEXECCommandContextName, mh.OSEXECCommandContextArgs)
}

// OSExecCommandContextOK mocks returning success from OSExecCommandContextCall
func (mh *MockHelper) OSExecCommandContextOK(m *wrp.MockLimitedOSExec) (
	*gomock.Call, *wrp.MockLimitedOSExecCmd) {
	cmdMock := wrp.NewMockLimitedOSExecCmd(mh.Ctrl)
	call := mh.OSExecCommandContextCall(m).Return(cmdMock)
	return call, cmdMock
}

// OSExecCmdCall mocks implementation of CombinedOutput method in LimitedOSExecCmd interface
func (mh *MockHelper) OSExecCmdCall(m *wrp.MockLimitedOSExecCmd) *gomock.Call {
	return m.EXPECT().CombinedOutput()
}

// OSExecCmdOK mocks returning success from OSExecCmdCall
func (mh *MockHelper) OSExecCmdOK(m *wrp.MockLimitedOSExecCmd) *gomock.Call {
	return mh.OSExecCmdCall(m).Return([]byte(mh.OSEXECCmdOKReturn), nil)
}

// OSExecCmdErr mocks returning error from OSExecCmdCall
func (mh *MockHelper) OSExecCmdErr(m *wrp.MockLimitedOSExecCmd) *gomock.Call {
	return mh.OSExecCmdCall(m).Return(nil, ErrTest)
}

// FilePathEvalSymlinksCall mocks implementation of EvalSymlinks method in LimitedFilepath interface
func (mh *MockHelper) FilePathEvalSymlinksCall(m *wrp.MockLimitedFilepath) *gomock.Call {
	return m.EXPECT().EvalSymlinks(mh.FilePathEvalSymlinksCallPath)
}

// FilePathEvalSymlinksErr mocks returning error from FilePathEvalSymlinksCall
func (mh *MockHelper) FilePathEvalSymlinksErr(m *wrp.MockLimitedFilepath) *gomock.Call {
	return mh.FilePathEvalSymlinksCall(m).Return("", ErrTest)
}

// FilePathEvalSymlinksOK mocks returning success from FilePathEvalSymlinksCall
func (mh *MockHelper) FilePathEvalSymlinksOK(m *wrp.MockLimitedFilepath) *gomock.Call {
	return mh.FilePathEvalSymlinksCall(m).Return(mh.FilePathEvalSymlinksOKReturn, nil)
}

// ISCSILibGetInitiatorsCall mocks implementation of GetInitiators method in ISCSILib interface
func (mh *MockHelper) ISCSILibGetInitiatorsCall(m *wrp.MockISCSILib) *gomock.Call {
	return m.EXPECT().GetInitiators(mh.ISCSILibGetInitiatorsCallFilename)
}

// ISCSILibGetInitiatorsOK mocks returning success from ISCSILibGetInitiatorsCall
func (mh *MockHelper) ISCSILibGetInitiatorsOK(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibGetInitiatorsCall(m).Return(mh.ISCSILibGetInitiatorsOKReturn, nil)
}

// ISCSILibGetInitiatorsErr mocks returning error from ISCSILibGetInitiatorsCall
func (mh *MockHelper) ISCSILibGetInitiatorsErr(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibGetInitiatorsCall(m).Return(nil, ErrTest)
}

// ISCSILibPerformLoginCall mocks implementation of PerformLogin method in ISCSILib interface
func (mh *MockHelper) ISCSILibPerformLoginCall(m *wrp.MockISCSILib) *gomock.Call {
	return m.EXPECT().PerformLogin(mh.ISCSILibPerformLoginCallTarget)
}

// ISCSILibPerformLoginOK mocks returning success from ISCSILibPerformLoginCall
func (mh *MockHelper) ISCSILibPerformLoginOK(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibPerformLoginCall(m).Return(nil)
}

// ISCSILibPerformLoginErr mocks returning error from ISCSILibPerformLoginCall
func (mh *MockHelper) ISCSILibPerformLoginErr(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibPerformLoginCall(m).Return(ErrTest)
}

// ISCSILibGetSessionsCall mocks implementation of GetSessions method from ISCSILib interface
func (mh *MockHelper) ISCSILibGetSessionsCall(m *wrp.MockISCSILib) *gomock.Call {
	return m.EXPECT().GetSessions()
}

// ISCSILibGetSessionsOK mocks returning success from ISCSILibGetSessionsCall
func (mh *MockHelper) ISCSILibGetSessionsOK(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibGetSessionsCall(m).Return(mh.ISCSILibGetSessionsOKReturn, nil)
}

// ISCSILibGetSessionsErr mocks returning error from ISCSILibGetSessionsCall
func (mh *MockHelper) ISCSILibGetSessionsErr(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibGetSessionsCall(m).Return(nil, ErrTest)
}

// ISCSILibCreateOrUpdateNodeCall mocks implementation of CreateOrUpdateNode method in ISCSILib interface
func (mh *MockHelper) ISCSILibCreateOrUpdateNodeCall(m *wrp.MockISCSILib) *gomock.Call {
	return m.EXPECT().CreateOrUpdateNode(mh.ISCSILibCreateOrUpdateNodeCallTarget,
		mh.ISCSILibCreateOrUpdateNodeCallOptions)
}

// ISCSILibCreateOrUpdateNodeOK mocks returning success from ISCSILibCreateOrUpdateNodeCall
func (mh *MockHelper) ISCSILibCreateOrUpdateNodeOK(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibCreateOrUpdateNodeCall(m).Return(nil)
}

// ISCSILibCreateOrUpdateNodeErr mocks returning err from ISCSILibCreateOrUpdateNodeCall
func (mh *MockHelper) ISCSILibCreateOrUpdateNodeErr(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibCreateOrUpdateNodeCall(m).Return(ErrTest)
}
