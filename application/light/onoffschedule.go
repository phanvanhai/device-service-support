package light

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

// get OnOff-Schedules latest
func (l *Light) GetOnOffSchedulesFromDevice(devName string) error {
	reqs := make([]*sdkModel.CommandRequest, 1)
	request, ok := appModels.NewCommandRequest(devName, OnOffScheduleDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return fmt.Errorf("khong tim thay resource")
	}

	reqs[0] = request
	cmvl, err := l.nw.ReadCommands(devName, reqs)
	if err != nil {
		l.lc.Error(err.Error())
		l.updateOpStateAndConnectdStatus(devName, false)
		return err
	}
	repCmvlValue, _ := cmvl[0].StringValue()
	repConverted, err := appModels.NetValueToOnOffSchedule(l.nw, repCmvlValue, OnOffScheduleLimit, devName)
	if err != nil {
		return err
	}

	// trong DB, luon su dung ID thay Name
	repStr := appModels.OnOffScheduleToStringID(repConverted)
	pp := make(models.ProtocolProperties)
	pp[OnOffSchedulePropertyName] = repStr
	db.DB().UpdateProperty(devName, ScheduleProtocolName, pp)
	return nil
}

func combineOnOffSchedule(devName string, schs []appModels.EdgeOnOffSchedule) []appModels.EdgeOnOffSchedule {
	currentSchs := l.getOnOffSchedulesFromDB(devName)

	var owner string
	isDelete := false
	if len(schs) == 0 {
		owner = devName
		isDelete = true
	} else {
		owner = schs[0].OwnerName
		if appModels.CheckScheduleTime(schs[0].Time) == false {
			isDelete = true
		}
	}

	result := make([]appModels.EdgeOnOffSchedule, 0, OnOffScheduleLimit)
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

func (l *Light) OnOffScheduleWriteHandler(deviceName string, cmReq *sdkModel.CommandRequest, scheduleStr string) error {
	// chuyen doi noi dung string -> schedules
	schedules := appModels.StringNameToOnOffSchedule(scheduleStr)

	// loai bo nhung schedule loi (Owner = "")
	j := 0
	for _, s := range schedules {
		if s.OwnerName != "" {
			schedules[j] = s
			j++
		}
	}
	schedules = schedules[:j]

	newSchs := combineOnOffSchedule(deviceName, schedules)
	if len(newSchs) > OnOffScheduleLimit {
		return fmt.Errorf("loi vuot qua so luong lap lich cho phep")
	}
	reqConverted := appModels.OnOffScheduleEdgeToNetValue(l.nw, newSchs, deviceName, OnOffScheduleLimit)

	// tao CommandValue moi voi r.Value da duoc chuyen doi
	cmvlConverted := sdkModel.NewStringValue(OnOffScheduleDr, 0, reqConverted)
	param := make([]*sdkModel.CommandValue, 1)
	param[0] = cmvlConverted

	req := make([]*sdkModel.CommandRequest, 1)
	req[0] = cmReq

	// Gui lenh
	err := l.nw.WriteCommands(deviceName, req, param)
	if err != nil {
		l.lc.Error(err.Error())
		l.updateOpStateAndConnectdStatus(deviceName, false)
		return err
	}

	// Neu thanh cong, cap nhap lai thong tin trong Support Database
	// truoc khi luu vao DB, can chuyen Name -> ID
	newStr := appModels.OnOffScheduleToStringID(newSchs)
	pp, ok := db.DB().GetProperty(deviceName, ScheduleProtocolName)
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[OnOffSchedulePropertyName] = newStr
	db.DB().UpdateProperty(deviceName, ScheduleProtocolName, pp)

	return nil
}

func (l *Light) getOnOffSchedulesFromDB(deviceName string) []appModels.EdgeOnOffSchedule {
	// Lay thong tin tu Support Database va tao ket qua
	pp, ok := db.DB().GetProperty(deviceName, ScheduleProtocolName)
	onoffs := appModels.ScheduleNilStr
	if ok {
		onoffs, ok = pp[OnOffSchedulePropertyName]
		if !ok {
			onoffs = appModels.ScheduleNilStr
		}
	}

	return appModels.StringIDToOnOffSchedule(onoffs)
}

func (l *Light) SyncOnOffScheduleDBByGroups(deviceName string, groups []string) {
	schedules := l.getOnOffSchedulesFromDB(deviceName)
	if len(schedules) <= 0 {
		return
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
		}
	}
	schedules = schedules[:j]

	// cap nhap vao DB
	repStr := appModels.OnOffScheduleToStringID(schedules)
	pp := make(models.ProtocolProperties)
	pp[OnOffSchedulePropertyName] = repStr
	db.DB().UpdateProperty(deviceName, ScheduleProtocolName, pp)
}
