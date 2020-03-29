package models

type Header struct {
	Cmd uint8 `json:"cmd, omitempty"`
	// PacketId uint16 `json:"hid, omitempty"`
}

type Response struct {
	StatusCode    uint8  `json:"sc, omitempty"`
	StatusMessage string `json:"sm, omitempty"`
}

type CommandPacket struct {
	Header
	NetEvent
	Response
}

// DevicePacket : chi ho tro Add-Delete, khong su dung Update
type DevicePacket struct {
	Header
	MAC       string `json:"mac, omitempty"` // "AABBCCDD012345678"
	LinkKey   string `json:"lk, omitempty"`
	NetDevice string `json:"dev, omitempty"` // Address = ObjectType(1B) + Endpoint(1B) + Object Address(2B)
	Response
}

type Discovery struct {
	Header
	TimeOut uint8  `json:"to, omitempty"`
	Content string `json:"ct, omitempty"`
	Response
}
