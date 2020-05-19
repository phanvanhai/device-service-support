package light

import (
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (l *Light) Provision(dev *models.Device) (continueFlag bool, err error) {
	l.lc.Debug("Bat dau tien trinh kiem tra cap phep")
	defer l.lc.Debug("Ket thuc tien trinh kiem tra cap phep")

	provision := l.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState

	if (provision == false && opstate == models.Disabled) || (provision == true) {
		l.lc.Debug(fmt.Sprintf("Ket thuc tien trinh kiem tra cap phep vi: provision=%t", provision))
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

			// Khoi tao Schedule, VersionConfig trong Database
			appModels.FillOnOffScheduleToDB(newdev, appModels.ScheduleNilStr)
			appModels.FillDimmingScheduleToDB(newdev, appModels.ScheduleNilStr)
			appModels.FillVerisonToDB(newdev, appModels.VersionConfigInitStringValueConst)

			return false, sv.UpdateDevice(*newdev)
		}
	}
	return true, nil
}

// ConnectAndUpdate : versionInDev : neu != nil -> yeu cau phai tien hanh dong bo cau hinh
func (l *Light) ConnectAndUpdate(dev *models.Device, versionInDev *uint64) (err error) {
	l.lc.Debug("Bat dau tien trinh kiem tra ket noi va dong bo thiet bi")
	defer l.lc.Debug("Ket thuc tien trinh kiem tra ket noi va dong bo thiet bi")

	opstate := dev.OperatingState
	connected := db.DB().GetConnectedStatus(dev.Name)

	if versionInDev == nil {
		if (connected == false && opstate == models.Disabled) || (connected == true && opstate == models.Enabled) {
			l.lc.Debug(fmt.Sprintf("Ket thuc tien trinh kiem tra ket noi thiet bi vi: connected=%t & opstate=%s", connected, opstate))
			return nil
		}

		// do something: ...

		ver, err := appModels.ReadVersionConfigFromDevice(l, dev, PingDr)
		if err != nil {
			l.lc.Error(fmt.Sprintf("Ket thuc tien trinh kiem tra ket noi va dong bo thiet bi vi:%s", err.Error()))
			return err
		}
		versionInDev = &ver
	}

	if appModels.GetVersionFromDB(*dev) != *versionInDev {
		err = l.syncConfig(dev)
		if err != nil {
			return
		}
	}
	l.lc.Debug(fmt.Sprintf("cau hinh da duoc dong bo"))
	// dong bo cau hinh la buoc cuoi cung de ket luan: OpState co = true hay khong
	// OpState chi = true khi da duoc cap phep, kiem tra ket noi, cap nhap cau hinh ma khong co bat ky loi gi
	_, err = appModels.UpdateOpState(dev.Name, true)
	return
}

func (l *Light) syncConfig(dev *models.Device) error {
	var err error
	l.lc.Debug("update Realtime")
	err = appModels.UpdateRealtimeToDevice(l, dev, RealtimeDr)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	// update Groups latest
	l.lc.Debug("update groups to Device")
	err = appModels.UpdateGroupToDevice(l, l.nw, dev, GroupDr, GroupLimit)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	l.lc.Debug("update on-off schedule of Device")
	// get OnOff-Schedules latest
	err = appModels.UpdateOnOffSchedulesConfigToDevice(l, l.nw, dev, OnOffScheduleDr, OnOffScheduleLimit)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	l.lc.Debug("update dimming schedule to Device")
	// get Dimming-Schedules latest
	err = appModels.UpdateDimmingSchedulesConfigToDevice(l, l.nw, dev, DimmingScheduleDr, DimmingScheduleLimit)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	// update versionconfig
	l.lc.Debug("update version config to Device")
	err = appModels.UpdateVersionConfigToDevice(l, dev, PingDr, appModels.GetVersionFromDB(*dev))
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	// vi cac buoc tren khong thay doi noi dung DB nen khong can update vao DB
	return nil
}
