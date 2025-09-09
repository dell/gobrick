/*
 *
 * Copyright © 2020 Dell Inc. or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package multipath

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	mh "github.com/dell/gobrick/internal/mockhelper"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/golang/mock/gomock"
)

type mpFields struct {
	chroot   string
	osexec   *wrp.MockLimitedOSExec
	filePath *wrp.MockLimitedFilepath
}

func getDefaultMPFields(ctrl *gomock.Controller) mpFields {
	filePathMock := wrp.NewMockLimitedFilepath(ctrl)
	osExecMock := wrp.NewMockLimitedOSExec(ctrl)
	return mpFields{
		filePath: filePathMock,
		osexec:   osExecMock,
		chroot:   "chroot",
	}
}

func TestNewMultipath(t *testing.T) {
	mp := NewMultipath("")
	if mp == nil {
		t.Errorf("NewMultipath() should not return nil")
	}
}

type t struct {
	name        string
	fields      mpFields
	args        interface{}
	stateSetter func(fields mpFields)
	wantErr     bool
}

func getMultipathCommonTestCases(mocks mh.MockHelper, defaultArgs interface{},
	ctrl *gomock.Controller,
) []t {
	return []t{
		{
			name:   "ok",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: false,
		},
		{
			name:   "unexpected output",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "timeout"
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: true,
		},
		{
			name:   "exec error",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args:    defaultArgs,
			wantErr: true,
		},
	}
}

func Test_multipath_AddPath(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx, path: mh.ValidDevicePath}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: multipathDaemon,
		OSEXECCommandContextArgs: []string{"add", "path", mh.ValidDevicePath},
		OSEXECCmdOKReturn:        "ok",
	}

	tests := getMultipathCommonTestCases(mocks, defaultArgs, ctrl)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if err := mp.AddPath(tt.args.(args).ctx, tt.args.(args).path); (err != nil) != tt.wantErr {
				t.Errorf("AddPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_multipath_AddWWID(t *testing.T) {
	type args struct {
		ctx  context.Context
		wwid string
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx, wwid: mh.ValidWWID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: multipathTool,
		OSEXECCommandContextArgs: []string{"-a", mh.ValidWWID},
		OSEXECCmdOKReturn:        fmt.Sprintf("wwid '%s' added", mh.ValidWWID),
	}

	tests := getMultipathCommonTestCases(mocks, defaultArgs, ctrl)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if err := mp.AddWWID(tt.args.(args).ctx, tt.args.(args).wwid); (err != nil) != tt.wantErr {
				t.Errorf("AddWWID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_multipath_DelPath(t *testing.T) {
	type args struct {
		ctx  context.Context
		path string
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx, path: mh.ValidDevicePath}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: multipathDaemon,
		OSEXECCommandContextArgs: []string{"del", "path", mh.ValidDevicePath},
		OSEXECCmdOKReturn:        "ok",
	}

	tests := getMultipathCommonTestCases(mocks, defaultArgs, ctrl)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if err := mp.DelPath(tt.args.(args).ctx, tt.args.(args).path); (err != nil) != tt.wantErr {
				t.Errorf("DelPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_multipath_FlushDevice(t *testing.T) {
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
		OSEXECCommandContextName: multipathTool,
		OSEXECCommandContextArgs: []string{"-f", mh.ValidDMName},
		OSEXECCmdOKReturn:        "ok",
	}

	tests := []struct {
		name        string
		fields      mpFields
		stateSetter func(fields mpFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "ok"
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: false,
		},
		{
			name:   "exec error",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args:    defaultArgs,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if err := mp.FlushDevice(tt.args.ctx, tt.args.deviceMapName); (err != nil) != tt.wantErr {
				t.Errorf("FlushDevice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_multipath_GetDMWWID(t *testing.T) {
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
		OSEXECCommandContextName: multipathDaemon,
		OSEXECCommandContextArgs: []string{"show", "maps"},
		OSEXECCmdOKReturn: fmt.Sprintf(
			`mpathdtpy dm-27 368ccf098001f6b4aed52e38178e00a9a
mpathdtpq %s %s
mpathdtps dm-15 368ccf09800e8113e0a8cc37fa467b1ad
mpathdtpu dm-17 368ccf09800563738e15a11c113e1968e`, mh.ValidDMName, mh.ValidWWID),
	}

	tests := []struct {
		name        string
		fields      mpFields
		args        args
		stateSetter func(fields mpFields)
		want        string
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    mh.ValidWWID,
			wantErr: false,
		},
		{
			name:   "exec error",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args:    defaultArgs,
			want:    "",
			wantErr: true,
		},
		{
			name:   "retry on timeout",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				call, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				call.Times(2)
				validResp := mocks.OSEXECCmdOKReturn
				mocks.OSEXECCmdOKReturn = "timeout"
				mocks.OSExecCmdOK(cmdMock)
				// retry
				mocks.OSEXECCmdOKReturn = validResp
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    mh.ValidWWID,
			wantErr: false,
		},
		{
			name:   "not found",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "mpathdtpu dm-17 368ccf09800563738e15a11c113e1968e"
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    "",
			wantErr: true,
		},
		{
			name:   "invalid data in response",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = " " + mh.ValidDMName + " "
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			got, err := mp.GetDMWWID(tt.args.ctx, tt.args.deviceMapName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDMWWID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetDMWWID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_multipath_IsDaemonRunning(t *testing.T) {
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
		OSEXECCommandContextName: multipathDaemon,
		OSEXECCommandContextArgs: []string{"show", "status"},
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
		fields      mpFields
		stateSetter func(fields mpFields)
		args        args
		want        bool
	}{
		{
			name:   "ok",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args: defaultArgs,
			want: true,
		},
		{
			name:   "exec error",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args: defaultArgs,
			want: false,
		},
		{
			name:   "invalid resp",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "error receiving packet"
				mocks.OSExecCmdOK(cmdMock)
			},
			args: defaultArgs,
			want: false,
		},
		{
			name:   "canceled context",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
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
			mp := &Multipath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if got := mp.IsDaemonRunning(tt.args.ctx); got != tt.want {
				t.Errorf("IsDaemonRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_multipath_runCommand(t *testing.T) {
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

	chrootFields := getDefaultMPFields(ctrl)
	chrootFields.chroot = chRoot

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: testCMD,
		OSEXECCommandContextArgs: testArgs,
		OSEXECCmdOKReturn:        validResp,
	}

	tests := []struct {
		name        string
		fields      mpFields
		stateSetter func(fields mpFields)
		args        args
		want        []byte
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    []byte(validResp),
			wantErr: false,
		},
		{
			name:   "exec err",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
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
			stateSetter: func(fields mpFields) {
				mocks.OSEXECCommandContextName = "chroot"
				mocks.OSEXECCommandContextArgs = append([]string{chRoot, testCMD}, testArgs...)
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			want:    []byte(validResp),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			got, err := mp.runCommand(tt.args.ctx, tt.args.command, tt.args.args)
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

func TestMultipath_RemoveDeviceFromWWIDSFile(t *testing.T) {
	type args struct {
		ctx  context.Context
		wwid string
	}
	ctx := context.Background()

	defaultArgs := args{ctx: ctx, wwid: mh.ValidWWID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: multipathDaemon,
		OSEXECCommandContextArgs: []string{"-w", mh.ValidWWID},
		OSEXECCmdOKReturn:        "ok",
	}

	tests := []struct {
		name        string
		fields      mpFields
		stateSetter func(fields mpFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "ok",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "ok"
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: false,
		},
		{
			name:   "exec error",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSExecCmdErr(cmdMock)
			},
			args:    defaultArgs,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				chroot:   tt.fields.chroot,
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			if err := mp.RemoveDeviceFromWWIDSFile(tt.args.ctx, tt.args.wwid); (err != nil) != tt.wantErr {
				t.Errorf("Multipath.RemoveDeviceFromWWIDSFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetMultipathNameAndPaths(t *testing.T) {

	type args struct {
		ctx  context.Context
		wwid string
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx, wwid: mh.ValidWWID}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: multipathDaemon,
		OSEXECCommandContextArgs: []string{"show", "paths", "raw", "format", "%d %w %m"},
		OSEXECCmdOKReturn: fmt.Sprintf("%s %s %s",
			`"path1" "path2"`, "360000970000120001598533030384533", "mpath-name"),
	}

	tests := []struct {
		name        string
		fields      mpFields
		stateSetter func(fields mpFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "Successful call",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "ok"
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			_, _, err := mp.GetMultipathNameAndPaths(tt.args.ctx, tt.args.wwid)

			if (err != nil) != tt.wantErr {
				t.Errorf("Multipath.GetMultipathNameAndPaths() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetMpathMinorByMpathName(t *testing.T) {

	type args struct {
		ctx   context.Context
		mpath string
	}

	ctx := context.Background()

	defaultArgs := args{ctx: ctx, mpath: "mpatha"}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := mh.MockHelper{
		Ctrl:                     ctrl,
		OSEXECCommandContextName: dmsetupTool,
		OSEXECCommandContextArgs: []string{"info", defaultArgs.mpath, "-C", "--noheadings", "--separator", ":", "-o", "name,minor"},
		OSEXECCmdOKReturn:        "mpatha:0",
	}

	tests := []struct {
		name        string
		fields      mpFields
		stateSetter func(fields mpFields)
		args        args
		wantErr     bool
	}{
		{
			name:   "Successful call",
			fields: getDefaultMPFields(ctrl),
			stateSetter: func(fields mpFields) {
				_, cmdMock := mocks.OSExecCommandContextOK(fields.osexec)
				mocks.OSEXECCmdOKReturn = "mpatha:0"
				mocks.OSExecCmdOK(cmdMock)
			},
			args:    defaultArgs,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &Multipath{
				osexec:   tt.fields.osexec,
				filePath: tt.fields.filePath,
			}
			tt.stateSetter(tt.fields)
			_, _, err := mp.GetMpathMinorByMpathName(tt.args.ctx, tt.args.mpath)

			if (err != nil) != tt.wantErr {
				t.Errorf("Multipath.GetMpathMinorByMpathName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
