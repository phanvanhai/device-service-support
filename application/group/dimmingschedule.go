package group

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"

	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func (gr *LightGroup) updateDimmingScheduleElement(groupName string, element string) error {
	sv := sdk.RunningService()
	group, _ := sv.GetDeviceByName(groupName)

	pp, ok := group.Protocols[ScheduleProtocolName]
	schedulesStr := "[]"
	if ok {
		schedulesStr, ok = pp[DimmingSchedulePropertyName]
		if !ok {
			schedulesStr = "[]"
		}
	}

	return gr.WriteCommandByResource(groupName, DimmingScheduleDr, schedulesStr, element)
}

func validateDimmingSchedule(name string, schedulesStr string) (string, error) {
	var schedules []appModels.EdgeDimmingSchedule
	err := json.Unmarshal([]byte(schedulesStr), &schedules)
	if err != nil {
		str := fmt.Sprintf("Loi phan tich thanh chuoi lap lich. Loi:%s", err.Error())
		gr.lc.Error(str)
		return "", fmt.Errorf(str)
	}
	for i := range schedules {
		schedules[i].OwnerName = name
	}

	out, err := json.Marshal(schedules)
	if err != nil {
		str := fmt.Sprintf("Loi tao chuoi lap lich. Loi:%s", err.Error())
		gr.lc.Error(str)
		return "", fmt.Errorf(str)
	}
	return string(out), nil
}

func (gr *LightGroup) dimmingScheduleWriteHandler(groupName string, dimmingStr string) error {
	value, err := validateDimmingSchedule(groupName, dimmingStr)
	if err != nil {
		return err
	}

	sv := sdk.RunningService()
	group, _ := sv.GetDeviceByName(groupName)

	// cap nhap vao DB cua Group
	pp, ok := group.Protocols[ScheduleProtocolName]
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[DimmingSchedulePropertyName] = value
	group.Protocols[ScheduleProtocolName] = pp
	err = sv.UpdateDevice(group)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	// Gui lenh Unicast toi cac device
	errInfos := gr.WriteCommandToAll(groupName, DimmingScheduleDr, value)
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
