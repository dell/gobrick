package powerpath

import "context"

//go:generate ./generate_mock.sh

// Powerpath defines methods for powerpath related operations
type Powerpath interface {
	FlushDevice(ctx context.Context) error
	IsDaemonRunning(ctx context.Context) bool
	GetPowerPathDevices(ctx context.Context, devices []string) (string, error)
}
