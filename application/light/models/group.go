package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/phanvanhai/device-service-support/application/light/cm"

	nw "github.com/phanvanhai/device-service-support/network"
	zigbeeCm "github.com/phanvanhai/device-service-support/network/zigbee/cm"
	db "github.com/phanvanhai/device-service-support/support/db"
)

type EdgeGroup struct {
	Group string
}

type NetGroup struct {
	Address uint16
}

func groupNetToEdge(nw nw.Network, net NetGroup) EdgeGroup {
	grInt := uint32(zigbeeCm.PrefixHexValueNetGroupID<<16) | uint32(net.Address)
	netID := fmt.Sprintf("%04X", grInt)
	name := nw.DeviceNameByNetID(netID)

	return EdgeGroup{
		Group: name,
	}
}

func groupEdgeToNet(nw nw.Network, edge EdgeGroup) NetGroup {
	netID := nw.NetIDByDeviceName(edge.Group)
	grInt64, _ := strconv.ParseUint(netID, 16, 32)
	address := uint16(grInt64 & 0xFFFF)
	return NetGroup{
		Address: address,
	}
}

// input: ex 2 group: [0x1234, 0xABCD]
// output: ex 2 group: base64([]byte{0x12, 0x34, 0xAB, 0xCD}), kich thuoc dung = size
func netGroupToString(groups []NetGroup, size int) string {
	if len(groups) < size {
		g := make([]NetGroup, size-len(groups))
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
func stringToNetGroup(groups string, size int) ([]NetGroup, error) {
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

	result := make([]NetGroup, 0, size)
	for i := 0; i < size; i++ {
		if gr[i] == 0x0000 {
			continue
		}
		ngr := NetGroup{
			Address: gr[i],
		}
		result = append(result, ngr)
	}
	return result, nil
}

func GroupToNetValue(nw nw.Network, grInfo []db.RelationContent) string {
	netGrs := make([]NetGroup, 0, len(grInfo))
	for _, gr := range grInfo {
		netGroup := groupEdgeToNet(nw, EdgeGroup{Group: gr.Parent})
		netGrs = append(netGrs, netGroup)
	}
	return netGroupToString(netGrs, cm.GroupLimit)
}

func NetValueToGroup(nw nw.Network, value string) ([]string, error) {
	netGroups, err := stringToNetGroup(value, cm.GroupLimit)
	if err != nil {
		return nil, err
	}

	edgeGrs := make([]EdgeGroup, 0, len(netGroups))
	for _, ng := range netGroups {
		eg := groupNetToEdge(nw, ng)
		edgeGrs = append(edgeGrs, eg)
	}

	grs := make([]string, 0, len(edgeGrs))
	for _, eg := range edgeGrs {
		grs = append(grs, eg.Group)
	}

	return grs, nil
}
