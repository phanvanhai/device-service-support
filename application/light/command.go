package light

import (
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func (l *Light) NormalWriteCommand(dev *models.Device, cmReq *sdkModel.CommandRequest, cmvl *sdkModel.CommandValue) error {
	param := make([]*sdkModel.CommandValue, 1)
	param[0] = cmvl

	req := make([]*sdkModel.CommandRequest, 1)
	req[0] = cmReq

	// Gui lenh
	err := l.nw.WriteCommands(dev.Name, req, param)
	if err != nil {
		appModels.UpdateOpState(dev.Name, false)
		return err
	}
	return nil
}

func (l *Light) NormalReadCommand(dev *models.Device, cmReq *sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	// Gui lenh
	req := make([]*sdkModel.CommandRequest, 1)
	req[0] = cmReq

	cmvl, err := l.nw.ReadCommands(dev.Name, req)
	if err != nil {
		appModels.UpdateOpState(dev.Name, false)
		return nil, err
	}
	return cmvl, nil
}
