package light

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdk "github.com/phanvanhai/device-sdk-go"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"

	nw "github.com/phanvanhai/device-service-support/network"
	db "github.com/phanvanhai/device-service-support/support/db"
	tc "github.com/phanvanhai/device-service-support/transceiver"
)

func initialize(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transceiver) (*Light, error) {
	l := &Light{
		lc:      lc,
		asyncCh: asyncCh,
		nw:      nw,
		tc:      tc,
	}
	return l, nil
}

func (l *Light) EventCallback(async sdkModel.AsyncValues) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(async.DeviceName)
	if err != nil {
		return err
	}
	_, err = l.Connect(&dev)
	if err != nil {
		l.lc.Error(err.Error())
	}

	// ....
	// doi voi nhung truong hop can gui lenh toi Device, co the kiem tra err ben tren:
	// neu loi thi khong phai gui nua
	return nil
}

func (l *Light) Provision(dev *models.Device) (continueFlag bool, err error) {
	provision := l.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState

	if (provision == false && opstate == models.Disabled) || (provision == true && opstate == models.Enabled) {
		return true, nil
	}

	sv := sdk.RunningService()
	if provision == false { // opstate = true
		newdev, err := l.nw.AddObject(dev)
		if err != nil {
			l.lc.Error(err.Error())
			dev.OperatingState = models.Disabled
			return false, sv.UpdateDevice(*dev)
		}
		if newdev != nil {
			return false, sv.UpdateDevice(*newdev)
		}
	}

	return true, nil
}

func (l *Light) Connect(dev *models.Device) (continueFlag bool, err error) {
	opstate := dev.OperatingState
	connected := db.DB().GetConnectedStatus(dev.Name)

	if (connected == false && opstate == models.Disabled) || (connected == true && opstate == models.Enabled) {
		return true, nil
	}

	sv := sdk.RunningService()
	err = l.initDevice(dev.Name)
	if err != nil {
		db.DB().SetConnectedStatus(dev.Name, false)
		if dev.OperatingState == models.Enabled {
			dev.OperatingState = models.Disabled
			return false, sv.UpdateDevice(*dev)
		}
		return true, err
	}
	db.DB().SetConnectedStatus(dev.Name, true)
	if dev.OperatingState == models.Disabled {
		dev.OperatingState = models.Enabled
		return false, sv.UpdateDevice(*dev)
	}

	return true, nil
}

func (l *Light) initDevice(devName string) error {
	// grs := db.DB().ElementDotGroups(devName)
	return nil
}

func (l *Light) AddDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	l.lc.Debug("a new Device is added in MetaData: %s", deviceName)

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return err
	}
	l.nw.UpdateObjectCallback(&dev)
	db.DB().UpdateObject(&dev)

	isContinue, err := l.Provision(&dev)
	if isContinue == false {
		return err
	}

	isContinue, err = l.Connect(&dev)
	return err
}

func (l *Light) UpdateDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	l.lc.Debug("a Device is updated in MetaData: %s", deviceName)

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return err
	}
	l.nw.UpdateObjectCallback(&dev)
	db.DB().UpdateObject(&dev)

	isContinue, err := l.Provision(&dev)
	if isContinue == false {
		return err
	}

	isContinue, err = l.Connect(&dev)
	return err
}

func (l *Light) RemoveDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties) error {
	l.lc.Debug("a Device is deleted in MetaData: %s", deviceName)

	err := l.nw.DeleteObject(deviceName, protocols)

	l.nw.DeleteObjectCallback(deviceName)
	db.DB().DeleteDevice(deviceName)
	return err
}

func (l *Light) HandleReadCommands(objectID string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	return nil, nil
}

func (l *Light) HandleWriteCommands(objectID string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	return nil
}
