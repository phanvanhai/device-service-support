package models

import (
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
)

const (
	GetCmdMethod string = "get"
	SetCmdMethod string = "set"
)

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
