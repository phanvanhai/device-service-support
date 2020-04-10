package gateway

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
)

func (g *Gateway) EventCallback(async sdkModel.AsyncValues) error {
	// gateway do not have event
	return nil
}

func (g *Gateway) Initialize(dev *models.Device) error {
	return nil
}

func (g *Gateway) AddDeviceCallback(gatewayName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	g.lc.Debug(fmt.Sprintf("a new gateway is added in MetaData:%s", gatewayName))
	return nil
}

func (g *Gateway) UpdateDeviceCallback(gatewayName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	g.lc.Debug(fmt.Sprintf("a gateway is updated in MetaData:%s", gatewayName))
	return nil
}

func (g *Gateway) RemoveDeviceCallback(gatewayName string, protocols map[string]models.ProtocolProperties) error {
	g.lc.Debug(fmt.Sprintf("a gateway is deleted in MetaData:%s", gatewayName))
	return nil
}

func (g *Gateway) HandleReadCommands(gatewayName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	res := make([]*sdkModel.CommandValue, 0, len(reqs))
	for i, r := range reqs {
		g.lc.Info(fmt.Sprintf("GatewayApplication.HandleReadCommands: protocols: %v, resource: %v, request: %v", protocols, reqs[i].DeviceResourceName, reqs[i]))
		switch r.DeviceResourceName {
		case OnOffRelay1Dr:
			value := g.getRelay()
			res[i], _ = sdkModel.NewBoolValue(r.DeviceResourceName, 0, value)
		case EventDr:
			res[i] = sdkModel.NewStringValue(r.DeviceResourceName, 0, "implement me!")
		default:
			strErr := fmt.Sprintf("Khong ho tro doc Resource: %s", r.DeviceResourceName)
			g.lc.Error(strErr)
			return nil, fmt.Errorf(strErr)
		}
	}
	return nil, nil
}

func (g *Gateway) HandleWriteCommands(gatewayName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	for i, p := range params {
		g.lc.Info(fmt.Sprintf("GatewayApplication.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[i].DeviceResourceName, params[i]))
		switch p.DeviceResourceName {
		case OnOffRelay1Dr:
			value, _ := p.BoolValue()
			g.setRelay(value)
		case UpdateDeviceFirmwareDr:
			value, _ := p.StringValue()
			return g.updateFirmware(value)
		default:
			strErr := fmt.Sprintf("Khong ho tro ghi Resource: %s", p.DeviceResourceName)
			g.lc.Error(strErr)
			return fmt.Errorf(strErr)
		}
	}
	return nil
}
