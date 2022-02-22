package gobrick

import (
	"context"
	"fmt"
	"github.com/dell/gobrick/internal/mockhelper"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/golang/mock/gomock"
	"golang.org/x/sync/semaphore"
	"reflect"
	"testing"
	"time"
)

var (
	validWWPN1        = "5006016349e016fd"
	validWWPN2        = "5006016a49e016fd"
	validNodeName1    = "20000024ff5b8e27"
	validNodeName2    = "20000024ff5b8e26"
	validFCVolumeInfo = FCVolumeInfo{
		Targets: []FCTargetInfo{
			{WWPN: validWWPN1},
			{WWPN: validWWPN2},
		},
		Lun: validLunNumber,
	}
)

type fcFields struct {
	baseConnector             *baseConnector
	multipath                 *intmultipath.MockMultipath
	scsi                      *intscsi.MockSCSI
	filePath                  *wrp.MockLimitedFilepath
	os                        *wrp.MockLimitedOS
	ioutil                    *wrp.MockLimitedIOUtil
	limiter                   *semaphore.Weighted
	waitDeviceRegisterTimeout time.Duration
}

func getDefaultFCFields(ctrl *gomock.Controller) fcFields {
	con := NewFCConnector(FCConnectorParams{})
	bc := con.baseConnector
	mpMock := intmultipath.NewMockMultipath(ctrl)
	scsiMock := intscsi.NewMockSCSI(ctrl)
	bc.multipath = mpMock
	bc.scsi = scsiMock
	f := fcFields{
		baseConnector:             bc,
		multipath:                 mpMock,
		scsi:                      scsiMock,
		filePath:                  wrp.NewMockLimitedFilepath(ctrl),
		os:                        wrp.NewMockLimitedOS(ctrl),
		ioutil:                    wrp.NewMockLimitedIOUtil(ctrl),
		limiter:                   con.limiter,
		waitDeviceRegisterTimeout: con.waitDeviceRegisterTimeout,
	}
	return f
}

func getFCHBASInfoMock(mock *baseMockHelper,
	os *wrp.MockLimitedOS,
	ioutil *wrp.MockLimitedIOUtil, filepath *wrp.MockLimitedFilepath) {
	isFCSupportedMock(mock, os)
	sysPath := "/sys/class/fc_host/host"
	sysPathGlob := sysPath + "*"
	host1Path := sysPath + validSCSIHost1
	host2Path := sysPath + validSCSIHost2
	mock.FilePathGlobCallPattern = sysPathGlob
	mock.FilePathGlobOKReturn = []string{host1Path, host2Path}
	mock.FilePathGlobOK(filepath)

	mock.IOUTILReadFileCallPath = host1Path + "/port_name"
	mock.IOUTILReadFileOKReturn = "0x" + validWWPN1
	mock.IOUTILReadFileOK(ioutil)
	mock.IOUTILReadFileCallPath = host1Path + "/node_name"
	mock.IOUTILReadFileOKReturn = "0x" + validNodeName1
	mock.IOUTILReadFileOK(ioutil)

	mock.IOUTILReadFileCallPath = host2Path + "/port_name"
	mock.IOUTILReadFileOKReturn = "0x" + validWWPN2
	mock.IOUTILReadFileOK(ioutil)
	mock.IOUTILReadFileCallPath = host2Path + "/node_name"
	mock.IOUTILReadFileOKReturn = "0x" + validNodeName2
	mock.IOUTILReadFileOK(ioutil)
}

func isFCSupportedMock(mock *baseMockHelper, os *wrp.MockLimitedOS) {
	mock.OSStatCallPath = "/sys/class/fc_host"
	_, statMock := mock.OSStatCallOK(os)
	mock.OSStatFileInfoIsDirOKReturn = true
	mock.OSStatFileInfoIsDirOK(statMock)
}

func waitForDeviceWWNMock(mock *baseMockHelper,
	filepath *wrp.MockLimitedFilepath,
	ioutil *wrp.MockLimitedIOUtil,
	scsi *intscsi.MockSCSI) {
	findHCTLsForFCHBAMock(mock, filepath, ioutil)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1
	mock.SCSIGetDeviceNameByHCTLErr(scsi)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL2
	mock.SCSIGetDeviceNameByHCTLErr(scsi)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1Target1
	mock.SCSIGetDeviceNameByHCTLErr(scsi)

	findHCTLsForFCHBAMock(mock, filepath, ioutil)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1
	mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName
	mock.SCSIGetDeviceNameByHCTLOK(scsi)

	mock.SCSICheckDeviceIsValidCallDevice = mockhelper.ValidDevicePath
	mock.SCSICheckDeviceIsValidOKReturn = true
	mock.SCSICheckDeviceIsValidOK(scsi)

	mock.SCSIRescanSCSIHostByHCTLCallH = validHostOnlyHCTL2
	mock.SCSIRescanSCSIHostByHCTLOK(scsi)
	mock.SCSIRescanSCSIHostByHCTLCallH = validHCTL1
	mock.SCSIRescanSCSIHostByHCTLOK(scsi)
	mock.SCSIRescanSCSIHostByHCTLCallH = validHCTL1Target1
	mock.SCSIRescanSCSIHostByHCTLOK(scsi)

	mock.SCSIGetDeviceWWNCallDevices = []string{mockhelper.ValidDeviceName}
	mock.SCSIGetDeviceWWNOKReturn = mockhelper.ValidWWID
	mock.SCSIGetDeviceWWNOK(scsi)
}

func findHCTLsForFCHBAMock(mock *baseMockHelper,
	filepath *wrp.MockLimitedFilepath, ioutil *wrp.MockLimitedIOUtil) {

	sysPath := "/sys/class/fc_transport/target"
	sysPathGlob1 := sysPath + validSCSIHost1 + ":*"
	sysPathGlob2 := sysPath + validSCSIHost2 + ":*"
	sysPathMatch1 := fmt.Sprintf("%s%s:%s:%s",
		sysPath, validSCSIHost1, validHCTL1.Channel, validHCTL1.Target)
	sysPathMatch1Target1 := fmt.Sprintf("%s%s:%s:%s",
		sysPath, validSCSIHost1, validHCTL1Target1.Channel, validHCTL1Target1.Target)
	sysPathMatch2 := fmt.Sprintf("%s%s:%s:%s",
		sysPath, validSCSIHost2, validHCTL2.Channel, validHCTL2.Target)

	mock.FilePathGlobCallPattern = sysPathGlob1
	mock.FilePathGlobOKReturn = []string{sysPathMatch1, sysPathMatch1Target1}
	mock.FilePathGlobOK(filepath)

	mock.FilePathGlobCallPattern = sysPathGlob2
	mock.FilePathGlobOKReturn = []string{sysPathMatch2}
	mock.FilePathGlobOK(filepath)

	mock.IOUTILReadFileCallPath = sysPathMatch1 + "/port_name"
	mock.IOUTILReadFileOKReturn = "0x" + validWWPN1
	mock.IOUTILReadFileOK(ioutil)
	mock.IOUTILReadFileCallPath = sysPathMatch1Target1 + "/port_name"
	mock.IOUTILReadFileOKReturn = "0x" + validWWPN2
	mock.IOUTILReadFileOK(ioutil)

	mock.IOUTILReadFileCallPath = sysPathMatch2 + "/port_name"
	mock.IOUTILReadFileOKReturn = "0x" + validWWPN2
	mock.IOUTILReadFileOK(ioutil)
}

func cleanConnectionMock(mock *baseMockHelper,
	filepath *wrp.MockLimitedFilepath,
	ioutil *wrp.MockLimitedIOUtil,
	os *wrp.MockLimitedOS,
	scsi *intscsi.MockSCSI,
	multipath *intmultipath.MockMultipath) {

	getFCHBASInfoMock(mock, os, ioutil, filepath)
	findHCTLsForFCHBAMock(mock, filepath, ioutil)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1
	mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName
	mock.SCSIGetDeviceNameByHCTLOK(scsi)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL2
	mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName2
	mock.SCSIGetDeviceNameByHCTLOK(scsi)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1Target1
	mock.SCSIGetDeviceNameByHCTLErr(scsi).AnyTimes()

	BaseConnectorCleanDeviceMock(mock, scsi, multipath)
}

func TestFCConnector_ConnectVolume(t *testing.T) {
	type args struct {
		ctx  context.Context
		info FCVolumeInfo
	}

	ctx := context.Background()
	defaultArgs := args{ctx: ctx, info: validFCVolumeInfo}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := baseMockHelper{
		MockHelper: mockhelper.MockHelper{
			Ctrl: ctrl,
		},
		Ctx: gomock.Any(),
	}

	tests := []struct {
		name        string
		fields      fcFields
		stateSetter func(fields fcFields)
		args        args
		want        Device
		wantErr     bool
	}{
		{
			name:   "ok-multipath",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				getFCHBASInfoMock(&mock, fields.os, fields.ioutil, fields.filePath)
				waitForDeviceWWNMock(&mock, fields.filePath, fields.ioutil, fields.scsi)
				mock.SCSIGetDevicesByWWNCallWWN = mockhelper.ValidWWID
				mock.SCSIGetDevicesByWWNOKReturn = mockhelper.ValidDevices
				mock.SCSIGetDevicesByWWNOK(fields.scsi)

				mock.MultipathIsDaemonRunningOKReturn = true
				mock.MultipathIsDaemonRunningOK(fields.multipath)

				mock.MultipathAddWWIDCallWWID = mockhelper.ValidWWID
				mock.MultipathAddWWIDOK(fields.multipath)

				mock.MultipathAddPathCallPath = mockhelper.ValidDevicePath
				mock.MultipathAddPathOK(fields.multipath)

				mock.MultipathAddPathCallPath = mockhelper.ValidDevicePath2
				mock.MultipathAddPathOK(fields.multipath)

				mock.SCSIGetDMDeviceByChildrenCallDevices = mockhelper.ValidDevices
				mock.SCSIGetDMDeviceByChildrenOKReturn = mockhelper.ValidDMName
				mock.SCSIGetDMDeviceByChildrenOK(fields.scsi)

				mock.SCSIWaitUdevSymlinkCallWWN = mockhelper.ValidWWID
				mock.SCSIWaitUdevSymlinkCallDevice = mockhelper.ValidDMName
				mock.SCSIWaitUdevSymlinkOK(fields.scsi)

				mock.SCSICheckDeviceIsValidOKReturn = true
				mock.SCSICheckDeviceIsValidCallDevice = mockhelper.ValidDMPath
				mock.SCSICheckDeviceIsValidOK(fields.scsi)
			},
			args:    defaultArgs,
			want:    validDeviceMultipath,
			wantErr: false,
		},
		{
			name:   "ok-single",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				getFCHBASInfoMock(&mock, fields.os, fields.ioutil, fields.filePath)
				waitForDeviceWWNMock(&mock, fields.filePath, fields.ioutil, fields.scsi)
				mock.SCSIGetDevicesByWWNCallWWN = mockhelper.ValidWWID
				mock.SCSIGetDevicesByWWNOKReturn = mockhelper.ValidDevices
				mock.SCSIGetDevicesByWWNOK(fields.scsi)

				mock.MultipathIsDaemonRunningOKReturn = false
				mock.MultipathIsDaemonRunningOK(fields.multipath)

				mock.SCSIWaitUdevSymlinkCallWWN = mockhelper.ValidWWID
				mock.SCSIWaitUdevSymlinkCallDevice = mockhelper.ValidDeviceName
				mock.SCSIWaitUdevSymlinkOK(fields.scsi)

				mock.SCSICheckDeviceIsValidOKReturn = true
				mock.SCSICheckDeviceIsValidCallDevice = mockhelper.ValidDevicePath
				mock.SCSICheckDeviceIsValidOK(fields.scsi)
			},
			args:    defaultArgs,
			want:    validDevice,
			wantErr: false,
		},
		{
			name:   "connection error",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				getFCHBASInfoMock(&mock, fields.os, fields.ioutil, fields.filePath)
				waitForDeviceWWNMock(&mock, fields.filePath, fields.ioutil, fields.scsi)
				mock.SCSIGetDevicesByWWNCallWWN = mockhelper.ValidWWID
				mock.SCSIGetDevicesByWWNOKReturn = mockhelper.ValidDevices
				mock.SCSIGetDevicesByWWNOK(fields.scsi)

				mock.MultipathIsDaemonRunningOKReturn = false
				mock.MultipathIsDaemonRunningOK(fields.multipath)

				mock.SCSIWaitUdevSymlinkCallWWN = mockhelper.ValidWWID
				mock.SCSIWaitUdevSymlinkCallDevice = mockhelper.ValidDeviceName
				mock.SCSIWaitUdevSymlinkOK(fields.scsi)

				mock.SCSICheckDeviceIsValidOKReturn = false
				mock.SCSICheckDeviceIsValidCallDevice = mockhelper.ValidDevicePath
				mock.SCSICheckDeviceIsValidOK(fields.scsi)

				cleanConnectionMock(&mock, fields.filePath, fields.ioutil,
					fields.os, fields.scsi, fields.multipath)

			},
			args:    defaultArgs,
			want:    Device{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &FCConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				filePath:                  tt.fields.filePath,
				os:                        tt.fields.os,
				ioutil:                    tt.fields.ioutil,
				limiter:                   tt.fields.limiter,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
			}
			tt.stateSetter(tt.fields)
			got, err := fc.ConnectVolume(tt.args.ctx, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConnectVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConnectVolume() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFCConnector_DisconnectVolume(t *testing.T) {
	type args struct {
		ctx  context.Context
		info FCVolumeInfo
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx, info: validFCVolumeInfo}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := baseMockHelper{
		MockHelper: mockhelper.MockHelper{
			Ctrl: ctrl,
		},
		Ctx: gomock.Any(),
	}

	tests := []struct {
		name        string
		fields      fcFields
		stateSetter func(fields fcFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				cleanConnectionMock(
					&mock, fields.filePath,
					fields.ioutil, fields.os, fields.scsi, fields.multipath)
			},
			args:    defaultArgs,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &FCConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				filePath:                  tt.fields.filePath,
				os:                        tt.fields.os,
				ioutil:                    tt.fields.ioutil,
				limiter:                   tt.fields.limiter,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
			}
			tt.stateSetter(tt.fields)
			if err := fc.DisconnectVolume(tt.args.ctx, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("DisconnectVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFCConnector_DisconnectVolumeByDeviceName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx, name: mockhelper.ValidDMName}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := baseMockHelper{
		MockHelper: mockhelper.MockHelper{
			Ctrl: ctrl,
		},
		Ctx: gomock.Any(),
	}

	tests := []struct {
		name        string
		fields      fcFields
		stateSetter func(fields fcFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				BaserConnectorDisconnectDevicesByDeviceNameMock(
					&mock, fields.scsi, fields.multipath)
			},
			args:    defaultArgs,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &FCConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				filePath:                  tt.fields.filePath,
				os:                        tt.fields.os,
				ioutil:                    tt.fields.ioutil,
				limiter:                   tt.fields.limiter,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
			}
			tt.stateSetter(tt.fields)
			if err := fc.DisconnectVolumeByDeviceName(tt.args.ctx, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("DisconnectVolumeByDeviceName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFCConnector_GetInitiatorPorts(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	ctx := context.Background()

	defaultArgs := args{ctx: ctx}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := baseMockHelper{
		MockHelper: mockhelper.MockHelper{
			Ctrl: ctrl,
		},
		Ctx: gomock.Any(),
	}

	tests := []struct {
		name        string
		fields      fcFields
		stateSetter func(fields fcFields)
		args        args
		want        []string
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultFCFields(ctrl),
			args:   defaultArgs,
			stateSetter: func(fields fcFields) {
				getFCHBASInfoMock(&mock, fields.os, fields.ioutil, fields.filePath)
			},
			want:    []string{validWWPN1, validWWPN2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fc := &FCConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				filePath:                  tt.fields.filePath,
				os:                        tt.fields.os,
				ioutil:                    tt.fields.ioutil,
				limiter:                   tt.fields.limiter,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
			}
			tt.stateSetter(tt.fields)
			got, err := fc.GetInitiatorPorts(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInitiatorPorts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInitiatorPorts() got = %v, want %v", got, tt.want)
			}
		})
	}
}
