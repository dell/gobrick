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

// Package scsi is a generated GoMock package.
package scsi

import (
	"context"

	"github.com/dell/gobrick/pkg/scsi"
)

//go:generate ./generate_mock.sh

// SCSI defines methods for scsi operations
type SCSI interface {
	IsDeviceExist(ctx context.Context, device string) bool
	RescanSCSIHostByHCTL(ctx context.Context, h scsi.HCTL) error
	RescanSCSIDeviceByHCTL(ctx context.Context, h scsi.HCTL) error
	DeleteSCSIDeviceByHCTL(ctx context.Context, h scsi.HCTL) error
	DeleteSCSIDeviceByName(ctx context.Context, name string) error
	DeleteSCSIDeviceByPath(ctx context.Context, devPath string) error
	GetDeviceWWN(ctx context.Context, devices []string) (string, error)
	GetNVMEDeviceWWN(ctx context.Context, devices []string) (string, error)
	GetDevicesByWWN(ctx context.Context, wwn string) ([]string, error)
	GetDMDeviceByChildren(ctx context.Context, devices []string) (string, error)
	GetNVMEDMDeviceByChildren(ctx context.Context, devices []string) (string, error)
	GetNVMEMultipathDMName(device string, pattern string) ([]string, error)
	GetDMChildren(ctx context.Context, dmPath string) ([]string, error)
	CheckDeviceIsValid(ctx context.Context, device string) bool
	GetDeviceNameByHCTL(ctx context.Context, h scsi.HCTL) (string, error)
	WaitUdevSymlink(ctx context.Context, deviceName string, wwn string) error
	WaitUdevSymlinkNVMe(ctx context.Context, deviceName string, wwn string) error
	GetNVMESymlink(checkPath string) (string, error)
}
