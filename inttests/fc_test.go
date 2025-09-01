/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"log"
	"testing"
	"time"

	"github.com/dell/gobrick"
	"github.com/stretchr/testify/assert"
)

const (
	fileWithFCTestData = "fc_test_data.json"
)

type fcTestSettings struct {
	Params  gobrick.FCConnectorParams
	Volumes []fcVolumesTestData
}

type fcVolumesTestData struct {
	Targets []gobrick.FCTargetInfo
	Luns    string
}

var parsedFCVolumeInfo []gobrick.FCVolumeInfo

var fcTestConfig = fcTestSettings{}

var fcConnector *gobrick.FCConnector

var skipFCtests bool

func init() {
	loadTestData(fileWithFCTestData, &fcTestConfig)
	if len(fcTestConfig.Volumes) == 0 {
		log.Printf("Config for FC tests is invalid, skip FC tests")
		skipFCtests = true
	}
	for _, v := range fcTestConfig.Volumes {
		parsedLuns := parseLuns(v.Luns)
		for _, lun := range parsedLuns {
			parsedFCVolumeInfo = append(parsedFCVolumeInfo, gobrick.FCVolumeInfo{Lun: lun, Targets: v.Targets})
		}
	}
	fcConnector = gobrick.NewFCConnector(fcTestConfig.Params)
}

func TestFC_ConnectDisconnect(t *testing.T) {
	if skipFCtests {
		t.SkipNow()
	}
	ctx, cFunc := context.WithTimeout(context.Background(), time.Second*60)
	defer cFunc()
	device, err := fcConnector.ConnectVolume(ctx, parsedFCVolumeInfo[0])
	assert.Nil(t, err)
	assert.NotEmpty(t, device.Name, device.WWN)
}
