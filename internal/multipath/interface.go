/*
Copyright © 2020-2022 Dell Inc. or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package multipath is a generated GoMock package.
package multipath

import "context"

//go:generate ./generate_mock.sh

// Multipath defines methods for multipath related operations
type Multipath interface {
	AddWWID(ctx context.Context, wwid string) error
	AddPath(ctx context.Context, path string) error
	DelPath(ctx context.Context, path string) error
	FlushDevice(ctx context.Context, deviceMapName string) error
	RemoveDeviceFromWWIDSFile(ctx context.Context, wwid string) error
	IsDaemonRunning(ctx context.Context) bool
	GetDMWWID(ctx context.Context, deviceMapName string) (string, error)
	GetMultipathNameAndPaths(ctx context.Context, wwid string) (string, []string, error)
	GetMpathMinorByMpathName(ctx context.Context, mpath string) (string, bool, error)
}
