package light

import (
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
)

// get Dimming-Schedules latest
func (l *Light) UpdateDimmingSchedulesToDevice(deviceName string) error {
	schs := l.GetDimmingSchedulesFromDB(deviceName)
	reqConverted := appModels.DimmingScheduleEdgeToNetValue(l.nw, schs, deviceName, DimmingScheduleLimit)

	reqs := make([]*sdkModel.CommandRequest, 1)
	request, ok := appModels.NewCommandRequest(deviceName, DimmingScheduleDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return fmt.Errorf("khong tim thay resource")
	}
	reqs[0] = request

	cmvlConverted := sdkModel.NewStringValue(DimmingScheduleDr, 0, reqConverted)
	param := make([]*sdkModel.CommandValue, 0, 1)
	param = append(param, cmvlConverted)

	err := l.nw.WriteCommands(deviceName, reqs, param)
	return err
}

func combineDimmingSchedule(deviceName string, schs []appModels.EdgeDimmingSchedule) []appModels.EdgeDimmingSchedule {
	currentSchs := l.GetDimmingSchedulesFromDB(deviceName)

	var owner string
	isDelete := false
	if len(schs) == 0 {
		owner = deviceName
		isDelete = true
	} else {
		owner = schs[0].OwnerName
		if appModels.CheckScheduleTime(schs[0].Time) == false {
			isDelete = true
		}
	}

	result := make([]appModels.EdgeDimmingSchedule, 0, DimmingScheduleLimit)
	// loai bo schedules cu (co OwnerName = owner) trong danh sach hien tai:
	for _, s := range currentSchs {
		if s.OwnerName != owner {
			result = append(result, s)
		}
	}

	// neu truong hop la them schedule:
	if !isDelete {
		result = append(result, schs...)
	}
	return result
}

func (l *Light) DimmingScheduleWriteHandler(deviceName string, cmReq *sdkModel.CommandRequest, scheduleStr string) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}
	// chuyen doi noi dung string -> schedules
	schedules := appModels.StringNameToDimmingSchedule(scheduleStr)

	// loai bo nhung schedule loi (Owner = "")
	j := 0
	for _, s := range schedules {
		if s.OwnerName != "" {
			schedules[j] = s
			j++
		}
	}
	schedules = schedules[:j]

	newSchs := combineDimmingSchedule(deviceName, schedules)
	if len(newSchs) > DimmingScheduleLimit {
		return fmt.Errorf("loi vuot qua so luong lap lich cho phep")
	}
	reqConverted := appModels.DimmingScheduleEdgeToNetValue(l.nw, newSchs, deviceName, DimmingScheduleLimit)

	// tao CommandValue moi voi r.Value da duoc chuyen doi
	cmvlConverted := sdkModel.NewStringValue(DimmingScheduleDr, 0, reqConverted)
	param := make([]*sdkModel.CommandValue, 1)
	param[0] = cmvlConverted

	req := make([]*sdkModel.CommandRequest, 1)
	req[0] = cmReq

	// Gui lenh
	err = l.nw.WriteCommands(deviceName, req, param)
	if err != nil {
		l.lc.Error(err.Error())
		l.updateOpStateAndConnectdStatus(deviceName, false)
		return err
	}

	// Neu thanh cong, cap nhap lai thong tin trong Support Database
	// truoc khi luu vao DB, can chuyen Name -> ID
	newStr := appModels.DimmingScheduleToStringID(newSchs)
	pp, ok := dev.Protocols[ScheduleProtocolName]
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[DimmingSchedulePropertyName] = newStr
	dev.Protocols[ScheduleProtocolName] = pp
	err = sv.UpdateDevice(dev)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	return nil
}

func (l *Light) GetDimmingSchedulesFromDB(deviceName string) []appModels.EdgeDimmingSchedule {
	// Lay thong tin tu Support Database va tao ket qua
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		l.lc.Error(err.Error())
		return nil
	}

	pp, ok := dev.Protocols[ScheduleProtocolName]
	var schsID string
	if ok {
		schsID, _ = pp[DimmingSchedulePropertyName]
	}

	return appModels.StringIDToDimmingSchedule(schsID)
}

func (l *Light) SyncDimmingScheduleDBByGroups(deviceName string, groups []string) (string, bool) {
	var change = false
	schedules := l.GetDimmingSchedulesFromDB(deviceName)
	if len(schedules) <= 0 {
		return appModels.ScheduleNilStr, change
	}

	// loai bo nhung schudule khong lien quan den group hay chinh device
	j := 0
	for _, s := range schedules {
		var isExist = false
		if s.OwnerName == deviceName {
			isExist = true
		} else {
			for _, g := range groups {
				if s.OwnerName == g {
					isExist = true
					break
				}
			}
		}

		if isExist == true {
			schedules[j] = s
			j++
		} else {
			change = true
		}
	}
	schedules = schedules[:j]

	newStr := appModels.DimmingScheduleToStringID(schedules)
	return newStr, change
}
