package multipath

import "context"

//go:generate ./generate_mock.sh

// Multipath defines methods for multipath related operations
type Multipath interface {
	AddWWID(ctx context.Context, wwid string) error
	AddPath(ctx context.Context, path string) error
	DelPath(ctx context.Context, path string) error
	FlushDevice(ctx context.Context, deviceMapName string) error
	IsDaemonRunning(ctx context.Context) bool
	GetDMWWID(ctx context.Context, deviceMapName string) (string, error)
}
