package light

import (
	"fmt"

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

// implement me!
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

func (l *Light) Initialize(dev *models.Device) error {
	isContinue, err := l.Provision(dev)
	if isContinue == false {
		return err
	}

	isContinue, err = l.Connect(dev)
	// TODO: get config from device --> update db.DB().Protocols(dev) map[string]interface{}
	return err
}

func (l *Light) Provision(dev *models.Device) (continueFlag bool, err error) {
	l.lc.Debug("tien trinh cap phep")
	provision := l.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState
	l.lc.Debug("provison=", provision)

	if (provision == false && opstate == models.Disabled) || (provision == true && opstate == models.Enabled) {
		l.lc.Debug("thoat tien trinh cap phep vi: provision=", provision, "& opstate=", opstate)
		return true, nil
	}

	sv := sdk.RunningService()
	if provision == false { // opstate = true
		newdev, err := l.nw.AddObject(dev)
		if err != nil {
			l.lc.Error(err.Error())
			db.DB().SetConnectedStatus(dev.Name, false)
			dev.OperatingState = models.Disabled
			return false, sv.UpdateDevice(*dev)
		}
		if newdev != nil {
			l.lc.Debug("cap nhap lai thong tin device sau khi da cap phep")
			return false, sv.UpdateDevice(*newdev)
		}
		l.lc.Debug("newdev after provision = nil")
	}

	return true, nil
}

func (l *Light) Connect(dev *models.Device) (continueFlag bool, err error) {
	l.lc.Debug("tien trinh ket noi thiet bi")
	opstate := dev.OperatingState
	connected := db.DB().GetConnectedStatus(dev.Name)

	if (connected == false && opstate == models.Disabled) || (connected == true && opstate == models.Enabled) {
		l.lc.Debug("thoat tien trinh ket noi thiet bi vi: connected=", connected, "& opstate=", opstate)
		return true, nil
	}

	sv := sdk.RunningService()
	err = l.initDevice(dev.Name)
	if err != nil {
		l.lc.Error(err.Error())
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
		l.lc.Debug("cap nhap lai OpState sau khi da ket noi thanh cong")
		return false, sv.UpdateDevice(*dev)
	}

	return true, nil
}

// implement me!
func (l *Light) initDevice(devName string) error {
	// grs := db.DB().ElementDotGroups(devName)
	return nil
}

func (l *Light) AddDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	l.lc.Debug("a new Device is added in MetaData:", deviceName)

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return err
	}

	isContinue, err := l.Provision(&dev)
	if isContinue == false {
		return err
	}

	isContinue, err = l.Connect(&dev)
	return err
}

func (l *Light) UpdateDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	l.lc.Debug("a Device is updated in MetaData:", deviceName)

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return err
	}

	isContinue, err := l.Provision(&dev)
	if isContinue == false {
		return err
	}

	isContinue, err = l.Connect(&dev)
	return err
}

func (l *Light) RemoveDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties) error {
	l.lc.Debug("a Device is deleted in MetaData:", deviceName)

	err := l.nw.DeleteObject(deviceName, protocols)
	return err
}

// implement me!
func (l *Light) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	provision := l.nw.CheckExist(deviceName)
	if provision == false {
		l.lc.Error("thiet bi chua duoc cap phep")
		return nil, fmt.Errorf("thiet bi chua duoc cap phep")
	}
	connected := db.DB().GetConnectedStatus(deviceName)
	if connected == false {
		l.lc.Error("thiet bi chua duoc ket noi")
		return nil, fmt.Errorf("thiet bi chua duoc ket noi")
	}

	res := make([]*sdkModel.CommandValue, 0, len(reqs))
	for i, r := range reqs {
		l.lc.Info(fmt.Sprintf("SimpleDriver.HandleReadCommands: protocols: %v, resource: %v, request: %v", protocols, reqs[i].DeviceResourceName, reqs[i]))

		req := make([]*sdkModel.CommandRequest, 0, 1)
		req = append(req, &r)

		cmvl, err := l.nw.ReadCommands(deviceName, req)
		if err != nil {
			return nil, err
		}
		res = append(res, cmvl...)
	}
	return res, nil
}

// implement me!
func (l *Light) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	provision := l.nw.CheckExist(deviceName)
	if provision == false {
		l.lc.Error("thiet bi chua duoc cap phep")
		return fmt.Errorf("thiet bi chua duoc cap phep")
	}
	connected := db.DB().GetConnectedStatus(deviceName)
	if connected == false {
		l.lc.Error("thiet bi chua duoc ket noi")
		return fmt.Errorf("thiet bi chua duoc ket noi")
	}

	for i, r := range params {
		l.lc.Info(fmt.Sprintf("SimpleDriver.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[i].DeviceResourceName, params[i]))
		param := make([]*sdkModel.CommandValue, 0, 1)
		param = append(param, r)

		req := make([]*sdkModel.CommandRequest, 0, 1)
		req = append(req, &reqs[i])

		err := l.nw.WriteCommands(deviceName, req, param)
		if err != nil {
			return err
		}
	}
	return nil
}
