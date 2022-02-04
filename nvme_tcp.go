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
	"sync"
	"time"

	"github.com/dell/gobrick/internal/logger"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	"github.com/dell/gobrick/internal/tracer"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/gobrick/pkg/multipath"
	"github.com/dell/gobrick/pkg/scsi"
	"github.com/dell/goiscsi"
	"golang.org/x/sync/semaphore"
	"golang.org/x/sync/singleflight"
)

const (
	iSCSIFailedSessionMinimumLoginRetryInterval = time.Second * 300
	iSCSIWaitDeviceTimeoutDefault               = time.Second * 30
	iSCSIWaitDeviceRegisterTimeoutDefault       = time.Second * 10
	iSCSIMaxParallelOperationsDefault           = 5
)

type NVMeTCPConnectorParams struct {
	// nvmeLib command will run from this chroot
	Chroot string

	// timeouts
	// how long to wait for iSCSI session to become active after login
	WaitDeviceTimeout                      time.Duration
	WaitDeviceRegisterTimeout              time.Duration
	FailedSessionMinimumLoginRetryInterval time.Duration
	MultipathFlushTimeout                  time.Duration
	MultipathFlushRetryTimeout             time.Duration

	MultipathFlushRetries int
	MaxParallelOperations int
}

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
	setTimeouts(&conn.failedSessionMinimumLoginRetryInterval,
		params.FailedSessionMinimumLoginRetryInterval, iSCSIFailedSessionMinimumLoginRetryInterval)
	setTimeouts(&conn.waitDeviceTimeout,
		params.WaitDeviceTimeout, iSCSIWaitDeviceTimeoutDefault)
	setTimeouts(&conn.waitDeviceRegisterTimeout,
		params.WaitDeviceRegisterTimeout, iSCSIWaitDeviceRegisterTimeoutDefault)

	conn.loginLock = newRateLock()

	maxParallelOperations := params.MaxParallelOperations
	if maxParallelOperations == 0 {
		maxParallelOperations = iSCSIMaxParallelOperationsDefault
	}
	conn.limiter = semaphore.NewWeighted(int64(maxParallelOperations))
	conn.singleCall = &singleflight.Group{}
	return conn
}

type NVMeTCPConnector struct {
	baseConnector *baseConnector
	multipath     intmultipath.Multipath
	scsi          intscsi.SCSI
	nvmeTCPLib    wrp.NVMeTCP

	manualSessionManagement bool

	// timeouts
	waitDeviceTimeout                      time.Duration
	waitDeviceRegisterTimeout              time.Duration
	failedSessionMinimumLoginRetryInterval time.Duration

	loginLock  *rateLock
	limiter    *semaphore.Weighted
	singleCall *singleflight.Group

	// wrappers
	filePath wrp.LimitedFilepath
}

type NVMeTCPTargetInfo struct {
	Portal string
	Target string
}

type NVMeTCPVolumeInfo struct {
	Targets []NVMeTCPTargetInfo
	Lun     int
}

func singleCallKeyForNVMeTCPTargets(info NVMeTCPVolumeInfo) string {
	data := make([]string, len(info.Targets))
	for i, t := range info.Targets {
		target := strings.Join([]string{t.Portal, t.Target}, ":")
		data[i] = target
	}
	return strings.Join(data, ",")
}

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
	sessions := ret.([]goiscsi.ISCSISession)

	ret, _, _ = c.singleCall.Do(
		"IsDaemonRunning",
		func() (interface{}, error) { return c.multipath.IsDaemonRunning(ctx), nil })
	multipathIsEnabled := ret.(bool)
	// ---------------------------
	var d Device

	if multipathIsEnabled {
		logger.Info(ctx, "start multipath device connection")
		d, err = c.connectMultipathDevice(ctx, sessions, info)
	} else {
		logger.Info(ctx, "start single device connection")
		d, err = c.connectSingleDevice(ctx, sessions, info)
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

func (c *NVMeTCPConnector) DisconnectVolume(ctx context.Context, info ISCSIVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.DisconnectVolume")()
	if err := c.limiter.Acquire(ctx, 1); err != nil {
		return errors.New("too many parallel operations. try later")
	}
	defer c.limiter.Release(1)
	addDefaultNVMeTCPPortToVolumeInfoPortals(&info)
	return c.cleanConnection(ctx, false, info)
}

func (c *NVMeTCPConnector) DisconnectVolumeByDeviceName(ctx context.Context, name string) error {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.DisconnectVolumeByDeviceName")()
	if err := c.limiter.Acquire(ctx, 1); err != nil {
		return errors.New("too many parallel operations. try later")
	}
	defer c.limiter.Release(1)
	return c.baseConnector.disconnectDevicesByDeviceName(ctx, name)
}

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
			info.Targets[i].Portal += ":4420"
		}
	}
}

func (c *NVMeTCPConnector) cleanConnection(ctx context.Context, force bool, info ISCSIVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.cleanConnection")()
	var devices []string
	for _, t := range info.Targets {
		session, found, err := c.getSessionByTargetInfo(ctx, t)
		if err != nil {
			logger.Error(ctx, "failed to get iSCSI session: %s", err.Error())
			if force {
				continue
			}
			return err
		}
		if !found {
			continue
		}
		hctl, err := c.findHCTLByISCSISessionID(
			ctx, session.SID, strconv.FormatInt(int64(info.Lun), 10))
		if err != nil {
			logger.Error(ctx, "failed to get scsi info for iSCSI session: %s", err.Error())
			continue

		}
		dev, err := c.scsi.GetDeviceNameByHCTL(ctx, hctl)
		if err != nil {
			logger.Error(ctx, "failed to resolve iSCSI device name: %s", err.Error())
			continue
		}
		devices = append(devices, dev)
	}
	if len(devices) == 0 {
		return nil
	}
	return c.baseConnector.cleanDevices(ctx, force, devices)
}

func (c *NVMeTCPConnector) connectSingleDevice(
	ctx context.Context, sessions []goiscsi.ISCSISession, info ISCSIVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.connectSingleDevice")()
	devCH := make(chan string, len(sessions))
	wg := sync.WaitGroup{}
	discoveryCtx, cFunc := context.WithTimeout(ctx, c.waitDeviceTimeout)
	defer cFunc()

	for _, s := range sessions {
		wg.Add(1)
		go c.discoverDevice(discoveryCtx, 4, &wg, devCH, s, info)
	}
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
		devices = readDevicesFromResultCH(devCH, devices)
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
			var err error
			wwn, err = c.scsi.GetDeviceWWN(ctx, devices)
			if err != nil {
				logger.Info(ctx, "wwn for devices %s not found", devices)
			}
		}
		if wwn != "" {
			for _, d := range devices {
				if err := c.scsi.WaitUdevSymlink(ctx, d, wwn); err == nil {
					logger.Info(ctx, "registered device found: %s", d)
					return Device{Name: d, WWN: wwn}, nil
				}
			}
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

func readDevicesFromResultCH(ch chan string, result []string) []string {
	for {
		select {
		case d := <-ch:
			result = append(result, d)
		default:
			return result
		}
	}
}

func (c *NVMeTCPConnector) connectMultipathDevice(
	ctx context.Context, sessions []gonvme.NVMESession, info NVMeTCPVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.connectMultipathDevice")()
	devCH := make(chan string, len(sessions))
	wg := sync.WaitGroup{}
	discoveryCtx, cFunc := context.WithTimeout(ctx, c.waitDeviceTimeout)
	defer cFunc()
	//--------------------------
	for _, s := range sessions {
		wg.Add(1)
		go c.discoverDevice(discoveryCtx, 4, &wg, devCH, s, info)
	}
	// for non blocking wg wait
	wgCH := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgCH)
	}()

	var devices []string
	var wwn, mpath string
	var wwnAdded, discoveryComplete, lastTry bool
	var endTime time.Time
	for {
		// get discovered devices
		select {
		case <-ctx.Done():
			return Device{}, errors.New("connectMultipathDevice canceled")
		default:
		}
		devices = readDevicesFromResultCH(devCH, devices)
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
			var err error
			wwn, err = c.scsi.GetDeviceWWN(ctx, devices)
			if err != nil {
				logger.Info(ctx, "wwn for devices %s not found", devices)
			}
		}
		if wwn != "" && mpath == "" {
			var err error
			mpath, err = c.scsi.GetDMDeviceByChildren(ctx, devices)
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
			if err := c.scsi.WaitUdevSymlink(ctx, mpath, wwn); err == nil {
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

	return nil
}

func (c *NVMeTCPConnector) discoverDevice(
	ctx context.Context, rescans int, wg *sync.WaitGroup, result chan string,
	session gonvme.NVMESession, info NVMeTCPVolumeInfo) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.discoverDevice")()
	defer wg.Done()
	lun := strconv.FormatInt(int64(info.Lun), 10)

	var numRescans, secondsNextScan int

	numRescans = 0
	secondsNextScan = 1

	var hctl scsi.HCTL
	var hctlFound bool
	doScans := true
	for doScans {
		if !hctlFound || !hctl.IsFullInfo() {
			resp, err := c.findHCTLByISCSISessionID(ctx, session.SID, lun)
			if err == nil {
				hctlFound = true
				hctl = resp
			} else {
				logger.Error(ctx, err.Error())
			}
		}
		if hctlFound {
			if secondsNextScan <= 0 {
				numRescans++
				err := c.scsi.RescanSCSIHostByHCTL(ctx, hctl)
				if err != nil {
					logger.Error(ctx, err.Error())
				}
				secondsNextScan = int(math.Pow(float64(numRescans+2), 2))
			}
			if hctl.IsFullInfo() {
				dev, err := c.scsi.GetDeviceNameByHCTL(ctx, hctl)
				if err == nil {
					// we found device without rescans, information could be outdated, try to refresh it
					if numRescans == 0 {
						logger.Debug(ctx, "device %s found without scanning, "+
							"try to refresh device information", dev)
						err := c.scsi.RescanSCSIDeviceByHCTL(ctx, hctl)
						if err != nil {
							logger.Error(ctx, err.Error())
						}
					}
					logger.Info(ctx, "device found: %s", dev)
					result <- dev
					return
				}
				logger.Error(ctx, err.Error())

			}
		}
		select {
		case <-ctx.Done():
			logger.Info(ctx, "device discovery canceled")
			doScans = false
		default:
			doScans = numRescans <= rescans
		}
		if doScans {
			// skip wait for first iteration in manual scan mode
			if numRescans >= 0 {
				time.Sleep(time.Second)
			}
			secondsNextScan--
		}
	}
}

func (c *NVMeTCPConnector) checkNVMeTCPSessions(
	ctx context.Context, info ISCSIVolumeInfo) ([]gonvme.NVMESession, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.checkNVMeTCPSessions")()
	var activeSessions []gonvme.NVMESession
	var targetsToLogin []NVMeTCPTargetInfo
	for _, t := range info.Targets {
		tgt := gonvme.NVMeTarget{Portal: t.Portal, Target: t.Target}
		logger.Info(ctx,
			"check NVMe session for %s %s", t.Portal, t.Target)
		opt := make(map[string]string)
		opt["node.session.initial_login_retry_max"] = "1"
		if c.chapEnabled {
			opt["node.session.auth.authmethod"] = "CHAP"
			opt["node.session.auth.username"] = c.chapUser
			opt["node.session.auth.password"] = c.chapPassword
		}
		/*err := c.iscsiLib.CreateOrUpdateNode(tgt, opt)
		if err != nil {
			logger.Error(ctx,
				"can't create or update target for %s %s: %s", t.Portal, t.Target, err)
			continue
		}
		err = c.tryEnableManualISCSISessionMGMT(ctx, t)
		if err != nil {
			continue
		}*/
		session, found, err := c.getSessionByTargetInfo(ctx, t)
		if err != nil {
			logger.Error(ctx,
				"unable to get iSCSI session info: %s", err.Error())
			continue
		}
		if !found {
			targetsToLogin = append(targetsToLogin, t)
		}
		if c.isISCSISessionActive(ctx, session) {
			activeSessions = append(activeSessions, session)
		}
	}

	errMsg := "can't find active iSCSI session"

	// try login to Targets without sessions
	if len(targetsToLogin) != 0 {
		newSession, err := c.tryISCSILogin(ctx, targetsToLogin, len(activeSessions) == 0)
		if err != nil && len(activeSessions) == 0 {
			logger.Error(ctx, errMsg)
			return nil, errors.New(errMsg)
		}
		activeSessions = append(activeSessions, newSession...)
	}
	if len(activeSessions) == 0 {
		logger.Error(ctx, errMsg)
		return nil, errors.New(errMsg)
	}
	logger.Info(ctx, "found active iSCSI session")
	return activeSessions, nil
}

func (c *NVMeTCPConnector) tryISCSILogin(
	ctx context.Context, targets []ISCSITargetInfo, force bool) ([]goiscsi.ISCSISession, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.tryISCSILogin")()
	var sessions []goiscsi.ISCSISession
	for _, t := range targets {
		logPrefix := fmt.Sprintf("Portal: %s, Target: %s :", t.Portal, t.Target)
		tgt := goiscsi.ISCSITarget{Portal: t.Portal, Target: t.Target}
		logger.Info(ctx, logPrefix+"trying login to iSCSI target")
		mutexKey := strings.Join([]string{t.Portal, t.Target}, ":")
		if !(c.loginLock.RateCheck(
			mutexKey,
			c.failedSessionMinimumLoginRetryInterval) || force) {
			logger.Error(ctx, logPrefix+"rate limit - skip login")
			continue
		}
		err := c.nvmeTCPLib.PerformLogin(tgt)
		if err != nil {
			logger.Error(ctx, logPrefix+"can't login to iSCSI target: %s", err.Error())
			continue
		}

		s, found, err := c.getSessionByTargetInfo(ctx, t)
		if err != nil || !found {
			logger.Error(ctx, logPrefix+"can't read session info after login")
			continue
		}
		logger.Info(ctx, logPrefix+"successfully login to iSCSI target")
		sessions = append(sessions, s)
	}
	if len(targets) != 0 && len(sessions) == 0 {
		msg := "can't login to all Targets"
		logger.Error(ctx, msg)
		return nil, errors.New(msg)
	}
	return sessions, nil
}

/*
func (c *NVMeTCPConnector) tryEnableManualISCSISessionMGMT(ctx context.Context, target ISCSITargetInfo) error {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.tryEnableManualISCSISessionMGMT")()
	logPrefix := fmt.Sprintf("Portal: %s, Target: %s :", target.Portal, target.Target)
	tgt := goiscsi.ISCSITarget{Portal: target.Portal, Target: target.Target}
	if c.manualSessionManagement {
		opt := make(map[string]string)
		opt["node.session.scan"] = "manual"
		if c.chapEnabled {
			opt["node.session.auth.authmethod"] = "CHAP"
			opt["node.session.auth.username"] = c.chapUser
			opt["node.session.auth.password"] = c.chapPassword
		}
		err := c.iscsiLib.CreateOrUpdateNode(tgt, opt)
		if err != nil {
			var expectedErr bool
			if exitError, ok := err.(*exec.ExitError); ok {
				// if exit code == 7 manual session management not supported
				expectedErr = exitError.ExitCode() == 7
			}
			if !expectedErr {
				// handle other errors here
				logger.Error(ctx, logPrefix+"unable to update iSCSI session settings: %s", err.Error())
				return err
			}
			c.manualSessionManagement = false
		}
	}
	if !c.manualSessionManagement {
		logger.Error(ctx, logPrefix+"manual session management not supported")
	}
	return nil
}
*/

func (c *NVMeTCPConnector) isISCSISessionActive(
	ctx context.Context, session goiscsi.ISCSISession) bool {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.isISCSISessionActive")()
	return session.ISCSISessionState == goiscsi.ISCSISessionState_LOGGED_IN &&
		session.ISCSIConnectionState == goiscsi.ISCSIConnectionState_LOGGED_IN
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

func (c *NVMeTCPConnector) findHCTLByISCSISessionID(
	ctx context.Context, sessionID string, lun string) (scsi.HCTL, error) {
	defer tracer.TraceFuncCall(ctx, "NVMeTCPConnector.findHCTLByISCSISessionID")()
	result := scsi.HCTL{}
	sessionPattern := "/sys/class/iscsi_host/host*/device/session%s"
	matches, err := c.filePath.Glob(fmt.Sprintf(sessionPattern, sessionID) + "/target*")
	if err != nil {
		logger.Error(ctx, "error while try to resolve glob pattern %s: %s", sessionPattern, err.Error())
		return result, err
	}
	if len(matches) != 0 {
		_, fileName := path.Split(matches[0])
		targetData := strings.Split(fileName, ":")
		if len(targetData) != 3 {
			msg := "can't parse values from filename"
			logger.Info(ctx, msg)
			return result, errors.New(msg)
		}
		result.Host = targetData[0][6:]
		result.Channel = targetData[1]
		result.Target = targetData[2]
		result.Lun = lun
		return result, nil
	}
	// target not found try to resolve host only
	logger.Info(ctx, "can't find session data for %s in sysfs host only data will be resolved", sessionID)
	matches, err = c.filePath.Glob(fmt.Sprintf(sessionPattern, sessionID))
	if err != nil || len(matches) == 0 {
		return result, errors.New("can't resolve host data from sysfs")
	}
	result.Host = strings.Split(matches[0], "/")[4][4:]
	result.Channel = "-"
	result.Target = "-"
	result.Lun = "-"

	return result, nil
}
