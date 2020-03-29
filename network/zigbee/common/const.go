package common

const (
	ZigbeeNetworkIDConst = 0
)

const (
	TIMEPUB                      = 60 // second
	CHANSIZEPUB                  = 10
	SendRequestTimeoutConst      = 60000 // int64: milisecond
	ReceiverResponseTimeoutConst = 60000 // int64: milisecond
)

const (
	//CommandCmdConst :
	CommandCmdConst = uint8(iota)

	//PushEventCmdConst :
	ReportConst

	//AddObjectCmdConst :
	AddObjectCmdConst

	//DeleteObjectCmdConst :
	DeleteObjectCmdConst

	//ScanCmdConst :
	ScanCmdConst
)

type StatusCode uint8

const (
	Success StatusCode = 0
	Error   StatusCode = 1
)

// Const of Protocols.General
const (
	ProtocolNameConst    = "General"
	MACPropertyConst     = "MAC"
	LinkKeyPropertyConst = "LinkKey"
	IDPropertyConst      = "ID"
	TypePropertyConst    = "Type"
	AttributeNetResource = "AttZigbee"

	DeviceTypeConst   = "Device"
	GroupTypeConst    = "Group"
	ScenarioTypeConst = "Scenario"

	PrefixHexValueNetGroupIDConst = 0x0100
)
