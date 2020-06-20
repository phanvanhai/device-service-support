package group

import (
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	appModels "github.com/phanvanhai/device-service-support/application/models"
	"github.com/phanvanhai/device-service-support/common"
	"github.com/phanvanhai/device-service-support/support/db"
)

func (gr *LightGroup) Provision(dev *models.Device) (continueFlag bool, err error) {
	gr.lc.Debug("tien trinh cap phep")
	provision := gr.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState
	gr.lc.Debug(fmt.Sprintf("provison=%t", provision))

	if (provision == false && opstate == models.Disabled) || (provision == true) {
		gr.lc.Debug(fmt.Sprintf("thoat tien trinh cap phep vi: provision=%t & opstate=%s", provision, opstate))
		return true, nil
	}

	sv := sdk.RunningService()
	if provision == false { // opstate = true
		newdev, err := gr.nw.AddObject(dev)
		if err != nil {
			gr.lc.Error(err.Error())
			continueFlag, err = appModels.UpdateOpState(dev.Name, false)
			return continueFlag, err
		}
		if newdev != nil {
			gr.lc.Debug("cap nhap lai thong tin device sau khi da cap phep")

			// Khoi tao Schedule trong Database
			appModels.FillOnOffScheduleToDB(newdev, appModels.ScheduleNilStr)
			appModels.FillDimmingScheduleToDB(newdev, appModels.ScheduleNilStr)

			return false, sv.UpdateDevice(*newdev)
		}
	}
	return true, nil
}

// ConnectAndUpdate luon duoc thuc hien, khong quan tam den OpState
func (gr *LightGroup) ConnectAndUpdate(group *models.Device) error {
	gr.lc.Debug("Bat dau tien trinh kiem tra ket dong bo nhom")
	defer gr.lc.Debug("Ket thuc tien trinh kiem tra ket dong bo nhom")

	relations := db.DB().GroupDotElement(group.Name)

	needUpdate := false
	oldpp, ok := group.Protocols[common.RelationProtocolNameConst]
	if !ok && (len(relations) != 0) {
		needUpdate = true
	} else {
		if len(oldpp) != len(relations) {
			needUpdate = true
		}
	}

	if needUpdate {
		gr.lc.Debug(fmt.Sprintf("Cap nhap lai Database cua Group:%s", group.Name))
		pp := make(models.ProtocolProperties)
		for _, r := range relations {
			id := db.DB().NameToID(r.Element)
			pp[id] = r.Content
		}
		group.Protocols[common.RelationProtocolNameConst] = pp
		sv := sdk.RunningService()
		return sv.UpdateDevice(*group)
	}

	return nil
}
