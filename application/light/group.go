package light

import (
	"encoding/json"
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
	db "github.com/phanvanhai/device-service-support/support/db"
)

// update Groups latest
func (l *Light) UpdateGroupToDevice(deviceName string) error {
	reqs := make([]*sdkModel.CommandRequest, 1)
	groups := db.DB().ElementDotGroups(deviceName)
	netGroups := appModels.RelationGroupToNetValue(l.nw, groups, GroupLimit)

	request, ok := appModels.NewCommandRequest(deviceName, GroupDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return fmt.Errorf("khong tim thay resource")
	}

	cmvlConverted := sdkModel.NewStringValue(GroupDr, 0, netGroups)
	param := make([]*sdkModel.CommandValue, 0, 1)
	param = append(param, cmvlConverted)

	reqs[0] = request
	err := l.nw.WriteCommands(deviceName, reqs, param)

	return err
}

func (l *Light) GroupWriteHandler(deviceName string, cmReq *sdkModel.CommandRequest, groupStr string) error {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(deviceName)
	if err != nil {
		l.lc.Error(err.Error())
		return nil
	}

	var groups []string
	err = json.Unmarshal([]byte(groupStr), &groups)
	if err != nil {
		return err
	}

	if len(groups) > GroupLimit {
		return fmt.Errorf("loi vuot qua so luong nhom cho phep")
	}
	reqConverted := appModels.GroupToNetValue(l.nw, groups, GroupLimit)

	// tao CommandValue moi voi r.Value da duoc chuyen doi
	cmvlConverted := sdkModel.NewStringValue(GroupDr, 0, reqConverted)
	param := make([]*sdkModel.CommandValue, 1)
	param[0] = cmvlConverted

	req := make([]*sdkModel.CommandRequest, 1)
	req[0] = cmReq

	// Gui lenh
	err = l.nw.WriteCommands(deviceName, req, param)
	if err != nil {
		l.lc.Error(err.Error())
		appModels.UpdateOpState(deviceName, false)
		return err
	}
	str := fmt.Sprintf("Cap nhap thanh cong danh sach Group cua Device:%s", deviceName)
	l.lc.Debug(str)

	l.lc.Debug("sync on-off schedules of Device in DB")
	newOnOff, change1 := l.SyncOnOffScheduleDBByGroups(deviceName, groups)

	l.lc.Debug("sync dimming schedules of Device in DB")
	newDimming, change2 := l.SyncDimmingScheduleDBByGroups(deviceName, groups)

	if !change1 && !change2 {
		return nil
	}

	pp, ok := dev.Protocols[ScheduleProtocolName]
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[OnOffSchedulePropertyName] = newOnOff
	pp[DimmingSchedulePropertyName] = newDimming
	err = sv.UpdateDevice(dev)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}

	return nil
}
