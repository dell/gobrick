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

//go:generate ./generate_mock.sh

package gobrickutils

import (
	"errors"
	"regexp"
	"strings"
)

const (
	deviceNameRegEx = `[a-z]+[a-z0-9]*`
)

// ValidateDeviceName - Validates device names
func ValidateDeviceName(deviceName string) error {
	r := regexp.MustCompile(deviceNameRegEx)
	if !r.MatchString(deviceName) {
		return errors.New("error invalid device name")
	}
	return nil
}

// ValidateCommandInput - Validates command input
func ValidateCommandInput(input string) error {
	if strings.Contains(input, "|") {
		return errors.New("error invalid command input")
	}
	return nil
}
