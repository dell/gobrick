/*
 *
 * Copyright © 2020-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
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
package gobrick

import (
	"context"
	"errors"
	"reflect"
	"testing"

	mh "github.com/dell/gobrick/internal/mockhelper"
	"github.com/golang/mock/gomock"
)

func TestConnectRDMVolume(t *testing.T) {
	type args struct {
		ctx  context.Context
		info RDMVolumeInfo
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name        string
		fields      fcFields
		stateSetter func(fields fcFields)
		args        args
		want        Device
		wantErr     bool
	}{
		{
			name:        "Trying to connect without any target",
			fields:      getDefaultFCFields(ctrl),
			stateSetter: func(_ fcFields) {},
			args: args{
				ctx: context.Background(),
				info: RDMVolumeInfo{
					Targets: []FCTargetInfo{},
					Lun:     0,
					WWN:     "",
				},
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:        "Trying to connect without wwn",
			fields:      getDefaultFCFields(ctrl),
			stateSetter: func(_ fcFields) {},
			args: args{
				ctx: context.Background(),
				info: RDMVolumeInfo{
					Targets: []FCTargetInfo{
						{
							WWPN: "",
						},
					},
					Lun: 0,
					WWN: "",
				},
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "No device found",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{}, errors.New("No device found")).AnyTimes()
				fields.scsi.EXPECT().WaitUdevSymlink(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().CheckDeviceIsValid(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			},
			args: args{
				ctx: context.Background(),
				info: RDMVolumeInfo{
					Targets: []FCTargetInfo{
						{
							WWPN: "",
						},
					},
					Lun: 0,
					WWN: mh.ValidWWID,
				},
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "Device is invalid",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{mh.ValidDeviceName}, nil).AnyTimes()
				fields.scsi.EXPECT().WaitUdevSymlink(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().CheckDeviceIsValid(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
				fields.scsi.EXPECT().GetDMDeviceByChildren(gomock.Any(), gomock.Any()).Return(mh.ValidDeviceName, nil).AnyTimes()
				fields.multipath.EXPECT().FlushDevice(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().IsDeviceExist(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
				fields.multipath.EXPECT().RemoveDeviceFromWWIDSFile(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().DeleteSCSIDeviceByName(gomock.Any(), gomock.Any()).AnyTimes()
				fields.multipath.EXPECT().DelPath(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			},
			args: args{
				ctx: context.Background(),
				info: RDMVolumeInfo{
					Targets: []FCTargetInfo{
						{
							WWPN: "",
						},
					},
					Lun: 0,
					WWN: mh.ValidWWID,
				},
			},
			want:    Device{},
			wantErr: true,
		},
		{
			name:   "Successfully connected",
			fields: getDefaultFCFields(ctrl),
			stateSetter: func(fields fcFields) {
				fields.scsi.EXPECT().GetDevicesByWWN(gomock.Any(), gomock.Any()).Return([]string{mh.ValidDeviceName}, nil).AnyTimes()
				fields.scsi.EXPECT().WaitUdevSymlink(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				fields.scsi.EXPECT().CheckDeviceIsValid(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
			},
			args: args{
				ctx: context.Background(),
				info: RDMVolumeInfo{
					Targets: []FCTargetInfo{
						{
							WWPN: "",
						},
					},
					Lun: 0,
					WWN: mh.ValidWWID,
				},
			},
			want: Device{
				WWN:  "3" + mh.ValidWWID,
				Name: mh.ValidDeviceName,
			},
			wantErr: false,
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
			got, err := fc.ConnectRDMVolume(tt.args.ctx, tt.args.info)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConnectRDMVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ConnectRDMVolume() got = %v, want %v", got, tt.want)
			}
		})
	}
}
