package transfer

import (
	"fmt"
	"strings"

	"github.com/phanvanhai/device-service-support/transfer/serial"
)

const TransferTypeConfigConst = "TransferType"
const (
	SERIAL = "serial"
)

// Transfer interface
type Transfer interface {
	// Sender gui payload trong khoang thoi gian toi da timeout
	Sender(payload []byte, timeout int64) error

	// Listen nhan payload []byte voi bo loc filter
	Listen(filter func(v interface{}) bool) chan interface{}

	// CancelListen huy lang nghe nhan du lieu
	CancelListen(sub chan interface{})

	Close() error
}

// NewTransferClient func
func NewTransferClient(transferType string, config map[string]string) (Transfer, error) {
	switch tc := strings.ToLower(transferType); tc {
	case SERIAL:
		return serial.NewSerialClient(config)
	default:
		return nil, fmt.Errorf("unknown transfer type '%s' requested", transferType)
	}
}
