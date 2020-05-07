package sensor

import (
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
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
			continueFlag, err = appModels.UpdateOpState(dev.Name, false)
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
		continueFlag, err = appModels.UpdateOpState(dev.Name, false)
		return
	}
	continueFlag, err = appModels.UpdateOpState(dev.Name, true)

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
