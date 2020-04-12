package group

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"

	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func (gr *LightGroup) UpdateOnOffScheduleToElement(groupName string, element string) error {
	sv := sdk.RunningService()
	group, _ := sv.GetDeviceByName(groupName)

	var schedulesStr string
	pp, ok := group.Protocols[ScheduleProtocolName]
	if ok {
		schedulesStr, _ = pp[OnOffSchedulePropertyName]
	}

	schs := appModels.StringIDToOnOffSchedule(schedulesStr)
	// khi gui toi Element, neu schudle = nil -> tao 1 schedule bieu dien gia tri nil
	if len(schs) == 0 {
		scheduleNil := appModels.EdgeOnOffSchedule{
			OwnerName: groupName,
			Time:      appModels.CreateScheuleTimeError()}
		schs = append(schs, scheduleNil)
	}

	schedulesStr = appModels.OnOffScheduleToStringName(schs)
	str := fmt.Sprintf("gui OnOff schedule toi cac device. OnOff=%s", schedulesStr)
	gr.lc.Debug(str)
	return gr.WriteCommandByResource(groupName, OnOffScheduleDr, schedulesStr, element)
}

func (gr *LightGroup) OnOffScheduleWriteHandler(groupName string, onoffStr string) error {
	sv := sdk.RunningService()
	group, _ := sv.GetDeviceByName(groupName)

	schs := appModels.StringNameToOnOffSchedule(onoffStr)
	// fill OwnerName
	for i := range schs {
		schs[i].OwnerName = groupName
	}

	// cap nhap vao DB cua Group
	// truoc khi luu vao DB, can chuyen Name -> ID
	strID := appModels.OnOffScheduleToStringID(schs)
	pp, ok := group.Protocols[ScheduleProtocolName]
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[OnOffSchedulePropertyName] = strID
	group.Protocols[ScheduleProtocolName] = pp
	err := sv.UpdateDevice(group)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	// Gui lenh Unicast toi cac device
	// khi gui toi Element, neu schudle = nil -> tao 1 schedule bieu dien gia tri nil
	schs = appModels.StringIDToOnOffSchedule(strID)
	if len(schs) == 0 {
		scheduleNil := appModels.EdgeOnOffSchedule{
			OwnerName: groupName,
			Time:      appModels.CreateScheuleTimeError()}
		schs = append(schs, scheduleNil)
	}
	strName := appModels.OnOffScheduleToStringName(schs)

	errInfos := gr.WriteCommandToAll(groupName, OnOffScheduleDr, strName)
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
