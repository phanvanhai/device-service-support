package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/common"

	"github.com/phanvanhai/device-service-support/support/db"

	nw "github.com/phanvanhai/device-service-support/network"
)

type EdgeOnOffSchedule struct {
	OwnerName string `json:"owner, omitempty"`
	Time      uint32 `json:"time, omitempty"`
	Value     bool   `json:"value, omitempty"`
}

type netOnOffSchedule struct {
	OwnerAddress uint16
	Time         uint32
	Value        bool
}

// Database: nil = '[]'
// 			     = '[{"owner":"id", "time":1234, "value":true},{"owner":"id", "time":4321, "value":false}]'
// Cloud <-> DS: '[{"owner":"dev1", "time":1234, "value":true},{"owner":"dev1", "time":4321, "value":false}]'
// 			: nil = '[]'
// Coord --> DS: base64('uint16uint32booluint16uint32bool')

func FillOnOffScheduleToDB(dev *models.Device, value string) {
	SetProperty(dev, common.ScheduleProtocolName, common.OnOffSchedulePropertyName, value)
}

func UpdateOnOffScheduleToElement(nw nw.Network, group *models.Device, resourceName string, element string) error {
	schs := onOffScheduleGetFromDB(group)
	// khi gui toi Element, neu schedule = nil -> tao 1 schedule bieu dien gia tri nil
	if len(schs) == 0 {
		scheduleNil := EdgeOnOffSchedule{
			OwnerName: group.Name,
			Time:      CreateScheuleTimeError()}
		schs = append(schs, scheduleNil)
	}

	schedulesStr := onOffScheduleToStringName(schs)
	return WriteCommandToOtherDeviceByResource(nw, group.Name, resourceName, schedulesStr, element)
}

func OnOffScheduleWriteHandlerForGroup(nw nw.Network, group *models.Device, cmReq *sdkModel.CommandRequest, scheduleStr string) error {
	groupName := group.Name

	schs := stringNameToOnOffSchedule(scheduleStr)
	// fill OwnerName
	for i := range schs {
		schs[i].OwnerName = groupName
	}

	// cap nhap vao DB cua Group
	// truoc khi luu vao DB, can chuyen Name -> ID
	strID := onOffScheduleToStringID(schs)
	SetProperty(group, common.ScheduleProtocolName, common.OnOffSchedulePropertyName, strID)
	sv := sdk.RunningService()
	err := sv.UpdateDevice(*group)
	if err != nil {
		return err
	}

	// Gui lenh Unicast toi cac device
	// khi gui toi Element, neu schedule = nil -> tao 1 schedule bieu dien gia tri nil
	schs = stringIDToOnOffSchedule(strID)
	if len(schs) == 0 {
		scheduleNil := EdgeOnOffSchedule{
			OwnerName: groupName,
			Time:      CreateScheuleTimeError()}
		schs = append(schs, scheduleNil)
	}
	strName := onOffScheduleToStringName(schs)

	errInfos := GroupWriteUnicastCommandToAll(nw, groupName, cmReq.DeviceResourceName, strName)
	for _, e := range errInfos {
		if e.Error != "" {
			errStr, _ := json.Marshal(errInfos)
			return fmt.Errorf("Loi gui lenh toi cac device. Loi:%s", string(errStr))
		}
	}
	return nil
}

func UpdateOnOffSchedulesToDevice(cm NormalWriteCommand, nw nw.Network, dev *models.Device, resourceName string, limit int) error {
	deviceName := dev.Name

	request, ok := NewCommandRequest(deviceName, resourceName)
	if !ok {
		return fmt.Errorf("khong tim thay resource")
	}

	schs := onOffScheduleGetFromDB(dev)
	reqConverted := onOffScheduleEdgeToNetValue(nw, schs, deviceName, limit)
	// tao CommandValue moi voi r.Value da duoc chuyen doi
	cmvlConverted := sdkModel.NewStringValue(resourceName, 0, reqConverted)

	return cm.NormalWriteCommand(dev, request, cmvlConverted)
}

func OnOffScheduleWriteHandlerForDevice(cm NormalWriteCommand, nw nw.Network, dev *models.Device, cmReq *sdkModel.CommandRequest, scheduleStr string, limit int, groups []string) error {
	deviceName := dev.Name
	// chuyen doi noi dung string -> schedules
	schedules := stringNameToOnOffSchedule(scheduleStr)

	// loai bo nhung schedule loi (Owner = "")
	j := 0
	for _, s := range schedules {
		if s.OwnerName != "" {
			schedules[j] = s
			j++
		}
	}
	schedules = schedules[:j]

	newSchs := combineOnOffSchedule(dev, schedules, limit, groups)
	if len(newSchs) > limit {
		return fmt.Errorf("loi vuot qua so luong lap lich cho phep")
	}

	reqConverted := onOffScheduleEdgeToNetValue(nw, newSchs, deviceName, limit)
	// tao CommandValue moi voi r.Value da duoc chuyen doi
	cmvlConverted := sdkModel.NewStringValue(cmReq.DeviceResourceName, 0, reqConverted)

	err := cm.NormalWriteCommand(dev, cmReq, cmvlConverted)
	if err != nil {
		return err
	}

	// Neu thanh cong, cap nhap lai thong tin trong Support Database
	// truoc khi luu vao DB, can chuyen Name -> ID
	newStr := onOffScheduleToStringID(newSchs)
	SetProperty(dev, common.ScheduleProtocolName, common.OnOffSchedulePropertyName, newStr)

	return nil
}

func OnOffScheduleReadHandler(dev *models.Device, resourceName string, groups []string) (cmvl *sdkModel.CommandValue) {
	// Lay thong tin tu Support Database va tao ket qua
	schs := onOffScheduleGetFromDB(dev)

	// combine with group:
	j := 0
	for _, s := range schs {
		if s.OwnerName == dev.Name {
			schs[j] = s
			j++
		} else {
			for _, g := range groups {
				if s.OwnerName == g {
					schs[j] = s
					j++
					break
				}
			}
		}
	}
	schs = schs[:j]

	schsStr := onOffScheduleToStringName(schs)
	cmvl = sdkModel.NewStringValue(resourceName, 0, schsStr)
	return
}

func combineOnOffSchedule(dev *models.Device, schs []EdgeOnOffSchedule, onOffScheduleLimit int, groups []string) []EdgeOnOffSchedule {
	deviceName := dev.Name
	currentSchs := onOffScheduleGetFromDB(dev)

	// combine with group:
	j := 0
	for _, s := range currentSchs {
		if s.OwnerName == dev.Name {
			currentSchs[j] = s
			j++
		} else {
			for _, g := range groups {
				if s.OwnerName == g {
					currentSchs[j] = s
					j++
					break
				}
			}
		}
	}
	currentSchs = currentSchs[:j]

	var owner string
	isDelete := false
	if len(schs) == 0 {
		owner = deviceName
		isDelete = true
	} else {
		owner = schs[0].OwnerName
		if CheckScheduleTime(schs[0].Time) == false {
			isDelete = true
		}
	}

	result := make([]EdgeOnOffSchedule, 0, onOffScheduleLimit)
	// loai bo schedules cu (co OwnerName = owner) trong danh sach hien tai:
	for _, s := range currentSchs {
		if s.OwnerName != owner {
			result = append(result, s)
		}
	}

	// neu truong hop la them schedule:
	if !isDelete {
		result = append(result, schs...)
	}
	return result
}

func edgeToNetOnOffSchedule(nw nw.Network, edge EdgeOnOffSchedule, owner string) netOnOffSchedule {
	result := netOnOffSchedule{
		Time:  edge.Time,
		Value: edge.Value,
	}

	if edge.OwnerName == owner {
		result.OwnerAddress = OwnerMe
	} else {
		netID := nw.NetIDByDeviceName(edge.OwnerName)
		grInt64, _ := strconv.ParseUint(netID, 16, 32)
		result.OwnerAddress = uint16(grInt64 & 0xFFFF)
	}

	return result
}

func encodeNetOnOffSchedules(schedules []netOnOffSchedule, size int) string {
	if len(schedules) < size {
		schs := make([]netOnOffSchedule, size-len(schedules))
		for i := range schs {
			schs[i].Time = TimeError
		}
		schedules = append(schedules, schs...)
	}
	if len(schedules) > size {
		schedules = schedules[:size]
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, schedules)
	schedulesByte := buf.Bytes()
	return base64.StdEncoding.EncodeToString(schedulesByte)
}

func onOffScheduleEdgeToNetValue(nw nw.Network, schedules []EdgeOnOffSchedule, owner string, size int) string {
	netSchs := make([]netOnOffSchedule, 0, len(schedules))
	for _, sch := range schedules {
		netSch := edgeToNetOnOffSchedule(nw, sch, owner)
		netSchs = append(netSchs, netSch)
	}
	return encodeNetOnOffSchedules(netSchs, size)
}

// String returns a JSON encoded string representation of the model
func onOffScheduleToStringName(schedules []EdgeOnOffSchedule) string {
	out, err := json.Marshal(schedules)
	if err != nil {
		return ScheduleNilStr
	}
	return string(out)
}

func stringNameToOnOffSchedule(schedulesStr string) []EdgeOnOffSchedule {
	var schedules []EdgeOnOffSchedule
	if schedulesStr == "" {
		schedulesStr = ScheduleNilStr
	}
	err := json.Unmarshal([]byte(schedulesStr), &schedules)
	if err != nil {
		return schedules
	}

	return schedules
}

func onOffScheduleToStringID(schedules []EdgeOnOffSchedule) string {
	for i := range schedules {
		schedules[i].OwnerName = db.DB().NameToID(schedules[i].OwnerName)
	}
	out, err := json.Marshal(schedules)
	if err != nil {
		return ScheduleNilStr
	}
	return string(out)
}

func stringIDToOnOffSchedule(schedulesStr string) []EdgeOnOffSchedule {
	var schedules []EdgeOnOffSchedule
	if schedulesStr == "" {
		schedulesStr = ScheduleNilStr
	}
	err := json.Unmarshal([]byte(schedulesStr), &schedules)
	if err != nil {
		return schedules
	}
	for i := range schedules {
		schedules[i].OwnerName = db.DB().IDToName(schedules[i].OwnerName)
	}
	return schedules
}

func onOffScheduleGetFromDB(dev *models.Device) []EdgeOnOffSchedule {
	sch, _ := GetProperty(dev, common.ScheduleProtocolName, common.OnOffSchedulePropertyName)
	return stringIDToOnOffSchedule(sch)
}
