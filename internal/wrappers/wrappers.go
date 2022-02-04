//go:generate ./generate_mock.sh

package wrappers

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dell/goiscsi"
)

type LimitedFileInfo interface {
	IsDir() bool
}

type LimitedFile interface {
	WriteString(s string) (n int, err error)
	Close() error
}

type LimitedOSExec interface {
	CommandContext(ctx context.Context, name string, arg ...string) LimitedOSExecCmd
}

type LimitedOSExecCmd interface {
	CombinedOutput() ([]byte, error)
}

type LimitedIOUtil interface {
	ReadFile(filename string) ([]byte, error)
}

type LimitedFilepath interface {
	Glob(pattern string) (matches []string, err error)
	EvalSymlinks(path string) (string, error)
}

type LimitedOS interface {
	OpenFile(name string, flag int, perm os.FileMode) (LimitedFile, error)
	Stat(name string) (LimitedFileInfo, error)
	IsNotExist(err error) bool
	Mkdir(name string, perm os.FileMode) error
	Remove(name string) error
}

type ISCSILib interface {
	GetInitiators(filename string) ([]string, error)
	PerformLogin(target goiscsi.ISCSITarget) error
	GetSessions() ([]goiscsi.ISCSISession, error)
	CreateOrUpdateNode(target goiscsi.ISCSITarget, options map[string]string) error
}

// wrappers

type OSExecWrapper struct{}

func (w *OSExecWrapper) CommandContext(ctx context.Context, name string, arg ...string) LimitedOSExecCmd {
	return exec.CommandContext(ctx, name, arg...)
}

type IOUTILWrapper struct{}

func (io *IOUTILWrapper) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename) // #nosec G304
}

type FilepathWrapper struct{}

func (io *FilepathWrapper) Glob(pattern string) (matches []string, err error) {
	return filepath.Glob(pattern)
}

func (io *FilepathWrapper) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

type OSWrapper struct{}

func (io *OSWrapper) OpenFile(name string, flag int, perm os.FileMode) (LimitedFile, error) {
	return os.OpenFile(filepath.Clean(name), flag, perm) // #nosec G304
}

func (io *OSWrapper) Stat(name string) (LimitedFileInfo, error) {
	return os.Stat(name)
}

func (io *OSWrapper) IsNotExist(err error) bool {
	return os.IsNotExist(err)
}

func (io *OSWrapper) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (io *OSWrapper) Remove(name string) error {
	return os.Remove(name)
}
