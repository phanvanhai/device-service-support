package sensor

import (
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (s *Sensor) Provision(dev *models.Device) (continueFlag bool, err error) {
	s.lc.Debug("Bat dau tien trinh kiem tra cap phep")
	defer s.lc.Debug("Ket thuc tien trinh kiem tra cap phep")
	provision := s.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState

	if (provision == false && opstate == models.Disabled) || (provision == true) {
		s.lc.Debug(fmt.Sprintf("Ket thuc tien trinh kiem tra cap phep vi: provision=%t & opstate=%s", provision, opstate))
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
	}
	return true, nil
}

// ConnectToDevice : versionInDev = neu != nil -> yeu cau phai tien hanh dong bo cau hinh
func (s *Sensor) ConnectAndUpdate(dev *models.Device, versionInDev *uint64) (err error) {
	s.lc.Debug("Bat dau tien trinh kiem tra ket noi va dong bo thiet bi")
	defer s.lc.Debug("Ket thuc tien trinh kiem tra ket noi va dong bo thiet bi")

	opstate := dev.OperatingState
	connected := db.DB().GetConnectedStatus(dev.Name)

	// Hien tai khong can xu ly update version config, vi khong co config
	if (connected == false && opstate == models.Disabled) || (connected == true && opstate == models.Enabled) {
		s.lc.Debug(fmt.Sprintf("Ket thuc tien trinh kiem tra ket noi thiet bi vi: connected=%t & opstate=%s", connected, opstate))
		return nil
	}

	err = s.syncConfig(dev, versionInDev)
	if err != nil {
		return
	}
	s.lc.Debug(fmt.Sprintf("cau hinh da duoc dong bo"))
	// dong bo cau hinh la buoc cuoi cung de ket luan: OpState co = true hay khong
	// OpState chi = true khi da duoc cap phep, kiem tra ket noi, cap nhap cau hinh ma khong co bat ky loi gi
	_, err = appModels.UpdateOpState(dev.Name, true)
	return
}

func (s *Sensor) syncConfig(dev *models.Device, versionInDev *uint64) error {
	// update Realtime
	err := appModels.UpdateRealtimeToDevice(s, dev, RealtimeDr)
	if err != nil {
		return err
	}
	return nil
}
