package group

import (
	"encoding/json"
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func (gr *LightGroup) NormalWriteCommand(groupName string, req *sdkModel.CommandRequest, param *sdkModel.CommandValue) error {
	value := param.ValueToString()

	params := make([]*sdkModel.CommandValue, 1)
	params[0] = param

	reqs := make([]*sdkModel.CommandRequest, 1)
	reqs[0] = req

	// Gui lenh broad-cast
	err := gr.nw.WriteCommands(groupName, reqs, params)
	if err != nil {
		return err
	}

	// Gui lenh Unicast toi cac device
	errInfos := appModels.GroupWriteUnicastCommandToAll(gr.nw, groupName, param.DeviceResourceName, value)
	for _, e := range errInfos {
		if e.Error != "" {
			errStr, _ := json.Marshal(errInfos)
			return fmt.Errorf("Loi gui lenh toi cac device. Loi:%s", string(errStr))
		}
	}
	return nil
}
