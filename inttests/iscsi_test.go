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

package inttests

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/dell/gobrick"
	"github.com/dell/gobrick/pkg/multipath"
	"github.com/stretchr/testify/assert"
)

const (
	fileWithISCSITestData = "iscsi_test_data.json"
)

var iscsiConnector *gobrick.ISCSIConnector

type iscsiTestSettings struct {
	Params  gobrick.ISCSIConnectorParams
	Volumes []iscsiVolumesTestData
}

type iscsiVolumesTestData struct {
	Targets []gobrick.ISCSITargetInfo
	Luns    string
}

var skipISCSItests bool

var iscsiTestConfig = iscsiTestSettings{}

var parsedISCSIVolumeInfo []gobrick.ISCSIVolumeInfo

func init() {
	loadTestData(fileWithISCSITestData, &iscsiTestConfig)
	if len(iscsiTestConfig.Volumes) == 0 {
		log.Printf("Config for iSCSI tests is invalid, skip iSCSI tests")
		skipISCSItests = true
	}
	for _, v := range iscsiTestConfig.Volumes {
		parsedLuns := parseLuns(v.Luns)
		for _, lun := range parsedLuns {
			parsedISCSIVolumeInfo = append(parsedISCSIVolumeInfo, gobrick.ISCSIVolumeInfo{Lun: lun, Targets: v.Targets})
		}
	}
	iscsiConnector = gobrick.NewISCSIConnector(iscsiTestConfig.Params)
}

func TestISCSI_ConnectDisconnect(t *testing.T) {
	if skipISCSItests {
		t.SkipNow()
	}
	wg := sync.WaitGroup{}
	count := 200
	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			ctx := context.Background()
			for {
				ctx, cFunc := context.WithTimeout(ctx, time.Second*60)
				device, err := iscsiConnector.ConnectVolume(ctx, parsedISCSIVolumeInfo[i])
				cFunc()
				if err != nil {
					log.Printf("ID: %d ERROR: %s", i, err.Error())
				} else {
					log.Printf("ID: %d Device Name: %s", i, device.Name)
					break
				}
				log.Printf("ID: %d RETRY!", i)
			}
			// assert.Nil(t, err)
			// assert.NotEmpty(t, device.Name, device.WWN)
			wg.Done()
		}()
	}
	wg.Wait()
	// assert.Nil(t, iscsiConnector.DisconnectVolume(ctx, volume))
}

func TestISCSI_ConnectDisconnectByName(t *testing.T) {
	if skipISCSItests {
		t.SkipNow()
	}
	ctx, cFunc := context.WithTimeout(context.Background(), time.Second*60)
	defer cFunc()
	volume := parsedISCSIVolumeInfo[0]
	device, err := iscsiConnector.ConnectVolume(ctx, volume)
	assert.Nil(t, err)
	assert.NotEmpty(t, device.Name, device.WWN)
	assert.Nil(t, iscsiConnector.DisconnectVolumeByDeviceName(ctx, device.Name))
}

func Test_readWWID(_ *testing.T) {
	m := multipath.NewMultipath("/noderoot")
	fmt.Println(m.GetDMWWID(context.Background(), "dm-80"))
}

func TestParseLuns(t *testing.T) {
	testCases := []struct {
		input    string
		expected []int
	}{
		{
			input:    "1-3,5,7-9",
			expected: []int{1, 2, 3, 5, 7, 8, 9},
		},
		{
			input:    "2,4,6",
			expected: []int{2, 4, 6},
		},
		{
			input:    "1,3,5,7,9",
			expected: []int{1, 3, 5, 7, 9},
		},
		{
			input:    "1-3,5,7-9,10-12",
			expected: []int{1, 2, 3, 5, 7, 8, 9, 10, 11, 12},
		},
		{
			input:    "1-3,5,7-9,10-12,14",
			expected: []int{1, 2, 3, 5, 7, 8, 9, 10, 11, 12, 14},
		},
		{
			input:    "1-3,5,7-9,10-12,14,15-17",
			expected: []int{1, 2, 3, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17},
		},
		{
			input:    "1-3,5,7-9,10-12,14,15-17,19-21",
			expected: []int{1, 2, 3, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 19, 20, 21},
		},
		{
			input:    "1-3,5,7-9,10-12,14,15-17,19-21,23",
			expected: []int{1, 2, 3, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 19, 20, 21, 23},
		},
		{
			input:    "1-3,5,7-9,10-12,14,15-17,19-21,23,25",
			expected: []int{1, 2, 3, 5, 7, 8, 9, 10, 11, 12, 14, 15, 16, 17, 19, 20, 21, 23, 25},
		},
	}

	for _, tc := range testCases {
		result := parseLuns(tc.input)
		assert.Equal(t, tc.expected, result, "Unexpected result for input: %s", tc.input)
	}
}
