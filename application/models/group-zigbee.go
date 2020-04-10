package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/phanvanhai/device-service-support/support/db"

	nw "github.com/phanvanhai/device-service-support/network"
	zigbeeConstants "github.com/phanvanhai/device-service-support/network/zigbee/cm"
)

// Cloud <-> DS: '[1234,4321]'
// Coord --> DS: base64('uint16uint16')

func convertEdgeToNetGroup(nw nw.Network, name string) uint16 {
	netID := nw.NetIDByDeviceName(name)
	grInt64, _ := strconv.ParseUint(netID, 16, 32)
	address := uint16(grInt64 & 0xFFFF)
	return address
}

func convertNetToEdgeGroup(nw nw.Network, addr uint16) string {
	shifPrefix := zigbeeConstants.PrefixHexValueNetGroupID
	grInt := uint32(shifPrefix<<16) | uint32(addr)
	netID := fmt.Sprintf("%04X", grInt)
	name := nw.DeviceNameByNetID(netID)
	return name
}

func encodeNetGroups(groups []uint16, size int) string {
	if len(groups) < size {
		g := make([]uint16, size-len(groups))
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
func decodeNetGroups(value string, size int) ([]uint16, error) {
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}

	gr := make([]uint16, size)
	reader := bytes.NewReader(decoded)
	err = binary.Read(reader, binary.BigEndian, gr)
	if err != nil {
		return nil, err
	}

	result := make([]uint16, 0, size)
	for i := 0; i < size; i++ {
		if gr[i] == 0x0000 {
			continue
		}
		result = append(result, gr[i])
	}
	return result, nil
}

func GroupToNetValue(nw nw.Network, groups []string, size int) string {
	netGrs := make([]uint16, 0, len(groups))
	for _, gr := range groups {
		netGroup := convertEdgeToNetGroup(nw, gr)
		netGrs = append(netGrs, netGroup)
	}
	return encodeNetGroups(netGrs, size)
}

func NetValueToGroup(nw nw.Network, value string, size int) ([]string, error) {
	netGroups, err := decodeNetGroups(value, size)
	if err != nil {
		return nil, err
	}

	edgeGrs := make([]string, 0, len(netGroups))
	for _, ng := range netGroups {
		eg := convertNetToEdgeGroup(nw, ng)
		edgeGrs = append(edgeGrs, eg)
	}

	return edgeGrs, nil
}

func RelationGroupToNetValue(nw nw.Network, relations []db.RelationContent, size int) string {
	groups := make([]string, 0, len(relations))
	for _, relation := range relations {
		groups = append(groups, relation.Parent)
	}
	return GroupToNetValue(nw, groups, size)
}

// String returns a JSON encoded string representation of the model
func RelationGroupToString(relations []db.RelationContent) (string, error) {
	groups := make([]string, 0, len(relations))
	for _, relation := range relations {
		groups = append(groups, relation.Parent)
	}

	out, err := json.Marshal(groups)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
