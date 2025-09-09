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
	"path"
	"strings"
	"time"

	"github.com/dell/gobrick/pkg/scsi"

	"github.com/dell/gobrick/internal/logger"
	intmultipath "github.com/dell/gobrick/internal/multipath"
	intpowerpath "github.com/dell/gobrick/internal/powerpath"
	intscsi "github.com/dell/gobrick/internal/scsi"
	"github.com/dell/gobrick/internal/tracer"
)

const (
	multipathFlushTimeoutDefault      = time.Second * 120
	multipathFlushRetriesDefault      = 10
	multipathFlushRetryTimeoutDefault = time.Second * 5
	deviceMapperPrefix                = "dm-"
	noDevicesFound                    = "no devices found"
	orphanMpathDevice                 = "orphan"
)

type baseConnectorParams struct {
	MultipathFlushRetries      int
	MultipathFlushTimeout      time.Duration
	MultipathFlushRetryTimeout time.Duration
}

// cleanVolumeReq - structure to keep volume details for clean up on the host
type cleanVolumeReq struct {
	wwn       string
	sdDisks   []string
	mpathName string
	dmName    string
}

func newBaseConnector(mp intmultipath.Multipath, pp intpowerpath.Powerpath, s intscsi.SCSI, params baseConnectorParams) *baseConnector {
	conn := &baseConnector{
		multipath: mp,
		powerpath: pp,
		scsi:      s,
	}

	if params.MultipathFlushRetries == 0 {
		conn.multipathFlushRetries = multipathFlushRetriesDefault
	} else {
		conn.multipathFlushRetries = params.MultipathFlushRetries
	}

	setTimeouts(&conn.multipathFlushTimeout,
		params.MultipathFlushTimeout, multipathFlushTimeoutDefault)
	setTimeouts(&conn.multipathFlushRetryTimeout,
		params.MultipathFlushRetryTimeout, multipathFlushRetryTimeoutDefault)

	return conn
}

type baseConnector struct {
	multipath intmultipath.Multipath
	powerpath intpowerpath.Powerpath
	scsi      intscsi.SCSI

	multipathFlushRetries      int
	multipathFlushTimeout      time.Duration
	multipathFlushRetryTimeout time.Duration
}

// identifyDevicesForWWN - get the multipath name and sd disks and minor
// this function should be only used for multipath enabled environment
func (bc *baseConnector) identifyDevicesForWWN(ctx context.Context, wwn string) (*cleanVolumeReq, error) {
	defer tracer.TraceFuncCall(ctx, "baseConnector.identifyDevicesForWWN")()
	// get the multipath name and sd disks
	mpathName, sdDisks, err := bc.multipath.GetMultipathNameAndPaths(ctx, wwn)
	if err != nil {
		logger.Error(ctx, "failed to get multipath name: %s", err.Error())
		return nil, err
	}
	if strings.Contains(mpathName, orphanMpathDevice) {
		logger.Debug(ctx, "orphan multipath device found for wwn: %s", wwn)
		mpathName = ""
	}
	// get the dm name if mpath not orphan
	dmName, err := bc.scsi.GetDMDeviceByChildren(ctx, sdDisks)
	if err != nil {
		if strings.Contains(err.Error(), scsi.DmNotFoundErr) {
			logger.Debug(ctx, "dm holder not found for wwn: %s", wwn)
			dmName = ""
		} else {
			logger.Error(ctx, "failed to get device mapper name: %s", err.Error())
			return nil, err
		}
	}
	if len(sdDisks) == 0 {
		// there is no multipath device, and we could not retrieve sdDisks
		// check if we have sdDisks for the wwn
		// append a 3 to wwn for formatting for scsi devices
		sdDisks, err := bc.scsi.GetDevicesByWWN(ctx, fmt.Sprintf("3%s", wwn))
		if err != nil {
			logger.Error(ctx, "failed to find devices by wwn: %s", err.Error())
			return nil, err
		}
		if len(sdDisks) == 0 {
			logger.Info(ctx, "no devices found for wwn %s", wwn)
			return nil, errors.New(noDevicesFound)
		}
	}
	return &cleanVolumeReq{
		wwn:       wwn,
		sdDisks:   sdDisks,
		mpathName: mpathName,
		dmName:    dmName,
	}, nil
}

func (bc *baseConnector) disconnectDevicesByWWN(ctx context.Context, wwn string) error {
	defer tracer.TraceFuncCall(ctx, "baseConnector.disconnectDevicesByWWN")()
	if bc.multipath.IsDaemonRunning(ctx) {
		// check if the system is on multipath
		mpathInfo, err := bc.identifyDevicesForWWN(ctx, wwn)
		if err != nil {
			if errors.Is(err, errors.New(noDevicesFound)) {
				logger.Info(ctx, "no devices found for wwn %s, nothing to clean", wwn)
				return nil
			}
			logger.Error(ctx, "failed to identify devices for wwn: %s", err.Error())
			return err
		}
		return bc.cleanDevicesByMpathInfo(ctx, false, mpathInfo)
	}
	// multipath is not running, check if we have sdDisks for the wwn
	// append a 3 to wwn for formatting for scsi devices
	devices, err := bc.scsi.GetDevicesByWWN(ctx, fmt.Sprintf("3%s", wwn))
	if err != nil {
		logger.Error(ctx, "failed to find devices by wwn: %s", err.Error())
		return err
	}
	for i, d := range devices {
		err := bc.scsi.DeleteSCSIDeviceByName(ctx, d)
		if err != nil {
			logger.Error(ctx, "can't delete block device: %s", err.Error())
			return err
		}
		logger.Debug(ctx, "%d. %s disk deleted for %s", i+1, d, wwn)
	}
	logger.Debug(ctx, "clean devices completed for %s", wwn)
	return nil
}

func (bc *baseConnector) disconnectDevicesByDeviceName(ctx context.Context, name string) error {
	defer tracer.TraceFuncCall(ctx, "baseConnector.disconnectDevicesByDeviceName")()
	if !bc.scsi.IsDeviceExist(ctx, name) {
		logger.Info(ctx, "device %s not found", name)
		return nil
	}
	var err error
	var wwn string
	if strings.HasPrefix(name, deviceMapperPrefix) {
		wwn, err = bc.getDMWWN(ctx, name)
	} else {
		wwn, err = bc.scsi.GetDeviceWWN(ctx, []string{name})
	}
	if err != nil {
		logger.Error(ctx, "can't find wwn for device: %s", err.Error())
		return err
	}

	devices, err := bc.scsi.GetDevicesByWWN(ctx, wwn)
	if err != nil {
		logger.Error(ctx, "failed to find devices by wwn: %s", err.Error())
		return err
	}
	return bc.cleanDevices(ctx, false, devices, wwn)
}

func (bc *baseConnector) disconnectNVMEDevicesByDeviceName(ctx context.Context, name string) error {
	defer tracer.TraceFuncCall(ctx, "baseConnector.disconnectNVMEDevicesByDeviceName")()
	if !bc.scsi.IsDeviceExist(ctx, name) {
		logger.Info(ctx, "device %s not found", name)
		return nil
	}
	var err error
	var wwn string
	if strings.HasPrefix(name, deviceMapperPrefix) {
		wwn, err = bc.getNVMEDMWWN(ctx, name)
	} else {
		wwn, err = bc.scsi.GetNVMEDeviceWWN(ctx, []string{name})
	}
	if err != nil {
		logger.Error(ctx, "can't find wwn for device: %s", err.Error())
		return err
	}

	devices, err := bc.scsi.GetDevicesByWWN(ctx, wwn)
	if err != nil {
		logger.Error(ctx, "failed to find devices by wwn: %s", err.Error())
		return err
	}
	return bc.cleanNVMeDevices(ctx, false, devices, wwn)
}

func (bc *baseConnector) cleanNVMeDevices(ctx context.Context,
	force bool, devices []string, wwn string,
) error {
	defer tracer.TraceFuncCall(ctx, "baseConnector.cleanNVMeDevices")()
	var newDevices []string
	for _, device := range devices {
		newDevice := strings.ReplaceAll(device, "/dev/", "")
		newDevices = append(newDevices, newDevice)
	}
	dm, err := bc.scsi.GetDMDeviceByChildren(ctx, newDevices)
	if err != nil {
		logger.Info(ctx, "multipath device not found: %s", err.Error())
	} else {
		err := bc.cleanMultipathDevice(ctx, dm, wwn)
		if err != nil {
			msg := fmt.Sprintf("failed to flush multipath device: %s", err.Error())
			logger.Error(ctx, msg)
			if !force {
				return err
			}
		}
	}
	for _, d := range newDevices {
		err := bc.scsi.DeleteSCSIDeviceByName(ctx, d)
		if err != nil {
			logger.Error(ctx, "can't delete block device: %s", err.Error())
			if !force {
				return err
			}
		}
		if dm != "" {
			_ = bc.multipath.DelPath(ctx, path.Join("/dev", d))
		}
	}
	return nil
}

func (bc *baseConnector) cleanDevicesByMpathInfo(ctx context.Context, force bool, req *cleanVolumeReq) error {
	defer tracer.TraceFuncCall(ctx, "baseConnector.cleanDevicesByMpathInfo")()
	logger.Info(ctx, "cleanup started for: %+v", *req)
	if req.mpathName != "" {
		err := bc.cleanMultipathDeviceByName(ctx, req.mpathName)
		if err != nil {
			msg := fmt.Sprintf("failed to flush multipath device: %s", err.Error())
			logger.Error(ctx, msg)
			if !force {
				return err
			}
		}
		logger.Debug(ctx, "mpath: %s deleted", req.mpathName)
		// verify the deletion of mpath
		logger.Debug(ctx, "Verifying mpath: %s existence", req.mpathName)
		_, exist, err := bc.multipath.GetMpathMinorByMpathName(ctx, req.mpathName)
		if err != nil || exist {
			if exist {
				err = errors.New("mpath still exists, retry deletion")
				logger.Debug(ctx, "mpath: %s still exists", req.mpathName)
			} else {
				logger.Error(ctx, "can't get multipath minor: %s", err.Error())
			}
			if !force {
				return err
			}
		}
	} else {
		logger.Debug(ctx, "No multipath map found for %s; continuing to sd path deletion", req.wwn)
	}
	for i, d := range req.sdDisks {
		err := bc.scsi.DeleteSCSIDeviceByName(ctx, d)
		if err != nil {
			logger.Error(ctx, "can't delete block device: %s", err.Error())
			if !force {
				return err
			}
		}
		logger.Debug(ctx, "%d. %s disk deleted for %s", i+1, d, req.wwn)

	}
	logger.Debug(ctx, "clean devices completed for %s, mpath: %s", req.wwn, req.mpathName)
	return nil
}

func (bc *baseConnector) cleanDevices(ctx context.Context,
	force bool, devices []string, wwn string,
) error {
	defer tracer.TraceFuncCall(ctx, "baseConnector.cleanDevices")()
	dm, err := bc.scsi.GetDMDeviceByChildren(ctx, devices)
	if err != nil {
		logger.Info(ctx, "multipath device not found: %s", err.Error())
	} else {
		err := bc.cleanMultipathDevice(ctx, dm, wwn)
		if err != nil {
			msg := fmt.Sprintf("failed to flush multipath device: %s", err.Error())
			logger.Error(ctx, msg)
			if !force {
				return err
			}
		}
	}
	for _, d := range devices {
		err := bc.scsi.DeleteSCSIDeviceByName(ctx, d)
		if err != nil {
			logger.Error(ctx, "can't delete block device: %s", err.Error())
			if !force {
				return err
			}
		}
		if dm != "" {
			_ = bc.multipath.DelPath(ctx, path.Join("/dev", d))
		}
	}
	if bc.powerpath.IsDaemonRunning(ctx) {
		err = bc.powerpath.FlushDevice(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bc *baseConnector) cleanMultipathDevice(ctx context.Context, dm, wwid string) error {
	defer tracer.TraceFuncCall(ctx, "baseConnector.cleanMultipathDevice")()
	ctx, cancelFunc := context.WithTimeout(ctx, bc.multipathFlushTimeout)
	defer cancelFunc()

	if len(wwid) == 0 {
		wwid, _ = bc.multipath.GetDMWWID(ctx, dm)
	}

	for i := 0; i < bc.multipathFlushRetries; i++ {
		logger.Info(ctx, "trying to flush multipath device with retries: retry %d", i)
		err := bc.retryFlushMultipathDevice(ctx, dm)
		if err == nil {
			err := bc.multipath.RemoveDeviceFromWWIDSFile(ctx, wwid)
			if err != nil {
				logger.Error(ctx, "failed to remove wwid %s from wwids file: %s", wwid, err.Error())
			}
			return nil
		}
	}

	return fmt.Errorf("can't flush multipath device, timed out after multiple attempts")
}

func (bc *baseConnector) cleanMultipathDeviceByName(ctx context.Context, mpathName string) error {
	defer tracer.TraceFuncCall(ctx, "baseConnector.cleanMultipathDevice")()
	ctx, cancelFunc := context.WithTimeout(ctx, bc.multipathFlushTimeout)
	defer cancelFunc()

	for i := 0; i < bc.multipathFlushRetries; i++ {
		logger.Info(ctx, "trying to flush multipath device with retries: retry %d", i)
		err := bc.multipath.FlushDevice(ctx, mpathName)
		if err == nil {
			return nil
		}
		logger.Error(ctx, "received err: %s flushing multipath device: %s at retry: %d", err.Error(), mpathName, i)
	}
	return fmt.Errorf("can't flush multipath device, timed out after multiple attempts")
}

func (bc *baseConnector) retryFlushMultipathDevice(ctx context.Context, dm string) error {
	smallCtx, cancel := context.WithTimeout(ctx, bc.multipathFlushRetryTimeout)
	defer cancel()
	err := bc.multipath.FlushDevice(smallCtx, path.Join("/dev/", dm))
	if !bc.scsi.IsDeviceExist(ctx, dm) {
		logger.Info(ctx, "device %s no longer exists", dm)
		return nil
	}
	return err
}

func (bc *baseConnector) getDMWWN(ctx context.Context, dm string) (string, error) {
	defer tracer.TraceFuncCall(ctx, "baseConnector.getDMWWN")()
	logger.Info(ctx, "resolve wwn for DM: %s", dm)
	children, err := bc.scsi.GetDMChildren(ctx, dm)
	if err == nil {
		logger.Debug(ctx, "children for DM %s: %s", dm, children)
		wwn, err := bc.scsi.GetDeviceWWN(ctx, children)
		if err != nil {
			logger.Error(ctx, "failed to read WWN for DM %s children: %s", dm, err.Error())
			return "", err
		}
		logger.Debug(ctx, "WWN for DM %s is: %s", dm, wwn)
		return wwn, nil
	}
	logger.Debug(ctx, "failed to get children for DM %s: %s", dm, err.Error())
	logger.Info(ctx, "can't resolve DM %s WWN from children devices, query multipathd", dm)
	wwn, err := bc.multipath.GetDMWWID(ctx, dm)
	if err != nil {
		msg := fmt.Sprintf("failed to resolve DM %s WWN: %s", dm, err.Error())
		logger.Error(ctx, msg)
		return "", errors.New(msg)
	}
	logger.Info(ctx, "WWN for DM %s is: %s", dm, wwn)
	return wwn, nil
}

func (bc *baseConnector) getNVMEDMWWN(ctx context.Context, dm string) (string, error) {
	defer tracer.TraceFuncCall(ctx, "baseConnector.getDMWWN")()
	logger.Info(ctx, "resolve wwn for DM: %s", dm)
	children, err := bc.scsi.GetDMChildren(ctx, dm)
	if err == nil {
		logger.Debug(ctx, "children for DM %s: %s", dm, children)
		wwn, err := bc.scsi.GetNVMEDeviceWWN(ctx, children)
		if err != nil {
			logger.Error(ctx, "failed to read WWN for DM %s children: %s", dm, err.Error())
			return "", err
		}
		logger.Debug(ctx, "WWN for DM %s is: %s", dm, wwn)
		return wwn, nil
	}
	logger.Debug(ctx, "failed to get children for DM %s: %s", dm, err.Error())
	logger.Info(ctx, "can't resolve DM %s WWN from children devices, query multipathd", dm)
	wwn, err := bc.multipath.GetDMWWID(ctx, dm)
	if err != nil {
		msg := fmt.Sprintf("failed to resolve DM %s WWN: %s", dm, err.Error())
		logger.Error(ctx, msg)
		return "", errors.New(msg)
	}
	logger.Info(ctx, "WWN for DM %s is: %s", dm, wwn)
	return wwn, nil
}
