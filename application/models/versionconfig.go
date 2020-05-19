package models

import (
	"fmt"
	"strconv"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/support/common"
)

const VersionConfigInitStringValueConst = "1"

func VersionConfigCheckUpdate(dev models.Device, versionInDev uint64) bool {
	verInDB := GetVersionFromDB(dev)
	return (verInDB == versionInDev)
}

func ReadVersionConfigFromDevice(cm NormalReadCommand, dev *models.Device, resourceName string) (uint64, error) {
	request, ok := NewCommandRequest(dev.Name, resourceName)
	if !ok {
		return 0, fmt.Errorf("khong tim thay resource")
	}
	cmvl, err := cm.NormalReadCommand(dev, request)
	if err != nil {
		return 0, err
	}
	return cmvl[0].Uint64Value()
}

func UpdateVersionConfigToDevice(cm NormalWriteCommand, dev *models.Device, resourceName string, versionInDev uint64) error {
	request, ok := NewCommandRequest(dev.Name, resourceName)
	if !ok {
		return fmt.Errorf("khong tim thay resource")
	}

	cmvlConverted, _ := sdkModel.NewUint64Value(resourceName, 0, versionInDev)

	return cm.NormalWriteCommand(dev, request, cmvlConverted)
}

func FillVerisonToDB(dev *models.Device, value string) {
	SetProperty(dev, common.GeneralProtocolNameConst, common.VerisonConfigConst, value)
}

func GenerateNewVersion(ver1 uint64, ver2 uint64) uint64 {
	max := ver1
	if ver2 > max {
		max = ver2
	}
	for {
		max++
		if max != ver1 && max != ver2 && max != 0 {
			break
		}
	}

	return max
}

func GetVersionFromDB(dev models.Device) (ver uint64) {
	ver, _ = strconv.ParseUint(VersionConfigInitStringValueConst, 10, 64)
	verStr, ok := GetProperty(&dev, common.GeneralProtocolNameConst, common.VerisonConfigConst)
	if !ok {
		return
	}

	ver, _ = strconv.ParseUint(verStr, 10, 64)
	return
}
