package group

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/models"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"

	appModels "github.com/phanvanhai/device-service-support/application/models"
	"github.com/phanvanhai/device-service-support/common"
	"github.com/phanvanhai/device-service-support/support/db"
)

func (gr *LightGroup) addElement(group *models.Device, elementName string) error {
	groupName := group.Name
	elementID := db.DB().NameToID(elementName)
	if elementID == "" {
		str := fmt.Sprintf("Khong ton tai Device:%s", elementName)
		gr.lc.Error(str)
		return fmt.Errorf(str)
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

	gr.lc.Debug(fmt.Sprintf("Bat dau gui yeu cau toi Device:%s", elementName))
	err = appModels.WriteCommandToOtherDeviceByResource(gr.nw, groupName, DeviceDr, string(grsStr), elementName)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	gr.lc.Debug(fmt.Sprintf("Gui yeu cau thanh cong toi Device:%s", elementName))
	gr.lc.Debug(fmt.Sprintf("Da them Device:%s vao Group:%s", elementName, groupName))

	// cap nhap vao DB cua Group
	// luon thay Name -> ID khi luu vao Database
	appModels.SetProperty(group, common.RelationProtocolNameConst, elementID, "")
	sv := sdk.RunningService()
	err = sv.UpdateDevice(*group)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	gr.lc.Debug(fmt.Sprintf("Them thanh cong thong tin Device:%s vao Database cua Group:%s", elementName, groupName))

	gr.lc.Debug(fmt.Sprintf("Bat dau cap nhap thong tin cau hinh cua Group:%s toi Device:%s", groupName, elementName))

	// Cap nhap OnOff Schedules
	err = appModels.UpdateOnOffScheduleToElement(gr.nw, group, OnOffScheduleDr, elementName)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	// Cap nhap Dimming Schedules
	err = appModels.UpdateDimmingScheduleToElement(gr.nw, group, DimmingScheduleDr, elementName)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	gr.lc.Debug(fmt.Sprintf("Cap nhap thanh cong thong tin cau hinh cua Group:%s toi Device:%s", groupName, elementName))
	return nil
}

func (gr *LightGroup) deleteElement(group *models.Device, elementName string, updateDB bool) error {
	groupName := group.Name
	elementID := db.DB().NameToID(elementName)
	if elementID == "" {
		gr.lc.Debug(fmt.Sprintf("Khong ton tai Device:%s", elementName))
		return nil
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
		gr.lc.Debug(fmt.Sprintf("Device:%s khong thuoc Group:%s", elementName, groupName))
		return nil
	}

	// gui lenh
	grsStr, err := json.Marshal(grs)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}

	gr.lc.Debug(fmt.Sprintf("Bat dau gui yeu cau toi Device:%s", elementName))
	err = appModels.WriteCommandToOtherDeviceByResource(gr.nw, groupName, DeviceDr, string(grsStr), elementName)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	gr.lc.Debug(fmt.Sprintf("Gui yeu cau thanh cong toi Device:%s", elementName))
	gr.lc.Debug(fmt.Sprintf("Da xoa Device:%s khoi Group:%s", elementName, groupName))

	if !updateDB {
		return nil
	}

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
	sv := sdk.RunningService()
	err = sv.UpdateDevice(*group)
	if err != nil {
		gr.lc.Error(err.Error())
		return err
	}
	gr.lc.Debug(fmt.Sprintf("Xoa thanh cong thong tin Device:%s vao Database cua Group:%s", elementName, groupName))

	return nil
}

func (gr *LightGroup) elementWriteHandler(group *models.Device, method string, elementName string) error {
	if db.DB().NameToID(elementName) == "" {
		strErr := fmt.Sprintf("khong ton tai Device: %s", elementName)
		gr.lc.Error(strErr)
		return fmt.Errorf(strErr)
	}
	switch strings.ToLower(method) {
	case PutMethod:
		err := gr.addElement(group, elementName)
		if err != nil {
			gr.lc.Error(fmt.Sprintf("Them Device:%s gap loi:%s", elementName, err.Error()))
			return err
		}
		gr.lc.Debug(fmt.Sprintf("Them thanh cong Device:%s vao Group:%s", elementName, group.Name))
		return nil
	case DeleteMethod:
		err := gr.deleteElement(group, elementName, true)
		if err != nil {
			gr.lc.Error(fmt.Sprintf("Xoa Device:%s gap loi:%s", elementName, err.Error()))
			return err
		}
		gr.lc.Debug(fmt.Sprintf("Xoa thanh cong Device:%s trong Group:%s", elementName, group.Name))
		return nil
	default:
		strErr := fmt.Sprintf("khong ho tro phuong thuc: %s", method)
		gr.lc.Error(strErr)
		return fmt.Errorf(strErr)
	}
}

func (gr *LightGroup) deleteAllElement(groupName string) {
	relations := db.DB().GroupDotElement(groupName)
	group := models.Device{
		Name: groupName,
	}

	for _, relation := range relations {
		gr.deleteElement(&group, relation.Element, false)
	}
}
