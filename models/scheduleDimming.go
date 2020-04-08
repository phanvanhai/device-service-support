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

type EdgeScheduleDimming struct {
	OwnerName string `json:"owner, omitempty"`
	Time      uint32 `json:"time, omitempty"`
	Value     uint16 `json:"value, omitempty"`
}

type netScheduleDimming struct {
	OwnerAddress uint16
	Time         uint32
	Value        uint16
}

func scheduleDimmingNetToEdge(nw nw.Network, net netScheduleDimming, ownerName string, shifPrefix uint16) EdgeScheduleDimming {
	result := EdgeScheduleDimming{
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

func scheduleDimmingEdgeToNet(nw nw.Network, edge EdgeScheduleDimming, ownerName string) netScheduleDimming {
	result := netScheduleDimming{
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

func netScheduleDimmingToString(schedules []netScheduleDimming, size int) string {
	if len(schedules) < size {
		s := make([]netScheduleDimming, size-len(schedules))
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
func stringToNetScheduleDimming(scheduleStr string, size int) ([]netScheduleDimming, error) {
	decode, err := base64.StdEncoding.DecodeString(scheduleStr)
	if err != nil {
		return nil, err
	}

	sch := make([]netScheduleDimming, size)
	reader := bytes.NewBuffer(decode)
	err = binary.Read(reader, binary.BigEndian, sch)
	if err != nil {
		return nil, err
	}

	result := make([]netScheduleDimming, 0, size)
	for i := 0; i < size; i++ {
		if cm.CheckScheduleTime(sch[i].Time) == false {
			continue
		}
		result = append(result, sch[i])
	}
	return result, nil
}

func ScheduleDimmingEdgeToNetValue(nw nw.Network, schedules []EdgeScheduleDimming, ownerName string, size int) string {
	netSchs := make([]netScheduleDimming, 0, len(schedules))
	for _, sch := range schedules {
		netSch := scheduleDimmingEdgeToNet(nw, sch, ownerName)
		netSchs = append(netSchs, netSch)
	}
	return netScheduleDimmingToString(netSchs, size)
}

func NetValueToScheduleDimming(nw nw.Network, value string, size int, ownerName string, shifPrefix uint16) ([]EdgeScheduleDimming, error) {
	netSchs, err := stringToNetScheduleDimming(value, size)
	if err != nil {
		return nil, err
	}
	edgeSchs := make([]EdgeScheduleDimming, 0, len(netSchs))
	for _, sch := range netSchs {
		eg := scheduleDimmingNetToEdge(nw, sch, ownerName, shifPrefix)
		edgeSchs = append(edgeSchs, eg)
	}
	return edgeSchs, nil
}
