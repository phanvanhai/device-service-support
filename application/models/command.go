package models

import (
	"encoding/json"
	"fmt"
	"time"

	sdkCommand "github.com/edgexfoundry/device-sdk-go"
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	nw "github.com/phanvanhai/device-service-support/network"
	"github.com/phanvanhai/device-service-support/support/db"
)

const (
	GetCmdMethod string = "get"
	SetCmdMethod string = "set"
)

type ElementError struct {
	Name  string
	Error string
}

type NormalWriteCommand interface {
	NormalWriteCommand(dev *models.Device, cmReq *sdkModel.CommandRequest, cmvl *sdkModel.CommandValue) error
}

type NormalReadCommand interface {
	NormalReadCommand(dev *models.Device, cmReq *sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error)
}

type WriteCommandToOtherDevice interface {
	WriteCommandToOtherDeviceByResource(nw nw.Network, group string, resouce string, body interface{}, device string) error
}

func NewCommandRequest(deviceName string, deviceResource string) (*sdkModel.CommandRequest, bool) {
	sv := sdk.RunningService()
	resource, ok := sv.DeviceResource(deviceName, deviceResource, "")
	if !ok {
		return nil, false
	}
	cmrq := sdkModel.CommandRequest{
		DeviceResourceName: deviceResource,
		Attributes:         resource.Attributes,
		Type:               sdkModel.ParseValueType(resource.Properties.Value.Type),
	}
	return &cmrq, true
}

func GroupWriteUnicastCommandToAll(nw nw.Network, group string, resouce string, body interface{}) []ElementError {
	relations := db.DB().GroupDotElement(group)
	errs := make([]ElementError, len(relations))

	for i, r := range relations {
		time.Sleep(2 * time.Second)
		err := WriteCommandToOtherDeviceByResource(nw, group, resouce, body, r.Element)
		errs[i].Name = r.Element
		if err == nil {
			errs[i].Error = ""
		} else {
			errs[i].Error = err.Error()
		}
	}
	return errs
}

func WriteCommandToOtherDeviceByResource(nw nw.Network, srcDev string, resouce string, body interface{}, destDev string) error {
	// tao & gui lenh toi Element
	newresource := nw.ConvertResourceByDevice(srcDev, resouce, destDev)
	if newresource == "" {
		return fmt.Errorf("Khong tim thay trong Device:%s resource lien quan den Resource:%s cua %s", destDev, resouce, srcDev)
	}
	params := make(map[string]interface{})
	params[newresource] = body
	newbody, err := json.Marshal(&params)
	if err != nil {
		return fmt.Errorf("Loi tao Body cho lenh. Loi:%s", err.Error())
	}
	_, err = sdkCommand.CommandRequest(destDev, newresource, string(newbody), SetCmdMethod, "")
	return err
}
