package light

import (
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
)

// get OnOff-Schedules latest
func (l *Light) UpdateOnOffSchedulesConfigToDevice(dev *models.Device) error {
	deviceName := dev.Name

	request, ok := appModels.NewCommandRequest(deviceName, OnOffScheduleDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return fmt.Errorf("khong tim thay resource")
	}

	schs := appModels.OnOffScheduleGetFromDB(dev)
	return l.WriteOnOffScheduleToDevice(deviceName, request, schs, OnOffScheduleLimit)
}

func (l *Light) WriteOnOffScheduleToDevice(deviceName string, cmReq *sdkModel.CommandRequest, schs []appModels.EdgeOnOffSchedule, onOffScheduleLimit int) error {
	reqConverted := appModels.OnOffScheduleEdgeToNetValue(l.nw, schs, deviceName, onOffScheduleLimit)
	// tao CommandValue moi voi r.Value da duoc chuyen doi
	cmvlConverted := sdkModel.NewStringValue(OnOffScheduleDr, 0, reqConverted)
	param := make([]*sdkModel.CommandValue, 1)
	param[0] = cmvlConverted

	req := make([]*sdkModel.CommandRequest, 1)
	req[0] = cmReq

	// Gui lenh
	err := l.nw.WriteCommands(deviceName, req, param)
	if err != nil {
		l.lc.Error(err.Error())
		appModels.UpdateOpState(deviceName, false)
		return err
	}

	return nil
}

func (l *Light) SyncOnOffScheduleDBByGroups(deviceName string, groups []string) (string, bool) {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		l.lc.Error(err.Error())
		return appModels.ScheduleNilStr, false
	}

	var change = false
	schedules := appModels.OnOffScheduleGetFromDB(&dev)
	if len(schedules) <= 0 {
		return appModels.ScheduleNilStr, change
	}

	// loai bo nhung schudule khong lien quan den group hay chinh device
	j := 0
	for _, s := range schedules {
		var isExist = false
		if s.OwnerName == deviceName {
			isExist = true
		} else {
			for _, g := range groups {
				if s.OwnerName == g {
					isExist = true
					break
				}
			}
		}

		if isExist == true {
			schedules[j] = s
			j++
		} else {
			change = true
		}
	}
	schedules = schedules[:j]

	newStr := appModels.OnOffScheduleToStringID(schedules)
	return newStr, change
}
