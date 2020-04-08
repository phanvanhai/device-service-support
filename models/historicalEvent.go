package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
)

type EdgeHistoricalEvent struct {
	RealTime uint64 `json:"realtime, omitempty"`
	Event    string `json:"event, omitempty"`
}

type netHistoricalEvent struct {
	RealTime uint64 `json:"realtime, omitempty"`
	Event    string `json:"event, omitempty"`
}

func historicalEventNetToEdge(netEvent netHistoricalEvent) EdgeHistoricalEvent {
	return EdgeHistoricalEvent{
		RealTime: netEvent.RealTime,
		Event:    netEvent.Event,
	}
}

func historicalEventEdgeToNet(edge EdgeHistoricalEvent) netHistoricalEvent {
	return netHistoricalEvent{
		RealTime: edge.RealTime,
		Event:    edge.Event,
	}
}

// input: ex 2 group: [0x1234, 0xABCD]
// output: ex 2 group: base64([]byte{0x12, 0x34, 0xAB, 0xCD}), kich thuoc dung = size
func netHistoricalEventToString(events []netHistoricalEvent, size int) string {
	if len(events) < size {
		g := make([]netHistoricalEvent, size-len(events))
		events = append(events, g...)
	}
	if len(events) > size {
		events = events[:size]
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, events)
	eventByte := buf.Bytes()
	str := base64.StdEncoding.EncodeToString(eventByte)
	return str
}

// input: ex 2 group: base64([]byte{0x12, 0x34, 0xAB, 0xCD}), kich thuoc bieu dien phai dung = size
// output: ex 2 group: "01001234", "0100ABCD"
func stringToNetHistoricalEvent(events string, size int) ([]netHistoricalEvent, error) {
	decoded, err := base64.StdEncoding.DecodeString(events)
	if err != nil {
		return nil, err
	}

	result := make([]netHistoricalEvent, size)
	reader := bytes.NewReader(decoded)
	err = binary.Read(reader, binary.BigEndian, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func EdgeHistoricalEventToNetValue(events []EdgeHistoricalEvent, size int) string {
	netEvents := make([]netHistoricalEvent, 0, len(events))
	for _, events := range events {
		net := historicalEventEdgeToNet(events)
		netEvents = append(netEvents, net)
	}
	return netHistoricalEventToString(netEvents, size)
}

func NetValueToEdgeHistoricalEvent(value string, size int) ([]EdgeHistoricalEvent, error) {
	netEvents, err := stringToNetHistoricalEvent(value, size)
	if err != nil {
		return nil, err
	}

	edgeEvents := make([]EdgeHistoricalEvent, 0, len(netEvents))
	for _, event := range netEvents {
		eg := historicalEventNetToEdge(event)
		edgeEvents = append(edgeEvents, eg)
	}

	return edgeEvents, nil
}
