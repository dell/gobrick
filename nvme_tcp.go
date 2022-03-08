/*
 *
 * Copyright © 2022 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"path"
	"strings"
	"sync"
	"time"

	"github.com/dell/gobrick/internal/logger"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	"github.com/dell/gobrick/internal/tracer"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/gobrick/pkg/multipath"
	"github.com/dell/gobrick/pkg/scsi"
	"github.com/dell/gonvme"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

const (
	// NVMeWaitDeviceTimeoutDefault - NVMe default device time out
	NVMeWaitDeviceTimeoutDefault = time.Second * 30

	// NVMeWaitDeviceRegisterTimeoutDefault - NVMe default device register timeout
	NVMeWaitDeviceRegisterTimeoutDefault = time.Second * 10

	// NVMeMaxParallelOperationsDefault - max parallen NVMe operations
	NVMeMaxParallelOperationsDefault = 5

	// NVMePortDefault - NVMe TCP port
	NVMePortDefault = ":4420"
)

// NVMeTCPConnectorParams - type definition for NVMe TCP connector params
type NVMeTCPConnectorParams struct {
	// nvmeLib command will run from this chroot
	Chroot string

	// timeouts
	// how long to wait for nvme session to become active after login
	WaitDeviceTimeout                      time.Duration
	WaitDeviceRegisterTimeout              time.Duration
	FailedSessionMinimumLoginRetryInterval time.Duration
	MultipathFlushTimeout                  time.Duration
	MultipathFlushRetryTimeout             time.Duration

	MultipathFlushRetries int
	MaxParallelOperations int
}

// DevicePathResult - placeholder for nvme devicepaths
type DevicePathResult struct {
	devicePaths []string
	nguid       string
}

// NewNVMeTCPConnector - get new NVMeTCPConnector
func NewNVMeTCPConnector(params NVMeTCPConnectorParams) *NVMeTCPConnector {
	mp := multipath.NewMultipath(params.Chroot)
	s := scsi.NewSCSI(params.Chroot)

	conn := &NVMeTCPConnector{
		multipath: mp,
		scsi:      s,
		filePath:  &wrp.FilepathWrapper{},
		baseConnector: newBaseConnector(mp, s,
			baseConnectorParams{
				MultipathFlushTimeout:      params.MultipathFlushTimeout,
				MultipathFlushRetryTimeout: params.MultipathFlushRetryTimeout,
				MultipathFlushRetries:      params.MultipathFlushRetries}),
	}

	nvmeTCPOpts := make(map[string]string)
	nvmeTCPOpts["chrootDirectory"] = params.Chroot

	conn.nvmeTCPLib = gonvme.NewNVMeTCP(nvmeTCPOpts)

	// always try to use manual session management first
	conn.manualSessionManagement = true

	// timeouts
	setTimeouts(&conn.waitDeviceTimeout,
		params.WaitDeviceTimeout, NVMeWaitDeviceTimeoutDefault)
	setTimeouts(&conn.waitDeviceRegisterTimeout,
		params.WaitDeviceRegisterTimeout, NVMeWaitDeviceRegisterTimeoutDefault)

	conn.loginLock = newRateLock()

	maxParallelOperations := params.MaxParallelOperations
	if maxParallelOperations == 0 {
		maxParallelOperations = NVMeMaxParallelOperationsDefault
	}
	conn.limiter = semaphore.NewWeighted(int64(maxParallelOperations))
	conn.singleCall = &singleflight.Group{}
	return conn
}

// NVMeTCPConnector - type defenition for NVMe connector
type NVMeTCPConnector struct {
	baseConnector *baseConnector
	multipath     intmultipath.Multipath
	scsi          intscsi.SCSI
	nvmeTCPLib    wrp.NVMeTCP

	manualSessionManagement bool

	// timeouts
	waitDeviceTimeout         time.Duration
	waitDeviceRegisterTimeout time.Duration

	loginLock  *rateLock
	limiter    *semaphore.Weighted
	singleCall *singleflight.Group

	// wrappers
	filePath wrp.LimitedFilepath
}

// NVMeTCPTargetInfo - Placeholder for NVMe targets
type NVMeTCPTargetInfo struct {
	Portal string
	Target string
}

// NVMeTCPVolumeInfo - placeholder for NVMe TCP volume
type NVMeTCPVolumeInfo struct {
	Targets []NVMeTCPTargetInfo
	WWN     string
}

func singleCallKeyForNVMeTCPTargets(info NVMeTCPVolumeInfo) string {
	data := make([]string, len(info.Targets))
	for i, t := range info.Targets {
		target := strings.Join([]string{t.Portal, t.Target}, ":")
		data[i] = target
	}
	return strings.Join(data, ",")
}

// ConnectVolume - connect to nvme volume
func (c *NVMeTCPConnector) ConnectVolume(ctx context.Context, info NVMeTCPVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.ConnectVolume")()
	if err := c.limiter.Acquire(ctx, 1); err != nil {
		return Device{}, errors.New("too many parallel operations. try later")
	}
	defer c.limiter.Release(1)
	addDefaultNVMeTCPPortToVolumeInfoPortals(&info)

	if err := c.validateNVMeTCPVolumeInfo(ctx, info); err != nil {
		return Device{}, err
	}

	ret, err, _ := c.singleCall.Do(
		singleCallKeyForNVMeTCPTargets(info),
		func() (interface{}, error) { return c.checkNVMeTCPSessions(ctx, info) })
	if err != nil {
		return Device{}, err
	}
	sessions := ret.([]gonvme.NVMESession)

	ret, _, _ = c.singleCall.Do(
		"IsDaemonRunning",
		func() (interface{}, error) { return c.multipath.IsDaemonRunning(ctx), nil })
	multipathIsEnabled := ret.(bool)

	var d Device

	if multipathIsEnabled {
		logger.Info(ctx, "start multipath device connection")
		d, err = c.connectMultipathDevice(ctx, sessions, info)
		if err != nil {
			logger.Info(ctx, "start single device connection")
			d, err = c.connectSingleDevice(ctx, info)
		}
	} else {
		logger.Info(ctx, "start single device connection")
		d, err = c.connectSingleDevice(ctx, info)
	}

	if err == nil {
		if c.scsi.CheckDeviceIsValid(ctx, path.Join("/dev/", d.Name)) {
			return d, nil
		}
		msg := fmt.Sprintf("device %s found but failed to read data from it", d.Name)
		logger.Error(ctx, msg)
		err = errors.New(msg)
	}
	logger.Error(ctx, "failed to connect volume, try to cleanup: %s", err.Error())
	_ = c.cleanConnection(ctx, true, info)
	return Device{}, err
}

// DisconnectVolume - disconnect a given nvme volume
func (c *NVMeTCPConnector) DisconnectVolume(ctx context.Context, info NVMeTCPVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.DisconnectVolume")()
	if err := c.limiter.Acquire(ctx, 1); err != nil {
		return errors.New("too many parallel operations. try later")
	}
	defer c.limiter.Release(1)
	addDefaultNVMeTCPPortToVolumeInfoPortals(&info)
	return c.cleanConnection(ctx, false, info)
}

// DisconnectVolumeByDeviceName - disconnect from a given device
func (c *NVMeTCPConnector) DisconnectVolumeByDeviceName(ctx context.Context, name string) error {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.DisconnectVolumeByDeviceName")()
	if err := c.limiter.Acquire(ctx, 1); err != nil {
		return errors.New("too many parallel operations. try later")
	}
	defer c.limiter.Release(1)
	return c.baseConnector.disconnectNVMEDevicesByDeviceName(ctx, name)
}

// GetInitiatorName - returns nqn
func (c *NVMeTCPConnector) GetInitiatorName(ctx context.Context) ([]string, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.GetInitiatorName")()
	logger.Info(ctx, "get initiator name")
	data, err := c.nvmeTCPLib.GetInitiators("")
	if err != nil {
		logger.Error(ctx, "failed to read initiator name: %s", err.Error())
	}
	logger.Info(ctx, "initiator name is: %s", data)
	return data, nil
}

func addDefaultNVMeTCPPortToVolumeInfoPortals(info *NVMeTCPVolumeInfo) {
	for i, t := range info.Targets {
		if !strings.Contains(t.Portal, ":") {
			info.Targets[i].Portal += NVMePortDefault
		}
	}
}

func (c *NVMeTCPConnector) cleanConnection(ctx context.Context, force bool, info NVMeTCPVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.cleanConnection")()
	var devices []string
	wwn := info.WWN

	DevicePathsAndNamespaces := c.nvmeTCPLib.ListNamespaceDevices()
	var devicePath string
	var namespace string

	for DevicePathAndNamespace := range DevicePathsAndNamespaces {
		devicePath = DevicePathAndNamespace.DevicePath
		namespace = DevicePathAndNamespace.Namespace
		for _, namespaceID := range DevicePathsAndNamespaces[DevicePathAndNamespace] {
			nguid, newnamespace, _ := c.nvmeTCPLib.GetNamespaceData(devicePath, namespaceID)

			if c.wwnMatches(nguid, wwn) && namespace == newnamespace {
				devices = append(devices, devicePath)
			}
		}
	}
	if len(devices) == 0 {
		return nil
	}
	return c.baseConnector.cleanDevices(ctx, force, devices)
}

func (c *NVMeTCPConnector) connectSingleDevice(ctx context.Context, info NVMeTCPVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.connectSingleDevice")()
	devCH := make(chan DevicePathResult)
	wg := sync.WaitGroup{}
	discoveryCtx, cFunc := context.WithTimeout(ctx, c.waitDeviceTimeout)
	defer cFunc()

	wg.Add(1)
	go c.discoverDevice(discoveryCtx, &wg, devCH, info)
	// for non blocking wg wait
	wgCH := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgCH)
	}()

	var devices []string
	var wwn string
	var discoveryComplete, lastTry bool
	var endTime time.Time
	for {
		// get discovered devices
		select {
		case <-ctx.Done():
			return Device{}, errors.New("connectSingleDevice canceled")
		default:
		}
		devices, nguid := readNVMeDevicesFromResultCH(devCH, devices)
		// check all discovery gorutines finished
		if !discoveryComplete {
			select {
			case <-wgCH:
				discoveryComplete = true
				logger.Info(ctx, "all discovery goroutines complete")
			default:
				logger.Info(ctx, "discovery goroutines are still running")
			}
		}
		if discoveryComplete && len(devices) == 0 {
			msg := "discovery complete but devices not found"
			logger.Error(ctx, msg)
			return Device{}, errors.New(msg)
		}
		if wwn == "" && len(devices) != 0 {
			logger.Info(ctx, "Invalid WWN provided ")
		}
		if wwn != "" && nguid != "" {
			if len(devices) > 1 {
				logger.Debug(ctx, "Multiple nvme devices found for the given wwn %s", wwn)
			}
			logger.Info(ctx, "------- %s", devices[0])
			return Device{Name: devices[0], WWN: wwn}, nil
		}
		if discoveryComplete && !lastTry {
			logger.Info(ctx, "discovery finished, wait %f seconds for device registration",
				c.waitDeviceRegisterTimeout.Seconds())
			lastTry = true
			endTime = time.Now().Add(c.waitDeviceRegisterTimeout)
		}
		if lastTry && time.Now().After(endTime) {
			msg := "registered device not found"
			logger.Error(ctx, msg)
			return Device{}, errors.New(msg)
		}
		time.Sleep(time.Second)
	}
}

func (c *NVMeTCPConnector) connectMultipathDevice(
	ctx context.Context, sessions []gonvme.NVMESession, info NVMeTCPVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.connectMultipathDevice")()
	devCH := make(chan DevicePathResult)
	wg := sync.WaitGroup{}
	discoveryCtx, cFunc := context.WithTimeout(ctx, c.waitDeviceTimeout)
	defer cFunc()

	wg.Add(1)
	go c.discoverDevice(discoveryCtx, &wg, devCH, info)
	// for non blocking wg wait
	wgCH := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgCH)
	}()

	var devices []string
	var mpath string
	wwn := info.WWN
	var wwnAdded, discoveryComplete, lastTry bool
	var endTime time.Time
	for {
		// get discovered devices
		select {
		case <-ctx.Done():
			return Device{}, errors.New("connectMultipathDevice canceled")
		default:
		}
		devices, nguid := readNVMeDevicesFromResultCH(devCH, devices)

		// check all discovery gorutines finished
		if !discoveryComplete {
			select {
			case <-wgCH:
				discoveryComplete = true
				logger.Info(ctx, "all discover goroutines complete")
			default:
				logger.Info(ctx, "discover goroutines are still running")
			}
		}
		if discoveryComplete && len(devices) == 0 {
			msg := "discover complete but devices not found"
			logger.Error(ctx, msg)
			return Device{}, errors.New(msg)
		}
		if wwn == "" && len(devices) != 0 {
			logger.Info(ctx, "Invalid WWN provided ")
		}

		if wwn != "" && mpath == "" {
			var err error
			mpath, err = c.scsi.GetNVMEDMDeviceByChildren(ctx, devices)
			if err != nil {
				logger.Debug(ctx, "failed to get DM by children: %s", err.Error())
			}
			if mpath == "" && !wwnAdded {
				if err := c.multipath.AddWWID(ctx, wwn); err == nil {
					wwnAdded = true
				} else {
					logger.Info(ctx, err.Error())
				}
			}
		}
		if mpath != "" {
			//use nguid as wwn for nvme devices
			var err error
			if err = c.scsi.WaitUdevSymlinkNVMe(ctx, mpath, nguid); err == nil {
				logger.Info(ctx, "multipath device found: %s", mpath)
				return Device{WWN: wwn, Name: mpath, MultipathID: wwn}, nil
			}
		}
		if discoveryComplete && !lastTry {
			logger.Info(ctx, "discovery finished, wait %f seconds for DM to appear",
				c.waitDeviceRegisterTimeout.Seconds())
			lastTry = true
			for _, d := range devices {
				if err := c.multipath.AddPath(ctx, path.Join("/dev/", d)); err != nil {
					logger.Error(ctx, err.Error())
				}
			}
			endTime = time.Now().Add(c.waitDeviceRegisterTimeout)
		}
		if lastTry && time.Now().After(endTime) {
			msg := "registered multipath device not found"
			logger.Error(ctx, msg)
			return Device{}, errors.New(msg)
		}
		time.Sleep(time.Second)
	}
}

func (c *NVMeTCPConnector) validateNVMeTCPVolumeInfo(ctx context.Context, info NVMeTCPVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.validateNVMeTCPVolumeInfo")()
	if len(info.Targets) == 0 {
		return errors.New("at least one NVMe target required")
	}
	for _, t := range info.Targets {
		if t.Target == "" || t.Portal == "" {
			return errors.New("invalid target info")
		}
	}

	if info.WWN == "" {
		return errors.New("invalid volume wwn")
	}

	return nil
}

func (c *NVMeTCPConnector) discoverDevice(ctx context.Context, wg *sync.WaitGroup, result chan DevicePathResult, info NVMeTCPVolumeInfo) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.findDevice")()
	defer wg.Done()
	wwn := info.WWN

	DevicePathsAndNamespaces := c.nvmeTCPLib.ListNamespaceDevices()

	var devicePaths []string
	var devicePath string
	var namespace string

	nguidResult := ""

	for DevicePathAndNamespace := range DevicePathsAndNamespaces {

		devicePath = DevicePathAndNamespace.DevicePath
		namespace = DevicePathAndNamespace.Namespace

		for _, namespaceID := range DevicePathsAndNamespaces[DevicePathAndNamespace] {
			nguid, newnamespace, _ := c.nvmeTCPLib.GetNamespaceData(devicePath, namespaceID)

			if c.wwnMatches(nguid, wwn) && namespace == newnamespace {
				devicePaths = append(devicePaths, devicePath)
				nguidResult = nguid
			}
		}
	}
	devicePathResult := DevicePathResult{devicePaths: devicePaths, nguid: nguidResult}

	result <- devicePathResult
}

func (c *NVMeTCPConnector) wwnMatches(nguid, wwn string) bool {

	/*
		Sample wwn : naa.68ccf098001111a2222b3d4444a1b23c
		wwn1 : 1111a2222b3d4444
		wwn2 : a1b23c

		Sample nguid : 1111a2222b3d44448ccf096800a1b23c
	*/
	if len(wwn) < 32 {
		return false
	}
	wwn1 := wwn[13 : len(wwn)-7]
	wwn2 := wwn[len(wwn)-6 : len(wwn)-1]

	if strings.Contains(nguid, wwn1) && strings.Contains(nguid, wwn2) {
		return true
	}
	return false
}

func readNVMeDevicesFromResultCH(ch chan DevicePathResult, result []string) ([]string, string) {

	devicePathResult := <-ch
	var devicePaths []string
	for _, path := range devicePathResult.devicePaths {
		// modify path /dev/nvme0n1 -> nvme0n1
		newpath := strings.ReplaceAll(path, "/dev/", "")
		devicePaths = append(devicePaths, newpath)
	}
	return devicePaths, devicePathResult.nguid
}

func (c *NVMeTCPConnector) checkNVMeTCPSessions(
	ctx context.Context, info NVMeTCPVolumeInfo) ([]gonvme.NVMESession, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.checkNVMeTCPSessions")()
	var activeSessions []gonvme.NVMESession
	//var targetsToLogin []NVMeTCPTargetInfo
	for _, t := range info.Targets {
		logger.Info(ctx,
			"check NVMe session for %s %s", t.Portal, t.Target)

		session, _, err := c.getSessionByTargetInfo(ctx, t)
		if err != nil {
			logger.Error(ctx,
				"unable to get nvme session info: %s", err.Error())
			continue
		} else {
			activeSessions = append(activeSessions, session)
		}
	}

	errMsg := "can't find active nvme session"

	if len(activeSessions) == 0 {
		logger.Error(ctx, errMsg)
		return nil, errors.New(errMsg)
	}
	logger.Info(ctx, "found active nvme sessions")
	return activeSessions, nil
}

func (c *NVMeTCPConnector) getSessionByTargetInfo(ctx context.Context,
	target NVMeTCPTargetInfo) (gonvme.NVMESession, bool, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.getSessionByTargetInfo")()
	r := gonvme.NVMESession{}
	logPrefix := fmt.Sprintf("Portal: %s, Target: %s :", target.Portal, target.Target)
	sessions, err := c.nvmeTCPLib.GetSessions()
	if err != nil {
		logger.Error(ctx, logPrefix+"unable to get nvme sessions: %s", err.Error())
		return r, false, err
	}
	var found bool
	//TODO: check if comparision needs contains check
	for _, s := range sessions {
		if s.Target == target.Target && s.Portal == target.Portal {
			r = s
			found = true
			break
		}
	}
	if found {
		logger.Info(ctx, logPrefix+"nvme session found")
	} else {
		logger.Info(ctx, logPrefix+"nvme session not found")
	}
	return r, found, nil
}
