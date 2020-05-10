package light

import (
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
	"github.com/phanvanhai/device-service-support/support/common"
	db "github.com/phanvanhai/device-service-support/support/db"
)

// update Groups latest
func (l *Light) UpdateGroupToDevice(dev *models.Device) error {
	deviceName := dev.Name

	request, ok := appModels.NewCommandRequest(deviceName, GroupDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return fmt.Errorf("khong tim thay resource")
	}

	relations := db.DB().ElementDotGroups(deviceName)
	groups := make([]string, len(relations))
	for i, r := range relations {
		groups[i] = r.Parent
	}

	_, err := l.WriteGroupToDevice(dev, request, groups, GroupLimit)

	return err
}

func (l *Light) WriteGroupToDevice(dev *models.Device, cmReq *sdkModel.CommandRequest, groups []string, grouplimit int) (bool, error) {
	deviceName := dev.Name
	needUpdate := false

	if len(groups) > grouplimit {
		return needUpdate, fmt.Errorf("loi vuot qua so luong nhom cho phep")
	}

	reqConverted := appModels.GroupToNetValue(l.nw, groups, GroupLimit)

	// tao CommandValue moi voi r.Value da duoc chuyen doi
	cmvlConverted := sdkModel.NewStringValue(GroupDr, 0, reqConverted)
	param := make([]*sdkModel.CommandValue, 1)
	param[0] = cmvlConverted

	req := make([]*sdkModel.CommandRequest, 1)
	req[0] = cmReq

	// Gui lenh
	err := l.nw.WriteCommands(deviceName, req, param)
	if err != nil {
		l.lc.Error(err.Error())
		appModels.UpdateOpState(deviceName, false)
		return needUpdate, err
	}
	str := fmt.Sprintf("Cap nhap thanh cong danh sach Group cua Device:%s", deviceName)
	l.lc.Debug(str)

	l.lc.Debug("sync on-off schedules of Device in DB")
	newOnOff, change1 := l.SyncOnOffScheduleDBByGroups(deviceName, groups)

	l.lc.Debug("sync dimming schedules of Device in DB")
	newDimming, change2 := l.SyncDimmingScheduleDBByGroups(deviceName, groups)

	if !change1 && !change2 {
		return needUpdate, nil
	}

	appModels.SetProperty(dev, common.ScheduleProtocolName, common.OnOffSchedulePropertyName, newOnOff)
	appModels.SetProperty(dev, common.ScheduleProtocolName, common.DimmingSchedulePropertyName, newDimming)
	needUpdate = true

	return needUpdate, nil
}
