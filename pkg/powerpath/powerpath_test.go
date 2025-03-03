/*
 *
 * Copyright © 2020-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in coppliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or ipplied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package powerpath

import (
	"context"
	"reflect"
	"testing"

	mh "github.com/dell/gobrick/internal/mockhelper"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/golang/mock/gomock"
)

type ppFields struct {
	chroot   string
	osexec   *wrp.MockLimitedOSExec
	filePath *wrp.MockLimitedFilepath
}

func getDefaultPPFields(ctrl *gomock.Controller) ppFields {
	filePathMock := wrp.NewMockLimitedFilepath(ctrl)
	osExecMock := wrp.NewMockLimitedOSExec(ctrl)
	return ppFields{
		filePath: filePathMock,
		osexec:   osExecMock,
	}
}

func TestNewPowerpath(t *testing.T) {
	pp := NewPowerpath("")
	if pp == nil {
		t.Errorf("NewPowerpath() should not return nil")
	}
}

type t struct {
	name        string
	fields      ppFields
	args        interface{}
	stateSetter func(fields ppFields)
	wantErr     bool
}

func getPowerpathCommonTestCases(mocks mh.MockHelper, defaultArgs interface{},
	ctrl *gomock.Controller,
) []t {
	return []t{
		{
			name:   "ok",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: false,
		},
		{
			name:   "unexpected output",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "timeout"
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: true,
		},
		{
			name:   "exec error",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args:    defaultArgs,
			wantErr: true,
		},
	}
}

func Test_powerpath_FlushDevice(t *testing.T) {
	type args struct {
		ctx           context.Context
		deviceMapName string
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx, deviceMapName: mh.ValidDMName}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: powerpathTool,
		OSEXECCommandContextArgs: []string{"check", "force"},
		OSEXECCmdOKReturn:        "ok",
	}

	tests := []struct {
		name        string
		fields      ppFields
		stateSetter func(fields ppFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "ok"
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: false,
		},
		{
			name:   "exec error",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args:    defaultArgs,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := &Powerpath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if err := pp.FlushDevice(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("FlushDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_powerpath_IsDaemonRunning(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	canceledCtx, cFunc := context.WithCancel(context.Background())
	cFunc()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: powerpathDaemon,
		OSEXECCommandContextArgs: []string{"version"},
		OSEXECCmdOKReturn: `
path checker states:
down                2
up                  10

paths: 12
busy: False
`,
	}

	tests := []struct {
		name        string
		fields      ppFields
		stateSetter func(fields ppFields)
		args        args
		want        bool
	}{
		{
			name:   "ok",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
				_, cmdMock2 := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock2)
			},
			args: defaultArgs,
			want: true,
		},
		{
			name:   "exec error",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args: defaultArgs,
			want: false,
		},
		{
			name:   "invalid resp",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "error receiving packet"
				mocks.OSExecCmdOK(cmdMock)
			},
			args: defaultArgs,
			want: false,
		},
		{
			name:   "canceled context",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "canceled"
				mocks.OSExecCmdErr(cmdMock)
			},
			args: args{ctx: canceledCtx},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := &Powerpath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if got := pp.IsDaemonRunning(tt.args.ctx); got != tt.want {
				t.Errorf("IsDaemonRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_powerpath_runCommand(t *testing.T) {
	type args struct {
		ctx     context.Context
		command string
		args    []string
	}

	ctx := context.Background()

	testCMD := "foo"
	testArgs := []string{"-a", "bar"}
	validResp := "ok"
	chRoot := "/fakeroot"

	defaultArgs := args{ctx: ctx, command: testCMD, args: testArgs}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	chrootFields := getDefaultPPFields(ctrl)
	chrootFields.chroot = chRoot

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: testCMD,
		OSEXECCommandContextArgs: testArgs,
		OSEXECCmdOKReturn:        validResp,
	}

	tests := []struct {
		name        string
		fields      ppFields
		stateSetter func(fields ppFields)
		args        args
		want        []byte
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    []byte(validResp),
			wantErr: false,
		},
		{
			name:   "exec err",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args:    defaultArgs,
			want:    nil,
			wantErr: true,
		},
		{
			name:   "chroot",
			fields: chrootFields,
			stateSetter: func(fields ppFields) {
				mocks.OSEXECCommandContextName = "chroot"
				mocks.OSEXECCommandContextArgs = append([]string{chRoot, testCMD}, testArgs...)
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    []byte(validResp),
			wantErr: false,
		},
		{
			name:        "invalid command",
			fields:      chrootFields,
			stateSetter: func(_ ppFields) {},
			args:        args{ctx: ctx, command: "ls | cd", args: testArgs},
			want:        []byte(nil),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := &Powerpath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			got, err := pp.runCommand(tt.args.ctx, tt.args.command, tt.args.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("runCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("runCommand() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_powerpath_GetPowerPathDevices(t *testing.T) {
	type args struct {
		ctx     context.Context
		devices []string
	}

	ctx := context.Background()

	defaultArgs := args{
		ctx:     ctx,
		devices: []string{"device1"},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: powerpathDaemon,
		OSEXECCommandContextArgs: []string{"display", "dev=device1"},
		OSEXECCmdOKReturn:        "Pseudo name=emc_power\n",
	}

	tests := []struct {
		name        string
		fields      ppFields
		stateSetter func(fields ppFields)
		args        args
		want        string
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    "emc_power",
			wantErr: false,
		},
		{
			name:   "exec error",
			fields: getDefaultPPFields(ctrl),
			stateSetter: func(fields ppFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args:    defaultArgs,
			want:    "",
			wantErr: true,
		},
		{
			name:        "No devices",
			fields:      getDefaultPPFields(ctrl),
			stateSetter: func(_ ppFields) {},
			args: args{
				ctx:     ctx,
				devices: []string{},
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := &Powerpath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			got, err := pp.GetPowerPathDevices(tt.args.ctx, tt.args.devices)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPowerPathDevices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetPowerPathDevices() = %v, want %v", got, tt.want)
			}
		})
	}
}
