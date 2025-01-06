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
	"reflect"
	"testing"

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
			stateSetter: func(fields fcFields) {},
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
			stateSetter: func(fields fcFields) {},
			args: args{
				ctx: context.Background(),
				info: RDMVolumeInfo{
					Targets: []FCTargetInfo{
						{
							WWPN: "3000000000000000",
						},
					},
					Lun: 0,
					WWN: "",
				},
			},
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
