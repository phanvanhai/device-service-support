package light

import (
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (l *Light) Provision(dev *models.Device) (continueFlag bool, err error) {
	l.lc.Debug("tien trinh kiem tra cap phep")
	provision := l.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState
	l.lc.Debug(fmt.Sprintf("provison=%t", provision))

	if (provision == false && opstate == models.Disabled) || (provision == true) {
		l.lc.Debug(fmt.Sprintf("thoat tien trinh kiem tra cap phep vi: provision=%t & opstate=%s", provision))
		return true, nil
	}

	sv := sdk.RunningService()
	if provision == false { // opstate = true
		newdev, err := l.nw.AddObject(dev)
		if err != nil {
			l.lc.Error(err.Error())
			continueFlag, err = appModels.UpdateOpState(dev.Name, false)
			return continueFlag, err
		}
		if newdev != nil {
			l.lc.Debug("cap nhap lai thong tin device sau khi da cap phep")

			// Khoi tao Schedule trong Database
			pp := make(models.ProtocolProperties)
			pp[OnOffSchedulePropertyName] = appModels.ScheduleNilStr
			pp[DimmingSchedulePropertyName] = appModels.ScheduleNilStr
			newdev.Protocols[ScheduleProtocolName] = pp

			return false, sv.UpdateDevice(*newdev)
		}
		l.lc.Debug("newdev after provision = nil")
	}

	return true, nil
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
		continueFlag, err = appModels.UpdateOpState(dev.Name, false)
		return
	}
	continueFlag, err = appModels.UpdateOpState(dev.Name, true)

	return
}

func (l *Light) initDevice(deviceName string) error {
	// l.lc.Debug("update on-off schedule of Device")
	// // get OnOff-Schedules latest
	// err := l.UpdateOnOffSchedulesToDevice(deviceName)
	// if err != nil {
	// 	l.lc.Error(err.Error())
	// 	return err
	// }

	// l.lc.Debug("update schedule to Device")
	// // get Dimming-Schedules latest
	// err = l.UpdateDimmingSchedulesToDevice(deviceName)
	// if err != nil {
	// 	l.lc.Error(err.Error())
	// 	return err
	// }

	// // update Groups latest
	// l.lc.Debug("update groups to Device")
	// err = l.UpdateGroupToDevice(deviceName)
	// if err != nil {
	// 	l.lc.Error(err.Error())
	// 	return err
	// }

	return nil
}
