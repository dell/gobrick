/*
 *
 * Copyright © 2021-2024 Dell Inc. or its subsidiaries. All Rights Reserved.
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

// Package gobrick is a generated GoMock package.
package gobrick

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/dell/gobrick/internal/logger"
	"github.com/dell/gobrick/internal/tracer"
	"github.com/dell/gobrick/pkg/scsi"
	log "github.com/sirupsen/logrus"
)

// RDMVolumeInfo has request info for RDM device
type RDMVolumeInfo struct {
	Targets []FCTargetInfo
	Lun     int
	WWN     string
}

// ConnectRDMVolume attach RDM to a VM
func (fc *FCConnector) ConnectRDMVolume(ctx context.Context, info RDMVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.ConnectVsphereVolume")()
	if err := fc.limiter.Acquire(ctx, 1); err != nil {
		return Device{}, errors.New("too many parallel operations. try later")
	}
	defer fc.limiter.Release(1)
	if err := fc.validateRDMVolumeInfo(ctx, info); err != nil {
		return Device{}, err
	}
	device, err := fc.connectRDM(ctx, info)
	if err == nil {
		return device, nil
	}
	logger.Error(ctx, "failed to connect RDM volume: %s, try to cleanup", err.Error())
	_ = fc.cleanRDMConnection(ctx, true, info)
	return Device{}, err
}

func (fc *FCConnector) cleanRDMConnection(ctx context.Context, force bool, info RDMVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "FCConnector.cleanRDMConnection")()
	wwn := fmt.Sprintf("3%s", info.WWN)
	devices, err := fc.scsi.GetDevicesByWWN(ctx, wwn)
	if err != nil || len(devices) == 0 {
		msg := "failed to get devices by WWN: " + wwn
		logger.Error(ctx, msg)
		return errors.New(msg)
	}
	logger.Info(ctx, "devices found: %s", devices)
	return fc.baseConnector.cleanDevices(ctx, force, devices, wwn)
}

func (fc *FCConnector) validateRDMVolumeInfo(ctx context.Context, info RDMVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "FCConnector.validateRDMVolumeInfo")()
	if len(info.Targets) == 0 {
		return errors.New("at least one FC target required")
	}
	if info.WWN == "" {
		return errors.New("wwn of RDM required")
	}
	return nil
}

func (fc *FCConnector) connectRDM(ctx context.Context, info RDMVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.connectRDM")()
	err := fc.rescanAllHosts(ctx)
	if err != nil {
		return Device{}, err
	}
	wwn := fmt.Sprintf("3%s", info.WWN)
	devices, err := fc.scsi.GetDevicesByWWN(ctx, wwn)
	if err != nil || len(devices) == 0 {
		msg := "failed to get devices by WWN: " + wwn
		logger.Error(ctx, msg)
		return Device{}, errors.New(msg)
	}
	var device string
	device, err = fc.waitSingleDevice(ctx, wwn, devices)
	if err != nil {
		return Device{}, err
	}

	if !fc.scsi.CheckDeviceIsValid(ctx, path.Join("/dev/", device)) {
		msg := "device invalid"
		logger.Error(ctx, msg)
		return Device{}, errors.New(msg)
	}
	d := Device{WWN: wwn, Name: device}
	return d, nil
}

func (fc *FCConnector) rescanAllHosts(ctx context.Context) error {
	defer tracer.TraceFuncCall(ctx, "FCConnector.rescanAllHosts")()
	hostsDir := "/sys/class/scsi_host"
	hostFiles, err := os.ReadDir(fmt.Sprintf("%s/", hostsDir))
	if err != nil {
		log.Errorf("rescanSCSIHOSTALL failed to read scsi_host dir, err: %s", err.Error())
		return err
	}
	log.Infof("found (%d) files in hostsDir (%s)", len(hostFiles), hostsDir)
	scsiHost := scsi.NewSCSI("")
	for host := 0; host < len(hostFiles); host++ {
		// at least one target port not found, do full scsi rescan
		hctl := scsi.HCTL{
			Host:    strconv.Itoa(host),
			Lun:     "-",
			Channel: "-",
			Target:  "-",
		}
		err := scsiHost.RescanSCSIHostByHCTL(ctx, hctl)
		if err != nil {
			log.Error(ctx, err.Error())
			continue
		}
	}
	return nil
}
