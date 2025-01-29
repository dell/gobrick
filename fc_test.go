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
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	intpowerpath "github.com/dell/gobrick/internal/powerpath"

	"github.com/dell/gobrick/internal/mockhelper"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/golang/mock/gomock"
	"golang.org/x/sync/semaphore"
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
	powerpath                 *intpowerpath.MockPowerpath
	scsi                      *intscsi.MockSCSI
	filePath                  *wrp.MockLimitedFilepath
	os                        *wrp.MockLimitedOS
	limiter                   *semaphore.Weighted
	waitDeviceRegisterTimeout time.Duration
}

func getDefaultFCFields(ctrl *gomock.Controller) fcFields {
	con := NewFCConnector(FCConnectorParams{})
	bc := con.baseConnector
	mpMock := intmultipath.NewMockMultipath(ctrl)
	ppMock := intpowerpath.NewMockPowerpath(ctrl)
	scsiMock := intscsi.NewMockSCSI(ctrl)
	bc.multipath = mpMock
	bc.scsi = scsiMock
	f := fcFields{
		baseConnector:             bc,
		multipath:                 mpMock,
		powerpath:                 ppMock,
		scsi:                      scsiMock,
		filePath:                  wrp.NewMockLimitedFilepath(ctrl),
		os:                        wrp.NewMockLimitedOS(ctrl),
		limiter:                   con.limiter,
		waitDeviceRegisterTimeout: con.waitDeviceRegisterTimeout,
	}
	return f
}

func getFCHBASInfoMock(mock *baseMockHelper,
	os *wrp.MockLimitedOS, filepath *wrp.MockLimitedFilepath,
) {
	isFCSupportedMock(mock, os)
	sysPath := "/sys/class/fc_host/host"
	sysPathGlob := sysPath + "*"
	host1Path := sysPath + validSCSIHost1
	host2Path := sysPath + validSCSIHost2
	mock.FilePathGlobCallPattern = sysPathGlob
	mock.FilePathGlobOKReturn = []string{host1Path, host2Path}
	mock.FilePathGlobOK(filepath)

	mock.OSReadFileCallPath = host1Path + "/port_name"
	mock.OSReadFileOKReturn = "0x" + validWWPN1
	mock.OSReadFileOK(os)
	mock.OSReadFileCallPath = host1Path + "/node_name"
	mock.OSReadFileOKReturn = "0x" + validNodeName1
	mock.OSReadFileOK(os)

	mock.OSReadFileCallPath = host2Path + "/port_name"
	mock.OSReadFileOKReturn = "0x" + validWWPN2
	mock.OSReadFileOK(os)
	mock.OSReadFileCallPath = host2Path + "/node_name"
	mock.OSReadFileOKReturn = "0x" + validNodeName2
	mock.OSReadFileOK(os)
}

func isFCSupportedMock(mock *baseMockHelper, os *wrp.MockLimitedOS) {
	mock.OSStatCallPath = "/sys/class/fc_host"
	_, statMock := mock.OSStatCallOK(os)
	mock.OSStatFileInfoIsDirOKReturn = true
	mock.OSStatFileInfoIsDirOK(statMock)
}

func waitForDeviceWWNMock(mock *baseMockHelper,
	filepath *wrp.MockLimitedFilepath,
	os *wrp.MockLimitedOS,
	scsi *intscsi.MockSCSI,
) {
	findHCTLsForFCHBAMock(mock, filepath, os)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1
	mock.SCSIGetDeviceNameByHCTLErr(scsi)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL2
	mock.SCSIGetDeviceNameByHCTLErr(scsi)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1Target1
	mock.SCSIGetDeviceNameByHCTLErr(scsi)

	findHCTLsForFCHBAMock(mock, filepath, os)

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
	filepath *wrp.MockLimitedFilepath, os *wrp.MockLimitedOS,
) {
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

	mock.OSReadFileCallPath = sysPathMatch1 + "/port_name"
	mock.OSReadFileOKReturn = "0x" + validWWPN1
	mock.OSReadFileOK(os)
	mock.OSReadFileCallPath = sysPathMatch1Target1 + "/port_name"
	mock.OSReadFileOKReturn = "0x" + validWWPN2
	mock.OSReadFileOK(os)

	mock.OSReadFileCallPath = sysPathMatch2 + "/port_name"
	mock.OSReadFileOKReturn = "0x" + validWWPN2
	mock.OSReadFileOK(os)
}

func cleanConnectionMock(mock *baseMockHelper,
	filepath *wrp.MockLimitedFilepath,
	os *wrp.MockLimitedOS,
	scsi *intscsi.MockSCSI,
	multipath *intmultipath.MockMultipath,
) {
	getFCHBASInfoMock(mock, os, filepath)
	findHCTLsForFCHBAMock(mock, filepath, os)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1
	mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName
	mock.SCSIGetDeviceNameByHCTLOK(scsi)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL2
	mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName2
	mock.SCSIGetDeviceNameByHCTLOK(scsi)

	mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1Target1
	mock.SCSIGetDeviceNameByHCTLErr(scsi).AnyTimes()

	if mock.MultipathIsDaemonRunningOKReturn {
		BaseConnectorCleanMultiPathDeviceMock(mock, scsi, multipath)
	} else {
		BaseConnectorCleanDeviceMock(mock, scsi)
	}
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
			name:        "at least one FC target required",
			fields:      getDefaultFCFields(ctrl),
			stateSetter: func(_ fcFields) {},
			args: args{
				ctx: ctx,
				info: FCVolumeInfo{
					Targets: []FCTargetInfo{},
					Lun:     validLunNumber,
				},
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:        "invalid target info",
			fields:      getDefaultFCFields(ctrl),
			stateSetter: func(_ fcFields) {},
			args: args{
				ctx: ctx,
				info: FCVolumeInfo{
					Targets: []FCTargetInfo{
						{
							WWPN: "",
						},
					},
					Lun: validLunNumber,
				},
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "failed to get FC hbas info AS FC is not supported for the host",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				fields.os.EXPECT().Stat(gomock.Any()).Return(nil, errors.New("FC is not supported for this host")).AnyTimes()
			},
			args:    defaultArgs,
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "ok-multipath",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				getFCHBASInfoMock(&mock, fields.os, fields.filePath)
				waitForDeviceWWNMock(&mock, fields.filePath, fields.os, fields.scsi)
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
				getFCHBASInfoMock(&mock, fields.os, fields.filePath)
				waitForDeviceWWNMock(&mock, fields.filePath, fields.os, fields.scsi)
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
				getFCHBASInfoMock(&mock, fields.os, fields.filePath)
				waitForDeviceWWNMock(&mock, fields.filePath, fields.os, fields.scsi)
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

				cleanConnectionMock(&mock, fields.filePath, fields.os, fields.scsi, fields.multipath)
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
				powerpath:                 tt.fields.powerpath,
				scsi:                      tt.fields.scsi,
				filePath:                  tt.fields.filePath,
				os:                        tt.fields.os,
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
					&mock, fields.filePath, fields.os, fields.scsi, fields.multipath)
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
					&mock, fields.scsi)
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
				getFCHBASInfoMock(&mock, fields.os, fields.filePath)
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
