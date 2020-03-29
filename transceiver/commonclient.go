package transceiver

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/device-service-package/transceiver/serial"
)

type Frame struct {
	Preambel byte
	Length   uint16
	Payload  []byte
	Crc      byte
}

const (
	SERIAL = "serial"
)

// Transceiver interface
type Transceiver interface {
	// Sender gui payload trong khoang thoi gian toi da timeout
	Sender(payload []byte, timeout int64) error

	// Listen nhan payload []byte voi bo loc filter
	Listen(filter func(v interface{}) bool) chan interface{}

	// CancelListen huy lang nghe nhan du lieu
	CancelListen(sub chan interface{})

	Close() error
}

// NewTransceiverClient func
func NewTransceiverClient(transceiverType string, config map[string]string) (Transceiver, error) {
	switch tc := strings.ToLower(transceiverType); tc {
	case SERIAL:
		return serial.NewSerialClient(config)
	default:
		return nil, fmt.Errorf("unknown transceiver type '%s' requested", transceiverType)
	}
}
