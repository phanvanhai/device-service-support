package models

import (
	"strconv"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/support/common"
)

const VersionConfigInitStringValueConst = "1"

type VersionConfig interface {
	// UpdateVersionConfigToDevice update version config to device
	UpdateVersionConfigToDevice(deviceName string, versionConfig uint64) error
	// ReadVersionConfigFromDevice read current version config to device
	ReadVersionConfigFromDevice(deviceName string) (uint64, error)
}

func VersionConfigCheckUpdate(dev models.Device, versionInDev uint64) bool {
	verInDB := getVersionFromDB(dev)
	return (verInDB == versionInDev)
}

func VersionConfigUpdate(versioner VersionConfig, dev models.Device, currentVersionInDev *uint64) error {
	currentVersionInDb := getVersionFromDB(dev)
	if currentVersionInDev == nil {
		currentVersionInDev = &currentVersionInDb
	}
	newVersion := generateVersion(currentVersionInDb, *currentVersionInDev)
	err := versioner.UpdateVersionConfigToDevice(dev.Name, newVersion)
	if err != nil {
		return err
	}
	return updateVersionConfigToDB(dev, newVersion)
}

func generateVersion(ver1 uint64, ver2 uint64) uint64 {
	max := ver1
	if ver2 > max {
		max = ver2
	}
	for {
		max++
		if max != ver1 && max != ver2 {
			break
		}
	}

	return max
}

func updateVersionConfigToDB(dev models.Device, version uint64) error {
	pp, ok := dev.Protocols[common.GeneralProtocolNameConst]
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[common.VerisonConfigConst] = strconv.FormatUint(version, 10)

	sv := sdk.RunningService()
	return sv.UpdateDevice(dev)
}

func getVersionFromDB(dev models.Device) (ver uint64) {
	ver, _ = strconv.ParseUint(VersionConfigInitStringValueConst, 10, 64)
	pp, ok := dev.Protocols[common.GeneralProtocolNameConst]
	if !ok {
		return
	}
	verStr, ok := pp[common.VerisonConfigConst]
	if !ok {
		return
	}

	ver, _ = strconv.ParseUint(verStr, 10, 64)
	return
}
