package scsi

import (
	"context"
	"github.com/dell/gobrick/pkg/scsi"
)

//go:generate ./generate_mock.sh

type SCSI interface {
	IsDeviceExist(ctx context.Context, device string) bool
	RescanSCSIHostByHCTL(ctx context.Context, h scsi.HCTL) error
	RescanSCSIDeviceByHCTL(ctx context.Context, h scsi.HCTL) error
	DeleteSCSIDeviceByHCTL(ctx context.Context, h scsi.HCTL) error
	DeleteSCSIDeviceByName(ctx context.Context, name string) error
	DeleteSCSIDeviceByPath(ctx context.Context, devPath string) error
	GetDeviceWWN(ctx context.Context, devices []string) (string, error)
	GetDevicesByWWN(ctx context.Context, wwn string) ([]string, error)
	GetDMDeviceByChildren(ctx context.Context, devices []string) (string, error)
	GetDMChildren(ctx context.Context, dmPath string) ([]string, error)
	CheckDeviceIsValid(ctx context.Context, device string) bool
	GetDeviceNameByHCTL(ctx context.Context, h scsi.HCTL) (string, error)
	WaitUdevSymlink(ctx context.Context, deviceName string, wwn string) error
}
