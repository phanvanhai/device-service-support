package serial

import (
	"github.com/phanvanhai/device-service-support/support/pubsub"
	"github.com/tarm/serial"
)

var driver *Serial

type Serial struct {
	serial     *serial.Port
	bus        *pubsub.Publisher
	enableSend chan bool
}

const (
	PortSerialConfigName = "TransferPortSerial"
	BaudSerialConfigName = "TransferBaudSerial"
	DEFAULTBAUD          = 9600
	TimePubDefault       = 30 // second
	ChanSizeDefault      = 10

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
