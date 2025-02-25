/*
Copyright © 2020-2025 Dell Inc. or its subsidiaries. All Rights Reserved.

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

	"github.com/dell/gonvme"

	mh "github.com/dell/gobrick/internal/mockhelper"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/golang/mock/gomock"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

var (
	validNVMEPortal1     = "1.1.1.1:3260"
	validNVMETarget1     = "nqn.2014-08.org.nvmexpress:uuid:csi_master"
	validNVMEPortal2     = "1.1.1.1:3260"
	validNVMETarget2     = "nqn.2014-08.org.nvmexpress:uuid:csi_worker"
	validNVMETargetInfo1 = NVMeTargetInfo{
		Portal: validNVMEPortal1,
		Target: validNVMETarget1,
	}
	validNVMETargetInfo2 = NVMeTargetInfo{
		Portal: validNVMEPortal2,
		Target: validNVMETarget2,
	}
	validNVMEVolumeInfo = NVMeVolumeInfo{
		Targets: []NVMeTargetInfo{validNVMETargetInfo1, validNVMETargetInfo2},
		WWN:     validNQN,
	}

	validLibNVMETarget1 = gonvme.NVMeTarget{
		TargetNqn: validNVMETarget1,
		Portal:    validNVMEPortal1,
	}

	validLibNVMETarget2 = gonvme.NVMeTarget{
		TargetNqn: validNVMETarget2,
		Portal:    validNVMEPortal2,
	}

	validNVMEInitiatorName = "nqn.2014-08.org.nvmexpress:uuid:csi_worker:e16da41ba075"

	validLibNVMESession1 = gonvme.NVMESession{
		Target:            validNVMETarget1,
		Portal:            validNVMEPortal1,
		Name:              "nvme1",
		NVMESessionState:  "live",
		NVMETransportName: "tcp",
	}
	validLibNVMESession2 = gonvme.NVMESession{
		Target:            validNVMETarget2,
		Portal:            validNVMEPortal2,
		Name:              "nvme2",
		NVMESessionState:  "live",
		NVMETransportName: "tcp",
	}
	validLibNVMESessions = []gonvme.NVMESession{validLibNVMESession1, validLibNVMESession2}

	validDevicePathsAndNamespacesWithTwoDevices = []gonvme.DevicePathAndNamespace{
		{DevicePath: "/dev/nvme0n1", Namespace: "ns1"},
		{DevicePath: "/dev/nvme1n1", Namespace: "ns2"},
	}
)

type NVMEFields struct {
	baseConnector                          *baseConnector
	multipath                              *intmultipath.MockMultipath
	scsi                                   *intscsi.MockSCSI
	nvmeLib                                *wrp.MockNVMe
	filePath                               *wrp.MockLimitedFilepath
	os                                     *wrp.MockLimitedOS
	manualSessionManagement                bool
	waitDeviceTimeout                      time.Duration
	waitDeviceRegisterTimeout              time.Duration
	failedSessionMinimumLoginRetryInterval time.Duration
	loginLock                              *rateLock
	limiter                                *semaphore.Weighted
	singleCall                             *singleflight.Group
}

func getDefaultNVMEFields(ctrl *gomock.Controller) NVMEFields {
	con := NewNVMeConnector(NVMeConnectorParams{MultipathFlushTimeout: 1})
	bc := con.baseConnector
	mpMock := intmultipath.NewMockMultipath(ctrl)
	scsiMock := intscsi.NewMockSCSI(ctrl)
	bc.multipath = mpMock
	bc.scsi = scsiMock
	return NVMEFields{
		baseConnector:                          bc,
		multipath:                              mpMock,
		scsi:                                   scsiMock,
		nvmeLib:                                wrp.NewMockNVMe(ctrl),
		filePath:                               wrp.NewMockLimitedFilepath(ctrl),
		manualSessionManagement:                con.manualSessionManagement,
		waitDeviceTimeout:                      con.waitDeviceTimeout,
		waitDeviceRegisterTimeout:              con.waitDeviceRegisterTimeout,
		failedSessionMinimumLoginRetryInterval: con.waitDeviceTimeout,
		loginLock:                              con.loginLock,
		limiter:                                con.limiter,
		singleCall:                             con.singleCall,
		os:                                     wrp.NewMockLimitedOS(ctrl),
	}
}

func TestNVME_Connector_ConnectVolume(t *testing.T) {
	type args struct {
		ctx   context.Context
		info  NVMeVolumeInfo
		useFc bool
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		fields      NVMEFields
		args        args
		stateSetter func(fields NVMEFields)
		want        Device
		wantErr     bool
	}{
		{
			name:        "Without any targets",
			fields:      getDefaultNVMEFields(ctrl),
			stateSetter: func(_ NVMEFields) {},
			args: args{
				ctx:   ctx,
				info:  NVMeVolumeInfo{},
				useFc: false,
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:        "Incorrect targets",
			fields:      getDefaultNVMEFields(ctrl),
			stateSetter: func(_ NVMEFields) {},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "", Target: ""},
					},
				},
				useFc: false,
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:        "Invalid volume wwn",
			fields:      getDefaultNVMEFields(ctrl),
			stateSetter: func(_ NVMEFields) {},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "test-portal", Target: "test-target"},
					},
					WWN: "",
				},
				useFc: false,
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "unable to get nvme sessions",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.multipath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.nvmeLib.EXPECT().GetSessions().Return(nil, errors.New("unable to get nvme sessions")).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "test-portal", Target: "test-target"},
					},
					WWN: validNQN,
				},
				useFc: false,
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "nvme session not found",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.multipath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.nvmeLib.EXPECT().GetSessions().Return(validLibNVMESessions, nil).AnyTimes()
				fields.nvmeLib.EXPECT().ListNVMeDeviceAndNamespace().Return(validDevicePathsAndNamespacesWithTwoDevices, nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetNVMeDeviceData(gomock.Any()).Return("0f8da909812540628ccf09680039914f", "ns1", nil).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("naa.68ccf098000f8da9098125406239914f", nil).AnyTimes()
				fields.scsi.EXPECT().WaitUdevSymlinkNVMe(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().CheckDeviceIsValid(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "test-portal", Target: "test-target"},
					},
					WWN: "naa.68ccf098000f8da9098125406239914f",
				},
				useFc: false,
			},
			want: Device{
				WWN:         "naa.68ccf098000f8da9098125406239914f",
				Name:        "naa.68ccf098000f8da9098125406239914f",
				MultipathID: "naa.68ccf098000f8da9098125406239914f",
			},
			wantErr: false,
		},
		{
			name:   "nvme session found",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.multipath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.nvmeLib.EXPECT().GetSessions().Return(validLibNVMESessions, nil).AnyTimes()
				fields.nvmeLib.EXPECT().ListNVMeDeviceAndNamespace().Return(validDevicePathsAndNamespacesWithTwoDevices, nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetNVMeDeviceData(gomock.Any()).Return("0f8da909812540628ccf09680039914f", "ns1", nil).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("naa.68ccf098000f8da9098125406239914f", nil).AnyTimes()
				fields.scsi.EXPECT().WaitUdevSymlinkNVMe(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().CheckDeviceIsValid(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{validNVMETargetInfo1},
					WWN:     "naa.68ccf098000f8da9098125406239914f",
				},
				useFc: false,
			},
			want: Device{
				WWN:         "naa.68ccf098000f8da9098125406239914f",
				Name:        "naa.68ccf098000f8da9098125406239914f",
				MultipathID: "naa.68ccf098000f8da9098125406239914f",
			},
			wantErr: false,
		},
		{
			name:   "nvme session found AND Couldn't find the nvme namespaces",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.multipath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.nvmeLib.EXPECT().GetSessions().Return(validLibNVMESessions, nil).AnyTimes()
				fields.nvmeLib.EXPECT().ListNVMeDeviceAndNamespace().Return(validDevicePathsAndNamespacesWithTwoDevices, nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetNVMeDeviceData(gomock.Any()).Return("0f8da909812540628ccf09680039914f", "ns1", nil).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("naa.68ccf098000f8da9098125406239914f", errors.New("Couldn't find the nvme namespaces")).AnyTimes()
				fields.scsi.EXPECT().WaitUdevSymlinkNVMe(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().CheckDeviceIsValid(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("multipath device not found")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(errors.New("can not delete block device")).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{validNVMETargetInfo1},
					WWN:     "naa.68ccf098000f8da9098125406239914f",
				},
				useFc: false,
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "nvme session found and Device is not valid",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.multipath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.nvmeLib.EXPECT().GetSessions().Return(validLibNVMESessions, nil).AnyTimes()
				fields.nvmeLib.EXPECT().ListNVMeDeviceAndNamespace().Return(validDevicePathsAndNamespacesWithTwoDevices, nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetNVMeDeviceData(gomock.Any()).Return("0f8da909812540628ccf09680039914f", "ns1", nil).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("naa.68ccf098000f8da9098125406239914f", nil).AnyTimes()
				fields.scsi.EXPECT().WaitUdevSymlinkNVMe(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().CheckDeviceIsValid(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("multipath device not found")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(errors.New("can not delete block device")).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{validNVMETargetInfo1},
					WWN:     "naa.68ccf098000f8da9098125406239914f",
				},
				useFc: false,
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "multipath IsDaemonRunning false and Failed to connect volume",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.multipath.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				fields.nvmeLib.EXPECT().ListNVMeDeviceAndNamespace().Return(validDevicePathsAndNamespacesWithTwoDevices, nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetNVMeDeviceData(gomock.Any()).Return("0f8da909812540628ccf09680039914f", "ns1", nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetSessions().Return(validLibNVMESessions, nil).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "test-portal", Target: "test-target"},
					},
					WWN: validNQN,
				},
				useFc: false,
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "multipath IsDaemonRunning false and able to connect volume",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.multipath.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				fields.nvmeLib.EXPECT().ListNVMeDeviceAndNamespace().Return(validDevicePathsAndNamespacesWithTwoDevices, nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetNVMeDeviceData(gomock.Any()).Return("0f8da909812540628ccf09680039914f", "ns1", nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetSessions().Return(validLibNVMESessions, nil).AnyTimes()
				fields.scsi.EXPECT().CheckDeviceIsValid(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{validNVMETargetInfo1},
					WWN:     "naa.68ccf098000f8da9098125406239914f",
				},
				useFc: false,
			},
			want: Device{
				Name: "nvme0n1",
				WWN:  "naa.68ccf098000f8da9098125406239914f",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &NVMeConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				nvmeLib:                   tt.fields.nvmeLib,
				manualSessionManagement:   tt.fields.manualSessionManagement,
				waitDeviceTimeout:         tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
				loginLock:                 tt.fields.loginLock,
				limiter:                   tt.fields.limiter,
				singleCall:                tt.fields.singleCall,
				filePath:                  tt.fields.filePath,
			}

			tt.stateSetter(tt.fields)
			got, err := c.ConnectVolume(tt.args.ctx, tt.args.info, tt.args.useFc)
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

func TestNVME_Connector_DisconnectVolume(t *testing.T) {
	type args struct {
		ctx  context.Context
		info NVMeVolumeInfo
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		fields      NVMEFields
		args        args
		stateSetter func(fields NVMEFields)
		wantErr     bool
	}{
		{
			name:   "request",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.nvmeLib.EXPECT().ListNVMeDeviceAndNamespace().Return(validDevicePathsAndNamespacesWithTwoDevices, nil).AnyTimes()
				fields.nvmeLib.EXPECT().GetNVMeDeviceData(gomock.Any()).Return("0f8da909812540628ccf09680039914f", "ns1", nil).AnyTimes()
			},
			args:    args{ctx: ctx, info: NVMeVolumeInfo{}},
			wantErr: false,
		},
		{
			name:   "Error in finding the nvme namespaces",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.nvmeLib.EXPECT().ListNVMeDeviceAndNamespace().Return([]gonvme.DevicePathAndNamespace{}, errors.New("Couldn't find the nvme namespaces")).AnyTimes()
			},
			args:    args{ctx: ctx, info: NVMeVolumeInfo{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &NVMeConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				nvmeLib:                   tt.fields.nvmeLib,
				manualSessionManagement:   tt.fields.manualSessionManagement,
				waitDeviceTimeout:         tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
				loginLock:                 tt.fields.loginLock,
				limiter:                   tt.fields.limiter,
				singleCall:                tt.fields.singleCall,
				filePath:                  tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			err := c.DisconnectVolume(tt.args.ctx, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("DisconnectVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNVME_Connector_DisconnectVolumeByDeviceName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		fields      NVMEFields
		args        args
		stateSetter func(fields NVMEFields)
		wantErr     bool
	}{
		{
			name:   "empty request",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
			},
			args:    args{ctx: ctx, name: ""},
			wantErr: false,
		},
		{
			name:   "Disconnect with device mapper name",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{mh.ValidDeviceName}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			args:    args{ctx: ctx, name: mh.ValidDMName},
			wantErr: false,
		},
		{
			name:   "Disconnect with device mapper name AND failed to get children for DM",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("failed to get children for DM")).AnyTimes()
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{mh.ValidDeviceName}, nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			args:    args{ctx: ctx, name: mh.ValidDMName},
			wantErr: false,
		},
		{
			name:   "Disconnect with device name AND can't find wwn for device",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), gomock.Any()).Return("", errors.New("can't find wwn for device")).AnyTimes()
			},
			args:    args{ctx: ctx, name: mh.ValidDeviceName},
			wantErr: true,
		},
		{
			name:   "Disconnect with device name AND failed to find devices by wwn",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("failed to find devices by wwn")).AnyTimes()
			},
			args:    args{ctx: ctx, name: mh.ValidDeviceName},
			wantErr: true,
		},
		{
			name:   "Disconnect with device name",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{mh.ValidDeviceName}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			args:    args{ctx: ctx, name: mh.ValidDeviceName},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &NVMeConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				nvmeLib:                   tt.fields.nvmeLib,
				manualSessionManagement:   tt.fields.manualSessionManagement,
				waitDeviceTimeout:         tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
				loginLock:                 tt.fields.loginLock,
				limiter:                   tt.fields.limiter,
				singleCall:                tt.fields.singleCall,
				filePath:                  tt.fields.filePath,
			}

			tt.stateSetter(tt.fields)
			err := c.DisconnectVolumeByDeviceName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("DisconnectVolumeByDeviceName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNVME_Connector_GetInitiatorName(t *testing.T) {
	type args struct {
		ctx  context.Context
		info NVMeVolumeInfo
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		fields      NVMEFields
		args        args
		stateSetter func(fields NVMEFields)
		want        []string
		wantErr     bool
	}{
		{
			name:   "request",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.nvmeLib.EXPECT().GetInitiators(gomock.Any()).Return([]string{"nqn.1988-11.com.dell.mock:01:0000000000000"}, nil)
			},
			args:    args{ctx: ctx, info: NVMeVolumeInfo{}},
			want:    []string{"nqn.1988-11.com.dell.mock:01:0000000000000"},
			wantErr: false,
		},
		{
			name:   "failed to read initiator name",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.nvmeLib.EXPECT().GetInitiators(gomock.Any()).Return([]string{"nqn.1988-11.com.dell.mock:01:0000000000000"}, errors.New("failed to read initiator name"))
			},
			args:    args{ctx: ctx, info: NVMeVolumeInfo{}},
			want:    []string{"nqn.1988-11.com.dell.mock:01:0000000000000"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &NVMeConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				nvmeLib:                   tt.fields.nvmeLib,
				manualSessionManagement:   tt.fields.manualSessionManagement,
				waitDeviceTimeout:         tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
				loginLock:                 tt.fields.loginLock,
				limiter:                   tt.fields.limiter,
				singleCall:                tt.fields.singleCall,
				filePath:                  tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			got, err := c.GetInitiatorName(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInitiatorName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInitiatorName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNVME_wwnMatches(t *testing.T) {
	tests := []struct {
		nguid string
		wwn   string
		want  bool
	}{
		{nguid: "0f8da909812540628ccf09680039914f", wwn: "naa.68ccf098000f8da9098125406239914f", want: true},
		{nguid: "0f8da909812540628ccf09680039914f", wwn: "NAA.68ccf098000f8da9098125406239914f", want: true},
		{nguid: "0f8da909812540628ccf09680039914f", wwn: "68ccf098000f8da9098125406239914f", want: true},
		{nguid: "0f8da909812540628ccf09680039914f", wwn: "68CCF098000F8DA9098125406239914F", want: true},
		{nguid: "0f8da909812540628ccf09680039914f", wwn: "60000978000f8da9098125406239914f", want: false},
		{nguid: "12635330303134340000976000012000", wwn: "60000970000120001263533030313434", want: true},
		{nguid: "12635330303134340000976000012000", wwn: "68ccf070000120001263533030313434", want: false},
		{nguid: "12635330303134340000976000012000", wwn: "8ccf070000120001263533030313434", want: false},
		{nguid: "12635330303134340000976000012000", wwn: "68CCF070000120001263533030313434", want: false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("(%s,%s) should be %v", tt.nguid, tt.wwn, tt.want), func(t *testing.T) {
			c := &NVMeConnector{}
			if c.wwnMatches(tt.nguid, tt.wwn) != tt.want {
				t.Errorf("wwnMatches(%v, %v) = %v, want %v", tt.nguid, tt.wwn, !tt.want, tt.want)
			}
		})
	}
}

func TestNVME_Connector_tryNVMeConnect(t *testing.T) {
	type args struct {
		ctx   context.Context
		info  NVMeVolumeInfo
		useFC bool
	}

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		fields      NVMEFields
		stateSetter func(fields NVMEFields)
		args        args
		wantErr     bool
	}{
		{
			name:        "Without FC",
			fields:      getDefaultNVMEFields(ctrl),
			stateSetter: func(_ NVMEFields) {},
			args: args{
				ctx:   ctx,
				info:  NVMeVolumeInfo{},
				useFC: false,
			},
			wantErr: false,
		},
		{
			name:   "connection with FC - Error in reading file",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.filePath.EXPECT().Glob(gomock.Any()).Return([]string{}, errors.New("error in matching files with pattern")).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "192.168.0.1", Target: "nqn-1"},
					},
					WWN: "",
				},
				useFC: true,
			},
			wantErr: true,
		},
		{
			name:   "connection with FC - No fc host matches",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.filePath.EXPECT().Glob(gomock.Any()).Return([]string{}, nil).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "192.168.0.1", Target: "nqn-1"},
					},
					WWN: "",
				},
				useFC: true,
			},
			wantErr: false,
		},
		{
			name:   "connection with FC - multiple FCHostsInfo",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.filePath.EXPECT().Glob(gomock.Any()).Return([]string{"host"}, nil).AnyTimes()
				fields.os.EXPECT().ReadFile(gomock.Any()).Return([]byte("testName"), nil).AnyTimes()
				fields.nvmeLib.EXPECT().NVMeFCConnect(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "192.168.0.1", Target: "nqn-1"},
					},
					WWN: "",
				},
				useFC: true,
			},
			wantErr: false,
		},
		{
			name:   "connection with FC - multiple FCHostsInfo AND Couldn't connect to NVMeFC target",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.filePath.EXPECT().Glob(gomock.Any()).Return([]string{"host"}, nil).AnyTimes()
				fields.os.EXPECT().ReadFile(gomock.Any()).Return([]byte("testName"), nil).AnyTimes()
				fields.nvmeLib.EXPECT().NVMeFCConnect(gomock.Any(), gomock.Any()).Return(errors.New("Couldn't connect to NVMeFC target")).AnyTimes()
			},
			args: args{
				ctx: ctx,
				info: NVMeVolumeInfo{
					Targets: []NVMeTargetInfo{
						{Portal: "192.168.0.1", Target: "nqn-1"},
					},
					WWN: "",
				},
				useFC: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &NVMeConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				nvmeLib:                   tt.fields.nvmeLib,
				manualSessionManagement:   tt.fields.manualSessionManagement,
				waitDeviceTimeout:         tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
				loginLock:                 tt.fields.loginLock,
				limiter:                   tt.fields.limiter,
				singleCall:                tt.fields.singleCall,
				filePath:                  tt.fields.filePath,
				os:                        tt.fields.os,
			}

			tt.stateSetter(tt.fields)
			err := c.tryNVMeConnect(tt.args.ctx, tt.args.info, tt.args.useFC)
			if (err != nil) != tt.wantErr {
				t.Errorf("tryNVMeConnect() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNVME_readNVMeDevicesFromResultCH(t *testing.T) {
	tests := []struct {
		name             string
		devicePathResult DevicePathResult
		expectedPaths    []string
		expectedNguid    string
	}{
		{
			name: "Single device path",
			devicePathResult: DevicePathResult{
				devicePaths: []string{"/dev/nvme0n1"},
				nguid:       "test-nguid-1",
			},
			expectedPaths: []string{"nvme0n1"},
			expectedNguid: "test-nguid-1",
		},
		{
			name: "Multiple device paths",
			devicePathResult: DevicePathResult{
				devicePaths: []string{"/dev/nvme0n1", "/dev/nvme1n1"},
				nguid:       "test-nguid-2",
			},
			expectedPaths: []string{"nvme0n1", "nvme1n1"},
			expectedNguid: "test-nguid-2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := make(chan DevicePathResult, 1)
			ch <- tt.devicePathResult
			gotPaths, gotNguid := readNVMeDevicesFromResultCH(ch, nil)
			if !reflect.DeepEqual(gotPaths, tt.expectedPaths) {
				t.Errorf("readNVMeDevicesFromResultCH() gotPaths = %v, expectedPaths %v", gotPaths, tt.expectedPaths)
			}
			if gotNguid != tt.expectedNguid {
				t.Errorf("readNVMeDevicesFromResultCH() gotNguid = %v, expectedNguid %v", gotNguid, tt.expectedNguid)
			}
		})
	}
}

func TestNVME_Connector_getFCHostInfo(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		fields      NVMEFields
		args        args
		stateSetter func(fields NVMEFields)
		want        []FCHBAInfo
		wantErr     bool
	}{
		{
			name:   "Successfully retrieve FC host info",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.filePath.EXPECT().Glob("/sys/class/fc_host/host*").Return([]string{
					"/sys/class/fc_host/host0",
				}, nil)
				fields.os.EXPECT().ReadFile("/sys/class/fc_host/host0/port_name").Return([]byte("tcp"), nil)
				fields.os.EXPECT().ReadFile("/sys/class/fc_host/host0/node_name").Return([]byte("nqn-1"), nil)
			},
			args: args{ctx: ctx},
			want: []FCHBAInfo{
				{PortName: "tcp", NodeName: "nqn-1"},
			},
			wantErr: false,
		},
		{
			name:   "Error globbing FC hosts",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.filePath.EXPECT().Glob("/sys/class/fc_host/host*").Return(nil, errors.New("glob error"))
			},
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Error reading port_name",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.filePath.EXPECT().Glob("/sys/class/fc_host/host*").Return([]string{
					"/sys/class/fc_host/host0",
				}, nil)
				fields.os.EXPECT().ReadFile("/sys/class/fc_host/host0/port_name").Return(nil, errors.New("read port error"))
			},
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: false,
		},
		{
			name:   "Error reading node_name",
			fields: getDefaultNVMEFields(ctrl),
			stateSetter: func(fields NVMEFields) {
				fields.filePath.EXPECT().Glob("/sys/class/fc_host/host*").Return([]string{
					"/sys/class/fc_host/host0",
				}, nil)
				fields.os.EXPECT().ReadFile("/sys/class/fc_host/host0/port_name").Return([]byte("tcp"), nil)
				fields.os.EXPECT().ReadFile("/sys/class/fc_host/host0/node_name").Return(nil, errors.New("read node error"))
			},
			args:    args{ctx: ctx},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &NVMeConnector{
				baseConnector:             tt.fields.baseConnector,
				multipath:                 tt.fields.multipath,
				scsi:                      tt.fields.scsi,
				nvmeLib:                   tt.fields.nvmeLib,
				filePath:                  tt.fields.filePath,
				os:                        tt.fields.os,
				manualSessionManagement:   tt.fields.manualSessionManagement,
				waitDeviceTimeout:         tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout: tt.fields.waitDeviceRegisterTimeout,
				loginLock:                 tt.fields.loginLock,
				limiter:                   tt.fields.limiter,
				singleCall:                tt.fields.singleCall,
			}
			tt.stateSetter(tt.fields)
			got, err := c.getFCHostInfo(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFCHostInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFCHostInfo() got = %v, want %v", got, tt.want)
			}
		})
	}
}
