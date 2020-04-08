package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/phanvanhai/device-service-support/application/light/cm"
	nw "github.com/phanvanhai/device-service-support/network"
	zigbeeConstants "github.com/phanvanhai/device-service-support/network/zigbee/cm"
)

type EdgeDimmingSchedule struct {
	OwnerName string `json:"owner, omitempty"`
	Time      uint32 `json:"time, omitempty"`
	Value     uint16 `json:"value, omitempty"`
}

type netDimmingSchedule struct {
	OwnerAddress uint16
	Time         uint32
	Value        uint16
}

func convertNetToEdgeDimmingSchedule(nw nw.Network, net netDimmingSchedule, ownerName string) EdgeDimmingSchedule {
	shifPrefix := zigbeeConstants.PrefixHexValueNetGroupID
	result := EdgeDimmingSchedule{
		Time:  net.Time,
		Value: net.Value,
	}
	if net.OwnerAddress == 0x0000 {
		result.OwnerName = ownerName
	} else {
		grInt := uint32(shifPrefix<<16) | uint32(net.OwnerAddress)
		netID := fmt.Sprintf("%04X", grInt)
		result.OwnerName = nw.DeviceNameByNetID(netID)
	}
	return result
}

func convertEdgeToNetDimmingSchedule(nw nw.Network, edge EdgeDimmingSchedule, ownerName string) netDimmingSchedule {
	result := netDimmingSchedule{
		Time:  edge.Time,
		Value: edge.Value,
	}

	if edge.OwnerName == ownerName {
		result.OwnerAddress = 0x0000
	} else {
		netID := nw.NetIDByDeviceName(edge.OwnerName)
		grInt64, _ := strconv.ParseUint(netID, 16, 32)
		result.OwnerAddress = uint16(grInt64 & 0xFFFF)
	}

	return result
}

func encodeNetDimmingSchedules(schedules []netDimmingSchedule, size int) string {
	if len(schedules) < size {
		s := make([]netDimmingSchedule, size-len(schedules))
		schedules = append(schedules, s...)
	}
	if len(schedules) > size {
		schedules = schedules[:size]
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, schedules)
	schedulesByte := buf.Bytes()
	return base64.StdEncoding.EncodeToString(schedulesByte)
}

// kich thuoc bieu dien phai dung = size
func decodeNetDimmingSchedules(scheduleStr string, size int) ([]netDimmingSchedule, error) {
	decode, err := base64.StdEncoding.DecodeString(scheduleStr)
	if err != nil {
		return nil, err
	}

	sch := make([]netDimmingSchedule, size)
	reader := bytes.NewBuffer(decode)
	err = binary.Read(reader, binary.BigEndian, sch)
	if err != nil {
		return nil, err
	}

	result := make([]netDimmingSchedule, 0, size)
	for i := 0; i < size; i++ {
		if cm.CheckScheduleTime(sch[i].Time) == false {
			continue
		}
		result = append(result, sch[i])
	}
	return result, nil
}

func DimmingScheduleEdgeToNetValue(nw nw.Network, schedules []EdgeDimmingSchedule, ownerName string, size int) string {
	netSchs := make([]netDimmingSchedule, 0, len(schedules))
	for _, sch := range schedules {
		netSch := convertEdgeToNetDimmingSchedule(nw, sch, ownerName)
		netSchs = append(netSchs, netSch)
	}
	return encodeNetDimmingSchedules(netSchs, size)
}

func NetValueToDimmingSchedule(nw nw.Network, value string, size int, ownerName string) ([]EdgeDimmingSchedule, error) {
	netSchs, err := decodeNetDimmingSchedules(value, size)
	if err != nil {
		return nil, err
	}
	edgeSchs := make([]EdgeDimmingSchedule, 0, len(netSchs))
	for _, sch := range netSchs {
		eg := convertNetToEdgeDimmingSchedule(nw, sch, ownerName)
		edgeSchs = append(edgeSchs, eg)
	}
	return edgeSchs, nil
}

// String returns a JSON encoded string representation of the model
func DimmingScheduleToString(schedules []EdgeDimmingSchedule) string {
	out, err := json.Marshal(schedules)
	if err != nil {
		return err.Error()
	}
	return string(out)
}
