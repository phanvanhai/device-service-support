package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/phanvanhai/device-service-support/application/light/cm"
	nw "github.com/phanvanhai/device-service-support/network"
)

type EdgeScheduleOnOff struct {
	OwnerName string `json:"owner, omitempty"`
	Time      uint32 `json:"time, omitempty"`
	Value     bool   `json:"value, omitempty"`
}

type netScheduleOnOff struct {
	OwnerAddress uint16
	Time         uint32
	Value        bool
}

func scheduleOnOffNetToEdge(nw nw.Network, net netScheduleOnOff, ownerName string, shifPrefix uint16) EdgeScheduleOnOff {
	result := EdgeScheduleOnOff{
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

func scheduleOnOffEdgeToNet(nw nw.Network, edge EdgeScheduleOnOff, ownerName string) netScheduleOnOff {
	result := netScheduleOnOff{
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

func netScheduleOnOffToString(schedules []netScheduleOnOff, size int) string {
	if len(schedules) < size {
		s := make([]netScheduleOnOff, size-len(schedules))
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
func stringToNetScheduleOnOff(scheduleStr string, size int) ([]netScheduleOnOff, error) {
	decode, err := base64.StdEncoding.DecodeString(scheduleStr)
	if err != nil {
		return nil, err
	}

	sch := make([]netScheduleOnOff, size)
	reader := bytes.NewBuffer(decode)
	err = binary.Read(reader, binary.BigEndian, sch)
	if err != nil {
		return nil, err
	}

	result := make([]netScheduleOnOff, 0, size)
	for i := 0; i < size; i++ {
		if cm.CheckScheduleTime(sch[i].Time) == false {
			continue
		}
		result = append(result, sch[i])
	}
	return result, nil
}

func ScheduleOnOffEdgeToNetValue(nw nw.Network, schedules []EdgeScheduleOnOff, ownerName string, size int) string {
	netSchs := make([]netScheduleOnOff, 0, len(schedules))
	for _, sch := range schedules {
		netSch := scheduleOnOffEdgeToNet(nw, sch, ownerName)
		netSchs = append(netSchs, netSch)
	}
	return netScheduleOnOffToString(netSchs, size)
}

func NetValueToScheduleOnOff(nw nw.Network, value string, size int, ownerName string, shifPrefix uint16) ([]EdgeScheduleOnOff, error) {
	netSchs, err := stringToNetScheduleOnOff(value, size)
	if err != nil {
		return nil, err
	}
	edgeSchs := make([]EdgeScheduleOnOff, 0, len(netSchs))
	for _, sch := range netSchs {
		eg := scheduleOnOffNetToEdge(nw, sch, ownerName, shifPrefix)
		edgeSchs = append(edgeSchs, eg)
	}
	return edgeSchs, nil
}
