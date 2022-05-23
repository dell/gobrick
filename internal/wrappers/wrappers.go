//go:generate ./generate_mock.sh

package wrappers

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dell/goiscsi"
	"github.com/dell/gonvme"
)

// LimitedFileInfo defines limited file info interface
type LimitedFileInfo interface {
	IsDir() bool
}

// LimitedFile defines limited file interface
type LimitedFile interface {
	WriteString(s string) (n int, err error)
	Close() error
}

// LimitedOSExec defines limited os exec interface
type LimitedOSExec interface {
	CommandContext(ctx context.Context, name string, arg ...string) LimitedOSExecCmd
}

// LimitedOSExecCmd defines limited os exec command interface
type LimitedOSExecCmd interface {
	CombinedOutput() ([]byte, error)
}

// LimitedIOUtil defines limited ioutil interface
type LimitedIOUtil interface {
	ReadFile(filename string) ([]byte, error)
}

// LimitedFilepath defines limited file path interface
type LimitedFilepath interface {
	Glob(pattern string) (matches []string, err error)
	EvalSymlinks(path string) (string, error)
}

// LimitedOS defines limited os interface
type LimitedOS interface {
	OpenFile(name string, flag int, perm os.FileMode) (LimitedFile, error)
	Stat(name string) (LimitedFileInfo, error)
	IsNotExist(err error) bool
	Mkdir(name string, perm os.FileMode) error
	Remove(name string) error
}

// ISCSILib defines iscsi function spec
type ISCSILib interface {
	GetInitiators(filename string) ([]string, error)
	PerformLogin(target goiscsi.ISCSITarget) error
	GetSessions() ([]goiscsi.ISCSISession, error)
	CreateOrUpdateNode(target goiscsi.ISCSITarget, options map[string]string) error
}

// NVMe defines NVMe function spec
type NVMe interface {
	DiscoverNVMeTCPTargets(address string, login bool) ([]gonvme.NVMeTarget, error)
	DiscoverNVMeFCTargets(address string, login bool) ([]gonvme.NVMeTarget, error)
	GetInitiators(filename string) ([]string, error)
	NVMeTCPConnect(target gonvme.NVMeTarget, duplicateConnect bool) error
	NVMeFCConnect(target gonvme.NVMeTarget, duplicateConnect bool) error
	NVMeDisconnect(target gonvme.NVMeTarget) error
	GetSessions() ([]gonvme.NVMESession, error)
	ListNamespaceDevices() (map[gonvme.DevicePathAndNamespace][]string, error)
	GetNamespaceData(path string, namespaceID string) (string, string, error)
}

// wrappers

// OSExecWrapper contains implementation of LimitedOSExec interface
type OSExecWrapper struct{}

// CommandContext is a wrapper of exec.CommandContext
func (w *OSExecWrapper) CommandContext(ctx context.Context, name string, arg ...string) LimitedOSExecCmd {
	return exec.CommandContext(ctx, name, arg...)
}

// IOUTILWrapper contains implementation of LimitedIOUtil interface
type IOUTILWrapper struct{}

// ReadFile is a wrapper of ioutil.ReadFile
func (io *IOUTILWrapper) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename) // #nosec G304
}

// FilepathWrapper contains implementation of LimitedFilePath interface
type FilepathWrapper struct{}

// Glob is a wrapper of filepath.Glob
func (io *FilepathWrapper) Glob(pattern string) (matches []string, err error) {
	return filepath.Glob(pattern)
}

// EvalSymlinks is a wrapper of filepath.EvalSymlinks
func (io *FilepathWrapper) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

// OSWrapper contains implementation of LimitedOS interface
type OSWrapper struct{}

// OpenFile is a wrapper of os.OpenFile
func (io *OSWrapper) OpenFile(name string, flag int, perm os.FileMode) (LimitedFile, error) {
	return os.OpenFile(filepath.Clean(name), flag, perm) // #nosec G304
}

// Stat is a wrapper of os.Stat
func (io *OSWrapper) Stat(name string) (LimitedFileInfo, error) {
	return os.Stat(name)
}

// IsNotExist is a wrapper of os.IsNotExist
func (io *OSWrapper) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

// Mkdir is a wrapper of os.Mkdir
func (io *OSWrapper) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

// Remove is a wrapper of os.Remove
func (io *OSWrapper) Remove(name string) error {
	return os.Remove(name)
}
