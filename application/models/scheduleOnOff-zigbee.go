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
	"github.com/phanvanhai/device-service-support/support/common"

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

// func convertNetToEdgeOnOffSchedule(nw nw.Network, net netOnOffSchedule, owner string) EdgeOnOffSchedule {
// 	shifPrefix := zigbeeConstants.PrefixHexValueNetGroupID
// 	result := EdgeOnOffSchedule{
// 		Time:  net.Time,
// 		Value: net.Value,
// 	}
// 	if net.OwnerAddress == 0x0000 {
// 		result.OwnerName = owner
// 	} else {
// 		grInt := uint32(shifPrefix<<16) | uint32(net.OwnerAddress)
// 		netID := fmt.Sprintf("%04X", grInt)
// 		result.OwnerName = nw.DeviceNameByNetID(netID)
// 	}
// 	return result
// }

type OnOffSchedule interface {
	WriteOnOffScheduleToDevice(deviceName string, cmReq *sdkModel.CommandRequest, schs []EdgeOnOffSchedule, onOffScheduleLimit int) error
}

func OnOffScheduleWriteHandler(scher OnOffSchedule, dev *models.Device, cmReq *sdkModel.CommandRequest, scheduleStr string, onOffScheduleLimit int) error {
	deviceName := dev.Name
	// chuyen doi noi dung string -> schedules
	schedules := StringNameToOnOffSchedule(scheduleStr)

	// loai bo nhung schedule loi (Owner = "")
	j := 0
	for _, s := range schedules {
		if s.OwnerName != "" {
			schedules[j] = s
			j++
		}
	}
	schedules = schedules[:j]

	newSchs := combineOnOffSchedule(dev, schedules, onOffScheduleLimit)
	if len(newSchs) > onOffScheduleLimit {
		return fmt.Errorf("loi vuot qua so luong lap lich cho phep")
	}
	err := scher.WriteOnOffScheduleToDevice(deviceName, cmReq, newSchs, onOffScheduleLimit)
	if err != nil {
		return err
	}

	// Neu thanh cong, cap nhap lai thong tin trong Support Database
	// truoc khi luu vao DB, can chuyen Name -> ID
	newStr := OnOffScheduleToStringID(newSchs)
	SetProperty(dev, common.ScheduleProtocolName, common.OnOffSchedulePropertyName, newStr)
	sv := sdk.RunningService()
	err = sv.UpdateDevice(*dev)
	if err != nil {
		return err
	}
	return nil
}

func combineOnOffSchedule(dev *models.Device, schs []EdgeOnOffSchedule, onOffScheduleLimit int) []EdgeOnOffSchedule {
	deviceName := dev.Name
	currentSchs := OnOffScheduleGetFromDB(dev)

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
		result.OwnerAddress = 0x0000
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

// // kich thuoc bieu dien phai dung = size
// func decodeNetOnOffSchedules(scheduleStr string, size int) ([]netOnOffSchedule, error) {
// 	decode, err := base64.StdEncoding.DecodeString(scheduleStr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	sch := make([]netOnOffSchedule, size)
// 	reader := bytes.NewBuffer(decode)
// 	err = binary.Read(reader, binary.BigEndian, sch)
// 	if err != nil {
// 		return nil, err
// 	}

// 	result := make([]netOnOffSchedule, 0, size)
// 	for i := 0; i < size; i++ {
// 		if CheckScheduleTime(sch[i].Time) == false {
// 			continue
// 		}
// 		result = append(result, sch[i])
// 	}
// 	return result, nil
// }

func OnOffScheduleEdgeToNetValue(nw nw.Network, schedules []EdgeOnOffSchedule, owner string, size int) string {
	netSchs := make([]netOnOffSchedule, 0, len(schedules))
	for _, sch := range schedules {
		netSch := edgeToNetOnOffSchedule(nw, sch, owner)
		netSchs = append(netSchs, netSch)
	}
	return encodeNetOnOffSchedules(netSchs, size)
}

// func NetValueToOnOffSchedule(nw nw.Network, value string, size int, owner string) ([]EdgeOnOffSchedule, error) {
// 	netSchs, err := decodeNetOnOffSchedules(value, size)
// 	if err != nil {
// 		return nil, err
// 	}
// 	edgeSchs := make([]EdgeOnOffSchedule, 0, len(netSchs))
// 	for _, sch := range netSchs {
// 		eg := convertNetToEdgeOnOffSchedule(nw, sch, owner)
// 		edgeSchs = append(edgeSchs, eg)
// 	}
// 	return edgeSchs, nil
// }

// String returns a JSON encoded string representation of the model
func OnOffScheduleToStringName(schedules []EdgeOnOffSchedule) string {
	out, err := json.Marshal(schedules)
	if err != nil {
		return ScheduleNilStr
	}
	return string(out)
}

func StringNameToOnOffSchedule(schedulesStr string) []EdgeOnOffSchedule {
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

func OnOffScheduleToStringID(schedules []EdgeOnOffSchedule) string {
	for i := range schedules {
		schedules[i].OwnerName = db.DB().NameToID(schedules[i].OwnerName)
	}
	out, err := json.Marshal(schedules)
	if err != nil {
		return ScheduleNilStr
	}
	return string(out)
}

func StringIDToOnOffSchedule(schedulesStr string) []EdgeOnOffSchedule {
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

func OnOffScheduleGetFromDB(dev *models.Device) []EdgeOnOffSchedule {
	sch, _ := GetProperty(dev, common.ScheduleProtocolName, common.OnOffSchedulePropertyName)
	return StringIDToOnOffSchedule(sch)
}
