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

package gobrick

import (
	"time"

	"github.com/dell/gobrick/internal/logger"
	"github.com/dell/gobrick/internal/tracer"
)

// Device defines device info
type Device struct {
	WWN         string
	Name        string
	MultipathID string
	PowerpathID string
}

// Logger defines logger interface used
type Logger = logger.Logger

// Tracer defines tracing interface used
type Tracer = tracer.Tracer

// SetLogger set custom logger for gobrick
func SetLogger(customLogger Logger) {
	logger.SetLogger(customLogger)
}

// SetTracer set custom tracer for gobrick
func SetTracer(customTracer Tracer) {
	tracer.SetTracer(customTracer)
}

func setTimeouts(prop *time.Duration, value time.Duration, defaultVal time.Duration) {
	if value == 0 {
		*prop = defaultVal
	} else {
		*prop = value
	}
}
