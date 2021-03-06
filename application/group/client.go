package group

import (
	"encoding/json"
	"fmt"

	"github.com/phanvanhai/device-service-support/support/db"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func (gr *LightGroup) EventCallback(async sdkModel.AsyncValues) error {
	// group do not have event
	return nil
}

func (gr *LightGroup) Initialize(dev *models.Device) error {
	isContinue, err := gr.Provision(dev)
	if isContinue == false {
		return err
	}

	err = gr.ConnectAndUpdate(dev)
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

	gr.deleteAllElement(groupName)

	err := gr.nw.DeleteObject(groupName, protocols)
	return err
}

func (gr *LightGroup) HandleReadCommands(groupName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	sv := sdk.RunningService()
	group, _ := sv.GetDeviceByName(groupName)

	res := make([]*sdkModel.CommandValue, len(reqs))
	for i, r := range reqs {
		gr.lc.Info(fmt.Sprintf("LightGroupApplication.HandleReadCommands: resource: %v, request: %v", reqs[i].DeviceResourceName, reqs[i]))

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
			res[i] = appModels.OnOffScheduleReadHandler(&group, OnOffScheduleDr, nil)
		case DimmingScheduleDr:
			res[i] = appModels.DimmingScheduleRead(&group, DimmingScheduleDr, nil)
		default:
			strErr := fmt.Sprintf("Khong ho tro doc Resource: %s", r.DeviceResourceName)
			gr.lc.Error(strErr)
			return nil, fmt.Errorf(strErr)
		}
	}
	return res, nil
}

func (gr *LightGroup) HandleWriteCommands(groupName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(groupName)
	if err != nil {
		return err
	}

	if params[0].DeviceResourceName == MethodDr && params[1].DeviceResourceName == DeviceDr {
		method, _ := params[0].StringValue()
		elementName, _ := params[1].StringValue()
		gr.lc.Debug(fmt.Sprintf("Manager Device: method:%s - Device:%s", method, elementName))
		err := gr.elementWriteHandler(&dev, method, elementName)
		if err != nil {
			gr.lc.Error(err.Error())
			return err
		}
		return nil
	}

	for i, p := range params {
		gr.lc.Info(fmt.Sprintf("LightGroupApplication.HandleWriteCommands: resource: %v, parameters: %v", reqs[i].DeviceResourceName, params[i]))
		switch p.DeviceResourceName {
		case OnOffScheduleDr:
			reqValue, _ := p.StringValue()
			err = appModels.OnOffScheduleWriteHandlerForGroup(gr.nw, &dev, &reqs[i], reqValue)
			if err != nil {
				gr.lc.Error(err.Error())
				return err
			}
		case DimmingScheduleDr:
			reqValue, _ := p.StringValue()
			err = appModels.DimmingScheduleWriteHandlerForGroup(gr.nw, &dev, &reqs[i], reqValue)
			if err != nil {
				gr.lc.Error(err.Error())
				return err
			}
		default:
			err = gr.NormalWriteCommand(groupName, &reqs[i], p)
			if err != nil {
				gr.lc.Error(err.Error())
				return err
			}
		}
	}
	return nil
}
