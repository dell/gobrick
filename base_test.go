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
	"reflect"
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

	mock.MultipathGetDMWWIDCallMapName = mh.ValidDMName
	mock.MultipathGetDMWWIDOKReturn = mh.ValidWWID
	mock.MultipathGetDMWWIDOK(mp).AnyTimes()

	mock.MultipathFlushDeviceCallMapName = mh.ValidDMPath
	mock.MultipathFlushDeviceOK(mp).AnyTimes()

	mock.SCSIIsDeviceExistCallDevice = mh.ValidDMName
	mock.SCSIIsDeviceExistOKReturn = false
	mock.SCSIIsDeviceExistOK(scsi).AnyTimes()

	mp.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

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
	pp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
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

	bc := &baseConnector{
		multipath:             mp,
		powerpath:             pp,
		scsi:                  s,
		multipathFlushRetries: 1,
	}

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
				mp.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("mpath-name", []string{"path1", "path2"}, nil).AnyTimes()
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				mp.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mp.EXPECT().GetMpathMinorByMpathName(gomock.Any(), "mpath-name").Return("minor", false, nil).AnyTimes()
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"path1", "path2"}).Return("dm-device", nil).AnyTimes()
				s.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "Success when daemon is not running",
			args: args{
				ctx: context.Background(),
				wwn: "1234567893",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"path1", "path2"}, nil).AnyTimes()
				mp.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mp.EXPECT().GetMpathMinorByMpathName(gomock.Any(), "mpath-name").Return("minor", false, nil).AnyTimes()
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"path1", "path2"}).Return("dm-device", nil).AnyTimes()
				s.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "flushing device",
			args: args{
				ctx: context.Background(),
				wwn: "1234567891",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				mp.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("mpath-name", []string{"path1", "path2"}, nil).AnyTimes()
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"path1", "path2"}).Return("dm-device", nil).AnyTimes()
				mp.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "deleting device",
			args: args{
				ctx: context.Background(),
				wwn: "1234567892",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				mp.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("mpath-name", []string{"path1", "path2"}, nil).AnyTimes()
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"path1", "path2"}).Return("dm-device", nil).AnyTimes()
				mp.EXPECT().GetMpathMinorByMpathName(gomock.Any(), "mpath-name").Return("minor", true, nil).AnyTimes()
				mp.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				s.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(errors.New("delete error")).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "GetMultipathNameAndPaths error",
			args: args{
				ctx: context.Background(),
				wwn: "1234567894",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				mp.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("", []string{}, errors.New("multipath error")).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "GetDMDeviceByChildren error",
			args: args{
				ctx: context.Background(),
				wwn: "1234567895",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				mp.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("mpath-name", []string{"path1", "path2"}, nil).AnyTimes()
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"path1", "path2"}).Return("", errors.New("dm device error")).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "empty paths",
			args: args{
				ctx: context.Background(),
				wwn: "1234567896",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				mp.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("mpath-name", []string{}, nil).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "Powerpath daemon running",
			args: args{
				ctx: context.Background(),
				wwn: "1234567894",
			},
			setupMocks: func() {
				pp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				pp.EXPECT().FlushDevice(gomock.Any()).Return(nil).AnyTimes()
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"path1"}, nil).AnyTimes()
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"path1"}).Return("", nil).AnyTimes()
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "path1").Return(nil).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "no devices found when daemon running",
			args: args{
				ctx: context.Background(),
				wwn: "1234567897",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				mp.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("", nil, errors.New(noDevicesFound)).AnyTimes()
			},
			expectedError: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

func TestDisconnectDevicesByWWNNoMultipathError(t *testing.T) {
	type args struct {
		ctx context.Context
		wwn string
	}

	tests := []struct {
		name          string
		args          args
		setupMocks    func(mp *intmultipath.MockMultipath, pp *intpowerpath.MockPowerpath, s *intscsi.MockSCSI)
		expectedError string
	}{
		{
			name: "Error in getting devices by WWN",
			args: args{
				ctx: context.Background(),
				wwn: "1234567892",
			},
			setupMocks: func(mp *intmultipath.MockMultipath, _ *intpowerpath.MockPowerpath, s *intscsi.MockSCSI) {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("failed to find devices by wwn")).AnyTimes()
			},
			expectedError: "failed to find devices by wwn",
		},
		{
			name: "Error in GetDMDeviceByChildren",
			args: args{
				ctx: context.Background(),
				wwn: "1234567893",
			},
			setupMocks: func(mp *intmultipath.MockMultipath, _ *intpowerpath.MockPowerpath, s *intscsi.MockSCSI) {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"path1"}, nil).AnyTimes()
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "path1").Return(nil).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "Error in DeleteSCSIDeviceByName",
			args: args{
				ctx: context.Background(),
				wwn: "1234567894",
			},
			setupMocks: func(mp *intmultipath.MockMultipath, _ *intpowerpath.MockPowerpath, s *intscsi.MockSCSI) {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"path1"}, nil).AnyTimes()
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "path1").Return(errors.New("delete error")).AnyTimes()
			},
			expectedError: "delete error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mp := intmultipath.NewMockMultipath(ctrl)
			pp := intpowerpath.NewMockPowerpath(ctrl)
			s := intscsi.NewMockSCSI(ctrl)
			bc := &baseConnector{
				multipath:             mp,
				powerpath:             pp,
				scsi:                  s,
				multipathFlushRetries: 1,
			}
			tt.setupMocks(mp, pp, s)

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

func TestBaseConnector_cleanMultipathDeviceByName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mp := intmultipath.NewMockMultipath(ctrl)

	bc := &baseConnector{
		multipath:             mp,
		multipathFlushRetries: 3,
		multipathFlushTimeout: time.Second * 10,
	}

	tests := []struct {
		name       string
		setupMocks func()
		wantErr    bool
	}{
		{
			name: "flush succeeds on first try",
			setupMocks: func() {
				mp.EXPECT().FlushDevice(gomock.Any(), "mpatha").Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "flush succeeds on retry",
			setupMocks: func() {
				mp.EXPECT().FlushDevice(gomock.Any(), "mpatha").Return(errors.New("flush error")).Times(2)
				mp.EXPECT().FlushDevice(gomock.Any(), "mpatha").Return(nil).Times(1)
			},
			wantErr: false,
		},
		{
			name: "flush fails after all retries",
			setupMocks: func() {
				mp.EXPECT().FlushDevice(gomock.Any(), "mpatha").Return(errors.New("flush error")).Times(3)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			err := bc.cleanMultipathDeviceByName(context.Background(), "mpatha")
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanMultipathDeviceByName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDisconnectDevicesByWWNNoMultipathSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mp := intmultipath.NewMockMultipath(ctrl)
	pp := intpowerpath.NewMockPowerpath(ctrl)
	s := intscsi.NewMockSCSI(ctrl)

	bc := &baseConnector{
		multipath:             mp,
		powerpath:             pp,
		scsi:                  s,
		multipathFlushRetries: 1,
	}

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
				wwn: "1234567892",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"path1", "path2"}, nil).AnyTimes()
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "GetDMDeviceByChildren returns device",
			args: args{
				ctx: context.Background(),
				wwn: "1234567893",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"path1"}, nil).AnyTimes()
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"path1"}).Return("dm-device", nil).AnyTimes()
				mp.EXPECT().FlushDevice(gomock.Any(), "dm-device").Return(nil).AnyTimes()
				s.EXPECT().IsDeviceExist(gomock.Any(), "path1").Return(true).AnyTimes()
				s.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "path1").Return(nil).AnyTimes()
			},
			expectedError: "",
		},
		{
			name: "device does not exist",
			args: args{
				ctx: context.Background(),
				wwn: "1234567894",
			},
			setupMocks: func() {
				mp.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"path1"}, nil).AnyTimes()
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"path1"}).Return("", nil).AnyTimes()
				s.EXPECT().IsDeviceExist(gomock.Any(), "path1").Return(false).AnyTimes()
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
	type args struct {
		ctx        context.Context
		DeviceName string
	}

	tests := []struct {
		name        string
		args        args
		stateSetter func(fields BaseConnectorFields)
		expectedErr bool
	}{
		{
			name: "Device not found",
			args: args{
				ctx:        context.Background(),
				DeviceName: "non-existent-device",
			},
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
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDeviceWWN(gomock.Any(), gomock.Any()).Return("", errors.New("failed to read WWN for DM")).AnyTimes()
			},
			expectedErr: true,
		},
		{
			name: "Device does not have device mapper prefix",
			args: args{
				ctx:        context.Background(),
				DeviceName: "test-device",
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetDeviceWWN(gomock.Any(), gomock.Any()).Return("test-wwn", nil).AnyTimes()
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("failed to find devices by wwn")).AnyTimes()
			},
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath: fields.multipath,
				powerpath: fields.powerpath,
				scsi:      fields.scsi,
			}

			test.stateSetter(fields)

			err := bc.disconnectDevicesByDeviceName(test.args.ctx, test.args.DeviceName)

			if (err != nil) != test.expectedErr {
				t.Errorf("disconnectDevicesByDeviceName() error = %v, wantErr %v", err, test.expectedErr)
				return
			}
		})
	}
}

func TestCleanDevicesByMpathInfo(t *testing.T) {
	type args struct {
		ctx   context.Context
		force bool
		req   *cleanVolumeReq
	}

	tests := []struct {
		name        string
		args        args
		stateSetter func(fields BaseConnectorFields)
		expectedErr bool
	}{
		{
			name: "fail to verify multipath device existence",
			args: args{
				ctx:   context.Background(),
				force: false,
				req: &cleanVolumeReq{
					wwn:       "1234567892",
					sdDisks:   []string{"path1"},
					mpathName: "mpatha",
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().GetMpathMinorByMpathName(gomock.Any(), gomock.Any()).Return("", true, errors.New("failed to verify multipath device existence")).AnyTimes()
			},
			expectedErr: true,
		},
		{
			name: "fail to verify multipath device existence",
			args: args{
				ctx:   context.Background(),
				force: false,
				req: &cleanVolumeReq{
					wwn:       "1234567892",
					sdDisks:   []string{"path1"},
					mpathName: "mpatha",
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().GetMpathMinorByMpathName(gomock.Any(), gomock.Any()).Return("", true, errors.New("failed to verify multipath device existence")).AnyTimes()
			},
			expectedErr: true,
		},
		{
			name: "delete sd no mpath found",
			args: args{
				ctx:   context.Background(),
				force: false,
				req: &cleanVolumeReq{
					wwn:     "1234567892",
					sdDisks: []string{"path1"},
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name: "force cleanup on multipath flush error",
			args: args{
				ctx:   context.Background(),
				force: true,
				req: &cleanVolumeReq{
					wwn:       "1234567892",
					sdDisks:   []string{"path1"},
					mpathName: "mpatha",
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.multipath.EXPECT().GetMpathMinorByMpathName(gomock.Any(), gomock.Any()).Return("", false, nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name: "force cleanup on verification error",
			args: args{
				ctx:   context.Background(),
				force: true,
				req: &cleanVolumeReq{
					wwn:       "1234567892",
					sdDisks:   []string{"path1"},
					mpathName: "mpatha",
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().GetMpathMinorByMpathName(gomock.Any(), gomock.Any()).Return("", true, errors.New("verification error")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name: "empty sdDisks",
			args: args{
				ctx:   context.Background(),
				force: false,
				req: &cleanVolumeReq{
					wwn:     "1234567893",
					sdDisks: []string{},
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name: "delete SCSI device error",
			args: args{
				ctx:   context.Background(),
				force: false,
				req: &cleanVolumeReq{
					wwn:     "1234567894",
					sdDisks: []string{"path1"},
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(errors.New("delete error")).AnyTimes()
			},
			expectedErr: true,
		},
		{
			name: "multipath device exists",
			args: args{
				ctx:   context.Background(),
				force: false,
				req: &cleanVolumeReq{
					wwn:       "1234567895",
					sdDisks:   []string{"path1"},
					mpathName: "mpatha",
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().GetMpathMinorByMpathName(gomock.Any(), "mpatha").Return("minor", false, nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name: "GetMpathMinorByMpathName error only (not exist)",
			args: args{
				ctx:   context.Background(),
				force: false,
				req: &cleanVolumeReq{
					wwn:       "1234567896",
					sdDisks:   []string{"path1"},
					mpathName: "mpatha",
				},
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().GetMpathMinorByMpathName(gomock.Any(), "mpatha").Return("", false, errors.New("minor error")).AnyTimes()
			},
			expectedErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:             fields.multipath,
				powerpath:             fields.powerpath,
				scsi:                  fields.scsi,
				multipathFlushRetries: 1,
			}

			test.stateSetter(fields)

			err := bc.cleanDevicesByMpathInfo(test.args.ctx, test.args.force, test.args.req)

			if (err != nil) != test.expectedErr {
				t.Errorf("cleanDevicesByMpathInfo() error = %v, wantErr %v", err, test.expectedErr)
				return
			}
		})
	}
}

func TestCleanNVMeDevices(t *testing.T) {
	type args struct {
		ctx     context.Context
		Force   bool
		Devices []string
		WWN     string
	}

	tests := []struct {
		name        string
		args        args
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
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("failed to flush multipath device")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(errors.New("failed to remove wwid")).AnyTimes()
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
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("failed to flush multipath device")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:             fields.multipath,
				powerpath:             fields.powerpath,
				scsi:                  fields.scsi,
				multipathFlushRetries: 1,
			}

			test.stateSetter(fields)

			err := bc.cleanNVMeDevices(test.args.ctx, test.args.Force, test.args.Devices, test.args.WWN)

			if (err != nil) != test.expectedErr {
				t.Errorf("cleanNVMeDevices() error = %v, wantErr %v", err, test.expectedErr)
				return
			}
		})
	}
}

func TestCleanDevices(t *testing.T) {
	type args struct {
		ctx     context.Context
		Force   bool
		Devices []string
		WWN     string
	}

	tests := []struct {
		name         string
		args         args
		stateSetter  func(fields BaseConnectorFields)
		expectedErr  bool
		flushRetries int
	}{
		{
			name: "Failed to flush multipath device",
			args: args{
				ctx:     context.Background(),
				Force:   false,
				Devices: []string{},
				WWN:     "test-wwn",
			},
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
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("failed to GetDMDeviceByChildren")).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.powerpath.EXPECT().FlushDevice(gomock.Any()).Return(errors.New("failed to flush device")).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name:         "Force cleanup on multipath flush error",
			flushRetries: 0,
			args: args{
				ctx:     context.Background(),
				Force:   true,
				Devices: []string{"sdb"},
				WWN:     "test-wwn",
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name:         "Force cleanup on delete SCSI device error",
			flushRetries: 0,
			args: args{
				ctx:     context.Background(),
				Force:   true,
				Devices: []string{"sdb"},
				WWN:     "test-wwn",
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sdb").Return(errors.New("delete error")).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name: "empty devices",
			args: args{
				ctx:     context.Background(),
				Force:   false,
				Devices: []string{},
				WWN:     "test-wwn",
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
			},
			expectedErr: true,
		},
		{
			name:         "Powerpath daemon running",
			flushRetries: 1,
			args: args{
				ctx:     context.Background(),
				Force:   false,
				Devices: []string{"sdb"},
				WWN:     "test-wwn",
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sdb").Return(nil).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.powerpath.EXPECT().FlushDevice(gomock.Any()).Return(nil).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name:         "DM device found",
			flushRetries: 1,
			args: args{
				ctx:     context.Background(),
				Force:   false,
				Devices: []string{"sdb"},
				WWN:     "test-wwn",
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sdb").Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name:         "DelPath error",
			flushRetries: 0,
			args: args{
				ctx:     context.Background(),
				Force:   true,
				Devices: []string{"sdb"},
				WWN:     "test-wwn",
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sdb").Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(errors.New("delpath error")).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
			},
			expectedErr: false,
		},
		{
			name:         "GetDMDeviceByChildren error with force",
			flushRetries: 0,
			args: args{
				ctx:     context.Background(),
				Force:   true,
				Devices: []string{"sdb"},
				WWN:     "test-wwn",
			},
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("dm error")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sdb").Return(nil).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
			},
			expectedErr: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:                  fields.multipath,
				powerpath:                  fields.powerpath,
				scsi:                       fields.scsi,
				multipathFlushRetries:      test.flushRetries,
				multipathFlushTimeout:      time.Second * 10,
				multipathFlushRetryTimeout: time.Second * 5,
			}

			test.stateSetter(fields)

			err := bc.cleanDevices(test.args.ctx, test.args.Force, test.args.Devices, test.args.WWN)

			if (err != nil) != test.expectedErr {
				t.Errorf("cleanDevices() error = %v, wantErr %v", err, test.expectedErr)
				return
			}
		})
	}
}

func TestGetNVMEDMWWN(t *testing.T) {
	type args struct {
		ctx        context.Context
		DeviceName string
	}

	tests := []struct {
		name        string
		args        args
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath: fields.multipath,
				powerpath: fields.powerpath,
				scsi:      fields.scsi,
			}

			test.stateSetter(fields)

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

func TestIdentifyDevicesForWWN(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(m *intmultipath.MockMultipath, s *intscsi.MockSCSI)
		wwn        string
		want       *cleanVolumeReq
		wantErr    bool
	}{
		{
			name: "orphan multipath device",
			setupMocks: func(m *intmultipath.MockMultipath, s *intscsi.MockSCSI) {
				m.EXPECT().GetMultipathNameAndPaths(gomock.Any(), "test-wwn").Return("orphan-mpath-name", []string{"path1", "path2"}, nil)
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("dm-device", nil)
			},
			wwn: "test-wwn",
			want: &cleanVolumeReq{
				wwn:       "test-wwn",
				sdDisks:   []string{"path1", "path2"},
				mpathName: "",
				dmName:    "dm-device",
			},
			wantErr: false,
		},
		{
			name: "Error Getting Multipath name and paths",
			setupMocks: func(m *intmultipath.MockMultipath, _ *intscsi.MockSCSI) {
				m.EXPECT().GetMultipathNameAndPaths(gomock.Any(), "test-wwn").Return("", []string{}, errors.New("failed to get multipath name"))
			},
			wwn:     "test-wwn",
			want:    nil,
			wantErr: true,
		},
		{
			name: "No Device mapper found",
			setupMocks: func(m *intmultipath.MockMultipath, s *intscsi.MockSCSI) {
				m.EXPECT().GetMultipathNameAndPaths(gomock.Any(), "test-wwn").Return("mpath-name", []string{"path1", "path2"}, nil)
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("dm not found"))
			},
			wwn: "test-wwn",
			want: &cleanVolumeReq{
				wwn:       "test-wwn",
				sdDisks:   []string{"path1", "path2"},
				mpathName: "mpath-name",
				dmName:    "",
			},
			wantErr: false,
		},
		{
			name: "No Device mapper found",
			setupMocks: func(m *intmultipath.MockMultipath, s *intscsi.MockSCSI) {
				m.EXPECT().GetMultipathNameAndPaths(gomock.Any(), "test-wwn").Return("mpath-name", []string{"path1", "path2"}, nil)
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("failed to get device mapper name"))
			},
			wwn:     "test-wwn",
			want:    nil,
			wantErr: true,
		},
		{
			name: "length of WWN sddisks is zero with error",
			setupMocks: func(m *intmultipath.MockMultipath, s *intscsi.MockSCSI) {
				m.EXPECT().GetMultipathNameAndPaths(gomock.Any(), "test-wwn").Return("mpath-name", []string{}, nil)
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("dm-device", nil)
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("failed to find devices by wwn"))
			},
			wwn:     "test-wwn",
			want:    nil,
			wantErr: true,
		},
		{
			name: "length of WWN sddisks is zero without error",
			setupMocks: func(m *intmultipath.MockMultipath, s *intscsi.MockSCSI) {
				m.EXPECT().GetMultipathNameAndPaths(gomock.Any(), "test-wwn").Return("mpath-name", []string{}, nil)
				s.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("dm-device", nil)
				s.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{}, nil)
			},
			wwn:     "test-wwn",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := intmultipath.NewMockMultipath(ctrl)
			s := intscsi.NewMockSCSI(ctrl)

			tt.setupMocks(m, s)

			bc := &baseConnector{
				multipath: m,
				scsi:      s,
			}

			got, err := bc.identifyDevicesForWWN(context.Background(), tt.wwn)
			if (err != nil) != tt.wantErr {
				t.Errorf("identifyDevicesForWWN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("identifyDevicesForWWN() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaseConnector_getDMWWN(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		stateSetter func(fields BaseConnectorFields)
		dm          string
		wantErr     bool
	}{
		{
			name: "GetDMChildren error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("get children error")).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", errors.New("get wwid error")).AnyTimes()
			},
			dm:      "dm-0",
			wantErr: true,
		},
		{
			name: "GetDeviceWWN error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{"sdb"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDeviceWWN(gomock.Any(), gomock.Any()).Return("", errors.New("wwn error")).AnyTimes()
			},
			dm:      "dm-0",
			wantErr: true,
		},
		{
			name: "success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{"sdb"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDeviceWWN(gomock.Any(), gomock.Any()).Return("test-wwn", nil).AnyTimes()
			},
			dm:      "dm-0",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:             fields.multipath,
				powerpath:             fields.powerpath,
				scsi:                  fields.scsi,
				multipathFlushRetries: 0,
			}
			tt.stateSetter(fields)
			_, err := bc.getDMWWN(ctx, tt.dm)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDMWWN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseConnector_retryFlushMultipathDevice(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		stateSetter func(fields BaseConnectorFields)
		dm          string
		wantErr     bool
	}{
		{
			name: "FlushDevice error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			},
			dm:      "dm-0",
			wantErr: true,
		},
		{
			name: "device no longer exists",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
			},
			dm:      "dm-0",
			wantErr: false,
		},
		{
			name: "success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
			},
			dm:      "dm-0",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:                  fields.multipath,
				powerpath:                  fields.powerpath,
				scsi:                       fields.scsi,
				multipathFlushRetries:      0,
				multipathFlushRetryTimeout: time.Second * 5,
			}
			tt.stateSetter(fields)
			err := bc.retryFlushMultipathDevice(ctx, tt.dm)
			if (err != nil) != tt.wantErr {
				t.Errorf("retryFlushMultipathDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseConnector_cleanMultipathDevice(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		stateSetter func(fields BaseConnectorFields)
		dm          string
		wwid        string
		wantErr     bool
	}{
		{
			name: "retry flush error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
			},
			dm:      "dm-0",
			wwid:    "",
			wantErr: true,
		},
		{
			name: "empty wwid - get DMWWID success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			dm:      "dm-0",
			wwid:    "",
			wantErr: false,
		},
		{
			name: "success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			dm:      "dm-0",
			wwid:    "test-wwid",
			wantErr: false,
		},
		{
			name: "RemoveDeviceFromWWIDSFile error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(errors.New("remove error")).AnyTimes()
			},
			dm:      "dm-0",
			wwid:    "test-wwid",
			wantErr: false,
		},
		{
			name: "GetDMWWID error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("", errors.New("get wwid error")).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			dm:      "dm-0",
			wwid:    "",
			wantErr: false,
		},
		{
			name: "device no longer exists during retry",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			dm:      "dm-0",
			wwid:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:                  fields.multipath,
				powerpath:                  fields.powerpath,
				scsi:                       fields.scsi,
				multipathFlushRetries:      1,
				multipathFlushTimeout:      time.Second * 10,
				multipathFlushRetryTimeout: time.Second * 5,
			}
			tt.stateSetter(fields)
			err := bc.cleanMultipathDevice(ctx, tt.dm, tt.wwid)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanMultipathDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseConnector_disconnectNVMEDevicesByDeviceName(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		stateSetter func(fields BaseConnectorFields)
		deviceName  string
		wantErr     bool
	}{
		{
			name: "device not found",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
			},
			deviceName: "nvme0n1",
			wantErr:    false,
		},
		{
			name: "device mapper - getNVMEDMWWN error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{"nvme0n1"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), gomock.Any()).Return("", errors.New("wwn error")).AnyTimes()
			},
			deviceName: "dm-0",
			wantErr:    true,
		},
		{
			name: "device mapper - success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetDMChildren(gomock.Any(), gomock.Any()).Return([]string{"nvme0n1"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), gomock.Any()).Return("test-wwn", nil).AnyTimes()
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"nvme0n1"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			deviceName: "dm-0",
			wantErr:    false,
		},
		{
			name: "non-device mapper - GetNVMEDeviceWWN error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), []string{"nvme0n1"}).Return("", errors.New("wwn error")).AnyTimes()
			},
			deviceName: "nvme0n1",
			wantErr:    true,
		},
		{
			name: "non-device mapper - success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().GetNVMEDeviceWWN(gomock.Any(), []string{"nvme0n1"}).Return("test-wwn", nil).AnyTimes()
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{"nvme0n1"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			deviceName: "nvme0n1",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:             fields.multipath,
				powerpath:             fields.powerpath,
				scsi:                  fields.scsi,
				multipathFlushRetries: 1,
			}
			tt.stateSetter(fields)
			err := bc.disconnectNVMEDevicesByDeviceName(ctx, tt.deviceName)
			if (err != nil) != tt.wantErr {
				t.Errorf("disconnectNVMEDevicesByDeviceName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseConnector_cleanNVMeDevices(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		stateSetter func(fields BaseConnectorFields)
		devices     []string
		wwn         string
		force       bool
		wantErr     bool
	}{
		{
			name: "GetDMDeviceByChildren error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"nvme0n1"}).Return("", errors.New("dm error")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "nvme0n1").Return(nil).AnyTimes()
			},
			devices: []string{"/dev/nvme0n1"},
			wwn:     "test-wwn",
			force:   false,
			wantErr: false,
		},
		{
			name: "cleanMultipathDevice error - force false",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"nvme0n1"}).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			},
			devices: []string{"/dev/nvme0n1"},
			wwn:     "test-wwn",
			force:   false,
			wantErr: true,
		},
		{
			name: "cleanMultipathDevice error - force true",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"nvme0n1"}).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "nvme0n1").Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			devices: []string{"/dev/nvme0n1"},
			wwn:     "test-wwn",
			force:   true,
			wantErr: false,
		},
		{
			name: "success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"nvme0n1"}).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "nvme0n1").Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			devices: []string{"/dev/nvme0n1"},
			wwn:     "test-wwn",
			force:   false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:                  fields.multipath,
				powerpath:                  fields.powerpath,
				scsi:                       fields.scsi,
				multipathFlushRetries:      1,
				multipathFlushTimeout:      time.Second * 10,
				multipathFlushRetryTimeout: time.Second * 5,
			}
			tt.stateSetter(fields)
			err := bc.cleanNVMeDevices(ctx, tt.force, tt.devices, tt.wwn)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanNVMeDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseConnector_identifyDevicesForWWN(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		stateSetter func(fields BaseConnectorFields)
		wwn         string
		wantErr     bool
	}{
		{
			name: "GetMultipathNameAndPaths error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("", nil, errors.New("multipath error")).AnyTimes()
			},
			wwn:     "test-wwn",
			wantErr: true,
		},
		{
			name: "orphan device",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("orphan-device", []string{"sda"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("dm-0", nil).AnyTimes()
			},
			wwn:     "test-wwn",
			wantErr: false,
		},
		{
			name: "GetDMDeviceByChildren error - not dm not found",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("mpath0", []string{"sda"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("", errors.New("dm error")).AnyTimes()
			},
			wwn:     "test-wwn",
			wantErr: true,
		},
		{
			name: "dm not found",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("mpath0", []string{"sda"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("", errors.New(scsi.DmNotFoundErr)).AnyTimes()
			},
			wwn:     "test-wwn",
			wantErr: false,
		},
		{
			name: "success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.multipath.EXPECT().GetMultipathNameAndPaths(gomock.Any(), gomock.Any()).Return("mpath0", []string{"sda"}, nil).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("dm-0", nil).AnyTimes()
			},
			wwn:     "test-wwn",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath: fields.multipath,
				powerpath: fields.powerpath,
				scsi:      fields.scsi,
			}
			tt.stateSetter(fields)
			_, err := bc.identifyDevicesForWWN(ctx, tt.wwn)
			if (err != nil) != tt.wantErr {
				t.Errorf("identifyDevicesForWWN() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBaseConnector_cleanDevices(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		stateSetter func(fields BaseConnectorFields)
		devices     []string
		wwn         string
		force       bool
		wantErr     bool
	}{
		{
			name: "GetDMDeviceByChildren error",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("", errors.New("dm error")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sda").Return(nil).AnyTimes()
			},
			devices: []string{"sda"},
			wwn:     "test-wwn",
			force:   false,
			wantErr: false,
		},
		{
			name: "cleanMultipathDevice error - force false",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			},
			devices: []string{"sda"},
			wwn:     "test-wwn",
			force:   false,
			wantErr: true,
		},
		{
			name: "cleanMultipathDevice error - force true",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(errors.New("flush error")).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sda").Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			devices: []string{"sda"},
			wwn:     "test-wwn",
			force:   true,
			wantErr: false,
		},
		{
			name: "DeleteSCSIDeviceByName error - force false",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sda").Return(errors.New("delete error")).AnyTimes()
			},
			devices: []string{"sda"},
			wwn:     "test-wwn",
			force:   false,
			wantErr: true,
		},
		{
			name: "powerpath daemon running",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("", errors.New("dm not found")).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sda").Return(nil).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(true).AnyTimes()
				fields.powerpath.EXPECT().FlushDevice(gomock.Any()).Return(nil).AnyTimes()
			},
			devices: []string{"sda"},
			wwn:     "test-wwn",
			force:   false,
			wantErr: false,
		},
		{
			name: "success",
			stateSetter: func(fields BaseConnectorFields) {
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), []string{"sda"}).Return("dm-0", nil).AnyTimes()
				fields.multipath.EXPECT().GetDMWWID(gomock.Any(), gomock.Any()).Return("test-wwid", nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), "sda").Return(nil).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.powerpath.EXPECT().IsDaemonRunning(gomock.Any()).Return(false).AnyTimes()
			},
			devices: []string{"sda"},
			wwn:     "test-wwn",
			force:   false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			fields := getTestBaseConnector(ctrl)
			bc := &baseConnector{
				multipath:                  fields.multipath,
				powerpath:                  fields.powerpath,
				scsi:                       fields.scsi,
				multipathFlushRetries:      1,
				multipathFlushTimeout:      time.Second * 10,
				multipathFlushRetryTimeout: time.Second * 5,
			}
			tt.stateSetter(fields)
			err := bc.cleanDevices(ctx, tt.force, tt.devices, tt.wwn)
			if (err != nil) != tt.wantErr {
				t.Errorf("cleanDevices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCleanDevices_PowerpathFlushError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mp := intmultipath.NewMockMultipath(ctrl)
	pp := intpowerpath.NewMockPowerpath(ctrl)
	s := intscsi.NewMockSCSI(ctrl)

	bc := &baseConnector{
		multipath: mp,
		powerpath: pp,
		scsi:      s,
	}

	s.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return("", errors.New("dm not found"))
	pp.EXPECT().IsDaemonRunning(gomock.Any()).Return(true)
	pp.EXPECT().FlushDevice(gomock.Any()).Return(errors.New("powerpath flush error"))

	err := bc.cleanDevices(context.Background(), false, []string{}, "test-wwn")
	if err == nil {
		t.Error("expected error from powerpath flush, got nil")
	}
}
