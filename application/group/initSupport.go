package group

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"

	"github.com/phanvanhai/device-service-support/support/common"
	"github.com/phanvanhai/device-service-support/support/db"
)

func (gr *LightGroup) provision(dev *models.Device) (continueFlag bool, err error) {
	gr.lc.Debug("tien trinh cap phep")
	provision := gr.nw.CheckExist(dev.Name)
	opstate := dev.OperatingState
	gr.lc.Debug(fmt.Sprintf("provison=%t", provision))

	if (provision == false && opstate == models.Disabled) || (provision == true && opstate == models.Enabled) {
		gr.lc.Debug(fmt.Sprintf("thoat tien trinh cap phep vi: provision=%t & opstate=%s", provision, opstate))
		return true, nil
	}

	sv := sdk.RunningService()
	if provision == false { // opstate = true
		newdev, err := gr.nw.AddObject(dev)
		if err != nil {
			gr.lc.Error(err.Error())
			continueFlag, err = gr.updateOpState(dev.Name, false)
			return continueFlag, err
		}
		if newdev != nil {
			gr.lc.Debug("cap nhap lai thong tin device sau khi da cap phep")
			return false, sv.UpdateDevice(*newdev)
		}
		gr.lc.Debug("newdev after provision = nil")
	}

	return true, nil
}

func (gr *LightGroup) updateDB(group models.Device) error {
	relations := db.DB().GroupDotElement(group.Name)
	needUpdate := false
	for i, r := range relations {
		if db.DB().NameToID(r.Element) == "" {
			needUpdate = true
			relations[i] = relations[len(relations)-1]
			relations = relations[:len(relations)-1]
			str := fmt.Sprintf("Can loai bo thong tin Device:%s trong Database", r.Element)
			gr.lc.Debug(str)
		}
	}

	if needUpdate {
		str := fmt.Sprintf("Cap nhap lai Database cua Group:%s", group.Name)
		gr.lc.Debug(str)
		pp := make(models.ProtocolProperties)
		for _, r := range relations {
			id := db.DB().NameToID(r.Element)
			pp[id] = r.Content
		}
		group.Protocols[common.RelationProtocolNameConst] = pp
		sv := sdk.RunningService()
		return sv.UpdateDevice(group)
	}

	return nil
}

func (gr *LightGroup) updateOpState(devName string, status bool) (bool, error) {
	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(devName)
	if err != nil {
		return false, err
	}
	var notUpdate = true
	if status == false {
		if dev.OperatingState == models.Enabled {
			dev.OperatingState = models.Disabled
			gr.lc.Debug("cap nhap lai OpState = Disable")
			return false, sv.UpdateDevice(dev)
		}
		return false, nil
	}

	if dev.OperatingState == models.Disabled {
		dev.OperatingState = models.Enabled
		gr.lc.Debug("cap nhap lai OpState = Enabled")
		return false, sv.UpdateDevice(dev)
	}
	return notUpdate, nil
}
