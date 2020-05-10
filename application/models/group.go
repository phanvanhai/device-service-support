package models

import (
	"encoding/json"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type Group interface {
	WriteGroupToDevice(dev *models.Device, cmReq *sdkModel.CommandRequest, groups []string, grouplimit int) (bool, error)
}

func GroupWriteHandler(grouper Group, dev *models.Device, cmReq *sdkModel.CommandRequest, groupStr string, grouplimit int) error {
	var groups []string
	err := json.Unmarshal([]byte(groupStr), &groups)
	if err != nil {
		return err
	}

	needUpdate, err := grouper.WriteGroupToDevice(dev, cmReq, groups, grouplimit)
	if err != nil {
		return err
	}

	if needUpdate {
		sv := sdk.RunningService()
		err = sv.UpdateDevice(*dev)
		if err != nil {
			return err
		}
	}

	return nil
}
