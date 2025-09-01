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
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func loadTestData(filePathStr string, testConfig interface{}) {
	data, err := os.ReadFile(filepath.Clean(filePathStr))
	if err != nil {
		log.Printf("can't read file with test data, "+
			"please check that file exist: %s",
			err.Error())
	}
	err = json.Unmarshal(data, testConfig)
	if err != nil {
		log.Printf("can't unmarshal test data, check that data is in valid format: %s",
			err.Error())
	}
}

func parseLuns(luns string) []int {
	msg := "bad test data: \"luns\" field is invalid"
	var position int
	data := make(map[int]int)
	lunsSplt := strings.Split(luns, ",")
	for i := 0; i < len(lunsSplt); i++ {
		lunsRange := strings.Split(lunsSplt[i], "-")
		if len(lunsRange) == 2 {
			startLun, err := strconv.Atoi(lunsRange[0])
			if err != nil {
				panic(msg)
			}
			endLun, err := strconv.Atoi(lunsRange[1])
			if err != nil {
				panic(msg)
			}
			for j := startLun; j <= endLun; j++ {
				_, found := data[j]
				if !found {
					position++
					data[j] = position
				}
			}
		} else {
			lun, err := strconv.Atoi(lunsSplt[i])
			if err != nil {
				panic(msg)
			}
			_, found := data[lun]
			if !found {
				position++
				data[lun] = position
			}
		}
	}
	result := make([]int, len(data))
	for k, v := range data {
		result[v-1] = k
	}
	return result
}
