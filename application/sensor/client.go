package sensor

import (
	"encoding/json"
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (s *Sensor) EventCallback(async sdkModel.AsyncValues) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(async.DeviceName)
	if err != nil {
		return err
	}

	var hasRealtime = false
	var versionInDev *uint64
	j := 0
	for _, a := range async.CommandValues {
		// loai bo report Realtime
		if a.DeviceResourceName == RealtimeDr {
			hasRealtime = true
			continue
		}

		// loai bo Ping & doc gia tri Version = Ping neu co
		if a.DeviceResourceName == PingDr {
			ver, err := a.Uint64Value()
			if err == nil {
				versionInDev = &ver
				continue
			}
		}

		async.CommandValues[j] = a
		j++
	}
	async.CommandValues = async.CommandValues[:j]

	// send event
	if len(async.CommandValues) > 0 {
		s.lc.Debug(fmt.Sprintf("Pushed event to core data: %+v", async))
		s.asyncCh <- &async
	}

	db.DB().SetConnectedStatus(dev.Name, true)
	err = s.ConnectAndUpdate(&dev, versionInDev)
	if err != nil {
		s.lc.Error(err.Error())
		return err
	}

	// update Realtime if have Realtime report
	if hasRealtime {
		return appModels.UpdateRealtimeToDevice(s, &dev, RealtimeDr)
	}

	return nil
}

func (s *Sensor) Initialize(dev *models.Device) error {
	isContinue, err := s.Provision(dev)
	if isContinue == false {
		return err
	}

	return s.ConnectAndUpdate(dev, nil)
}

func (s *Sensor) AddDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debug(fmt.Sprintf("a new Device is added in MetaData:%s", deviceName))

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return err
	}

	return s.Initialize(&dev)
}

func (s *Sensor) UpdateDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debug(fmt.Sprintf("a Device is updated in MetaData:%s", deviceName))

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return err
	}

	return s.Initialize(&dev)
}

func (s *Sensor) RemoveDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties) error {
	s.lc.Debug(fmt.Sprintf("a Device is deleted in MetaData:%s", deviceName))

	err := s.nw.DeleteObject(deviceName, protocols)
	return err
}

func (s *Sensor) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	res := make([]*sdkModel.CommandValue, len(reqs))
	for i, r := range reqs {
		if r.DeviceResourceName == ScenarioDr {
			relations := db.DB().ElementDotScenario(deviceName)
			relationStr, err := json.Marshal(relations)
			if err != nil {
				str := fmt.Sprintf("Loi phan tich noi dung. Loi:%s", err.Error())
				s.lc.Error(str)
				return nil, err
			}
			newCmvl := sdkModel.NewStringValue(ScenarioDr, 0, string(relationStr))
			res[i] = newCmvl
			continue
		}

		s.lc.Info(fmt.Sprintf("SensorApplication.HandleReadCommands: resource: %v, request: %v", reqs[i].DeviceResourceName, reqs[i]))
		req := make([]*sdkModel.CommandRequest, 1)

		// Gui lenh
		req[0] = &r
		cmvl, err := s.nw.ReadCommands(deviceName, req)
		if err != nil {
			s.lc.Error(err.Error())
			appModels.UpdateOpState(deviceName, false)
			return nil, err
		}
		res[i] = cmvl[0]
	}
	return res, nil
}

func (s *Sensor) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		s.lc.Error(err.Error())
		return err
	}

	for i, p := range params {
		s.lc.Info(fmt.Sprintf("SensorApplication.HandleWriteCommands: resource: %v, parameters: %v", reqs[i].DeviceResourceName, params[i]))

		param := make([]*sdkModel.CommandValue, 1)
		param[0] = p

		req := make([]*sdkModel.CommandRequest, 1)
		req[0] = &reqs[i]

		// Gui lenh
		err := s.nw.WriteCommands(deviceName, req, param)
		if err != nil {
			s.lc.Error(err.Error())
			appModels.UpdateOpState(deviceName, false)
			return err
		}

		err = s.NormalWriteCommand(&dev, &reqs[i], p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Sensor) NormalWriteCommand(dev *models.Device, cmReq *sdkModel.CommandRequest, cmvl *sdkModel.CommandValue) error {
	param := make([]*sdkModel.CommandValue, 1)
	param[0] = cmvl

	req := make([]*sdkModel.CommandRequest, 1)
	req[0] = cmReq

	// Gui lenh
	err := s.nw.WriteCommands(dev.Name, req, param)
	if err != nil {
		s.lc.Error(err.Error())
		appModels.UpdateOpState(dev.Name, false)
		return err
	}
	return nil
}
