package light

import (
	"encoding/json"
	"fmt"
	"strconv"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

func (l *Light) EventCallback(async sdkModel.AsyncValues) error {
	// -----------------------------------
	return nil
	// ------------------------------------------
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
		l.lc.Debug(fmt.Sprintf("Pushed event to core data: %+v", async))
		l.asyncCh <- &async
	}

	db.DB().SetConnectedStatus(dev.Name, true)
	err = l.ConnectAndUpdate(&dev, versionInDev)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	// update Realtime if have Realtime report
	if hasRealtime {
		return appModels.UpdateRealtimeToDevice(l, &dev, RealtimeDr)
	}

	return nil
}

func (l *Light) Initialize(dev *models.Device) error {
	isContinue, err := l.Provision(dev)
	if isContinue == false {
		return err
	}

	return l.ConnectAndUpdate(dev, nil)
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
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		l.lc.Error(err.Error())
		return nil, err
	}

	res := make([]*sdkModel.CommandValue, len(reqs))
	for i, r := range reqs {
		l.lc.Info(fmt.Sprintf("LightApplication.HandleReadCommands: resource: %v, request: %v", reqs[i].DeviceResourceName, reqs[i]))

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
			groups := appModels.GetGroupList(dev.Name)
			grsStr, err := json.Marshal(groups)
			if err != nil {
				str := fmt.Sprintf("Loi phan tich noi dung. Loi:%s", err.Error())
				l.lc.Error(str)
				return nil, err
			}
			newCmvl := sdkModel.NewStringValue(GroupDr, 0, string(grsStr))
			res[i] = newCmvl
		case OnOffScheduleDr:
			// Lay thong tin tu Support Database va tao ket qua
			groups := appModels.GetGroupList(dev.Name)
			res[i] = appModels.OnOffScheduleReadHandler(&dev, OnOffScheduleDr, groups)
		case DimmingScheduleDr:
			// Lay thong tin tu Support Database va tao ket qua
			groups := appModels.GetGroupList(dev.Name)
			res[i] = appModels.DimmingScheduleRead(&dev, DimmingScheduleDr, groups)
		default:
			cmvl, err := l.NormalReadCommand(&dev, &r)
			if err != nil {
				l.lc.Error(err.Error())
				appModels.UpdateOpState(deviceName, false)
				return nil, err
			}
			res[i] = cmvl[0]
		}
	}
	return res, nil
}

func (l *Light) HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	for i, p := range params {
		l.lc.Info(fmt.Sprintf("LightApplication.HandleWriteCommands: resource: %v, parameters: %v", reqs[i].DeviceResourceName, params[i]))

		switch p.DeviceResourceName {
		case OnOffScheduleDr:
			groups := appModels.GetGroupList(dev.Name)
			// chuyen doi noi dung r.Value
			reqValue, _ := p.StringValue()
			err := appModels.OnOffScheduleWriteHandlerForDevice(l, l.nw, &dev, &reqs[i], reqValue, OnOffScheduleLimit, groups)
			if err != nil {
				return err
			}

			// vi can luu vao DB -> thay doi DB -> update Version
			currentVersion := appModels.GetVersionFromDB(dev)
			newVersion := appModels.GenerateNewVersion(currentVersion, currentVersion)
			appModels.FillVerisonToDB(&dev, strconv.FormatUint(newVersion, 10))
			err = appModels.UpdateVersionConfigToDevice(l, &dev, PingDr, newVersion)
			if err != nil {
				l.lc.Error(err.Error())
				return err
			}

			// update lap lich, version moi vao DB
			err = sv.UpdateDevice(dev)
			if err != nil {
				l.lc.Error(err.Error())
				return err
			}
		case DimmingScheduleDr:
			groups := appModels.GetGroupList(dev.Name)
			// chuyen doi noi dung r.Value
			reqValue, _ := p.StringValue()
			err := appModels.DimmingScheduleWriteHandlerForDevice(l, l.nw, &dev, &reqs[i], reqValue, DimmingScheduleLimit, groups)
			if err != nil {
				l.lc.Error(err.Error())
				return err
			}

			// vi can luu vao DB -> thay doi DB -> update Version
			currentVersion := appModels.GetVersionFromDB(dev)
			newVersion := appModels.GenerateNewVersion(currentVersion, currentVersion)
			appModels.FillVerisonToDB(&dev, strconv.FormatUint(newVersion, 10))
			err = appModels.UpdateVersionConfigToDevice(l, &dev, PingDr, newVersion)
			if err != nil {
				l.lc.Error(err.Error())
				return err
			}

			// update lap lich, version moi vao DB
			err = sv.UpdateDevice(dev)
			if err != nil {
				l.lc.Error(err.Error())
				return err
			}
		case GroupDr:
			// chuyen doi noi dung r.Value
			reqValue, _ := p.StringValue()
			err := appModels.GroupListWriteHandler(l, l.nw, &dev, &reqs[i], reqValue, GroupLimit)
			if err != nil {
				l.lc.Error(err.Error())
				return err
			}
			// vi khong can luu vao DB -> khong thay doi DB -> khong update Version
		default:
			err = l.NormalWriteCommand(&dev, &reqs[i], p)
			if err != nil {
				l.lc.Error(err.Error())
				return err
			}
		}
	}
	return nil
}
