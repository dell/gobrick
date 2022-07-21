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

package powerpath

import (
	"context"
	"fmt"
	"github.com/dell/gobrick/internal/logger"
	"github.com/dell/gobrick/internal/tracer"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/gobrick/pkg/utils"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	powerpathTool   = "powermt"
	powerpathDaemon = "powermt"
)

// NewPowerpath initializes powerpath struct
func NewPowerpath(chroot string) *Powerpath {
	mp := &Powerpath{
		chroot:   chroot,
		osexec:   &wrp.OSExecWrapper{},
		filePath: &wrp.FilepathWrapper{},
	}
	return mp
}

// Powerpath defines implementation of LimitedOSExec and LimitedFile path interfaces
type Powerpath struct {
	chroot   string
	osexec   wrp.LimitedOSExec
	filePath wrp.LimitedFilepath
}

// FlushDevice flush powerpath device. To prevent stucking always use context.WithTimeout
func (mp *Powerpath) FlushDevice(ctx context.Context) error {
	defer tracer.TraceFuncCall(ctx, "powerpath.FlushDevice")()
	return mp.flushDevice(ctx)
}

// IsDaemonRunning check if powerpath daemon running
func (mp *Powerpath) IsDaemonRunning(ctx context.Context) bool {
	defer tracer.TraceFuncCall(ctx, "powerpath.IsDaemonRunning")()
	return mp.isDaemonRunning(ctx)
}

// GetPowerPathDevices fetches dell-emc power path device name
func (mp *Powerpath) GetPowerPathDevices(ctx context.Context, devices []string) (string, error) {
	defer tracer.TraceFuncCall(ctx, "powerpath.GetPowerPathDevices")()
	return mp.getPowerPathDevices(ctx, devices)
}

func (mp *Powerpath) isDaemonRunning(ctx context.Context) bool {
	logger.Info(ctx, "powerpath - check daemon is running")
	resp, err := mp.runCommand(ctx, powerpathDaemon, []string{"version"})
	if err != nil {
		if ctx.Err() != nil {
			// it is safer to return true if context was canceled
			logger.Error(ctx, "powerpath - daemon state unknown")
			return true
		}
		logger.Error(ctx, "powerpath - failed to check daemon state: %s", err.Error())
		return false
	}
	if strings.Contains(string(resp), "error receiving packet") {
		logger.Info(ctx, "powerpath - daemon not started")
		return false
	}
	return true
}
func (mp *Powerpath) getPowerPathDevices(ctx context.Context, devices []string) (string, error) {
	logger.Info(ctx, "powerpath - trying to find powerpath emc name inside powerpath module")
	var match string
	for _, dev := range devices {
		deviceName := fmt.Sprintf("dev=%s", dev)
		out, err := mp.runCommand(ctx, powerpathDaemon, []string{"display", deviceName})
		if err != nil {
			log.Errorf("Error powermt display %s: %v", dev, err)
			return "", err
		}
		op := strings.Split(string(out), "\n")
		//L#0 Pseudo name=emcpowerc
		//L#1 Symmetrix ID=000197901586
		//L#3 Logical device ID=00002DCE
		//L#4 Device WWN=60000970000197901586533032444345
		for _, line := range op {
			if strings.Contains(line, "Pseudo name") {
				tokens := strings.Split(line, "=")
				// make sure we got two tokens
				if len(tokens) == 2 {
					if match == "" {
						match = tokens[1]
					} else if match != tokens[1] {
						log.Debugf("wrong parent for: %s", tokens[1])
						return "", fmt.Errorf("wrong parent for: %s", tokens[1])
					}
				}
				break
			}
		}
	}
	if strings.Contains(match, "emc") {
		return match, nil
	}
	return "", fmt.Errorf("power path device not found")
}

func (mp *Powerpath) flushDevice(ctx context.Context) error {
	logger.Info(ctx, "powerpath - start flush devices:")
	_, err := mp.runCommand(ctx, powerpathTool, []string{"check", "force"})
	return err
}

func (mp *Powerpath) runCommand(ctx context.Context, command string, args []string) ([]byte, error) {
	err := utils.ValidateCommandInput(command)
	if err != nil {
		return nil, err
	}

	if mp.chroot != "" {
		args = append([]string{mp.chroot, command}, args...)
		command = "chroot"
	}
	logger.
		Info(ctx, "powerpath command: %s args: %s", command, strings.Join(args, " "))
	data, err := mp.osexec.CommandContext(ctx, command, args...).CombinedOutput()
	logger.Debug(ctx, "powerpath command output: %s", string(data))
	return data, err
}
