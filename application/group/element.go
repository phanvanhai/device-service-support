package group

import (
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/phanvanhai/device-service-support/support/common"
	"github.com/phanvanhai/device-service-support/support/db"
)

func (gr *LightGroup) addElement(groupName string, elementName string) error {
	elementID := db.DB().NameToID(elementName)
	if elementID == "" {
		str := fmt.Sprintf("Khong ton tai Device:%s", elementName)
		gr.lc.Error(str)
		return fmt.Errorf(str)
	}

	sv := sdk.RunningService()
	group, err := sv.GetDeviceByName(groupName)
	if err != nil {
		return err
	}

	// tao danh sach nhom cho Element
	// du element da luu group -> van gui nhu binh thuong
	// vi co the element loi o phan update cau hinh group, user can lam lai buoc them group
	relations := db.DB().ElementDotGroups(elementName)
	grs := make([]string, 0, len(relations)+1)
	var grExist = false
	for _, r := range relations {
		if r.Parent == groupName {
			str := fmt.Sprintf("Group:%s da co san Device:%s", groupName, elementName)
			gr.lc.Debug(str)
			grExist = true
		}
		grs = append(grs, r.Parent)
	}
	if grExist == false {
		grs = append(grs, groupName)
	}

	// gui lenh
	grsStr, err := json.Marshal(grs)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	gr.lc.Debug("Bat dau gui yeu cau toi Device:%s", elementName)
	err = gr.WriteCommandByResource(groupName, DeviceDr, string(grsStr), elementName)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	gr.lc.Debug("Gui yeu cau thanh cong toi Device:%s", elementName)
	gr.lc.Debug("Da them Device:%s vao Group:%s", elementName, groupName)

	// cap nhap vao DB cua Group
	// luon thay Name -> ID khi luu vao Database
	pp, ok := group.Protocols[common.RelationProtocolNameConst]
	if !ok {
		pp = make(models.ProtocolProperties)
	}
	pp[elementID] = ""
	group.Protocols[common.RelationProtocolNameConst] = pp
	err = sv.UpdateDevice(group)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	gr.lc.Debug("Them thanh cong thong tin Device:%s vao Database cua Group:%s", elementName, groupName)

	gr.lc.Debug("Bat dau cap nhap thong tin cau hinh cua Group:%s toi Device:%s", groupName, elementName)
	// Cap nhap OnOff Schedules
	err = gr.UpdateOnOffScheduleToElement(groupName, elementName)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	// Cap nhap Dimming Schedules
	err = gr.UpdateDimmingScheduleToElement(groupName, elementName)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	gr.lc.Debug("Cap nhap thanh cong thong tin cau hinh cua Group:%s toi Device:%s", groupName, elementName)
	return nil
}

func (gr *LightGroup) deleteElement(groupName string, elementName string) error {
	elementID := db.DB().NameToID(elementName)
	if elementID == "" {
		str := fmt.Sprintf("Khong ton tai Device:%s", elementName)
		gr.lc.Debug(str)
		return nil
	}

	sv := sdk.RunningService()
	group, err := sv.GetDeviceByName(groupName)
	if err != nil {
		return err
	}

	// tao danh sach nhom cho Element
	relations := db.DB().ElementDotGroups(elementName)
	grs := make([]string, 0, len(relations))
	for _, r := range relations {
		if r.Parent != groupName {
			grs = append(grs, r.Parent)
		}
	}
	if len(grs) == len(relations) {
		str := fmt.Sprintf("Device:%s khong thuoc Group:%s", elementName, groupName)
		gr.lc.Debug(str)
		return nil
	}

	// gui lenh
	grsStr, err := json.Marshal(grs)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	gr.lc.Debug("Bat dau gui yeu cau toi Device:%s", elementName)
	err = gr.WriteCommandByResource(groupName, DeviceDr, string(grsStr), elementName)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	gr.lc.Debug("Gui yeu cau thanh cong toi Device:%s", elementName)
	gr.lc.Debug("Da them Device:%s vao Group:%s", elementName, groupName)

	// cap nhap vao DB cua Group
	pp, ok := group.Protocols[common.RelationProtocolNameConst]
	if !ok {
		return nil
	}
	_, ok = pp[elementID]
	if !ok {
		return nil
	}
	delete(pp, elementID)
	group.Protocols[common.RelationProtocolNameConst] = pp
	err = sv.UpdateDevice(group)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	gr.lc.Debug("Xoa thanh cong thong tin Device:%s vao Database cua Group:%s", elementName, groupName)

	return nil
}

func (gr *LightGroup) elementWriteHandler(groupName string, method string, elementName string) error {
	if db.DB().NameToID(elementName) == "" {
		strErr := fmt.Sprintf("khong ton tai Device: %s", elementName)
		gr.lc.Error(strErr)
		return fmt.Errorf(strErr)
	}
	switch strings.ToLower(method) {
	case PutMethod:
		err := gr.addElement(groupName, elementName)
		if err != nil {
			str := fmt.Sprintf("Them Device:%s gap loi:%s", elementName, err.Error())
			gr.lc.Error(str)
			return err
		}
		str := fmt.Sprintf("Them thanh cong Device:%s vao Group:%s", elementName, groupName)
		gr.lc.Debug(str)
		return nil
	case DeleteMethod:
		err := gr.deleteElement(groupName, elementName)
		if err != nil {
			str := fmt.Sprintf("Xoa Device:%s gap loi:%s", elementName, err.Error())
			gr.lc.Error(str)
			return err
		}
		str := fmt.Sprintf("Xoa thanh cong Device:%s trong Group:%s", elementName, groupName)
		gr.lc.Debug(str)
		return nil
	default:
		strErr := fmt.Sprintf("khong ho tro phuong thuc: %s", method)
		gr.lc.Error(strErr)
		return fmt.Errorf(strErr)
	}
}
