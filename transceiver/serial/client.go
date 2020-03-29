package serial

import (
	"github.com/edgexfoundry/device-service-package/support/pubsub"
	"github.com/tarm/serial"
)

var driver *Serial

type Serial struct {
	serial     *serial.Port
	bus        *pubsub.Publisher
	enableSend chan bool
}

const (
	PORTSERIAL  = "PortSerial"
	BAUDSERIAL  = "BaudSerial"
	DEFAULTBAUD = 9600
	TIMEPUB     = 60 // second
	CHANSIZEPUB = 10

	PREMBEL = 0x55
)

func NewSerialClient(config map[string]string) (*Serial, error) {
	if driver == nil {
		var err error
		driver, err = initialize(config)
		return driver, err
	}
	return driver, nil
}
