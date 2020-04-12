package sensor

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (s *Sensor) Provision(dev *models.Device) (continueFlag bool, err error) {
	s.lc.Debug("tien trinh kiem tra cap phep")
	provision := s.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState
	s.lc.Debug(fmt.Sprintf("provison=%t", provision))

	if (provision == false && opstate == models.Disabled) || (provision == true && opstate == models.Enabled) {
		s.lc.Debug(fmt.Sprintf("thoat tien trinh kiem tra cap phep vi: provision=%t & opstate=%s", provision, opstate))
		return true, nil
	}

	sv := sdk.RunningService()
	if provision == false { // opstate = true
		newdev, err := s.nw.AddObject(dev)
		if err != nil {
			s.lc.Error(err.Error())
			continueFlag, err = s.updateOpStateAndConnectdStatus(dev.Name, false)
			return continueFlag, err
		}
		if newdev != nil {
			s.lc.Debug("cap nhap lai thong tin device sau khi da cap phep")
			return false, sv.UpdateDevice(*newdev)
		}
		s.lc.Debug("newdev after provision = nil")
	}

	return true, nil
}

func (s *Sensor) updateOpStateAndConnectdStatus(devName string, status bool) (bool, error) {
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
			s.lc.Debug("cap nhap lai OpState = Disable")
			return false, sv.UpdateDevice(dev)
		}
		return false, nil
	}
	db.DB().SetConnectedStatus(dev.Name, true)
	if dev.OperatingState == models.Disabled {
		dev.OperatingState = models.Enabled
		s.lc.Debug("cap nhap lai OpState = Enabled")
		return false, sv.UpdateDevice(dev)
	}
	return notUpdate, nil
}

func (s *Sensor) Connect(dev *models.Device) (continueFlag bool, err error) {
	s.lc.Debug("tien trinh kiem tra ket noi thiet bi")
	opstate := dev.OperatingState
	connected := db.DB().GetConnectedStatus(dev.Name)

	if (connected == false && opstate == models.Disabled) || (connected == true && opstate == models.Enabled) {
		s.lc.Debug(fmt.Sprintf("thoat tien trinh kiem tra ket noi thiet bi vi: connected=%t & opstate=%s", connected, opstate))
		return true, nil
	}

	err = s.initDevice(dev.Name)
	if err != nil {
		s.lc.Error(err.Error())
		continueFlag, err = s.updateOpStateAndConnectdStatus(dev.Name, false)
		return
	}
	continueFlag, err = s.updateOpStateAndConnectdStatus(dev.Name, true)

	return
}

func (s *Sensor) initDevice(devName string) error {
	// update Realtim
	err := s.UpdateRealtime(devName)
	if err != nil {
		return err
	}
	return nil
}
