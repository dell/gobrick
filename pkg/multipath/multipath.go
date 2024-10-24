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

package multipath

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/dell/gobrick/internal/logger"
	"github.com/dell/gobrick/internal/tracer"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/gobrick/pkg/utils"
)

const (
	multipathTool   = "multipath"
	multipathDaemon = "multipathd"
)

// NewMultipath initializes multipath struct
func NewMultipath(chroot string) *Multipath {
	mp := &Multipath{
		chroot:   chroot,
		osexec:   &wrp.OSExecWrapper{},
		filePath: &wrp.FilepathWrapper{},
	}
	return mp
}

// Multipath defines implementation of LimitedOSExec and LimitedFile path interfaces
type Multipath struct {
	chroot string

	osexec   wrp.LimitedOSExec
	filePath wrp.LimitedFilepath
}

// AddWWID add wwid to the list of know multipath wwids.
//
//	This has the effect of multipathd being willing to create a dm for a
//	multipath even when there's only 1 device.
func (mp *Multipath) AddWWID(ctx context.Context, wwid string) error {
	defer tracer.TraceFuncCall(ctx, "multipath.AddWWID")()
	return mp.addWWID(ctx, wwid)
}

// AddPath add a path to multipathd for monitoring.
//
//	This has the effect of multipathd checking an already checked device
//	for multipath.
//	Together with `multipath_add_wwid` we can create a multipath when
//	there's only 1 path.
func (mp *Multipath) AddPath(ctx context.Context, path string) error {
	defer tracer.TraceFuncCall(ctx, "multipath.AddPath")()
	return mp.changePath(ctx, "add", path)
}

// DelPath remove a path from multipathd mapping
func (mp *Multipath) DelPath(ctx context.Context, path string) error {
	defer tracer.TraceFuncCall(ctx, "multipath.DelPath")()
	return mp.changePath(ctx, "del", path)
}

// FlushDevice flush multipath device. To prevent stucking always use context.WithTimeout
func (mp *Multipath) FlushDevice(ctx context.Context, deviceMapName string) error {
	defer tracer.TraceFuncCall(ctx, "multipath.FlushDevice")()
	return mp.flushDevice(ctx, deviceMapName)
}

// RemoveDeviceFromWWIDSFile removes multipath device from /etc/multipath/wwids file.
func (mp *Multipath) RemoveDeviceFromWWIDSFile(ctx context.Context, wwid string) error {
	defer tracer.TraceFuncCall(ctx, "multipath.RemoveDeviceFromWWIDSFile")()
	return mp.removeDeviceFromWWIDSFile(ctx, wwid)
}

// IsDaemonRunning check if multipath daemon running
func (mp *Multipath) IsDaemonRunning(ctx context.Context) bool {
	defer tracer.TraceFuncCall(ctx, "multipath.IsDaemonRunning")()
	return mp.isDaemonRunning(ctx)
}

// GetDMWWID read DeviceMapper WWID from multipathd
func (mp *Multipath) GetDMWWID(ctx context.Context, deviceMapName string) (string, error) {
	defer tracer.TraceFuncCall(ctx, "multipath.GetDMWWID")()
	return mp.getDMWWID(ctx, deviceMapName)
}

func returnError(err error, resp []byte) error {
	if err != nil {
		return err
	}
	return fmt.Errorf("unexpected result: %s", string(resp))
}

func (mp *Multipath) addWWID(ctx context.Context, wwid string) error {
	logger.Info(ctx, "multipath - add wwid: %s", wwid)
	resp, err := mp.runCommand(ctx, multipathTool, []string{"-a", wwid})
	if strings.Contains(string(resp), fmt.Sprintf("wwid '%s' added", wwid)) {
		return nil
	}
	return returnError(err, resp)
}

func (mp *Multipath) changePath(ctx context.Context, action string, path string) error {
	logger.Info(ctx, "multipath - %s path: %s", action, path)
	resp, err := mp.runCommand(ctx, multipathDaemon, []string{action, "path", path})
	if strings.TrimSpace(string(resp)) == "ok" {
		return nil
	}
	return returnError(err, resp)
}

func (mp *Multipath) isDaemonRunning(ctx context.Context) bool {
	logger.Info(ctx, "multipath - check daemon is running")
	resp, err := mp.runCommand(ctx, multipathDaemon, []string{"show", "status"})
	if err != nil {
		if ctx.Err() != nil {
			// it is safer to return true if context was canceled
			logger.Error(ctx, "multipath - daemon state unknown")
			return true
		}
		logger.Error(ctx, "multipath - failed to check daemon state: %s", err.Error())
		return false
	}
	if strings.Contains(string(resp), "error receiving packet") {
		logger.Info(ctx, "multipath - daemon not started")
		return false
	}
	return true
}

func (mp *Multipath) flushDevice(ctx context.Context, deviceMapName string) error {
	logger.Info(ctx, "multipath - start flush dm: %s", deviceMapName)
	_, err := mp.runCommand(ctx, multipathTool, []string{"-f", deviceMapName})
	return err
}

func (mp *Multipath) removeDeviceFromWWIDSFile(ctx context.Context, wwid string) error {
	logger.Info(ctx, "multipath - remove from wwids file wwid: %s", wwid)
	_, err := mp.runCommand(ctx, multipathTool, []string{"-w", wwid})
	return err
}

func (mp *Multipath) getDMWWID(ctx context.Context, deviceMapName string) (string, error) {
	logger.Info(ctx, "multipath - resolve WWID: %s", deviceMapName)
	var dataRead bool
	var output string
	for i := 0; i < 3; i++ {
		data, err := mp.runCommand(ctx, multipathDaemon, []string{"show", "maps"})
		if err != nil {
			return "", err
		}
		output = string(data)
		if strings.TrimSpace(output) != "timeout" {
			dataRead = true
			break
		}
		logger.Debug(ctx, "timeout reading from multipathd, retry")
	}
	if dataRead {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, " "+deviceMapName+" ") {
				flds := strings.Fields(line)
				if len(flds) == 3 {
					return flds[2], nil
				}
				logger.Debug(ctx, "unexpected data: %s", line)
			}
		}
	}
	msg := fmt.Sprintf("WWN for DM %s not found", deviceMapName)
	return "", errors.New(msg)
}

func (mp *Multipath) runCommand(ctx context.Context, command string, args []string) ([]byte, error) {
	err := utils.ValidateCommandInput(command)
	if err != nil {
		return nil, err
	}

	if mp.chroot != "" {
		args = append([]string{mp.chroot + os.Getenv("MP_CHROOT_SUFFIX"), command}, args...)
		command = "chroot"
	}
	logger.
		Info(ctx, "multipath command: %s args: %s", command, strings.Join(args, " "))
	data, err := mp.osexec.CommandContext(ctx, command, args...).CombinedOutput()
	logger.Debug(ctx, "multipath command output: %s", string(data))
	return data, err
}
