package network

type Packet struct {
	ContentType uint8
	PacketID    uint16
	Content     string
}
