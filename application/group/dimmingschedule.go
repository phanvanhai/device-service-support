package group

import (
	"encoding/json"
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func (gr *LightGroup) UpdateDimmingScheduleToElement(groupName string, element string) error {
	sv := sdk.RunningService()
	group, _ := sv.GetDeviceByName(groupName)

	var schedulesStr string
	pp, ok := group.Protocols[ScheduleProtocolName]
	if ok {
		schedulesStr, _ = pp[DimmingSchedulePropertyName]
	}

	schs := appModels.StringIDToDimmingSchedule(schedulesStr)
	// khi gui toi Element, neu schudle = nil -> tao 1 schedule bieu dien gia tri nil
	if len(schs) == 0 {
		scheduleNil := appModels.EdgeDimmingSchedule{
			OwnerName: groupName,
			Time:      appModels.CreateScheuleTimeError()}
		schs = append(schs, scheduleNil)
	}

	schedulesStr = appModels.DimmingScheduleToStringName(schs)
	str := fmt.Sprintf("gui Dimming schedule toi cac device. Dimming=%s", schedulesStr)
	gr.lc.Debug(str)
	return gr.WriteCommandByResource(groupName, DimmingScheduleDr, schedulesStr, element)
}

func (gr *LightGroup) DimmingScheduleWriteHandler(groupName string, dimmingStr string) error {
	sv := sdk.RunningService()
	group, _ := sv.GetDeviceByName(groupName)

	schs := appModels.StringNameToDimmingSchedule(dimmingStr)
	// fill OwnerName
	for i := range schs {
		schs[i].OwnerName = groupName
	}

	// cap nhap vao DB cua Group
	// truoc khi luu vao DB, can chuyen Name -> ID
	strID := appModels.DimmingScheduleToStringID(schs)
	pp, ok := group.Protocols[ScheduleProtocolName]
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[DimmingSchedulePropertyName] = strID
	group.Protocols[ScheduleProtocolName] = pp
	err := sv.UpdateDevice(group)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	// Gui lenh Unicast toi cac device
	// khi gui toi Element, neu schudle = nil -> tao 1 schedule bieu dien gia tri nil
	schs = appModels.StringIDToDimmingSchedule(strID)
	if len(schs) == 0 {
		scheduleNil := appModels.EdgeDimmingSchedule{
			OwnerName: groupName,
			Time:      appModels.CreateScheuleTimeError()}
		schs = append(schs, scheduleNil)
	}
	strName := appModels.DimmingScheduleToStringName(schs)

	errInfos := gr.WriteCommandToAll(groupName, DimmingScheduleDr, strName)
	for _, e := range errInfos {
		if e.Error != "" {
			errStr, _ := json.Marshal(errInfos)
			str := fmt.Sprintf("Loi gui lenh toi cac device. Loi:%s", string(errStr))
			gr.lc.Error(str)
			return fmt.Errorf(str)
		}
	}
	return nil
}
