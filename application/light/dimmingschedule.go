package light

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

// get Dimming-Schedules latest
func (l *Light) getDimmingSchedulesFromDevice(devName string) error {
	reqs := make([]*sdkModel.CommandRequest, 1)
	request, ok := appModels.NewCommandRequest(devName, DimmingScheduleDr)
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
	repConverted, err := appModels.NetValueToDimmingSchedule(l.nw, repCmvlValue, DimmingScheduleLimit, devName)
	if err != nil {
		return err
	}
	repStr := appModels.DimmingScheduleToString(repConverted)
	pp := make(models.ProtocolProperties)
	pp[DimmingSchedulePropertyName] = repStr
	db.DB().UpdateProperty(devName, ScheduleProtocolName, pp)
	return nil
}

func combineDimmingSchedule(devName string, schs []appModels.EdgeDimmingSchedule) []appModels.EdgeDimmingSchedule {
	var currentSchs []appModels.EdgeDimmingSchedule
	pp, ok := db.DB().GetProperty(devName, ScheduleProtocolName)
	if !ok {
		return nil
	}
	str, ok := pp[DimmingSchedulePropertyName]
	if !ok {
		str = "[]"
	}
	json.Unmarshal([]byte(str), &currentSchs)

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

	result := make([]appModels.EdgeDimmingSchedule, 0, DimmingScheduleLimit)
	if isDelete {
		for _, s := range schs {
			if s.OwnerName != owner {
				result = append(result, s)
			}
		}
	} else {
		result = append(result, currentSchs...)
		result = append(result, schs...)
	}
	return result
}

func (l *Light) dimmingScheduleWriteHandler(deviceName string, cmReq *sdkModel.CommandRequest, scheduleStr string) error {
	// chuyen doi noi dung r.Value
	var schedules []appModels.EdgeDimmingSchedule
	err := json.Unmarshal([]byte(scheduleStr), &schedules)
	if err != nil {
		return err
	}

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
	newStr := appModels.DimmingScheduleToString(newSchs)
	pp := make(models.ProtocolProperties)
	pp[DimmingSchedulePropertyName] = newStr
	db.DB().UpdateProperty(deviceName, ScheduleProtocolName, pp)

	return nil
}

func (l *Light) getDimmingSchedulesFromDB(deviceName string) string {
	// Lay thong tin tu Support Database va tao ket qua
	pp, ok := db.DB().GetProperty(deviceName, ScheduleProtocolName)
	onoffs := "[]"
	if ok {
		onoffs, ok = pp[DimmingSchedulePropertyName]
		if !ok {
			onoffs = "[]"
		}
	}
	return onoffs
}
