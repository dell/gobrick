/*
 *
 * Copyright © 2020-2026 Dell Inc. or its subsidiaries. All Rights Reserved.
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
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dell/gobrick/internal/logger"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intpowerpath "github.com/dell/gobrick/internal/powerpath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	"github.com/dell/gobrick/internal/tracer"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/gobrick/pkg/multipath"
	"github.com/dell/gobrick/pkg/powerpath"
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

// ISCSIConnectorParams defines ISCSI connector parameters
type ISCSIConnectorParams struct {
	// iscsiLib command will run from this chroot
	Chroot string

	// chap
	ChapUser     string
	ChapPassword string
	ChapEnabled  bool

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

// NewISCSIConnector creates an ISCSI client and returns it
func NewISCSIConnector(params ISCSIConnectorParams) *ISCSIConnector {
	mp := multipath.NewMultipath(params.Chroot)
	pp := powerpath.NewPowerpath(params.Chroot)
	s := scsi.NewSCSI(params.Chroot)

	conn := &ISCSIConnector{
		multipath: mp,
		powerpath: pp,
		scsi:      s,
		filePath:  &wrp.FilepathWrapper{},
		baseConnector: newBaseConnector(mp, pp, s,
			baseConnectorParams{
				MultipathFlushTimeout:      params.MultipathFlushTimeout,
				MultipathFlushRetryTimeout: params.MultipathFlushRetryTimeout,
				MultipathFlushRetries:      params.MultipathFlushRetries,
			}),
		chapPassword: params.ChapPassword,
		chapUser:     params.ChapUser,
		chapEnabled:  params.ChapEnabled,
	}

	iSCSIOpts := make(map[string]string)
	iSCSIOpts["chrootDirectory"] = params.Chroot

	conn.iscsiLib = goiscsi.NewLinuxISCSI(iSCSIOpts)

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

// ISCSIConnector defines iscsi connector info
type ISCSIConnector struct {
	baseConnector *baseConnector
	multipath     intmultipath.Multipath
	powerpath     intpowerpath.Powerpath
	scsi          intscsi.SCSI
	iscsiLib      wrp.ISCSILib

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

	// chap
	chapUser     string
	chapPassword string
	chapEnabled  bool
}

// ISCSITargetInfo defines iscsi target info
type ISCSITargetInfo struct {
	Portal string
	Target string
	// NetworkID is the ID of the network that this target is reachable on.
	// This data is returned from the array API and filled in by the driver to manage discovery.
	NetworkID string
}

// ISCSIVolumeInfo defines iscsi volume info
type ISCSIVolumeInfo struct {
	Targets []ISCSITargetInfo
	Lun     int
}

func singleCallKeyForISCSITargets(info ISCSIVolumeInfo) string {
	data := make([]string, len(info.Targets))
	for i, t := range info.Targets {
		target := strings.Join([]string{t.Portal, t.Target}, ":")
		data[i] = target
	}
	return strings.Join(data, ",")
}

// ConnectVolume connects to the iscsi volume and returns device info
func (c *ISCSIConnector) ConnectVolume(ctx context.Context, info ISCSIVolumeInfo) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.ConnectVolume")()
	if err := c.limiter.Acquire(ctx, 1); err != nil {
		return Device{}, errors.New("too many parallel operations. try later")
	}
	defer c.limiter.Release(1)
	addDefaultISCSIPortToVolumeInfoPortals(&info)

	if err := c.validateISCSIVolumeInfo(ctx, info); err != nil {
		return Device{}, err
	}
	logger.Info(ctx, "validating sessions completed")
	ret, err, _ := c.singleCall.Do(
		singleCallKeyForISCSITargets(info),
		func() (interface{}, error) { return c.checkISCSISessions(ctx, info) })
	if err != nil {
		return Device{}, err
	}
	logger.Info(ctx, "getting sessions")
	sessions := ret.([]goiscsi.ISCSISession)
	logger.Info(ctx, "checking the daemon")
	ret, _, _ = c.singleCall.Do(
		"IsPowerPathDaemonRunning",
		func() (interface{}, error) { return c.powerpath.IsDaemonRunning(ctx), nil })
	powerpathIsEnabled := ret.(bool)

	ret, _, _ = c.singleCall.Do(
		"IsMultiPathDaemonRunning",
		func() (interface{}, error) { return c.multipath.IsDaemonRunning(ctx), nil })
	multipathIsEnabled := ret.(bool)
	logger.Info(ctx, "checked the daemon status")

	var d Device

	if powerpathIsEnabled {
		logger.Info(ctx, "start powerpath device connection")
		d, err = c.connectPowerpathDevice(ctx, sessions, info)
	} else if multipathIsEnabled {
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

// DisconnectVolume disconnects from iscsi volume
func (c *ISCSIConnector) DisconnectVolume(ctx context.Context, info ISCSIVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.DisconnectVolume")()
	if err := c.limiter.Acquire(ctx, 1); err != nil {
		return errors.New("too many parallel operations. try later")
	}
	defer c.limiter.Release(1)
	addDefaultISCSIPortToVolumeInfoPortals(&info)
	return c.cleanConnection(ctx, false, info)
}

// DisconnectVolumeByDeviceName disconnects from volume specified by WWN name
func (c *ISCSIConnector) DisconnectVolumeByDeviceName(ctx context.Context, name string) error {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.DisconnectVolumeByDeviceName")()
	if err := c.limiter.Acquire(ctx, 1); err != nil {
		return errors.New("too many parallel operations. try later")
	}
	defer c.limiter.Release(1)
	return c.baseConnector.disconnectDevicesByDeviceName(ctx, name)
}

// GetInitiatorName gets iscsi initiators and returns it
func (c *ISCSIConnector) GetInitiatorName(ctx context.Context) ([]string, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.GetInitiatorName")()
	logger.Info(ctx, "get initiator name")
	data, err := c.iscsiLib.GetInitiators("")
	if err != nil {
		logger.Error(ctx, "failed to read initiator name: %s", err.Error())
	}
	logger.Info(ctx, "initiator name is: %s", data)
	return data, nil
}

func addDefaultISCSIPortToVolumeInfoPortals(info *ISCSIVolumeInfo) {
	for i, t := range info.Targets {
		if !strings.Contains(t.Portal, ":") {
			info.Targets[i].Portal += ":3260"
		}
	}
}

func (c *ISCSIConnector) cleanConnection(ctx context.Context, force bool, info ISCSIVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.cleanConnection")()
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
	return c.baseConnector.cleanDevices(ctx, force, devices, "")
}

func (c *ISCSIConnector) connectSingleDevice(
	ctx context.Context, sessions []goiscsi.ISCSISession, info ISCSIVolumeInfo,
) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.connectSingleDevice")()
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

func (c *ISCSIConnector) connectPowerpathDevice(
	ctx context.Context, sessions []goiscsi.ISCSISession, info ISCSIVolumeInfo,
) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.connectPowerpathDevice")()
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
	var wwn, ppath string
	var discoveryComplete, lastTry bool
	var endTime time.Time
	for {
		// get discovered devices
		select {
		case <-ctx.Done():
			return Device{}, errors.New("connectPowerpathDevice canceled")
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
		if wwn != "" && ppath == "" {
			var err error
			ppath, err = c.powerpath.GetPowerPathDevices(ctx, devices)
			if err != nil {
				logger.Debug(ctx, "failed to get powerpath device: %s", err.Error())
			}
			log.Debugf("pp device: %s devices: %+v", ppath, devices)
		}
		if ppath != "" {
			if err := c.scsi.WaitUdevSymlink(ctx, ppath, wwn); err == nil {
				logger.Info(ctx, "powerpath device found: %s", ppath)
				return Device{WWN: wwn, Name: ppath, PowerpathID: wwn}, nil
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

func (c *ISCSIConnector) connectMultipathDevice(
	ctx context.Context, sessions []goiscsi.ISCSISession, info ISCSIVolumeInfo,
) (Device, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.connectMultipathDevice")()
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

func (c *ISCSIConnector) validateISCSIVolumeInfo(ctx context.Context, info ISCSIVolumeInfo) error {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.validateISCSIVolumeInfo")()
	if len(info.Targets) == 0 {
		return errors.New("at least one iSCSI target required")
	}
	for _, t := range info.Targets {
		if t.Target == "" || t.Portal == "" {
			return errors.New("invalid target info")
		}
	}

	return nil
}

func (c *ISCSIConnector) discoverDevice(
	ctx context.Context, rescans int, wg *sync.WaitGroup, result chan string,
	session goiscsi.ISCSISession, info ISCSIVolumeInfo,
) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.discoverDevice")()
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

func (c *ISCSIConnector) checkISCSISessions(
	ctx context.Context, info ISCSIVolumeInfo,
) ([]goiscsi.ISCSISession, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.checkISCSISessions")()
	var activeSessions []goiscsi.ISCSISession
	var targetsToLogin []ISCSITargetInfo
	for _, t := range info.Targets {
		tgt := goiscsi.ISCSITarget{Portal: t.Portal, Target: t.Target}
		logger.Info(ctx,
			"update iSCSI node for %s %s", t.Portal, t.Target)
		opt := make(map[string]string)
		opt["node.session.initial_login_retry_max"] = "1"
		if c.chapEnabled {
			opt["node.session.auth.authmethod"] = "CHAP"
			opt["node.session.auth.username"] = c.chapUser
			opt["node.session.auth.password"] = c.chapPassword
		}
		err := c.iscsiLib.CreateOrUpdateNode(tgt, opt)
		if err != nil {
			logger.Error(ctx,
				"can't create or update target for %s %s: %s", t.Portal, t.Target, err)
			continue
		}
		err = c.tryEnableManualISCSISessionMGMT(ctx, t)
		if err != nil {
			continue
		}
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

func (c *ISCSIConnector) tryISCSILogin(
	ctx context.Context, targets []ISCSITargetInfo, force bool,
) ([]goiscsi.ISCSISession, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.tryISCSILogin")()
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
		err := c.iscsiLib.PerformLogin(tgt)
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

func (c *ISCSIConnector) tryEnableManualISCSISessionMGMT(ctx context.Context, target ISCSITargetInfo) error {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.tryEnableManualISCSISessionMGMT")()
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

func (c *ISCSIConnector) isISCSISessionActive(
	ctx context.Context, session goiscsi.ISCSISession,
) bool {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.isISCSISessionActive")()
	return session.ISCSISessionState == goiscsi.ISCSISessionStateLOGGEDIN &&
		session.ISCSIConnectionState == goiscsi.ISCSIConnectionStateLOGGEDIN
}

func (c *ISCSIConnector) getSessionByTargetInfo(ctx context.Context,
	target ISCSITargetInfo,
) (goiscsi.ISCSISession, bool, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.getSessionByTargetInfo")()
	r := goiscsi.ISCSISession{}
	logPrefix := fmt.Sprintf("Portal: %s, Target: %s :", target.Portal, target.Target)
	sessions, err := c.iscsiLib.GetSessions()
	if err != nil {
		logger.Error(ctx, logPrefix+"unable to get iSCSI sessions: %s", err.Error())
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
		logger.Info(ctx, logPrefix+"iSCSI session found")
	} else {
		logger.Info(ctx, logPrefix+"iSCSI session not found")
	}
	return r, found, nil
}

func (c *ISCSIConnector) findHCTLByISCSISessionID(
	ctx context.Context, sessionID string, lun string,
) (scsi.HCTL, error) {
	defer tracer.TraceFuncCall(ctx, "ISCSIConnector.findHCTLByISCSISessionID")()
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
