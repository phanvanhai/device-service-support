package light

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (l *Light) Provision(dev *models.Device) (continueFlag bool, err error) {
	l.lc.Debug("tien trinh kiem tra cap phep")
	provision := l.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState
	l.lc.Debug(fmt.Sprintf("provison=%t", provision))

	if (provision == false && opstate == models.Disabled) || (provision == true && opstate == models.Enabled) {
		l.lc.Debug(fmt.Sprintf("thoat tien trinh kiem tra cap phep vi: provision=%t & opstate=%s", provision, opstate))
		return true, nil
	}

	sv := sdk.RunningService()
	if provision == false { // opstate = true
		newdev, err := l.nw.AddObject(dev)
		if err != nil {
			l.lc.Error(err.Error())
			continueFlag, err = l.updateOpStateAndConnectdStatus(dev.Name, false)
			return continueFlag, err
		}
		if newdev != nil {
			l.lc.Debug("cap nhap lai thong tin device sau khi da cap phep")
			return false, sv.UpdateDevice(*newdev)
		}
		l.lc.Debug("newdev after provision = nil")
	}

	return true, nil
}

func (l *Light) updateOpStateAndConnectdStatus(devName string, status bool) (bool, error) {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(devName)
	if err != nil {
		return false, err
	}
	var notUpdate = true
	if status == false {
		db.DB().SetConnectedStatus(devName, false)
		if dev.OperatingState == models.Enabled {
			dev.OperatingState = models.Disabled
			l.lc.Debug("cap nhap lai OpState = Disable")
			return false, sv.UpdateDevice(dev)
		}
		return false, nil
	}
	db.DB().SetConnectedStatus(dev.Name, true)
	if dev.OperatingState == models.Disabled {
		dev.OperatingState = models.Enabled
		l.lc.Debug("cap nhap lai OpState = Enabled")
		return false, sv.UpdateDevice(dev)
	}
	return notUpdate, nil
}

func (l *Light) Connect(dev *models.Device) (continueFlag bool, err error) {
	l.lc.Debug("tien trinh kiem tra ket noi thiet bi")
	opstate := dev.OperatingState
	connected := db.DB().GetConnectedStatus(dev.Name)

	if (connected == false && opstate == models.Disabled) || (connected == true && opstate == models.Enabled) {
		l.lc.Debug(fmt.Sprintf("thoat tien trinh kiem tra ket noi thiet bi vi: connected=%t & opstate=%s", connected, opstate))
		return true, nil
	}

	err = l.initDevice(dev.Name)
	if err != nil {
		l.lc.Error(err.Error())
		continueFlag, err = l.updateOpStateAndConnectdStatus(dev.Name, false)
		return
	}
	continueFlag, err = l.updateOpStateAndConnectdStatus(dev.Name, true)

	return
}

func (l *Light) initDevice(devName string) error {
	// update Groups latest
	err := l.updateGroupToDevice(devName)
	if err != nil {
		return err
	}

	// get OnOff-Schedules latest
	err = l.getOnOffSchedulesFromDevice(devName)
	if err != nil {
		return err
	}

	// get Dimming-Schedules latest
	err = l.getDimmingSchedulesFromDevice(devName)
	if err != nil {
		return err
	}
	return nil
}
