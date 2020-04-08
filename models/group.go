package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	nw "github.com/phanvanhai/device-service-support/network"
)

type EdgeGroup struct {
	Group string
}

type netGroup struct {
	Address uint16
}

func groupNetToEdge(nw nw.Network, net netGroup, shifPrefix uint16) EdgeGroup {
	grInt := uint32(shifPrefix<<16) | uint32(net.Address)
	netID := fmt.Sprintf("%04X", grInt)
	name := nw.DeviceNameByNetID(netID)

	return EdgeGroup{
		Group: name,
	}
}

func groupEdgeToNet(nw nw.Network, edge EdgeGroup) netGroup {
	netID := nw.NetIDByDeviceName(edge.Group)
	grInt64, _ := strconv.ParseUint(netID, 16, 32)
	address := uint16(grInt64 & 0xFFFF)
	return netGroup{
		Address: address,
	}
}

// input: ex 2 group: [0x1234, 0xABCD]
// output: ex 2 group: base64([]byte{0x12, 0x34, 0xAB, 0xCD}), kich thuoc dung = size
func netGroupToString(groups []netGroup, size int) string {
	if len(groups) < size {
		g := make([]netGroup, size-len(groups))
		groups = append(groups, g...)
	}
	if len(groups) > size {
		groups = groups[:size]
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, groups)
	grByte := buf.Bytes()
	str := base64.StdEncoding.EncodeToString(grByte)
	return str
}

// input: ex 2 group: base64([]byte{0x12, 0x34, 0xAB, 0xCD}), kich thuoc bieu dien phai dung = size
// output: ex 2 group: "01001234", "0100ABCD"
func stringToNetGroup(groups string, size int) ([]netGroup, error) {
	decoded, err := base64.StdEncoding.DecodeString(groups)
	if err != nil {
		return nil, err
	}

	gr := make([]uint16, size)
	reader := bytes.NewReader(decoded)
	err = binary.Read(reader, binary.BigEndian, gr)
	if err != nil {
		return nil, err
	}

	result := make([]netGroup, 0, size)
	for i := 0; i < size; i++ {
		if gr[i] == 0x0000 {
			continue
		}
		ngr := netGroup{
			Address: gr[i],
		}
		result = append(result, ngr)
	}
	return result, nil
}

func GroupToNetValue(nw nw.Network, groups []EdgeGroup, size int) string {
	netGrs := make([]netGroup, 0, len(groups))
	for _, gr := range groups {
		netGroup := groupEdgeToNet(nw, gr)
		netGrs = append(netGrs, netGroup)
	}
	return netGroupToString(netGrs, size)
}

func NetValueToGroup(nw nw.Network, value string, size int, shifPrefix uint16) ([]EdgeGroup, error) {
	netGroups, err := stringToNetGroup(value, size)
	if err != nil {
		return nil, err
	}

	edgeGrs := make([]EdgeGroup, 0, len(netGroups))
	for _, ng := range netGroups {
		eg := groupNetToEdge(nw, ng, shifPrefix)
		edgeGrs = append(edgeGrs, eg)
	}

	return edgeGrs, nil
}