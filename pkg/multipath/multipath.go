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
	"strings"

	"github.com/dell/gobrick/internal/logger"
	"github.com/dell/gobrick/internal/tracer"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/gobrick/pkg/gobrickutils"
)

const (
	multipathTool   = "multipath"
	multipathDaemon = "multipathd"
	dmsetupTool     = "dmsetup"
	chroot          = "chroot"
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
	chroot   string
	osexec   wrp.LimitedOSExec
	filePath wrp.LimitedFilepath
}

func (mp *Multipath) GetMultipathNameAndPaths(ctx context.Context, wwid string) (string, []string, error) {
	defer tracer.TraceFuncCall(ctx, "multipath.GetMultipathNameAndPaths")()
	return mp.getMultipathNameAndPaths(ctx, wwid)
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

func (mp *Multipath) GetMpathMinorByMpathName(ctx context.Context, mpath string) (string, bool, error) {
	defer tracer.TraceFuncCall(ctx, "multipath.GetGetMpathMinorByMpathName")()
	return mp.getMpathMinorByMpathName(ctx, mpath)
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

func (mp *Multipath) getMultipathNameAndPaths(ctx context.Context, wwid string) (string, []string, error) {
	// Run the multipathd show paths command with raw format and grep for the wwn
	var output string
	var data []byte
	var err error

	if mp.chroot != "" {
		data, err = mp.osexec.CommandContext(ctx, chroot, mp.chroot, multipathDaemon, "show", "paths", "raw", "format", "%d %w %m").CombinedOutput()
	} else {
		data, err = mp.osexec.CommandContext(ctx, multipathDaemon, "show", "paths", "raw", "format", "%d %w %m").CombinedOutput()
	}
	if err != nil {
		return "", nil, err
	}
	// Split the output into lines
	output = string(data)

	// Initialize variables to store the mpath name and sd paths
	mpathName := ""
	sdPaths := []string{}

	// Iterate over the lines and filter based on the wwn
	// sdb 360000970000120001598533030384533 mpatha
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Split(line, " ")
		if len(fields) < 3 {
			continue
		}
		if strings.Contains(fields[1], wwid) {
			mpathName = fields[2]
			sdPaths = append(sdPaths, fields[0])
		}
	}
	// Return the mpath name and sd paths
	return mpathName, sdPaths, nil
}

// getMpathMinorByMpathName returns a minor number, existence of minor and error
// For a given mpath=mpathxb, this returns minor: dm-85, existence: true, err: nil
// For a non-existing mpath, this return minor: "", existence: false, err: nil
func (mp *Multipath) getMpathMinorByMpathName(ctx context.Context, mpath string) (string, bool, error) {
	var output string
	data, err := mp.runCommand(ctx, dmsetupTool, []string{"info", mpath, "-C", "--noheadings", "--separator", ":", "-o", "name,minor"})
	// Split the output into lines
	output = string(data)
	if err != nil {
		// if output contains "does not exist", return false without error
		if strings.Contains(output, "does not exist") {
			return "", false, nil
		}
		return "", false, err
	}
	// Initialize variables to store the mpath name and sd paths
	minor := ""

	// Iterate over the lines and filter based on the wwn
	fields := strings.Split(output, ":")
	if len(fields) > 2 {
		return minor, false, fmt.Errorf("unexpected output format")
	}

	minor = fmt.Sprintf("dm-%s", fields[1])
	return minor, true, nil
}

func (mp *Multipath) runCommand(ctx context.Context, command string, args []string) ([]byte, error) {
	err := gobrickutils.ValidateCommandInput(command)
	if err != nil {
		return nil, err
	}

	if mp.chroot != "" {
		args = append([]string{mp.chroot, command}, args...)
		command = chroot
	}
	logger.
		Info(ctx, "multipath command: %s args: %s", command, strings.Join(args, " "))
	data, err := mp.osexec.CommandContext(ctx, command, args...).CombinedOutput()
	logger.Debug(ctx, "multipath command output: %s", string(data))
	return data, err
}
