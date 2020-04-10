package group

import (
	"encoding/json"
	"fmt"

	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	"github.com/phanvanhai/device-service-support/support/db"

	sdkCommand "github.com/phanvanhai/device-sdk-go"
	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func (gr *LightGroup) WriteCommandByResource(group string, resouce string, body interface{}, device string) error {
	// tao & gui lenh toi Element
	newresource := gr.nw.ConvertResourceByDevice(group, resouce, device)
	if newresource == "" {
		str := fmt.Sprintf("Khong tim thay trong Device:%s resource lien quan den Resource:%s cua %s", device, resouce, group)
		gr.lc.Error(str)
		return fmt.Errorf(str)
	}
	params := make(map[string]interface{})
	params[newresource] = body
	newbody, err := json.Marshal(&params)
	if err != nil {
		str := fmt.Sprintf("Loi tao Body cho lenh. Loi:%s", err.Error())
		gr.lc.Error(str)
		return fmt.Errorf(str)
	}
	_, err = sdkCommand.CommandRequest(device, newresource, string(newbody), appModels.SetCmdMethod, "")
	return err
}

func (gr *LightGroup) WriteCommandToAll(group string, resouce string, body interface{}) []ElementError {
	relations := db.DB().GroupDotElement(group)
	errs := make([]ElementError, len(relations))

	for i, r := range relations {
		err := gr.WriteCommandByResource(group, resouce, body, r.Element)
		errs[i].Name = r.Element
		if err == nil {
			errs[i].Error = ""
		} else {
			errs[i].Error = err.Error()
		}
	}
	return errs
}

func (gr *LightGroup) NormalWriteCommand(groupName string, req *sdkModel.CommandRequest, param *sdkModel.CommandValue, value interface{}) error {
	params := make([]*sdkModel.CommandValue, 1)
	params[0] = param

	reqs := make([]*sdkModel.CommandRequest, 1)
	reqs[0] = req

	// Gui lenh broad-cast
	err := gr.nw.WriteCommands(groupName, reqs, params)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	if err != nil {
		str := fmt.Sprintf("Loi lay gia tri cua lenh. Loi:%s", err.Error())
		gr.lc.Error(str)
		return fmt.Errorf(str)
	}

	// Gui lenh Unicast toi cac device
	errInfos := gr.WriteCommandToAll(groupName, param.DeviceResourceName, value)
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
