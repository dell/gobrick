package mockhelper

import (
	"errors"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/goiscsi"
	"github.com/golang/mock/gomock"
	"os"
)

var (
	TestErrMsg       = "test error"
	TestErr          = errors.New(TestErrMsg)
	ValidWWID        = "368ccf09800533dde4fa5e9ac6e9034ef"
	ValidSYSFCWWID   = "naa.68ccf09800533dde4fa5e9ac6e9034ef"
	ValidDMName      = "dm-1"
	ValidDMPath      = "/dev/dm-1"
	ValidDevicePath  = "/dev/sde"
	ValidDeviceName  = "sde"
	ValidDevicePath2 = "/dev/sdf"
	ValidDeviceName2 = "sdf"
	ValidDevices     = []string{ValidDeviceName, ValidDeviceName2}
)

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

func (mh *MockHelper) OSStatCall(m *wrp.MockLimitedOS) *gomock.Call {
	return m.EXPECT().Stat(mh.OSStatCallPath)
}

func (mh *MockHelper) OSStatCallErr(m *wrp.MockLimitedOS) *gomock.Call {
	return mh.OSStatCall(m).Return(wrp.NewMockLimitedFileInfo(mh.Ctrl), TestErr)
}

func (mh *MockHelper) OSStatCallOK(m *wrp.MockLimitedOS) (*gomock.Call, *wrp.MockLimitedFileInfo) {
	statMock := wrp.NewMockLimitedFileInfo(mh.Ctrl)
	return mh.OSStatCall(m).Return(statMock, nil), statMock
}

func (mh *MockHelper) OSStatFileInfoIsDirCall(m *wrp.MockLimitedFileInfo) *gomock.Call {
	return m.EXPECT().IsDir()
}

func (mh *MockHelper) OSStatFileInfoIsDirOK(m *wrp.MockLimitedFileInfo) *gomock.Call {
	return mh.OSStatFileInfoIsDirCall(m).Return(mh.OSStatFileInfoIsDirOKReturn)
}

func (mh *MockHelper) OSOpenFileCall(m *wrp.MockLimitedOS) *gomock.Call {
	return m.EXPECT().OpenFile(
		mh.OSOpenFileCallPath,
		os.O_APPEND|os.O_WRONLY, os.FileMode(0200))
}

func (mh *MockHelper) OSOpenFileCallErr(m *wrp.MockLimitedOS) *gomock.Call {
	return mh.OSOpenFileCall(m).Return(wrp.NewMockLimitedFile(mh.Ctrl), TestErr)
}

func (mh *MockHelper) OSIsNotExistCall(m *wrp.MockLimitedOS) *gomock.Call {
	return m.EXPECT().IsNotExist(gomock.Any())
}

func (mh *MockHelper) OSIsNotExistOK(m *wrp.MockLimitedOS) *gomock.Call {
	return mh.OSIsNotExistCall(m).Return(mh.OSIsNotExistOKReturn)
}

func (mh *MockHelper) OSOpenFileCallOK(m *wrp.MockLimitedOS) (*gomock.Call, *wrp.MockLimitedFile) {
	fileMock := wrp.NewMockLimitedFile(mh.Ctrl)
	call := mh.OSOpenFileCall(m).Return(fileMock, nil)
	return call, fileMock
}

func (mh *MockHelper) FileWriteStringCall(m *wrp.MockLimitedFile) *gomock.Call {
	return m.EXPECT().WriteString(mh.FileWriteStringCallData)
}

func (mh *MockHelper) FileWriteStringErr(m *wrp.MockLimitedFile) *gomock.Call {
	return mh.FileWriteStringCall(m).Return(0, TestErr)
}

func (mh *MockHelper) FileWriteStringOK(m *wrp.MockLimitedFile) *gomock.Call {
	return mh.FileWriteStringCall(m).Return(5, nil)
}

func (mh *MockHelper) FileCloseCall(m *wrp.MockLimitedFile) *gomock.Call {
	return m.EXPECT().Close()
}

func (mh *MockHelper) FileCloseOK(m *wrp.MockLimitedFile) *gomock.Call {
	return mh.FileCloseCall(m).Return(nil)
}

func (mh *MockHelper) FileCloseErr(m *wrp.MockLimitedFile) *gomock.Call {
	return mh.FileCloseCall(m).Return(TestErr)
}

func (mh *MockHelper) IOUTILReadFileCall(m *wrp.MockLimitedIOUtil) *gomock.Call {
	return m.EXPECT().ReadFile(mh.IOUTILReadFileCallPath)
}

func (mh *MockHelper) IOUTILReadFileErr(m *wrp.MockLimitedIOUtil) *gomock.Call {
	return mh.IOUTILReadFileCall(m).Return(nil, TestErr)
}

func (mh *MockHelper) IOUTILReadFileOK(m *wrp.MockLimitedIOUtil) *gomock.Call {
	return mh.IOUTILReadFileCall(m).Return([]byte(mh.IOUTILReadFileOKReturn), nil)
}

func (mh *MockHelper) FilePathGlobCall(m *wrp.MockLimitedFilepath) *gomock.Call {
	return m.EXPECT().Glob(mh.FilePathGlobCallPattern)
}

func (mh *MockHelper) FilePathGlobErr(m *wrp.MockLimitedFilepath) *gomock.Call {
	return mh.FilePathGlobCall(m).Return(nil, TestErr)
}

func (mh *MockHelper) FilePathGlobOK(m *wrp.MockLimitedFilepath) *gomock.Call {
	return mh.FilePathGlobCall(m).Return(mh.FilePathGlobOKReturn, nil)
}

func (mh *MockHelper) OSExecCommandContextCall(m *wrp.MockLimitedOSExec) *gomock.Call {
	return m.EXPECT().CommandContext(gomock.Any(),
		mh.OSEXECCommandContextName, mh.OSEXECCommandContextArgs)
}

func (mh *MockHelper) OSExecCommandContextOK(m *wrp.MockLimitedOSExec) (
	*gomock.Call, *wrp.MockLimitedOSExecCmd) {
	cmdMock := wrp.NewMockLimitedOSExecCmd(mh.Ctrl)
	call := mh.OSExecCommandContextCall(m).Return(cmdMock)
	return call, cmdMock
}

func (mh *MockHelper) OSExecCmdCall(m *wrp.MockLimitedOSExecCmd) *gomock.Call {
	return m.EXPECT().CombinedOutput()
}

func (mh *MockHelper) OSExecCmdOK(m *wrp.MockLimitedOSExecCmd) *gomock.Call {
	return mh.OSExecCmdCall(m).Return([]byte(mh.OSEXECCmdOKReturn), nil)
}

func (mh *MockHelper) OSExecCmdErr(m *wrp.MockLimitedOSExecCmd) *gomock.Call {
	return mh.OSExecCmdCall(m).Return(nil, TestErr)
}

func (mh *MockHelper) FilePathEvalSymlinksCall(m *wrp.MockLimitedFilepath) *gomock.Call {
	return m.EXPECT().EvalSymlinks(mh.FilePathEvalSymlinksCallPath)
}

func (mh *MockHelper) FilePathEvalSymlinksErr(m *wrp.MockLimitedFilepath) *gomock.Call {
	return mh.FilePathEvalSymlinksCall(m).Return("", TestErr)
}

func (mh *MockHelper) FilePathEvalSymlinksOK(m *wrp.MockLimitedFilepath) *gomock.Call {
	return mh.FilePathEvalSymlinksCall(m).Return(mh.FilePathEvalSymlinksOKReturn, nil)
}

func (mh *MockHelper) ISCSILibGetInitiatorsCall(m *wrp.MockISCSILib) *gomock.Call {
	return m.EXPECT().GetInitiators(mh.ISCSILibGetInitiatorsCallFilename)
}

func (mh *MockHelper) ISCSILibGetInitiatorsOK(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibGetInitiatorsCall(m).Return(mh.ISCSILibGetInitiatorsOKReturn, nil)
}

func (mh *MockHelper) ISCSILibGetInitiatorsErr(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibGetInitiatorsCall(m).Return(nil, TestErr)
}

func (mh *MockHelper) ISCSILibPerformLoginCall(m *wrp.MockISCSILib) *gomock.Call {
	return m.EXPECT().PerformLogin(mh.ISCSILibPerformLoginCallTarget)
}

func (mh *MockHelper) ISCSILibPerformLoginOK(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibPerformLoginCall(m).Return(nil)
}

func (mh *MockHelper) ISCSILibPerformLoginErr(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibPerformLoginCall(m).Return(TestErr)
}

func (mh *MockHelper) ISCSILibGetSessionsCall(m *wrp.MockISCSILib) *gomock.Call {
	return m.EXPECT().GetSessions()
}

func (mh *MockHelper) ISCSILibGetSessionsOK(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibGetSessionsCall(m).Return(mh.ISCSILibGetSessionsOKReturn, nil)
}

func (mh *MockHelper) ISCSILibGetSessionsErr(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibGetSessionsCall(m).Return(nil, TestErr)
}

func (mh *MockHelper) ISCSILibCreateOrUpdateNodeCall(m *wrp.MockISCSILib) *gomock.Call {
	return m.EXPECT().CreateOrUpdateNode(mh.ISCSILibCreateOrUpdateNodeCallTarget,
		mh.ISCSILibCreateOrUpdateNodeCallOptions)
}

func (mh *MockHelper) ISCSILibCreateOrUpdateNodeOK(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibCreateOrUpdateNodeCall(m).Return(nil)
}

func (mh *MockHelper) ISCSILibCreateOrUpdateNodeErr(m *wrp.MockISCSILib) *gomock.Call {
	return mh.ISCSILibCreateOrUpdateNodeCall(m).Return(TestErr)
}
