package models

import (
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/support/db"
)

const (
	ScheduleNilStr = "[]"
	TimeError      = 0x00000000
	OwnerMe        = 0xFFFE
)

func CheckScheduleTime(t uint32) bool {
	return (t != TimeError)
}

func CreateScheuleTimeError() uint32 {
	return TimeError
}

func UpdateOpState(deviceName string, status bool) (opStateIsReadyInDB bool, err error) {
	opStateIsReadyInDB = true
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return
	}

	if status == false {
		db.DB().SetConnectedStatus(deviceName, false)
		if dev.OperatingState == models.Enabled {
			dev.OperatingState = models.Disabled
			opStateIsReadyInDB = false
			return opStateIsReadyInDB, sv.UpdateDevice(dev)
		}
		return
	}
	db.DB().SetConnectedStatus(dev.Name, true)
	if dev.OperatingState == models.Disabled {
		dev.OperatingState = models.Enabled
		opStateIsReadyInDB = false
		return opStateIsReadyInDB, sv.UpdateDevice(dev)
	}
	return
}

func SetProperty(dev *models.Device, protocolName string, propertyName string, value string) {
	if dev == nil {
		return
	}

	pp, ok := dev.Protocols[protocolName]
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[protocolName] = propertyName
	dev.Protocols[protocolName] = pp
}

func GetProperty(dev *models.Device, protocolName string, propertyName string) (string, bool) {
	if dev == nil {
		return "", false
	}

	pp, ok := dev.Protocols[protocolName]
	if !ok {
		return "", false
	}
	str, ok := pp[protocolName]
	return str, ok
}
