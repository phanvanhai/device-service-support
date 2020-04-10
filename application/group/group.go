package group

import (
	"encoding/json"
	"fmt"

	"github.com/phanvanhai/device-service-support/support/db"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"
)

func (gr *LightGroup) EventCallback(async sdkModel.AsyncValues) error {
	// send event
	gr.asyncCh <- &async

	return nil
}

func (gr *LightGroup) Initialize(dev *models.Device) error {
	isContinue, err := gr.provision(dev)
	if isContinue == false {
		return err
	}

	err = gr.updateDB(*dev)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	return nil
}

func (gr *LightGroup) AddDeviceCallback(groupName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	gr.lc.Debug(fmt.Sprintf("a new Group is added in MetaData:%s", groupName))

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(groupName)
	if err != nil {
		return err
	}

	return gr.Initialize(&dev)
}

func (gr *LightGroup) UpdateDeviceCallback(groupName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	gr.lc.Debug(fmt.Sprintf("a Group is updated in MetaData:%s", groupName))

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(groupName)
	if err != nil {
		return err
	}

	return gr.Initialize(&dev)
}

func (gr *LightGroup) RemoveDeviceCallback(groupName string, protocols map[string]models.ProtocolProperties) error {
	gr.lc.Debug(fmt.Sprintf("a Group is deleted in MetaData:%s", groupName))

	err := gr.nw.DeleteObject(groupName, protocols)
	return err
}

func (gr *LightGroup) HandleReadCommands(groupName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	provision := gr.nw.CheckExist(groupName)
	if provision == false {
		gr.lc.Error("thiet bi chua duoc cap phep")
		return nil, fmt.Errorf("thiet bi chua duoc cap phep")
	}
	sv := sdk.RunningService()
	group, _ := sv.GetDeviceByName(groupName)

	res := make([]*sdkModel.CommandValue, 0, len(reqs))
	for i, r := range reqs {
		gr.lc.Info(fmt.Sprintf("LightGroupApplication.HandleReadCommands: protocols: %v, resource: %v, request: %v", protocols, reqs[i].DeviceResourceName, reqs[i]))

		switch r.DeviceResourceName {
		case ScenarioDr:
			// Lay thong tin tu Support Database va tao ket qua
			relations := db.DB().ElementDotScenario(groupName)
			relationStr, err := json.Marshal(relations)
			if err != nil {
				relationStr = []byte(err.Error())
			}
			newCmvl := sdkModel.NewStringValue(ScenarioDr, 0, string(relationStr))
			res[i] = newCmvl
		case ListDeviceDr:
			// Lay thong tin tu Support Database va tao ket qua
			relations := db.DB().GroupDotElement(groupName)
			devices := make([]string, 0, len(relations))
			for _, relation := range relations {
				devices = append(devices, relation.Element)
			}

			devicesStr, err := json.Marshal(devices)
			if err != nil {
				devicesStr = []byte(err.Error())
			}
			newCmvl := sdkModel.NewStringValue(ListDeviceDr, 0, string(devicesStr))
			res[i] = newCmvl
		case OnOffScheduleDr:
			// Lay thong tin tu Support Database va tao ket qua
			pp, ok := group.Protocols[ScheduleProtocolName]
			onoffs := "[]"
			if ok {
				onoffs, ok = pp[OnOffSchedulePropertyName]
				if !ok {
					onoffs = "[]"
				}
			}
			newCmvl := sdkModel.NewStringValue(OnOffScheduleDr, 0, onoffs)
			res[i] = newCmvl
		case DimmingScheduleDr:
			// Lay thong tin tu Support Database va tao ket qua
			pp, ok := group.Protocols[ScheduleProtocolName]
			dims := "[]"
			if ok {
				dims, ok = pp[DimmingSchedulePropertyName]
				if !ok {
					dims = "[]"
				}
			}
			newCmvl := sdkModel.NewStringValue(DimmingScheduleDr, 0, dims)
			res[i] = newCmvl
		default:
			strErr := fmt.Sprintf("Khong ho tro doc Resource: %s", r.DeviceResourceName)
			gr.lc.Error(strErr)
			return nil, fmt.Errorf(strErr)
		}
	}
	return res, nil
}

func (gr *LightGroup) HandleWriteCommands(groupName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	provision := gr.nw.CheckExist(groupName)
	if provision == false {
		gr.lc.Error("thiet bi chua duoc cap phep")
		return fmt.Errorf("thiet bi chua duoc cap phep")
	}

	if params[0].DeviceResourceName == MethodDr && params[1].DeviceResourceName == DeviceDr {
		method, _ := params[0].StringValue()
		elementName, _ := params[1].StringValue()
		err := gr.elementWriteHandler(groupName, method, elementName)
		return err
	}

	for i, p := range params {
		gr.lc.Info(fmt.Sprintf("LightGroupApplication.HandleWriteCommands: protocols: %v, resource: %v, parameters: %v", protocols, reqs[i].DeviceResourceName, params[i]))
		switch p.DeviceResourceName {
		case OnOffScheduleDr:
			reqValue, _ := p.StringValue()
			err := gr.onOffScheduleWriteHandler(groupName, reqValue)
			if err != nil {
				return err
			}
		case DimmingScheduleDr:
			reqValue, _ := p.StringValue()
			err := gr.dimmingScheduleWriteHandler(groupName, reqValue)
			if err != nil {
				return err
			}
		default:
			var value interface{}
			var err error
			switch p.DeviceResourceName {
			case OnOffDr:
				value, _ = p.BoolValue()
			case DimmingDr:
				value, _ = p.Uint16Value()
			default:
				str := fmt.Sprintf("Khong ho tro ghi Resource:%s", p.DeviceResourceName)
				gr.lc.Error(str)
				return fmt.Errorf(str)
			}

			err = gr.NormalWriteCommand(groupName, &reqs[i], p, value)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
