package light

import (
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func (l *Light) WriteRealtimeToDevice(deviceName string, time uint64) error {
	request, ok := appModels.NewCommandRequest(deviceName, RealtimeDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return fmt.Errorf("khong tim thay resource")
	}

	cmvlConverted, err := sdkModel.NewUint64Value(RealtimeDr, 0, time)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}
	param := make([]*sdkModel.CommandValue, 0, 1)
	param = append(param, cmvlConverted)

	reqs := make([]*sdkModel.CommandRequest, 1)
	reqs[0] = request
	err = l.nw.WriteCommands(deviceName, reqs, param)
	if err != nil {
		l.lc.Error(err.Error())
		appModels.UpdateOpState(deviceName, false)
		return err
	}
	return nil
}
