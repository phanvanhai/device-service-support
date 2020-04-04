package cm

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

// Const of Network
const (
	MACProperty          = "NetworkMAC"
	LinkKeyProperty      = "NetworkLinkKey"
	NetIDProperty        = "NetworkID"
	AttributeNetResource = "AttZigbee"

	PrefixHexValueNetGroupID = 0x0100

	EventPublishTimeDefault    = int64(10000) // milisecond
	EventBufferSizeDefault     = int(16)
	EventPublishTimeConfigName = "Network_EventPublishTime"
	EventBufferSizeConfigName  = "Network_EventBufferSize"

	RequestTimeoutDefault    = int64(30000) // milisecond
	ResponseTimoutDefault    = int64(30000) // milisecond
	RequestTimeoutConfigName = "Network_RequestTimeout"
	ResponseTimoutConfigName = "Network_ResponseTimeout"
)
