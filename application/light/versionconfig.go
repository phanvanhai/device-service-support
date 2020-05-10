package light

import (
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
)

// UpdateVersionConfigToDevice update version config to device
func (l *Light) UpdateVersionConfigToDevice(deviceName string, versionConfig uint64) error {
	reqs := make([]*sdkModel.CommandRequest, 1)

	request, ok := appModels.NewCommandRequest(deviceName, PingDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return fmt.Errorf("khong tim thay resource")
	}

	cmvlConverted, _ := sdkModel.NewUint64Value(PingDr, 0, versionConfig)
	param := make([]*sdkModel.CommandValue, 0, 1)
	param = append(param, cmvlConverted)

	reqs[0] = request
	err := l.nw.WriteCommands(deviceName, reqs, param)

	if err != nil {
		l.lc.Error(err.Error())
		appModels.UpdateOpState(deviceName, false)
		return err
	}

	return nil
}

// ReadVersionConfigFromDevice read current version config to device
func (l *Light) ReadVersionConfigFromDevice(deviceName string) (uint64, error) {
	reqs := make([]*sdkModel.CommandRequest, 1)
	request, ok := appModels.NewCommandRequest(deviceName, PingDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return 0, fmt.Errorf("khong tim thay resource")
	}

	reqs[0] = request
	cmvl, err := l.nw.ReadCommands(deviceName, reqs)
	if err != nil {
		l.lc.Error(err.Error())
		appModels.UpdateOpState(deviceName, false)
		return 0, err
	}

	return cmvl[0].Uint64Value()
}
