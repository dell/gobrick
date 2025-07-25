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

package scsi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/dell/gobrick/internal/logger"
	"github.com/dell/gobrick/internal/tracer"
	wrp "github.com/dell/gobrick/internal/wrappers"
	"github.com/dell/gobrick/pkg/gobrickutils"
	"golang.org/x/sync/singleflight"
)

// constants
const (
	diskByIDPath           = "/dev/disk/by-id/"
	diskByIDSCSIPath       = diskByIDPath + "scsi-"
	diskByIDDMPath         = diskByIDPath + "dm-uuid-mpath-"
	diskByIDDMPathNVMe     = diskByIDPath + "dm-uuid-mpath-eui."
	scsiIDPath             = "/lib/udev/scsi_id"
	maxRetryCount          = 10
	NVMEMultipathSleepTime = 500
	NVMESymlinkSleepTime   = 200
)

// NewSCSI initializes scsi struct
func NewSCSI(chroot string) *Scsi {
	scsi := &Scsi{
		chroot:   chroot,
		filePath: &wrp.FilepathWrapper{},
		os:       &wrp.OSWrapper{},
		osexec:   &wrp.OSExecWrapper{},
	}
	scsi.singleCall = &singleflight.Group{}

	return scsi
}

// HCTL defines host, channel, target, lun info
type HCTL struct {
	Host    string
	Channel string
	Target  string
	Lun     string
}

// IsFullInfo validates HCTL struct
func (h *HCTL) IsFullInfo() bool {
	if h.Channel == "" || h.Channel == "-" ||
		h.Target == "" || h.Target == "-" {
		return false
	}
	return true
}

// DevicesHaveDifferentParentsErr defines a custom error
type DevicesHaveDifferentParentsErr struct{}

func (dperr *DevicesHaveDifferentParentsErr) Error() string {
	return "device have different parent DMs"
}

// Scsi defines scsi info
type Scsi struct {
	chroot string

	filePath wrp.LimitedFilepath
	os       wrp.LimitedOS
	osexec   wrp.LimitedOSExec

	singleCall *singleflight.Group
}

// IsDeviceExist checks if scsi device exists
func (s *Scsi) IsDeviceExist(ctx context.Context, device string) bool {
	defer tracer.TraceFuncCall(ctx, "scsi.IsDeviceExist")()
	return s.checkExist(ctx, path.Join("/dev/", device))
}

// RescanSCSIHostByHCTL performs scsi host rescan
func (s *Scsi) RescanSCSIHostByHCTL(ctx context.Context, addr HCTL) error {
	defer tracer.TraceFuncCall(ctx, "scsi.RescanSCSIHostByHCTL")()
	return s.rescanSCSIHostByHCTL(ctx, addr)
}

// RescanSCSIDeviceByHCTL performs scsi device rescan
func (s *Scsi) RescanSCSIDeviceByHCTL(ctx context.Context, h HCTL) error {
	defer tracer.TraceFuncCall(ctx, "scsi.RescanSCSIDeviceByHCTL")()
	return s.rescanSCSIDeviceByHCTL(ctx, h)
}

// DeleteSCSIDeviceByHCTL deletes scsi device specified by HCTL
func (s *Scsi) DeleteSCSIDeviceByHCTL(ctx context.Context, h HCTL) error {
	defer tracer.TraceFuncCall(ctx, "scsi.DeleteSCSIDeviceByHCTL")()
	return s.deleteSCSIDeviceByHCTL(ctx, h)
}

// DeleteSCSIDeviceByName deletes scsi device specified by name
func (s *Scsi) DeleteSCSIDeviceByName(ctx context.Context, name string) error {
	defer tracer.TraceFuncCall(ctx, "scsi.DeleteSCSIDeviceByName")()
	return s.deleteSCSIDeviceByName(ctx, name)
}

// GetDeviceWWN fetches device WWN
func (s *Scsi) GetDeviceWWN(ctx context.Context, devices []string) (string, error) {
	defer tracer.TraceFuncCall(ctx, "scsi.GetDeviceWWN")()
	return s.getDeviceWWN(ctx, devices)
}

// GetNVMEDeviceWWN gets wwn of nvme device
func (s *Scsi) GetNVMEDeviceWWN(ctx context.Context, devices []string) (string, error) {
	defer tracer.TraceFuncCall(ctx, "scsi.GetNVMEDeviceWWN")()
	return s.getNVMEDeviceWWN(ctx, devices)
}

// GetDevicesByWWN fetches devices by wwn
func (s *Scsi) GetDevicesByWWN(ctx context.Context, wwn string) ([]string, error) {
	defer tracer.TraceFuncCall(ctx, "scsi.GetDevicesByWWN")()
	return s.getDevicesByWWN(ctx, wwn)
}

// DeleteSCSIDeviceByPath deletes device by specified "device folder" path
// Examples:
//
//		/sys/block/sde/device/
//		/sys/class/scsi_device/37:0:0:1/device/
//	    /sys/class/iscsi_session/session3/device/target37:0:0/37:0:0:1/
func (s *Scsi) DeleteSCSIDeviceByPath(ctx context.Context, devPath string) error {
	defer tracer.TraceFuncCall(ctx, "scsi.DeleteSCSIDeviceByPath")()
	return s.deleteSCSIDeviceByPath(ctx, devPath)
}

// GetDMDeviceByChildren fetches multipath device name
func (s *Scsi) GetDMDeviceByChildren(ctx context.Context, devices []string) (string, error) {
	defer tracer.TraceFuncCall(ctx, "scsi.GetDMDeviceByChildren")()
	return s.getDMDeviceByChildren(ctx, devices)
}

// GetNVMEDMDeviceByChildren fetches multipath device name
func (s *Scsi) GetNVMEDMDeviceByChildren(ctx context.Context, devices []string) (string, error) {
	defer tracer.TraceFuncCall(ctx, "scsi.GetNVMEDMDeviceByChildren")()
	return s.getNVMEDMDeviceByChildren(ctx, devices)
}

// GetDMChildren fetches multipath block devices
func (s *Scsi) GetDMChildren(ctx context.Context, dm string) ([]string, error) {
	defer tracer.TraceFuncCall(ctx, "scsi.GetDMChildren")()
	return s.getDMChildren(ctx, dm)
}

// CheckDeviceIsValid checks if device specified is valid by performing io operation on the device
func (s *Scsi) CheckDeviceIsValid(ctx context.Context, device string) bool {
	defer tracer.TraceFuncCall(ctx, "scsi.CheckDeviceIsValid")()
	return s.checkDeviceIsValid(ctx, device)
}

// GetDeviceNameByHCTL finds scsi device name by HCTL
func (s *Scsi) GetDeviceNameByHCTL(ctx context.Context, h HCTL) (string, error) {
	defer tracer.TraceFuncCall(ctx, "scsi.GetDeviceNameByHCTL")()
	return s.getDeviceNameByHCTL(ctx, h)
}

// WaitUdevSymlink checks if udev symlink for device specified by device name with WWN is found
func (s *Scsi) WaitUdevSymlink(ctx context.Context, deviceName string, wwn string) error {
	defer tracer.TraceFuncCall(ctx, "scsi.WaitUdevSymlink")()
	return s.waitUdevSymlink(ctx, deviceName, wwn)
}

// WaitUdevSymlinkNVMe checks if udev symlink for device specified by device name with WWN is found for nvme
func (s *Scsi) WaitUdevSymlinkNVMe(ctx context.Context, deviceName string, wwn string) error {
	defer tracer.TraceFuncCall(ctx, "scsi.WaitUdevSymlinkNVMe")()
	return s.waitUdevSymlinkNVMe(ctx, deviceName, wwn)
}

func (s *Scsi) rescanSCSIHostByHCTL(ctx context.Context, addr HCTL) error {
	hostsDir := "/sys/class/scsi_host"
	filePath := fmt.Sprintf("%s/host%s/scan", hostsDir, addr.Host)
	scanString := fmt.Sprintf("%s %s %s", addr.Channel, addr.Target, addr.Lun)
	logger.Info(ctx, "rescan scsi: %s %s", addr.Host, scanString)
	scanFile, err := s.os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0o200)
	if err != nil {
		logger.Error(ctx, "Failed to open %s", filePath)
		return err
	}
	if _, err := scanFile.WriteString(scanString); err != nil {
		logger.Error(ctx, "Failed to write %s to %s", scanString, scanFile)
		return err
	}
	return scanFile.Close()
}

func (s *Scsi) deleteSCSIDeviceByHCTL(ctx context.Context, h HCTL) error {
	devicePath := fmt.Sprintf("/sys/class/scsi_device/%s:%s:%s:%s/device/",
		h.Host, h.Channel, h.Target, h.Lun)
	return s.DeleteSCSIDeviceByPath(ctx, devicePath)
}

func (s *Scsi) deleteSCSIDeviceByName(ctx context.Context, name string) error {
	devicePath := fmt.Sprintf("/sys/class/block/%s/device",
		name)
	return s.DeleteSCSIDeviceByPath(ctx, devicePath)
}

func (s *Scsi) getDeviceWWN(ctx context.Context, devices []string) (string, error) {
	var err error
	for _, d := range devices {
		wwidFilePath := fmt.Sprintf("/sys/block/%s/device/wwid", d)
		var result string
		result, err = s.readWWIDFile(ctx, wwidFilePath)
		if err == nil {
			return result, nil
		} else if s.os.IsNotExist(err) {
			if result, err = s.getDeviceWWNWithSCSIID(ctx, d); err == nil {
				return result, nil
			}
		}
	}
	return "", err
}

func (s *Scsi) getNVMEDeviceWWN(ctx context.Context, devices []string) (string, error) {
	var err error
	for _, d := range devices {
		wwidFilePath := fmt.Sprintf("/sys/block/%s/wwid", d)
		var result string
		result, err = s.readWWIDFile(ctx, wwidFilePath)
		if err == nil {
			return result, nil
		} else if s.os.IsNotExist(err) {
			if result, err = s.getDeviceWWNWithSCSIID(ctx, d); err == nil {
				return result, nil
			}
		}
	}
	return "", err
}

func (s *Scsi) getDeviceWWNWithSCSIID(ctx context.Context, device string) (string, error) {
	logger.Debug(ctx, "get wwn with scsi_id for device: %s", device)

	err := gobrickutils.ValidateDeviceName(device)
	if err != nil {
		return "", err
	}
	command := scsiIDPath
	args := []string{"-g", "-p", "0x83", "/dev/" + device}
	if s.chroot != "" {
		args = append([]string{s.chroot, command}, args...)
		command = "chroot"
	}
	data, err := s.osexec.CommandContext(ctx, command, args...).CombinedOutput()
	if err != nil {
		logger.Error(ctx, "failed to read device %s wwn with scsi_id: %s", device, err.Error())
		return "", err
	}
	result := strings.TrimSpace(string(data))
	logger.Debug(ctx, "scsi_id output for device %s: %s", device, result)
	return result, nil
}

func (s *Scsi) readWWIDFile(ctx context.Context, path string) (string, error) {
	wwnTypes := map[string]string{"t10.": "1", "eui.": "2", "naa.": "3"}
	data, err := s.os.ReadFile(path)
	if err == nil {
		wwid := strings.TrimSpace(string(data))
		wwnType, ok := wwnTypes[wwid[:4]]
		if !ok {
			wwnType = "8"
		}
		return wwnType + wwid[4:], nil
	}
	logger.Debug(ctx, "failed to read wwn file %s: %s", path, err.Error())
	return "", err
}

// delete device by specified "device folder" path
// Examples:
//
//		/sys/block/sde/device/
//		/sys/class/scsi_device/37:0:0:1/device/
//	    /sys/class/iscsi_session/session3/device/target37:0:0/37:0:0:1/
func (s *Scsi) deleteSCSIDeviceByPath(ctx context.Context, devPath string) error {
	deletePath := path.Join(devPath, "delete")
	statePath := path.Join(devPath, "state")
	// delete command for device in "blocked" state could stuck on 3.x kernel
	stateContent, err := s.os.ReadFile(statePath)
	if err != nil {
		logger.Error(ctx, "can't read %s: %s", statePath, err.Error())
		return err
	}
	deviceState := strings.TrimSpace(string(stateContent))
	logger.Info(ctx, "device state is: %s", deviceState)
	if deviceState == "blocked" {
		msg := "device is in blocked state"
		logger.Error(ctx, msg)
		return errors.New(msg)
	}
	deleteFile, err := s.os.OpenFile(deletePath, os.O_APPEND|os.O_WRONLY, 0o200)
	if err != nil {
		logger.Error(ctx, "could not open %s path", deletePath)
		// ignore
		return nil
	}
	logger.Info(ctx, "writing '1' to %s", deletePath)
	if _, err := deleteFile.WriteString("1"); err != nil {
		logger.Error(ctx, "failed to write to %s: %s", deletePath, err.Error())
	}
	if err := deleteFile.Close(); err != nil {
		logger.Error(ctx, "error while close %s: %s", deletePath, err.Error())
		return err
	}
	return nil
}

func (s *Scsi) getDMDeviceByChildren(ctx context.Context, devices []string) (string, error) {
	logger.Info(ctx, "multipath - trying to find multipath DM name")

	pattern := "/sys/block/%s/holders/dm-*"

	var match string

	for _, d := range devices {
		matches, err := s.filePath.Glob(fmt.Sprintf(pattern, d))
		if err != nil {
			return "", err
		}
		for _, m := range matches {
			data, err := s.os.ReadFile(path.Join(m, "dm/uuid"))
			if err != nil {
				logger.Error(ctx, "multipath - failed to read dm id file: %s", err.Error())
				continue
			}
			if strings.HasPrefix(string(data), "mpath") {
				_, dm := path.Split(m)
				if match == "" {
					match = dm
				} else if dm != match {
					return "", &DevicesHaveDifferentParentsErr{}
				}
			}
		}
	}
	if match != "" {
		return match, nil
	}
	return "", errors.New("dm not found")
}

// GetNVMEMultipathDMName finds the multipath DM mame for NVMe
func (s *Scsi) GetNVMEMultipathDMName(device string, pattern string) ([]string, error) {
	retryCount := 0
	for {
		matches, err := s.filePath.Glob(fmt.Sprintf(pattern, device))
		if len(matches) > 0 || retryCount == maxRetryCount {
			return matches, err
		}
		time.Sleep(NVMEMultipathSleepTime * time.Millisecond)
		retryCount = retryCount + 1
	}
}

func (s *Scsi) getNVMEDMDeviceByChildren(ctx context.Context, devices []string) (string, error) {
	logger.Info(ctx, "multipath - trying to find multipath DM name")

	pattern := "/sys/block/%s/holders/dm-*"

	var match string

	for _, d := range devices {
		matches, err := s.GetNVMEMultipathDMName(d, pattern)
		if err != nil {
			return "", err
		}
		for _, m := range matches {
			data, err := s.os.ReadFile(path.Join(m, "dm/uuid"))
			if err != nil {
				logger.Error(ctx, "multipath - failed to read dm id file: %s", err.Error())
				continue
			}
			if strings.HasPrefix(string(data), "mpath") {
				_, dm := path.Split(m)
				if match == "" {
					match = dm
				} else if dm != match {
					return "", &DevicesHaveDifferentParentsErr{}
				}
			}
		}
	}
	if match != "" {
		return match, nil
	}
	return "", errors.New("dm not found")
}

func (s *Scsi) getDMChildren(ctx context.Context, dm string) ([]string, error) {
	logger.Info(ctx, "multipath - get block device included in DM")
	var devices []string
	pattern := "/sys/block/%s/slaves/*"
	matches, err := s.filePath.Glob(fmt.Sprintf(pattern, dm))
	if err != nil {
		return nil, err
	}
	for _, m := range matches {
		_, device := path.Split(m)
		devices = append(devices, device)
	}
	return devices, nil
}

func (s *Scsi) getDevicesByWWN(ctx context.Context, wwn string) ([]string, error) {
	logger.Info(ctx, "get devices by wwn %s", wwn)
	ret, err, _ := s.singleCall.Do("getDevicesByWWN", func() (interface{}, error) {
		result := make(map[string][]string)
		matches, err := s.filePath.Glob("/sys/block/sd*")
		if err != nil {
			logger.Error(ctx,
				"failed to get devices by wwn %s: %s", wwn, err.Error())
			return nil, err
		}
		for _, m := range matches {
			_, devName := path.Split(m)
			devWWN, err := s.getDeviceWWN(ctx, []string{devName})
			if err != nil {
				continue
			}
			if devs, ok := result[devWWN]; ok {
				result[devWWN] = append(devs, devName)
			} else {
				result[devWWN] = []string{devName}
			}
		}
		return result, nil
	})
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, err
	}
	devs := ret.(map[string][]string)
	if devsForWnn, ok := devs[wwn]; ok {
		logger.Info(ctx, "devices for WWN %s: %s", wwn, devsForWnn)
		return devsForWnn, nil
	}
	logger.Info(ctx, "devices for WWN %s not found", wwn)
	return nil, nil
}

func (s *Scsi) checkExist(_ context.Context, device string) bool {
	_, err := s.os.Stat(device)
	return err == nil
}

func (s *Scsi) checkDeviceIsValid(ctx context.Context, devicePath string) bool {
	ctx, cancelF := context.WithTimeout(ctx, time.Second*10)
	defer cancelF()
	exist := s.checkExist(ctx, devicePath)
	if !exist {
		return false
	}
	// we are using dd tool instead of os.OpenFile to be able to cancel IO if it stuck
	cmd := s.osexec.CommandContext(ctx, "dd",
		fmt.Sprintf("if=%s", devicePath),
		"of=/dev/null", "bs=1k", "count=1")
	data, err := cmd.CombinedOutput()
	if err != nil {
		logger.Error(ctx,
			"devicePath %s is invalid: err %s, output: %s", devicePath, err.Error(), string(data))
		return false
	}
	logger.Debug(ctx, "dd output for %s: %s", devicePath, string(data))
	return strings.Index(string(data), "1024") > 0
}

func (s *Scsi) getDeviceNameByHCTL(ctx context.Context, h HCTL) (string, error) {
	logger.Info(ctx, "find scsi device name by HCTL, %s %s %s %s",
		h.Host, h.Channel, h.Target, h.Lun)
	if !h.IsFullInfo() {
		return "", errors.New("full HCTL required")
	}
	pattern := fmt.Sprintf("/sys/class/scsi_device/%s:%s:%s:%s/device/block/*",
		h.Host, h.Channel, h.Target, h.Lun)
	matches, err := s.filePath.Glob(pattern)
	if err != nil {
		logger.Error(ctx, "error while try to resolve glob pattern %s", pattern)
		return "", err
	}
	if len(matches) == 0 {
		msg := fmt.Sprintf("can't match block device with provided HCTL, "+
			"%s %s %s %s",
			h.Host, h.Channel, h.Target, h.Lun)
		logger.Error(ctx, msg)
		return "", errors.New(msg)
	}
	// Sort devices and return the first so we don't return a partition
	sort.Strings(matches)
	_, devName := path.Split(matches[0])
	logger.Info(ctx, "device name for HCTL %s %s %s %s is: %s",
		h.Host, h.Channel, h.Target, h.Lun, devName)
	return devName, nil
}

func (s *Scsi) rescanSCSIDeviceByHCTL(ctx context.Context, h HCTL) error {
	logger.Info(ctx, "rescan scsi device by HCTL, %s %s %s %s",
		h.Host, h.Channel, h.Target, h.Lun)
	if !h.IsFullInfo() {
		return errors.New("full HCTL required")
	}
	devicePath := fmt.Sprintf("/sys/class/scsi_device/%s:%s:%s:%s/device/rescan",
		h.Host, h.Channel, h.Target, h.Lun)
	scanFile, err := s.os.OpenFile(devicePath, os.O_APPEND|os.O_WRONLY, 0o200)
	if err != nil {
		logger.Error(ctx, "failed to open %s: %s", devicePath, err.Error())
		return err
	}
	if _, err := scanFile.WriteString("1"); err != nil {
		logger.Error(ctx, "failed to write to rescan file %s: %s", devicePath, err.Error())
		return err
	}
	return scanFile.Close()
}

func (s *Scsi) waitUdevSymlink(ctx context.Context, deviceName string, wwn string) error {
	var checkPath string
	if strings.HasPrefix(deviceName, "dm-") {
		checkPath = diskByIDDMPath + wwn
	} else {
		checkPath = diskByIDSCSIPath + wwn
	}
	symlink, err := s.filePath.EvalSymlinks(checkPath)
	if err != nil {
		msg := fmt.Sprintf("symlink for path %s not found: %s", checkPath, err.Error())
		logger.Info(ctx, msg)
		return errors.New(msg)
	}
	log.Debugf("check path: %s, symlink: %s for wwn: %s", checkPath, symlink, wwn)
	if d := strings.Replace(symlink, "/dev/", "", 1); d != deviceName {
		msg := fmt.Sprintf("udev symlink point to unexpected device: %s", d)
		logger.Info(ctx, msg)
		return errors.New(msg)
	}
	logger.Info(ctx, "udev symlink for %s with WWN %s found", deviceName, wwn)
	return nil
}

// GetNVMESymlink return the NVMe symlink for the given path
func (s *Scsi) GetNVMESymlink(checkPath string) (string, error) {
	retryCount := 1
	for {
		symlink, err := s.filePath.EvalSymlinks(checkPath)
		if err == nil || retryCount == maxRetryCount {
			return symlink, err
		}
		time.Sleep(NVMESymlinkSleepTime * time.Millisecond)
		retryCount = retryCount + 1
	}
}

func (s *Scsi) waitUdevSymlinkNVMe(ctx context.Context, deviceName string, wwn string) error {
	var checkPath string
	if strings.HasPrefix(deviceName, "dm-") {
		checkPath = diskByIDDMPathNVMe + wwn
	} else {
		checkPath = diskByIDSCSIPath + wwn
	}
	symlink, err := s.GetNVMESymlink(checkPath)
	if err != nil {
		msg := fmt.Sprintf("symlink for path %s not found: %s", checkPath, err.Error())
		logger.Error(ctx, msg)
		return errors.New(msg)
	}
	if d := strings.Replace(symlink, "/dev/", "", 1); d != deviceName {
		msg := fmt.Sprintf("udev symlink point to unexpected device: %s", d)
		logger.Error(ctx, msg)
		return errors.New(msg)
	}
	logger.Info(ctx, "udev symlink for %s with WWN %s found", deviceName, wwn)
	return nil
}
