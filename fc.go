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

package gobrick

import (
	"context"
	"errors"
	"fmt"
	"math"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/dell/gobrick/internal/logger"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intpowerpath "github.com/dell/gobrick/internal/powerpath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	"github.com/dell/gobrick/internal/tracer"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/gobrick/pkg/multipath"
	"github.com/dell/gobrick/pkg/powerpath"
	"github.com/dell/gobrick/pkg/scsi"
	"golang.org/x/sync/semaphore"
)

const (
	fcWaitDeviceRegisterTimeoutDefault = time.Second * 15
	fcMaxParallelOperationsDefault     = 5
)

// FCConnectorParams options for FCConnector
type FCConnectorParams struct {
	// run commands inside chroot
	Chroot string

	// how long wait for DM appear
	WaitDeviceRegisterTimeout time.Duration
	// timeout for multipath flush command
	MultipathFlushTimeout time.Duration
	// timeout for each multipath flush retry
	MultipathFlushRetryTimeout time.Duration
	// how many retries for multipath flush
	MultipathFlushRetries int

	// how many parallel operations allowed
	MaxParallelOperations int
}

// NewFCConnector create new FCConnector
func NewFCConnector(params FCConnectorParams) *FCConnector {
	mp := multipath.NewMultipath(params.Chroot)
	pp := powerpath.NewPowerpath(params.Chroot)
	s := scsi.NewSCSI(params.Chroot)

	conn := &FCConnector{
		multipath: mp,
		powerpath: pp,
		scsi:      s,
		filePath:  &wrp.FilepathWrapper{},
		os:        &wrp.OSWrapper{},
		baseConnector: newBaseConnector(mp, pp, s,
			baseConnectorParams{
				MultipathFlushTimeout:      params.MultipathFlushTimeout,
				MultipathFlushRetryTimeout: params.MultipathFlushRetryTimeout,
				MultipathFlushRetries:      params.MultipathFlushRetries,
			}),
	}

	setTimeouts(&conn.waitDeviceRegisterTimeout, params.WaitDeviceRegisterTimeout,
		fcWaitDeviceRegisterTimeoutDefault)

	maxParallelOperations := params.MaxParallelOperations
	if maxParallelOperations == 0 {
		maxParallelOperations = fcMaxParallelOperationsDefault
	}
	conn.limiter = semaphore.NewWeighted(int64(maxParallelOperations))

	return conn
}

// FCTargetInfo holds information about remote FC ports
type FCTargetInfo struct {
	WWPN string
}

// FCVolumeInfo connection request for volume
type FCVolumeInfo struct {
	Targets []FCTargetInfo
	Lun     int
}

// FCHBA holds information about host local FC ports
type FCHBA struct {
	PortName   string
	NodeName   string
	HostDevice string
}

// FCConnector connector for FC transport
type FCConnector struct {
	baseConnector *baseConnector
	multipath     intmultipath.Multipath
	powerpath     intpowerpath.Powerpath
	scsi          intscsi.SCSI

	// wrappers
	filePath wrp.LimitedFilepath
	os       wrp.LimitedOS

	limiter *semaphore.Weighted

	waitDeviceRegisterTimeout time.Duration
}

// ConnectVolume attach volume to a node
func (fc *FCConnector) ConnectVolume(ctx context.Context, info FCVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.ConnectVolume")()
	if err := fc.limiter.Acquire(ctx, 1); err != nil {
		return Device{}, errors.New("too many parallel operations. try later")
	}
	defer fc.limiter.Release(1)
	if err := fc.validateFCVolumeInfo(ctx, info); err != nil {
		return Device{}, err
	}

	hbas, err := fc.getFCHBASInfo(ctx)
	if err != nil {
		logger.Error(ctx, "failed to get FC hbas info: %s", err.Error())
		return Device{}, err
	}
	if len(hbas) == 0 {
		msg := "FC HBAs not found. FC is not supported on this host"
		logger.Error(ctx, msg)
		return Device{}, errors.New(msg)
	}
	device, err := fc.connectDevice(ctx, hbas, info)
	if err == nil {
		return device, nil
	}
	logger.Error(ctx, "failed to connect FC volume: %s, try to cleanup", err.Error())
	_ = fc.cleanConnection(ctx, true, info)
	return Device{}, err
}

// DisconnectVolume disconnects volume from a node by full volume request
func (fc *FCConnector) DisconnectVolume(ctx context.Context, info FCVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "FCConnector.DisconnectVolume")()
	if err := fc.limiter.Acquire(ctx, 1); err != nil {
		return errors.New("too many parallel operations. try later")
	}
	defer fc.limiter.Release(1)
	return fc.cleanConnection(ctx, false, info)
}

// DisconnectVolumeByDeviceName disconnects volume from a node by device name
func (fc *FCConnector) DisconnectVolumeByDeviceName(ctx context.Context, name string) error {
	defer tracer.TraceFuncCall(ctx, "FCConnector.DisconnectVolumeByDeviceName")()
	if err := fc.limiter.Acquire(ctx, 1); err != nil {
		return errors.New("too many parallel operations. try later")
	}
	defer fc.limiter.Release(1)
	return fc.baseConnector.disconnectDevicesByDeviceName(ctx, name)
}

// GetInitiatorPorts return information about nodes local FC ports
func (fc *FCConnector) GetInitiatorPorts(ctx context.Context) ([]string, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.GetInitiatorPorts")()
	logger.Info(ctx, "get initiator name")
	hbas, err := fc.getFCHBASInfo(ctx)
	if err != nil {
		logger.Error(ctx, "failed to read initiator ports names: %s", err.Error())
	}
	var ports []string
	for _, hba := range hbas {
		ports = append(ports, hba.PortName)
	}
	logger.Info(ctx, "initiator FC ports names are: %s", ports)
	return ports, nil
}

func (fc *FCConnector) cleanConnection(ctx context.Context, force bool, info FCVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "FCConnector.cleanConnection")()
	hbas, err := fc.getFCHBASInfo(ctx)
	if err != nil {
		logger.Error(ctx, "failed to get FC hbas info: %s", err.Error())
		return err
	}
	var devices []string
	for _, hba := range hbas {
		_, hctls, err := fc.findHCTLsForFCHBA(ctx, hba, info)
		if err != nil {
			logger.Error(ctx, err.Error())
		}
		if len(hctls) == 0 {
			continue
		}
		for _, hctl := range hctls {
			if hctl.IsFullInfo() {
				device, err := fc.scsi.GetDeviceNameByHCTL(ctx, hctl)
				if err != nil {
					logger.Error(ctx, err.Error())
					continue
				}
				devices = append(devices, device)
			}
		}
	}
	logger.Info(ctx, "devices found: %s", devices)
	return fc.baseConnector.cleanDevices(ctx, force, devices, "")
}

func (fc *FCConnector) validateFCVolumeInfo(ctx context.Context, info FCVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "FCConnector.validateFCVolumeInfo")()
	if len(info.Targets) == 0 {
		return errors.New("at least one FC target required")
	}
	for _, t := range info.Targets {
		if t.WWPN == "" {
			return errors.New("invalid target info")
		}
	}
	return nil
}

func (fc *FCConnector) connectDevice(
	ctx context.Context, hbas []FCHBA, info FCVolumeInfo,
) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.connectDevice")()
	wwn, err := fc.waitForDeviceWWN(ctx, hbas, info)
	if err != nil {
		return Device{}, err
	}
	devices, err := fc.scsi.GetDevicesByWWN(ctx, wwn)
	if err != nil || len(devices) == 0 {
		msg := "failed to get devices by WWN: " + wwn
		logger.Error(ctx, msg)
		return Device{}, errors.New(msg)
	}

	var device string
	var isMP bool
	if fc.powerpath.IsDaemonRunning(ctx) {
		device, err = fc.waitPowerpathDevice(ctx, wwn, devices)
		if err != nil {
			msg := "failed to find powerpath device"
			logger.Error(ctx, msg)
			return Device{}, errors.New(msg)
		}
	} else if !fc.multipath.IsDaemonRunning(ctx) {
		device, err = fc.waitSingleDevice(ctx, wwn, devices)
		if err != nil {
			return Device{}, err
		}
	} else {
		isMP = true
		device, err = fc.waitMultipathDevice(ctx, wwn, devices)
		if err != nil {
			msg := "failed to find multipath device"
			logger.Error(ctx, msg)
			return Device{}, errors.New(msg)
		}
	}
	if !fc.scsi.CheckDeviceIsValid(ctx, path.Join("/dev/", device)) {
		msg := "multipath device was found but failed to read data from it"
		logger.Error(ctx, msg)
		return Device{}, errors.New(msg)
	}
	d := Device{WWN: wwn, Name: device}
	if isMP {
		d.MultipathID = wwn
	}
	return d, nil
}

func (fc *FCConnector) waitSingleDevice(ctx context.Context, wwn string, devices []string) (string, error) {
	for i := 0; i < int(fc.waitDeviceRegisterTimeout.Seconds()); i++ {
		select {
		case <-ctx.Done():
			return "", errors.New("waitDevice canceled")
		default:
		}
		for _, d := range devices {
			if err := fc.scsi.WaitUdevSymlink(ctx, d, wwn); err == nil {
				return d, nil
			}
		}
		time.Sleep(time.Second)
	}
	msg := fmt.Sprintf("timeout waiting device for wwn %s", wwn)
	logger.Info(ctx, msg)
	return "", errors.New(msg)
}

func (fc *FCConnector) waitMultipathDevice(
	ctx context.Context, wwn string, devices []string,
) (string, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.waitMultipathDevice")()
	err := fc.multipath.AddWWID(ctx, wwn)
	if err != nil {
		return "", err
	}

	for _, d := range devices {
		devPath := path.Join("/dev/", d)
		if err := fc.multipath.AddPath(ctx, devPath); err != nil {
			logger.Info(ctx, err.Error())
		}
	}

	var mpath string
	for i := 0; i < int(fc.waitDeviceRegisterTimeout.Seconds()); i++ {
		select {
		case <-ctx.Done():
			return "", errors.New("waitMultipathDevice canceled")
		default:
		}
		resp, err := fc.scsi.GetDMDeviceByChildren(ctx, devices)
		if err == nil {
			if err := fc.scsi.WaitUdevSymlink(ctx, resp, wwn); err == nil {
				mpath = resp
				break
			}
		}
		time.Sleep(time.Second)
	}
	if mpath == "" {
		msg := fmt.Sprintf("multipath device for WWN %s not found", wwn)
		logger.Error(ctx, msg)
		return "", errors.New(msg)
	}
	logger.Info(ctx, "multipath device for WWN %s found: %s", wwn, mpath)
	return mpath, nil
}

func (fc *FCConnector) waitPowerpathDevice(
	ctx context.Context, wwn string, devices []string,
) (string, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.waitPowerpathDevice")()
	var ppath string
	for i := 0; i < int(fc.waitDeviceRegisterTimeout.Seconds()); i++ {
		select {
		case <-ctx.Done():
			return "", errors.New("PowerpathDevice canceled")
		default:
		}
		resp, err := fc.powerpath.GetPowerPathDevices(ctx, devices)
		if err == nil {
			if err := fc.scsi.WaitUdevSymlink(ctx, resp, wwn); err == nil {
				ppath = resp
				break
			}
		}
		time.Sleep(time.Second)
	}
	if ppath == "" {
		msg := fmt.Sprintf("powerpath device for WWN %s not found", wwn)
		logger.Error(ctx, msg)
		return "", errors.New(msg)
	}
	logger.Info(ctx, "powerpath device for WWN %s found: %s", wwn, ppath)
	return ppath, nil
}

func (fc *FCConnector) waitForDeviceWWN(
	ctx context.Context, hbas []FCHBA, info FCVolumeInfo,
) (string, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.waitForDeviceWWN")()
	var numRescans, secondsNextScan int

	numRescans = 0
	secondsNextScan = 1

	doScans := true

	for doScans {
		var hctlsToRescan []scsi.HCTL
		var hctlsToDiscover []scsi.HCTL
		for _, hba := range hbas {
			htr, htd, err := fc.findHCTLsForFCHBA(ctx, hba, info)
			if err != nil {
				return "", err
			}
			hctlsToRescan = append(hctlsToRescan, htr...)
			hctlsToDiscover = append(hctlsToDiscover, htd...)
		}
		if secondsNextScan <= 0 {
			numRescans++
			for _, hctl := range hctlsToRescan {
				err := fc.scsi.RescanSCSIHostByHCTL(ctx, hctl)
				if err != nil {
					logger.Error(ctx, err.Error())
					continue
				}
			}
			secondsNextScan = int(math.Pow(float64(numRescans+2), 2))
		}
		for _, hctl := range hctlsToDiscover {
			if !hctl.IsFullInfo() {
				logger.Debug(ctx, "HCTL incomplete, skip device resolving")
				continue
			}
			d, err := fc.scsi.GetDeviceNameByHCTL(ctx, hctl)
			if err != nil {
				logger.Debug(ctx, "device for HCTL %s not found", hctl)
				continue
			}
			logger.Info(ctx, "device found: %s", d)
			if numRescans == 0 {
				logger.Debug(ctx, "device %s found without scanning, "+
					"try to refresh device information", d)
				err := fc.scsi.RescanSCSIDeviceByHCTL(ctx, hctl)
				if err != nil {
					logger.Error(ctx, err.Error())
				}
			}

			if fc.scsi.CheckDeviceIsValid(ctx, path.Join("/dev/", d)) {
				logger.Debug(ctx, "device %s is valid", d)
				wwn, err := fc.scsi.GetDeviceWWN(ctx, []string{d})
				if err != nil {
					logger.Error(ctx, "failed to get %s WWN: %s", d, err.Error())
					continue
				}
				logger.Info(ctx, "FC wwn found: %s", wwn)
				return wwn, nil
			}
			logger.Debug(ctx, "device %s is invalid", d)
		}
		select {
		case <-ctx.Done():
			doScans = false
		default:
			doScans = numRescans <= 3
		}
		if doScans {
			if numRescans > 0 {
				time.Sleep(time.Second)
			}
			secondsNextScan--
		}
	}
	msg := "wwn for FC device not found"
	logger.Error(ctx, msg)
	return "", errors.New(msg)
}

func (fc *FCConnector) findHCTLsForFCHBA(
	ctx context.Context, hba FCHBA, info FCVolumeInfo,
) ([]scsi.HCTL, []scsi.HCTL, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.findHCTLsForFCHBA")()
	hostDev := hba.HostDevice
	if len(hostDev) > 4 {
		hostDev = hostDev[4:]
	}
	pattern := fmt.Sprintf("/sys/class/fc_transport/target%s:*", hostDev)
	matches, err := fc.filePath.Glob(pattern)
	if err != nil {
		msg := fmt.Sprintf("HBA: %s failed to match FC target path: %s", hba, err.Error())
		logger.Error(ctx, msg)
		return nil, nil, errors.New(msg)
	}
	targetMap := make(map[string]string)
	for _, m := range matches {
		data, err := fc.os.ReadFile(path.Join(m, "port_name"))
		if err != nil {
			msg := fmt.Sprintf("HBA: %s failed to read port_name for FC target: %s", hba, err.Error())
			logger.Error(ctx, msg)
			return nil, nil, errors.New(msg)
		}
		targetMap[m] = strings.Replace(
			strings.TrimSpace(string(data)), "0x", "", 1)
	}
	lun := strconv.FormatInt(int64(info.Lun), 10)

	var hctlsToDiscover []scsi.HCTL

	var needFullRescan bool

	for _, fcTarget := range info.Targets {
		var tgtFound bool
		for pathMatch, portName := range targetMap {
			if portName != fcTarget.WWPN {
				continue
			}
			tgtFound = true
			targetPathPart := strings.Split(pathMatch, "/")[4]
			ct := strings.Split(targetPathPart, ":")
			if len(ct) >= 3 {
				hctl := scsi.HCTL{
					Host:    hostDev,
					Lun:     lun,
					Channel: ct[1], Target: ct[2],
				}
				logger.Info(ctx, "HBA: %s FC device HCTL: %s", hba, hctl)
				hctlsToDiscover = append(hctlsToDiscover, hctl)
			}
		}
		if !tgtFound {
			needFullRescan = true
		}
	}
	var hctlsToRescan []scsi.HCTL
	if needFullRescan {
		// at least one target port not found, do full scsi rescan
		hctlsToRescan = []scsi.HCTL{{
			Host: hostDev, Lun: lun,
			Channel: "-", Target: "-",
		}}
	} else {
		hctlsToRescan = make([]scsi.HCTL, len(hctlsToDiscover))
		copy(hctlsToRescan, hctlsToDiscover)
	}

	logger.Info(ctx, "hctls for scan: %s for device discovery: %s",
		hctlsToRescan, hctlsToDiscover)
	return hctlsToRescan, hctlsToDiscover, nil
}

func (fc *FCConnector) getFCHBASInfo(ctx context.Context) ([]FCHBA, error) {
	defer tracer.TraceFuncCall(ctx, "FCConnector.getFCHBASInfo")()
	logger.Info(ctx, "get FC hbas info")
	if !fc.isFCSupported(ctx) {
		return nil, errors.New("FC is not supported for this host")
	}
	match, err := fc.filePath.Glob("/sys/class/fc_host/host*")
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, err
	}
	var hbas []FCHBA
	for _, m := range match {
		hba := FCHBA{}
		data, err := fc.os.ReadFile(path.Join(m, "port_name"))
		if err != nil {
			logger.Error(ctx, "match: %s failed to read port_name file: %s", match, err.Error())
			continue
		}
		hba.PortName = strings.Replace(strings.TrimSpace(string(data)), "0x", "", 1)
		data, err = fc.os.ReadFile(path.Join(m, "node_name"))
		if err != nil {
			logger.Error(ctx, "match: %s failed to read node_name file: %s", match, err.Error())
			continue
		}
		hba.NodeName = strings.Replace(string(data), "0x", "", 1)
		_, hba.HostDevice = path.Split(m)
		hbas = append(hbas, hba)
	}
	logger.Info(ctx, "FC has been found: %s", hbas)
	return hbas, nil
}

func (fc *FCConnector) isFCSupported(ctx context.Context) bool {
	defer tracer.TraceFuncCall(ctx, "FCConnector.isFCSupported")()
	logger.Info(ctx, "check is FC supported")
	stat, err := fc.os.Stat("/sys/class/fc_host")
	if err != nil || !stat.IsDir() {
		logger.Info(ctx, "FC is not supported for this host")
		return false
	}
	logger.Info(ctx, "FC is supported")
	return true
}
