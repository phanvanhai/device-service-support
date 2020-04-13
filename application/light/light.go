package light

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"

	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (l *Light) EventCallback(async sdkModel.AsyncValues) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(async.DeviceName)
	if err != nil {
		return err
	}

	db.DB().SetConnectedStatus(dev.Name, true)
	_, err = l.Connect(&dev)
	if err != nil {
		l.lc.Error(err.Error())
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
	str := fmt.Sprintf("Pushed event to core data: %+v", async)
	l.lc.Debug(str)
	l.asyncCh <- &async

	// update Realtime if have Realtime report
	if hasRealtime {
		l.UpdateRealtime(async.DeviceName)
	}

	return nil
}

func (l *Light) Initialize(dev *models.Device) error {
	isContinue, err := l.Provision(dev)
	if isContinue == false {
		return err
	}

	isContinue, err = l.Connect(dev)
	return err
}

func (l *Light) AddDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	l.lc.Debug(fmt.Sprintf("a new Device is added in MetaData:%s", deviceName))

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return err
	}

	return l.Initialize(&dev)
}

func (l *Light) UpdateDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	l.lc.Debug(fmt.Sprintf("a Device is updated in MetaData:%s", deviceName))

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		return err
	}

	return l.Initialize(&dev)
}

func (l *Light) RemoveDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties) error {
	l.lc.Debug(fmt.Sprintf("a Device is deleted in MetaData:%s", deviceName))

	err := l.nw.DeleteObject(deviceName, protocols)
	return err
}

func (l *Light) HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	provision := l.nw.CheckExist(deviceName)
	if provision == false {
		l.lc.Error("thiet bi chua duoc cap phep")
		return nil, fmt.Errorf("thiet bi chua duoc cap phep")
	}
	connected := db.DB().GetConnectedStatus(deviceName)
	if connected == false {
		l.lc.Error("thiet bi chua duoc ket noi")
		return nil, fmt.Errorf("thiet bi chua duoc ket noi")
	}

	res := make([]*sdkModel.CommandValue, len(reqs))
	for i, r := range reqs {
		l.lc.Info(fmt.Sprintf("LightApplication.HandleReadCommands: resource: %v, request: %v", reqs[i].DeviceResourceName, reqs[i]))
		req := make([]*sdkModel.CommandRequest, 1)

		switch r.DeviceResourceName {
		case ScenarioDr:
			relations := db.DB().ElementDotScenario(deviceName)
			relationStr, err := json.Marshal(relations)
			if err != nil {
				str := fmt.Sprintf("Loi phan tich noi dung. Loi:%s", err.Error())
				l.lc.Error(str)
				return nil, err
			}
			newCmvl := sdkModel.NewStringValue(ScenarioDr, 0, string(relationStr))
			res[i] = newCmvl
		case GroupDr:
			// Lay thong tin tu Support Database va tao ket qua
			groups := db.DB().ElementDotGroups(deviceName)
			grsStr, err := appModels.RelationGroupToString(groups)
			if err != nil {
				str := fmt.Sprintf("Loi phan tich noi dung. Loi:%s", err.Error())
				l.lc.Error(str)
				return nil, err
			}
			newCmvl := sdkModel.NewStringValue(GroupDr, 0, grsStr)
			res[i] = newCmvl
		case OnOffScheduleDr:
			// Lay thong tin tu Support Database va tao ket qua
			onoffs := l.GetOnOffSchedulesFromDB(deviceName)
			onoffsStr := appModels.OnOffScheduleToStringName(onoffs)
			newCmvl := sdkModel.NewStringValue(OnOffScheduleDr, 0, string(onoffsStr))
			res[i] = newCmvl
		case DimmingScheduleDr:
			// Lay thong tin tu Support Database va tao ket qua
			dims := l.GetDimmingSchedulesFromDB(deviceName)
			dimsStr := appModels.DimmingScheduleToStringName(dims)
			newCmvl := sdkModel.NewStringValue(DimmingScheduleDr, 0, string(dimsStr))
			res[i] = newCmvl
		default:
			// Gui lenh
			req[0] = &r
			cmvl, err := l.nw.ReadCommands(deviceName, req)
			if err != nil {
				l.lc.Error(err.Error())
				l.updateOpStateAndConnectdStatus(deviceName, false)
				return nil, err
			}
			res[i] = cmvl[0]
		}
	}
	return res, nil
}

func (l *Light) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	provision := l.nw.CheckExist(deviceName)
	if provision == false {
		l.lc.Error("thiet bi chua duoc cap phep")
		return fmt.Errorf("thiet bi chua duoc cap phep")
	}
	connected := db.DB().GetConnectedStatus(deviceName)
	if connected == false {
		l.lc.Error("thiet bi chua duoc ket noi")
		return fmt.Errorf("thiet bi chua duoc ket noi")
	}

	for i, p := range params {
		l.lc.Info(fmt.Sprintf("LightApplication.HandleWriteCommands: resource: %v, parameters: %v", reqs[i].DeviceResourceName, params[i]))

		switch p.DeviceResourceName {
		case OnOffScheduleDr:
			// chuyen doi noi dung r.Value
			reqValue, _ := p.StringValue()
			err := l.OnOffScheduleWriteHandler(deviceName, &reqs[i], reqValue)
			if err != nil {
				return err
			}
		case DimmingScheduleDr:
			reqValue, _ := p.StringValue()
			err := l.DimmingScheduleWriteHandler(deviceName, &reqs[i], reqValue)
			if err != nil {
				return err
			}
		case GroupDr:
			// chuyen doi noi dung r.Value
			reqValue, _ := p.StringValue()
			err := l.GroupWriteHandler(deviceName, &reqs[i], reqValue)
			if err != nil {
				return err
			}
		default:
			param := make([]*sdkModel.CommandValue, 1)
			param[0] = p

			req := make([]*sdkModel.CommandRequest, 1)
			req[0] = &reqs[i]

			// Gui lenh
			err := l.nw.WriteCommands(deviceName, req, param)
			if err != nil {
				l.lc.Error(err.Error())
				l.updateOpStateAndConnectdStatus(deviceName, false)
				return err
			}
		}
	}
	return nil
}
