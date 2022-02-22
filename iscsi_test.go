package gobrick

import (
	"context"
	"fmt"
	"github.com/dell/gobrick/internal/mockhelper"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/goiscsi"
	"github.com/golang/mock/gomock"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
	"reflect"
	"testing"
	"time"
)

var (
	validISCSIPortal1     = "1.1.1.1:3260"
	validISCSITarget1     = "iqn.2015-10.com.dell:dellemc-foobar123"
	validISCSIPortal2     = "1.1.1.1:3260"
	validISCSITarget2     = "iqn.2015-10.com.dell:dellemc-spam789"
	validISCSITargetInfo1 = ISCSITargetInfo{
		Portal: validISCSIPortal1,
		Target: validISCSITarget1,
	}
	validISCSITargetInfo2 = ISCSITargetInfo{
		Portal: validISCSIPortal2,
		Target: validISCSITarget2,
	}
	validISCSIVolumeInfo = ISCSIVolumeInfo{
		Targets: []ISCSITargetInfo{validISCSITargetInfo1, validISCSITargetInfo2},
		Lun:     validLunNumber,
	}

	validLibISCSITarget1 = goiscsi.ISCSITarget{
		Target: validISCSITarget1,
		Portal: validISCSIPortal1,
	}

	validLibISCSITarget2 = goiscsi.ISCSITarget{
		Target: validISCSITarget2,
		Portal: validISCSIPortal2,
	}

	validISCSIInitiatorName = "iqn.1993-08.org.debian:01:e16da41ba075"

	validLibISCSISession1 = goiscsi.ISCSISession{
		Target:               validISCSITarget1,
		Portal:               validISCSIPortal1,
		SID:                  "12",
		IfaceTransport:       "tcp",
		IfaceInitiatorname:   validISCSIInitiatorName,
		IfaceIPaddress:       "192.168.100.10",
		ISCSISessionState:    goiscsi.ISCSISessionState_LOGGED_IN,
		ISCSIConnectionState: goiscsi.ISCSIConnectionState_LOGGED_IN,
	}
	validLibISCSISession2 = goiscsi.ISCSISession{
		Target:               validISCSITarget2,
		Portal:               validISCSIPortal2,
		SID:                  "14",
		IfaceTransport:       "tcp",
		IfaceInitiatorname:   validISCSIInitiatorName,
		IfaceIPaddress:       "192.168.100.11",
		ISCSISessionState:    goiscsi.ISCSISessionState_LOGGED_IN,
		ISCSIConnectionState: goiscsi.ISCSIConnectionState_LOGGED_IN,
	}
	validLibISCSISessions = []goiscsi.ISCSISession{validLibISCSISession1, validLibISCSISession2}
)

type iscsiFields struct {
	baseConnector *baseConnector
	multipath     *intmultipath.MockMultipath
	scsi          *intscsi.MockSCSI
	iscsiLib      *wrp.MockISCSILib
	filePath      *wrp.MockLimitedFilepath

	manualSessionManagement                bool
	waitDeviceTimeout                      time.Duration
	waitDeviceRegisterTimeout              time.Duration
	failedSessionMinimumLoginRetryInterval time.Duration
	loginLock                              *rateLock
	limiter                                *semaphore.Weighted
	singleCall                             *singleflight.Group
}

func getDefaultISCSIFields(ctrl *gomock.Controller) iscsiFields {
	con := NewISCSIConnector(ISCSIConnectorParams{})
	bc := con.baseConnector
	mpMock := intmultipath.NewMockMultipath(ctrl)
	scsiMock := intscsi.NewMockSCSI(ctrl)
	bc.multipath = mpMock
	bc.scsi = scsiMock
	return iscsiFields{
		baseConnector:                          bc,
		multipath:                              mpMock,
		scsi:                                   scsiMock,
		iscsiLib:                               wrp.NewMockISCSILib(ctrl),
		filePath:                               wrp.NewMockLimitedFilepath(ctrl),
		manualSessionManagement:                con.manualSessionManagement,
		waitDeviceTimeout:                      con.waitDeviceTimeout,
		waitDeviceRegisterTimeout:              con.waitDeviceRegisterTimeout,
		failedSessionMinimumLoginRetryInterval: con.failedSessionMinimumLoginRetryInterval,
		loginLock:                              con.loginLock,
		limiter:                                con.limiter,
		singleCall:                             con.singleCall,
	}
}

func TestISCSIConnector_ConnectVolume(t *testing.T) {
	type args struct {
		ctx  context.Context
		info ISCSIVolumeInfo
	}

	ctx := context.Background()
	defaultArgs := args{ctx: ctx, info: validISCSIVolumeInfo}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	iscsiOptions := map[string]string{"node.session.initial_login_retry_max": "1"}
	iscsiOptionsManual := map[string]string{"node.session.scan": "manual"}
	mock := baseMockHelper{
		Ctx: ctx,
	}

	iscsiSessionSetup := func(fields iscsiFields) {
		// checkISCSISessions
		mock.ISCSILibCreateOrUpdateNodeCallOptions = iscsiOptions

		mock.ISCSILibCreateOrUpdateNodeCallTarget = validLibISCSITarget1
		mock.ISCSILibCreateOrUpdateNodeOK(fields.iscsiLib)
		mock.ISCSILibCreateOrUpdateNodeCallTarget = validLibISCSITarget2
		mock.ISCSILibCreateOrUpdateNodeOK(fields.iscsiLib)

		// tryEnableManualISCSISessionMGMT
		mock.ISCSILibCreateOrUpdateNodeCallOptions = iscsiOptionsManual
		mock.ISCSILibCreateOrUpdateNodeCallTarget = validLibISCSITarget1
		mock.ISCSILibCreateOrUpdateNodeOK(fields.iscsiLib)
		mock.ISCSILibCreateOrUpdateNodeCallTarget = validLibISCSITarget2
		mock.ISCSILibCreateOrUpdateNodeOK(fields.iscsiLib)

		// getSessionByTargetInfo
		mock.ISCSILibGetSessionsOKReturn = []goiscsi.ISCSISession{validLibISCSISession1}
		mock.ISCSILibGetSessionsOK(fields.iscsiLib).Times(2)

		// tryISCSILogin
		mock.ISCSILibPerformLoginCallTarget = validLibISCSITarget2
		mock.ISCSILibPerformLoginOK(fields.iscsiLib)
		mock.ISCSILibGetSessionsOKReturn = validLibISCSISessions
		mock.ISCSILibGetSessionsOK(fields.iscsiLib)
	}

	deviceDiscovery := func(fields iscsiFields) {
		sessionPtrn := "/sys/class/iscsi_host/host*/device/session%s"
		sessionMatchPtrn := "/sys/class/iscsi_host/host%s/device/session%s"
		targetPtrn := sessionPtrn + "/target*"
		targetMatchPtrn := "/sys/class/iscsi_host/host%s/device/session%s/target%s:%s:%s"

		oldCtx := mock.Ctx
		mock.Ctx = gomock.Any()

		// findHCTLByISCSISessionID - no match on target pattern
		// first session
		mock.FilePathGlobOKReturn = []string{}
		mock.FilePathGlobCallPattern = fmt.Sprintf(targetPtrn, validLibISCSISession1.SID)
		mock.FilePathGlobOK(fields.filePath)
		mock.FilePathGlobOK(fields.filePath)
		// second session
		mock.FilePathGlobCallPattern = fmt.Sprintf(targetPtrn, validLibISCSISession2.SID)
		mock.FilePathGlobOK(fields.filePath)
		mock.FilePathGlobOK(fields.filePath)

		// findHCTLByISCSISessionID - match host only pattern
		// first session
		mock.FilePathGlobCallPattern = fmt.Sprintf(sessionPtrn, validLibISCSISession1.SID)
		mock.FilePathGlobOKReturn = []string{fmt.Sprintf(
			sessionMatchPtrn,
			validHCTL1.Host,
			validLibISCSISession1.SID)}
		mock.FilePathGlobOK(fields.filePath)
		mock.FilePathGlobOK(fields.filePath)

		// second session
		mock.FilePathGlobCallPattern = fmt.Sprintf(sessionPtrn, validLibISCSISession2.SID)
		mock.FilePathGlobOKReturn = []string{fmt.Sprintf(
			sessionMatchPtrn,
			validHCTL2.Host,
			validLibISCSISession2.SID)}
		mock.FilePathGlobOK(fields.filePath)
		mock.FilePathGlobOK(fields.filePath)

		// first session
		mock.SCSIRescanSCSIHostByHCTLCallH = validHostOnlyHCTL1
		mock.SCSIRescanSCSIHostByHCTLOK(fields.scsi)
		// second session
		mock.SCSIRescanSCSIHostByHCTLCallH = validHostOnlyHCTL2
		mock.SCSIRescanSCSIHostByHCTLOK(fields.scsi)

		// findHCTLByISCSISessionID - match on target path
		// first session
		mock.FilePathGlobCallPattern = fmt.Sprintf(targetPtrn, validLibISCSISession1.SID)
		mock.FilePathGlobOKReturn = []string{fmt.Sprintf(
			targetMatchPtrn,
			validHCTL1.Host,
			validLibISCSISession1.SID,
			validHCTL1.Host,
			validHCTL1.Channel,
			validHCTL1.Target)}
		mock.FilePathGlobOK(fields.filePath)

		// second session
		mock.FilePathGlobCallPattern = fmt.Sprintf(targetPtrn, validLibISCSISession2.SID)
		mock.FilePathGlobOKReturn = []string{fmt.Sprintf(
			targetMatchPtrn,
			validHCTL2.Host,
			validLibISCSISession2.SID,
			validHCTL2.Host,
			validHCTL2.Channel,
			validHCTL2.Target)}
		mock.FilePathGlobOK(fields.filePath)

		// GetDeviceNameByHCTL - err on first try
		// first session
		mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1
		mock.SCSIGetDeviceNameByHCTLErr(fields.scsi)
		// second session
		mock.SCSIGetDeviceNameByHCTLCallH = validHCTL2
		// second session failed to discover device multipile times
		mock.SCSIGetDeviceNameByHCTLErr(fields.scsi)
		mock.SCSIGetDeviceNameByHCTLErr(fields.scsi)
		mock.SCSIGetDeviceNameByHCTLErr(fields.scsi)

		// GetDeviceNameByHCTL - OK
		// first session
		mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1
		mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName
		mock.SCSIGetDeviceNameByHCTLOK(fields.scsi)
		// second session
		mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName2
		mock.SCSIGetDeviceNameByHCTLCallH = validHCTL2
		mock.SCSIGetDeviceNameByHCTLOK(fields.scsi)

		mock.Ctx = oldCtx
	}

	multipathConnect := func(fields iscsiFields) {
		mock.SCSIGetDeviceWWNCallDevices = []string{mockhelper.ValidDeviceName}
		mock.SCSIGetDeviceWWNErr(fields.scsi).MinTimes(1)

		mock.SCSIGetDeviceWWNCallDevices = []string{mockhelper.ValidDeviceName, mockhelper.ValidDeviceName2}
		mock.SCSIGetDeviceWWNOKReturn = mockhelper.ValidWWID
		mock.SCSIGetDeviceWWNOK(fields.scsi)

		// GetDMDeviceByChildren fail 2 times
		mock.SCSIGetDMDeviceByChildrenCallDevices = []string{mockhelper.ValidDeviceName, mockhelper.ValidDeviceName2}
		mock.SCSIGetDMDeviceByChildrenErr(fields.scsi)
		mock.SCSIGetDMDeviceByChildrenErr(fields.scsi)

		mock.MultipathAddWWIDCallWWID = mockhelper.ValidWWID
		mock.MultipathAddWWIDOK(fields.multipath)

		mock.MultipathAddPathCallPath = mockhelper.ValidDevicePath
		mock.MultipathAddPathOK(fields.multipath)
		mock.MultipathAddPathCallPath = mockhelper.ValidDevicePath2
		mock.MultipathAddPathOK(fields.multipath)

		// GetDMDeviceByChildren OK
		mock.SCSIGetDMDeviceByChildrenOKReturn = mockhelper.ValidDMName
		mock.SCSIGetDMDeviceByChildrenOK(fields.scsi)

		mock.SCSIWaitUdevSymlinkCallDevice = mockhelper.ValidDMName
		mock.SCSIWaitUdevSymlinkCallWWN = mockhelper.ValidWWID
		mock.SCSIWaitUdevSymlinkErr(fields.scsi)
		mock.SCSIWaitUdevSymlinkOK(fields.scsi)
	}

	singleConnect := func(fields iscsiFields) {
		mock.SCSIGetDeviceWWNCallDevices = []string{mockhelper.ValidDeviceName}
		mock.SCSIGetDeviceWWNErr(fields.scsi).MinTimes(1)

		mock.SCSIGetDeviceWWNCallDevices = []string{mockhelper.ValidDeviceName, mockhelper.ValidDeviceName2}
		mock.SCSIGetDeviceWWNOKReturn = mockhelper.ValidWWID
		mock.SCSIGetDeviceWWNOK(fields.scsi)

		mock.SCSIWaitUdevSymlinkCallDevice = mockhelper.ValidDeviceName
		mock.SCSIWaitUdevSymlinkCallWWN = mockhelper.ValidWWID
		mock.SCSIWaitUdevSymlinkErr(fields.scsi).Times(2)
		mock.SCSIWaitUdevSymlinkCallDevice = mockhelper.ValidDeviceName2
		mock.SCSIWaitUdevSymlinkErr(fields.scsi)
		mock.SCSIWaitUdevSymlinkOK(fields.scsi)
	}

	tests := []struct {
		name        string
		fields      iscsiFields
		args        args
		stateSetter func(fields iscsiFields)
		want        Device
		wantErr     bool
	}{
		{
			name:        "empty request",
			fields:      getDefaultISCSIFields(ctrl),
			stateSetter: func(fields iscsiFields) {},
			args:        args{ctx: ctx, info: ISCSIVolumeInfo{}},
			want:        Device{},
			wantErr:     true,
		},
		{
			name:   "ok-multipath",
			fields: getDefaultISCSIFields(ctrl),
			stateSetter: func(fields iscsiFields) {
				iscsiSessionSetup(fields)

				mock.MultipathIsDaemonRunningOKReturn = true
				mock.MultipathIsDaemonRunningOK(fields.multipath)

				deviceDiscovery(fields)
				multipathConnect(fields)

				mock.SCSICheckDeviceIsValidCallDevice = mockhelper.ValidDMPath
				mock.SCSICheckDeviceIsValidOKReturn = true
				mock.SCSICheckDeviceIsValidOK(fields.scsi)

			},
			args:    defaultArgs,
			want:    validDeviceMultipath,
			wantErr: false,
		},
		{
			name:   "ok-single device",
			fields: getDefaultISCSIFields(ctrl),
			stateSetter: func(fields iscsiFields) {
				iscsiSessionSetup(fields)
				mock.MultipathIsDaemonRunningOKReturn = false
				mock.MultipathIsDaemonRunningOK(fields.multipath)
				deviceDiscovery(fields)
				singleConnect(fields)

				mock.SCSICheckDeviceIsValidCallDevice = mockhelper.ValidDevicePath2
				mock.SCSICheckDeviceIsValidOKReturn = true
				mock.SCSICheckDeviceIsValidOK(fields.scsi)

			},
			args: defaultArgs,
			want: Device{
				WWN:  mockhelper.ValidWWID,
				Name: mockhelper.ValidDeviceName2},
			wantErr: false,
		},
		{
			name:   "invalid device",
			fields: getDefaultISCSIFields(ctrl),
			stateSetter: func(fields iscsiFields) {
				iscsiSessionSetup(fields)
				mock.MultipathIsDaemonRunningOKReturn = false
				mock.MultipathIsDaemonRunningOK(fields.multipath)
				deviceDiscovery(fields)
				singleConnect(fields)

				mock.SCSICheckDeviceIsValidCallDevice = mockhelper.ValidDevicePath2
				mock.SCSICheckDeviceIsValidOKReturn = false
				mock.SCSICheckDeviceIsValidOK(fields.scsi)
				// no sessions
				mock.ISCSILibGetSessionsOKReturn = nil
				mock.ISCSILibGetSessionsOK(fields.iscsiLib).MinTimes(2)
			},
			args:    defaultArgs,
			want:    Device{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ISCSIConnector{
				baseConnector:                          tt.fields.baseConnector,
				multipath:                              tt.fields.multipath,
				scsi:                                   tt.fields.scsi,
				iscsiLib:                               tt.fields.iscsiLib,
				manualSessionManagement:                tt.fields.manualSessionManagement,
				waitDeviceTimeout:                      tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout:              tt.fields.waitDeviceRegisterTimeout,
				failedSessionMinimumLoginRetryInterval: tt.fields.failedSessionMinimumLoginRetryInterval,
				loginLock:                              tt.fields.loginLock,
				limiter:                                tt.fields.limiter,
				singleCall:                             tt.fields.singleCall,
				filePath:                               tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			got, err := c.ConnectVolume(tt.args.ctx, tt.args.info)
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

func TestISCSIConnector_GetInitiatorName(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.Background()
	defaultArgs := args{ctx: ctx}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := baseMockHelper{
		Ctx: ctx,
	}

	tests := []struct {
		name        string
		fields      iscsiFields
		args        args
		stateSetter func(fields iscsiFields)
		want        []string
		wantErr     bool
	}{
		{
			name:   "empty resp",
			fields: getDefaultISCSIFields(ctrl),
			stateSetter: func(fields iscsiFields) {
				mock.ISCSILibGetInitiatorsOK(fields.iscsiLib)
			},
			args:    defaultArgs,
			want:    nil,
			wantErr: false,
		},
		{
			name:   "valid resp",
			fields: getDefaultISCSIFields(ctrl),
			stateSetter: func(fields iscsiFields) {
				mock.ISCSILibGetInitiatorsOKReturn = []string{validISCSIInitiatorName}
				mock.ISCSILibGetInitiatorsOK(fields.iscsiLib)
			},
			args:    defaultArgs,
			want:    []string{validISCSIInitiatorName},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ISCSIConnector{
				baseConnector:                          tt.fields.baseConnector,
				multipath:                              tt.fields.multipath,
				scsi:                                   tt.fields.scsi,
				iscsiLib:                               tt.fields.iscsiLib,
				manualSessionManagement:                tt.fields.manualSessionManagement,
				waitDeviceTimeout:                      tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout:              tt.fields.waitDeviceRegisterTimeout,
				failedSessionMinimumLoginRetryInterval: tt.fields.failedSessionMinimumLoginRetryInterval,
				loginLock:                              tt.fields.loginLock,
				limiter:                                tt.fields.limiter,
				singleCall:                             tt.fields.singleCall,
				filePath:                               tt.fields.filePath,
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

func TestISCSIConnector_DisconnectVolume(t *testing.T) {
	type args struct {
		ctx  context.Context
		info ISCSIVolumeInfo
	}

	ctx := context.Background()
	defaultArgs := args{ctx: ctx, info: validISCSIVolumeInfo}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := baseMockHelper{
		Ctx: gomock.Any(),
	}

	tests := []struct {
		name        string
		fields      iscsiFields
		stateSetter func(fields iscsiFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultISCSIFields(ctrl),
			stateSetter: func(fields iscsiFields) {
				mock.ISCSILibGetSessionsOKReturn = validLibISCSISessions
				mock.ISCSILibGetSessionsOK(fields.iscsiLib).Times(2)

				sessionPtrn := "/sys/class/iscsi_host/host*/device/session%s"
				targetPtrn := sessionPtrn + "/target*"
				targetMatchPtrn := "/sys/class/iscsi_host/host%s/device/session%s/target%s:%s:%s"

				// findHCTLByISCSISessionID - match on target path
				// first session
				mock.FilePathGlobCallPattern = fmt.Sprintf(targetPtrn, validLibISCSISession1.SID)
				mock.FilePathGlobOKReturn = []string{fmt.Sprintf(
					targetMatchPtrn,
					validHCTL1.Host,
					validLibISCSISession1.SID,
					validHCTL1.Host,
					validHCTL1.Channel,
					validHCTL1.Target)}
				mock.FilePathGlobOK(fields.filePath)

				// second session
				mock.FilePathGlobCallPattern = fmt.Sprintf(targetPtrn, validLibISCSISession2.SID)
				mock.FilePathGlobOKReturn = []string{fmt.Sprintf(
					targetMatchPtrn,
					validHCTL2.Host,
					validLibISCSISession2.SID,
					validHCTL2.Host,
					validHCTL2.Channel,
					validHCTL2.Target)}
				mock.FilePathGlobOK(fields.filePath)

				mock.SCSIGetDeviceNameByHCTLCallH = validHCTL1
				mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName
				mock.SCSIGetDeviceNameByHCTLOK(fields.scsi)
				// second session
				mock.SCSIGetDeviceNameByHCTLOKReturn = mockhelper.ValidDeviceName2
				mock.SCSIGetDeviceNameByHCTLCallH = validHCTL2
				mock.SCSIGetDeviceNameByHCTLOK(fields.scsi)

				BaseConnectorCleanDeviceMock(&mock, fields.scsi, fields.multipath)

			},
			args:    defaultArgs,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ISCSIConnector{
				baseConnector:                          tt.fields.baseConnector,
				multipath:                              tt.fields.multipath,
				scsi:                                   tt.fields.scsi,
				iscsiLib:                               tt.fields.iscsiLib,
				manualSessionManagement:                tt.fields.manualSessionManagement,
				waitDeviceTimeout:                      tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout:              tt.fields.waitDeviceRegisterTimeout,
				failedSessionMinimumLoginRetryInterval: tt.fields.failedSessionMinimumLoginRetryInterval,
				loginLock:                              tt.fields.loginLock,
				limiter:                                tt.fields.limiter,
				singleCall:                             tt.fields.singleCall,
				filePath:                               tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if err := c.DisconnectVolume(tt.args.ctx, tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("DisconnectVolume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestISCSIConnector_DisconnectVolumeByDeviceName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}

	ctx := context.Background()
	defaultArgs := args{ctx: ctx, name: mockhelper.ValidDMName}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := baseMockHelper{
		Ctx: gomock.Any(),
	}

	tests := []struct {
		name        string
		fields      iscsiFields
		stateSetter func(fields iscsiFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultISCSIFields(ctrl),
			stateSetter: func(fields iscsiFields) {
				BaserConnectorDisconnectDevicesByDeviceNameMock(&mock, fields.scsi, fields.multipath)
			},
			args:    defaultArgs,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ISCSIConnector{
				baseConnector:                          tt.fields.baseConnector,
				multipath:                              tt.fields.multipath,
				scsi:                                   tt.fields.scsi,
				iscsiLib:                               tt.fields.iscsiLib,
				manualSessionManagement:                tt.fields.manualSessionManagement,
				waitDeviceTimeout:                      tt.fields.waitDeviceTimeout,
				waitDeviceRegisterTimeout:              tt.fields.waitDeviceRegisterTimeout,
				failedSessionMinimumLoginRetryInterval: tt.fields.failedSessionMinimumLoginRetryInterval,
				loginLock:                              tt.fields.loginLock,
				limiter:                                tt.fields.limiter,
				singleCall:                             tt.fields.singleCall,
				filePath:                               tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if err := c.DisconnectVolumeByDeviceName(tt.args.ctx, tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("DisconnectVolumeByDeviceName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
