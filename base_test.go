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
	"testing"
	"time"

	mh "github.com/dell/gobrick/internal/mockhelper"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intpowerpath "github.com/dell/gobrick/internal/powerpath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	"github.com/dell/gobrick/pkg/scsi"
	"github.com/golang/mock/gomock"
)

type baseMockHelper struct {
	Ctx                                  interface{}
	MultipathAddWWIDCallWWID             string
	MultipathAddPathCallPath             string
	MultipathDelPathCallPath             string
	MultipathFlushDeviceCallMapName      string
	MultipathGetDMWWIDCallMapName        string
	MultipathIsDaemonRunningOKReturn     bool
	MultipathGetDMWWIDOKReturn           string
	SCSIIsDeviceExistCallDevice          string
	SCSIRescanSCSIHostByHCTLCallH        scsi.HCTL
	SCSIRescanSCSIDeviceByHCTLCallH      scsi.HCTL
	SCSIDeleteSCSIDeviceByHCTLCallH      scsi.HCTL
	SCSIDeleteSCSIDeviceByNameCallName   string
	SCSIDeleteSCSIDeviceByPathCallPath   string
	SCSIGetDeviceWWNCallDevices          []string
	SCSIGetDevicesByWWNCallWWN           string
	SCSIGetDMDeviceByChildrenCallDevices []string
	SCSIGetDMChildrenCallDmPath          string
	SCSICheckDeviceIsValidCallDevice     string
	SCSIGetDeviceNameByHCTLCallH         scsi.HCTL
	SCSIWaitUdevSymlinkCallDevice        string
	SCSIWaitUdevSymlinkCallWWN           string
	SCSICheckDeviceIsValidOKReturn       bool
	SCSIIsDeviceExistOKReturn            bool
	SCSIGetDeviceWWNOKReturn             string
	SCSIGetDevicesByWWNOKReturn          []string
	SCSIGetDMDeviceByChildrenOKReturn    string
	SCSIGetDMChildrenOKReturn            []string
	SCSIGetDeviceNameByHCTLOKReturn      string
	mh.MockHelper
}

func (bmh *baseMockHelper) MultipathAddWWIDCall(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return m.EXPECT().AddWWID(bmh.Ctx, bmh.MultipathAddWWIDCallWWID)
}

func (bmh *baseMockHelper) MultipathAddWWIDOK(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathAddWWIDCall(m).Return(nil)
}

func (bmh *baseMockHelper) MultipathAddWWIDErr(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathAddWWIDCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) MultipathAddPathCall(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return m.EXPECT().AddPath(bmh.Ctx, bmh.MultipathAddPathCallPath)
}

func (bmh *baseMockHelper) MultipathAddPathOK(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathAddPathCall(m).Return(nil)
}

func (bmh *baseMockHelper) MultipathAddPathErr(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathAddPathCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) MultipathDelPathCall(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return m.EXPECT().DelPath(bmh.Ctx, bmh.MultipathDelPathCallPath)
}

func (bmh *baseMockHelper) MultipathDelPathOK(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathDelPathCall(m).Return(nil)
}

func (bmh *baseMockHelper) MultipathDelPathErr(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathDelPathCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) MultipathFlushDeviceCall(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return m.EXPECT().FlushDevice(bmh.Ctx, bmh.MultipathFlushDeviceCallMapName)
}

func (bmh *baseMockHelper) MultipathFlushDeviceOK(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathFlushDeviceCall(m).Return(nil)
}

func (bmh *baseMockHelper) MultipathFlushDeviceErr(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathFlushDeviceCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) MultipathIsDaemonRunningCall(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return m.EXPECT().IsDaemonRunning(bmh.Ctx)
}

func (bmh *baseMockHelper) MultipathIsDaemonRunningOK(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathIsDaemonRunningCall(m).Return(bmh.MultipathIsDaemonRunningOKReturn)
}

func (bmh *baseMockHelper) MultipathGetDMWWIDCall(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return m.EXPECT().GetDMWWID(bmh.Ctx, bmh.MultipathGetDMWWIDCallMapName)
}

func (bmh *baseMockHelper) MultipathGetDMWWIDOK(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathGetDMWWIDCall(m).Return(bmh.MultipathGetDMWWIDOKReturn, nil)
}

func (bmh *baseMockHelper) MultipathGetDMWWIDErr(
	m *intmultipath.MockMultipath,
) *gomock.Call {
	return bmh.MultipathGetDMWWIDCall(m).Return("", mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIIsDeviceExistCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().IsDeviceExist(bmh.Ctx, bmh.SCSIIsDeviceExistCallDevice)
}

func (bmh *baseMockHelper) SCSIIsDeviceExistOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIIsDeviceExistCall(m).Return(bmh.SCSIIsDeviceExistOKReturn)
}

func (bmh *baseMockHelper) SCSIRescanSCSIHostByHCTLCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().RescanSCSIHostByHCTL(bmh.Ctx, bmh.SCSIRescanSCSIHostByHCTLCallH)
}

func (bmh *baseMockHelper) SCSIRescanSCSIHostByHCTLOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIRescanSCSIHostByHCTLCall(m).Return(nil)
}

func (bmh *baseMockHelper) SCSIRescanSCSIHostByHCTLErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIRescanSCSIHostByHCTLCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIRescanSCSIDeviceByHCTLCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().RescanSCSIDeviceByHCTL(bmh.Ctx, bmh.SCSIRescanSCSIDeviceByHCTLCallH)
}

func (bmh *baseMockHelper) SCSIRescanSCSIDeviceByHCTLOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIRescanSCSIDeviceByHCTLCall(m).Return(nil)
}

func (bmh *baseMockHelper) SCSIRescanSCSIDeviceByHCTLErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIRescanSCSIDeviceByHCTLCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByHCTLCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().DeleteSCSIDeviceByHCTL(bmh.Ctx, bmh.SCSIDeleteSCSIDeviceByHCTLCallH)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByHCTLOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIDeleteSCSIDeviceByHCTLCall(m).Return(nil)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByHCTLErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIDeleteSCSIDeviceByHCTLCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByNameCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().DeleteSCSIDeviceByName(bmh.Ctx, bmh.SCSIDeleteSCSIDeviceByNameCallName)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByNameOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIDeleteSCSIDeviceByNameCall(m).Return(nil)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByNameErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIDeleteSCSIDeviceByNameCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByPathCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().DeleteSCSIDeviceByPath(bmh.Ctx, bmh.SCSIDeleteSCSIDeviceByPathCallPath)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByPathOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIDeleteSCSIDeviceByPathCall(m).Return(nil)
}

func (bmh *baseMockHelper) SCSIDeleteSCSIDeviceByPathErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIDeleteSCSIDeviceByPathCall(m).Return(mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIGetDeviceWWNCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().GetDeviceWWN(bmh.Ctx, bmh.SCSIGetDeviceWWNCallDevices)
}

func (bmh *baseMockHelper) SCSIGetDeviceWWNOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDeviceWWNCall(m).Return(bmh.SCSIGetDeviceWWNOKReturn, nil)
}

func (bmh *baseMockHelper) SCSIGetDeviceWWNErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDeviceWWNCall(m).Return("", mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIGetDevicesByWWNCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().GetDevicesByWWN(bmh.Ctx, bmh.SCSIGetDevicesByWWNCallWWN)
}

func (bmh *baseMockHelper) SCSIGetDevicesByWWNOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDevicesByWWNCall(m).Return(bmh.SCSIGetDevicesByWWNOKReturn, nil)
}

func (bmh *baseMockHelper) SCSIGetDevicesByWWNErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDevicesByWWNCall(m).Return(nil, mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIGetDMDeviceByChildrenCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().GetDMDeviceByChildren(bmh.Ctx, bmh.SCSIGetDMDeviceByChildrenCallDevices)
}

func (bmh *baseMockHelper) SCSIGetDMDeviceByChildrenOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDMDeviceByChildrenCall(m).Return(bmh.SCSIGetDMDeviceByChildrenOKReturn, nil)
}

func (bmh *baseMockHelper) SCSIGetDMDeviceByChildrenErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDMDeviceByChildrenCall(m).Return("", mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIGetDMChildrenCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().GetDMChildren(bmh.Ctx, bmh.SCSIGetDMChildrenCallDmPath)
}

func (bmh *baseMockHelper) SCSIGetDMChildrenOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDMChildrenCall(m).Return(bmh.SCSIGetDMChildrenOKReturn, nil)
}

func (bmh *baseMockHelper) SCSIGetDMChildrenErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDMChildrenCall(m).Return(nil, mh.ErrTest)
}

func (bmh *baseMockHelper) SCSICheckDeviceIsValidCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().CheckDeviceIsValid(bmh.Ctx, bmh.SCSICheckDeviceIsValidCallDevice)
}

func (bmh *baseMockHelper) SCSICheckDeviceIsValidOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSICheckDeviceIsValidCall(m).Return(bmh.SCSICheckDeviceIsValidOKReturn)
}

func (bmh *baseMockHelper) SCSIGetDeviceNameByHCTLCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().GetDeviceNameByHCTL(bmh.Ctx, bmh.SCSIGetDeviceNameByHCTLCallH)
}

func (bmh *baseMockHelper) SCSIGetDeviceNameByHCTLOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDeviceNameByHCTLCall(m).Return(bmh.SCSIGetDeviceNameByHCTLOKReturn, nil)
}

func (bmh *baseMockHelper) SCSIGetDeviceNameByHCTLErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIGetDeviceNameByHCTLCall(m).Return("", mh.ErrTest)
}

func (bmh *baseMockHelper) SCSIWaitUdevSymlinkCall(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return m.EXPECT().WaitUdevSymlink(
		bmh.Ctx, bmh.SCSIWaitUdevSymlinkCallDevice, bmh.SCSIWaitUdevSymlinkCallWWN)
}

func (bmh *baseMockHelper) SCSIWaitUdevSymlinkOK(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIWaitUdevSymlinkCall(m).Return(nil)
}

func (bmh *baseMockHelper) SCSIWaitUdevSymlinkErr(
	m *intscsi.MockSCSI,
) *gomock.Call {
	return bmh.SCSIWaitUdevSymlinkCall(m).Return(mh.ErrTest)
}

func BaseConnectorCleanMultiPathDeviceMock(mock *baseMockHelper,
	scsi *intscsi.MockSCSI, mp *intmultipath.MockMultipath,
) {
	mock.SCSIGetDMDeviceByChildrenCallDevices = []string{
		mh.ValidDeviceName, mh.ValidDeviceName2,
	}
	mock.SCSIGetDMDeviceByChildrenOKReturn = mh.ValidDMName
	mock.SCSIGetDMDeviceByChildrenOK(scsi)

	mock.MultipathFlushDeviceCallMapName = mh.ValidDMPath
	mock.MultipathFlushDeviceOK(mp)

	mock.SCSIDeleteSCSIDeviceByNameCallName = mh.ValidDeviceName
	mock.SCSIDeleteSCSIDeviceByNameOK(scsi)
	mock.SCSIDeleteSCSIDeviceByNameCallName = mh.ValidDeviceName2
	mock.SCSIDeleteSCSIDeviceByNameOK(scsi)

	mock.MultipathDelPathCallPath = mh.ValidDevicePath
	mock.MultipathDelPathErr(mp)

	mock.MultipathDelPathCallPath = mh.ValidDevicePath2
	mock.MultipathDelPathOK(mp)
}

func BaseConnectorCleanDeviceMock(mock *baseMockHelper,
	scsi *intscsi.MockSCSI,
) {
	mock.SCSIGetDMDeviceByChildrenCallDevices = []string{
		mh.ValidDeviceName, mh.ValidDeviceName2,
	}
	mock.SCSIGetDMDeviceByChildrenErr(scsi)

	mock.SCSIDeleteSCSIDeviceByNameCallName = mh.ValidDeviceName
	mock.SCSIDeleteSCSIDeviceByNameOK(scsi)
	mock.SCSIDeleteSCSIDeviceByNameCallName = mh.ValidDeviceName2
	mock.SCSIDeleteSCSIDeviceByNameOK(scsi)
}

func BaserConnectorDisconnectDevicesByDeviceNameMock(mock *baseMockHelper,
	scsi *intscsi.MockSCSI,
) {
	mock.SCSIIsDeviceExistCallDevice = mh.ValidDMName
	mock.SCSIIsDeviceExistOKReturn = true
	mock.SCSIIsDeviceExistOK(scsi)

	mock.SCSIGetDMChildrenCallDmPath = mh.ValidDMName
	mock.SCSIGetDMChildrenOKReturn = mh.ValidDevices
	mock.SCSIGetDMChildrenOK(scsi)

	mock.SCSIGetDeviceWWNCallDevices = mh.ValidDevices
	mock.SCSIGetDeviceWWNOKReturn = mh.ValidWWID
	mock.SCSIGetDeviceWWNOK(scsi)

	mock.SCSIGetDevicesByWWNCallWWN = mh.ValidWWID
	mock.SCSIGetDevicesByWWNOKReturn = mh.ValidDevices
	mock.SCSIGetDevicesByWWNOK(scsi)

	BaseConnectorCleanDeviceMock(mock, scsi)
}

type BaseConnectorFields struct {
	multipath *intmultipath.MockMultipath
	powerpath *intpowerpath.MockPowerpath
	scsi      *intscsi.MockSCSI
}

func getTestBaseConnector(ctrl *gomock.Controller) BaseConnectorFields {
	scsi := intscsi.NewMockSCSI(ctrl)
	mp := intmultipath.NewMockMultipath(ctrl)
	pp := intpowerpath.NewMockPowerpath(ctrl)
	return BaseConnectorFields{
		multipath: mp,
		powerpath: pp,
		scsi:      scsi,
	}
}

func TestNewBaseConnector(t *testing.T) {
	mp := &intmultipath.MockMultipath{}
	pp := &intpowerpath.MockPowerpath{}
	s := &intscsi.MockSCSI{}

	tests := []struct {
		name                 string
		params               baseConnectorParams
		expectedFlushRetries int
		expectedFlushTimeout time.Duration
		expectedRetryTimeout time.Duration
	}{
		{
			name:                 "default values",
			params:               baseConnectorParams{},
			expectedFlushRetries: multipathFlushRetriesDefault,
			expectedFlushTimeout: multipathFlushTimeoutDefault,
			expectedRetryTimeout: multipathFlushRetryTimeoutDefault,
		},
		{
			name: "custom values",
			params: baseConnectorParams{
				MultipathFlushRetries:      20,
				MultipathFlushTimeout:      time.Second * 10,
				MultipathFlushRetryTimeout: time.Second * 3,
			},
			expectedFlushRetries: 20,
			expectedFlushTimeout: time.Second * 10,
			expectedRetryTimeout: time.Second * 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := newBaseConnector(mp, pp, s, tt.params)

			if conn.multipathFlushRetries != tt.expectedFlushRetries {
				t.Errorf("expected multipathFlushRetries to be %d, got %d", tt.expectedFlushRetries, conn.multipathFlushRetries)
			}

			if conn.multipathFlushTimeout != tt.expectedFlushTimeout {
				t.Errorf("expected multipathFlushTimeout to be %v, got %v", tt.expectedFlushTimeout, conn.multipathFlushTimeout)
			}

			if conn.multipathFlushRetryTimeout != tt.expectedRetryTimeout {
				t.Errorf("expected multipathFlushRetryTimeout to be %v, got %v", tt.expectedRetryTimeout, conn.multipathFlushRetryTimeout)
			}
		})
	}
}

func TestDisconnectDevicesByWWN(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mp := intmultipath.NewMockMultipath(ctrl)
	pp := intpowerpath.NewMockPowerpath(ctrl)
	s := intscsi.NewMockSCSI(ctrl)

	type args struct {
		ctx context.Context
		wwn string
	}

	tests := []struct {
		name          string
		args          args
		setupMocks    func()
		expectedError string
	}{
		{
			name: "success",
			args: args{
				ctx: context.Background(),
				wwn: "1234567890",
			},
			setupMocks: func() {
				mp.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil)
				s.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true)
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil)
			},
			expectedError: "",
		},
		{
			name: "error flushing device",
			args: args{
				ctx: context.Background(),
				wwn: "1234567890",
			},
			setupMocks: func() {
				mp.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error"))
			},
			expectedError: "flush error",
		},
		{
			name: "error deleting device",
			args: args{
				ctx: context.Background(),
				wwn: "1234567890",
			},
			setupMocks: func() {
				mp.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil)
				s.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true)
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(errors.New("delete error"))
			},
			expectedError: "delete error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bc := &baseConnector{
				multipath: mp,
				powerpath: pp,
				scsi:      s,
			}

			tt.setupMocks()

			err := bc.disconnectDevicesByWWN(tt.args.ctx, tt.args.wwn)

			if err != nil && tt.expectedError == "" {
				t.Errorf("expected no error, got %v", err)
			} else if err == nil && tt.expectedError != "" {
				t.Errorf("expected error %v, got nil", tt.expectedError)
			} else if err != nil && tt.expectedError != "" && err.Error() != tt.expectedError {
				t.Errorf("expected error %v, got %v", tt.expectedError, err)
			}
		})
	}
}

func TestDisconnectDevicesByDeviceName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx        context.Context
		DeviceName string
	}

	tests := []struct {
		name        string
		args        args
		fields      BaseConnectorFields
		stateSetter func(fields BaseConnectorFields)
		expectedErr bool
	}{
		{
			name: "Device not found",
			args: args{
				ctx:        context.Background(),
				DeviceName: "non-existent-device",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name: "Device has device mapper prefix",
			args: args{
				ctx:        context.Background(),
				DeviceName: deviceMapperPrefix + "test-device",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true)
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{}, nil)
				fields.scsi.EXPECT().GetDeviceWWN(gomock.Any(), gomock.Any()).Return("", errors.New("failed to read WWN for DM"))
			},
			expectedErr: true,
		},
		{
			name: "Device does not have device mapper prefix",
			args: args{
				ctx:        context.Background(),
				DeviceName: "test-device",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetDeviceWWN(gomock.Any(), gomock.Any()).Return("test-wwn", nil).AnyTimes()
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("failed to find devices by wwn"))
			},
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bc := &baseConnector{
				multipath: test.fields.multipath,
				powerpath: test.fields.powerpath,
				scsi:      test.fields.scsi,
			}

			test.stateSetter(test.fields)

			err := bc.disconnectDevicesByDeviceName(test.args.ctx, test.args.DeviceName)

			if (err != nil) != test.expectedErr {
				t.Errorf("disconnectDevicesByDeviceName() error = %v, wantErr %v", err, test.expectedErr)
				return
			}
		})
	}
}

func TestCleanNVMeDevices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx     context.Context
		Force   bool
		Devices []string
		WWN     string
	}

	tests := []struct {
		name        string
		args        args
		fields      BaseConnectorFields
		stateSetter func(fields BaseConnectorFields)
		expectedErr bool
	}{
		{
			name: "Flush multipath device",
			args: args{
				ctx:     context.Background(),
				Force:   false,
				Devices: []string{},
				WWN:     "",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", nil)
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("failed to flush multipath device"))
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false)
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(errors.New("failed to remove wwid"))
			},
			expectedErr: false,
		},
		{
			name: "Failed to flush multipath device",
			args: args{
				ctx:     context.Background(),
				Force:   false,
				Devices: []string{},
				WWN:     "",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", nil)
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("failed to flush multipath device"))
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true)
			},
			expectedErr: true,
		},
		{
			name: "Flush multipath device with some devices - cant delete block device",
			args: args{
				ctx:   context.Background(),
				Force: false,
				Devices: []string{
					"/dev/sda",
					"/dev/sdb",
				},
				WWN: "",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("failed to GetDMDeviceByChildren")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(errors.New("can't delete block device")).AnyTimes()
			},
			expectedErr: true,
		},
		{
			name: "Flush multipath device with some devices",
			args: args{
				ctx:   context.Background(),
				Force: false,
				Devices: []string{
					"/dev/sda",
					"/dev/sdb",
				},
				WWN: "",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("test-name", errors.New("failed to GetDMDeviceByChildren")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bc := &baseConnector{
				multipath:             test.fields.multipath,
				powerpath:             test.fields.powerpath,
				scsi:                  test.fields.scsi,
				multipathFlushRetries: 1,
			}

			test.stateSetter(test.fields)

			err := bc.cleanNVMeDevices(test.args.ctx, test.args.Force, test.args.Devices, test.args.WWN)

			if (err != nil) != test.expectedErr {
				t.Errorf("cleanNVMeDevices() error = %v, wantErr %v", err, test.expectedErr)
				return
			}
		})
	}
}

func TestCleanDevices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx     context.Context
		Force   bool
		Devices []string
		WWN     string
	}

	tests := []struct {
		name        string
		args        args
		fields      BaseConnectorFields
		stateSetter func(fields BaseConnectorFields)
		expectedErr bool
	}{
		{
			name: "Failed to flush multipath device",
			args: args{
				ctx:     context.Background(),
				Force:   false,
				Devices: []string{},
				WWN:     "test-wwn",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
			},
			expectedErr: true,
		},
		{
			name: "Flush multipath device with some devices AND cant delete block device",
			args: args{
				ctx:     context.Background(),
				Force:   false,
				Devices: []string{},
				WWN:     "",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("failed to GetDMDeviceByChildren")).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.powerpath.EXPECT().FlushDevice(gomock.Any()).Return(errors.New("failed to flush device")).AnyTimes()
			},
			expectedErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bc := &baseConnector{
				multipath:             test.fields.multipath,
				powerpath:             test.fields.powerpath,
				scsi:                  test.fields.scsi,
				multipathFlushRetries: 0,
			}

			test.stateSetter(test.fields)

			err := bc.cleanDevices(test.args.ctx, test.args.Force, test.args.Devices, test.args.WWN)

			if (err != nil) != test.expectedErr {
				t.Errorf("cleanDevices() error = %v, wantErr %v", err, test.expectedErr)
				return
			}
		})
	}
}

func TestGetNVMEDMWWN(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type args struct {
		ctx        context.Context
		DeviceName string
	}

	tests := []struct {
		name        string
		args        args
		fields      BaseConnectorFields
		stateSetter func(fields BaseConnectorFields)
		expectedWWN string
		expectedErr bool
	}{
		{
			name: "Failed to read WWN for DM",
			args: args{
				ctx:        context.Background(),
				DeviceName: "non-existent-device",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()

				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), gomock.Any()).Return("", errors.New("failed to read WWN for DM")).AnyTimes()
			},
			expectedWWN: "",
			expectedErr: true,
		},
		{
			name: "Failed to resolve DM",
			args: args{
				ctx:        context.Background(),
				DeviceName: "non-existent-device",
			},
			fields: getTestBaseConnector(ctrl),
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("failed to get children for DM")).AnyTimes()

				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", errors.New("failed to resolve DM")).AnyTimes()
			},
			expectedWWN: "",
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bc := &baseConnector{
				multipath: test.fields.multipath,
				powerpath: test.fields.powerpath,
				scsi:      test.fields.scsi,
			}

			test.stateSetter(test.fields)

			wwn, err := bc.getNVMEDMWWN(test.args.ctx, test.args.DeviceName)

			if (err != nil) != test.expectedErr {
				t.Errorf("getNVMEDMWWN() error = %v, wantErr %v", err, test.expectedErr)
				return
			}

			if wwn != test.expectedWWN {
				t.Errorf("getNVMEDMWWN() = %v, want %v", wwn, test.expectedWWN)
			}
		})
	}
}
