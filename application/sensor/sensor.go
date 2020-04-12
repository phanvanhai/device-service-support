package sensor

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (s *Sensor) EventCallback(async sdkModel.AsyncValues) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(async.DeviceName)
	if err != nil {
		return err
	}
	_, err = s.Connect(&dev)
	if err != nil {
		s.lc.Error(err.Error())
	}

	var hasRealtime = false
	// loai bo report Realtime
	j := 0
	for _, a := range async.CommandValues {
		if a.DeviceResourceName != RealtimeDr {
			async.CommandValues[j] = a
			j++
		} else {
			hasRealtime = true
		}
	}
	async.CommandValues = async.CommandValues[:j]

	// send event
	s.asyncCh <- &async

	// update Realtime if have Realtime report
	if hasRealtime {
		s.UpdateRealtime(async.DeviceName)
	}

	return nil
}

func (s *Sensor) Initialize(dev *models.Device) error {
	isContinue, err := s.Provision(dev)
	if isContinue == false {
		return err
	}

	isContinue, err = s.Connect(dev)
	return err
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
	provision := s.nw.CheckExist(deviceName)
	if provision == false {
		s.lc.Error("thiet bi chua duoc cap phep")
		return nil, fmt.Errorf("thiet bi chua duoc cap phep")
	}
	connected := db.DB().GetConnectedStatus(deviceName)
	if connected == false {
		s.lc.Error("thiet bi chua duoc ket noi")
		return nil, fmt.Errorf("thiet bi chua duoc ket noi")
	}

	res := make([]*sdkModel.CommandValue, len(reqs))
	for i, r := range reqs {
		s.lc.Info(fmt.Sprintf("SensorApplication.HandleReadCommands: resource: %v, request: %v", reqs[i].DeviceResourceName, reqs[i]))
		req := make([]*sdkModel.CommandRequest, 1)

		// Gui lenh
		req[0] = &r
		cmvl, err := s.nw.ReadCommands(deviceName, req)
		if err != nil {
			s.lc.Error(err.Error())
			s.updateOpStateAndConnectdStatus(deviceName, false)
			return nil, err
		}
		res[i] = cmvl[0]
	}
	return res, nil
}

func (s *Sensor) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	provision := s.nw.CheckExist(deviceName)
	if provision == false {
		s.lc.Error("thiet bi chua duoc cap phep")
		return fmt.Errorf("thiet bi chua duoc cap phep")
	}
	connected := db.DB().GetConnectedStatus(deviceName)
	if connected == false {
		s.lc.Error("thiet bi chua duoc ket noi")
		return fmt.Errorf("thiet bi chua duoc ket noi")
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
			s.updateOpStateAndConnectdStatus(deviceName, false)
			return err
		}
	}
	return nil
}
