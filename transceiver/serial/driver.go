package serial

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/edgexfoundry/device-service-package/support/pubsub"
	"github.com/tarm/serial"
)

func initialize(config map[string]string) (*Serial, error) {
	port, ok := config[PORTSERIAL]
	if ok != true {
		return nil, fmt.Errorf("Khong co thong tin Port cho Serial")
	}

	var baud int
	var err error
	baudStr, ok := config[BAUDSERIAL]
	if ok != true {
		baud = DEFAULTBAUD
	} else {
		baud64, err := strconv.ParseInt(baudStr, 10, 32)
		if err != nil {
			return nil, err
		}
		baud = int(baud64)
	}

	configSerial := &serial.Config{Name: port, Baud: baud}
	serialPort, err := serial.OpenPort(configSerial)
	if err != nil {
		return nil, err
	}

	p := pubsub.NewPublisher(TIMEPUB*time.Second, CHANSIZEPUB)

	dr := &Serial{
		serial:     serialPort,
		bus:        p,
		enableSend: make(chan bool, 1),
	}

	go dr.receiverFrameRoutine()

	return dr, nil
}

func (dr *Serial) Sender(payload []byte, timeout int64) error {
	var err error
	frame, err := newFrameFrom(payload)
	if err != nil {
		return err
	}

	timeOut := time.After(time.Duration(timeout) * time.Millisecond)
	select {
	case <-timeOut:
		err = fmt.Errorf("Loi: Timeout")
	case dr.enableSend <- true:
		chanSendErr := make(chan error, 1)
		go dr.sendSerial(frame, chanSendErr)
		err, _ = <-chanSendErr
	}
	return err
}

// Listen nhan payload []byte voi bo loc filter
func (dr *Serial) Listen(filter func(v interface{}) bool) chan interface{} {
	return dr.bus.SubscribeTopic(filter)
}

func (dr *Serial) CancelListen(sub chan interface{}) {
	dr.bus.Evict(sub)
}

func (dr *Serial) Close() error {
	close(dr.enableSend)
	dr.bus.Close()
	return dr.serial.Close()
}

func newFrameFrom(payload []byte) ([]byte, error) {
	if len(payload) > math.MaxUint16 {
		return nil, fmt.Errorf("Kich thuoc payload vuot qua gioi han")
	}
	lengthPayload := len(payload)
	highLength := byte(lengthPayload >> 8)
	lowLength := byte(lengthPayload & 0x00FF)

	frame := make([]byte, 0, lengthPayload+4) // 4 = preambel + length + crc
	frame = append(frame, PREMBEL)
	frame = append(frame, highLength)
	frame = append(frame, lowLength)
	frame = append(frame, payload...)

	calCRC := PREMBEL + highLength + lowLength
	for _, c := range payload {
		calCRC += c
	}

	frame = append(frame, calCRC)
	return frame, nil
}

func (dr *Serial) sendSerial(payload []byte, chanSendErr chan error) {
	dr.serial.Flush()
	_, err := dr.serial.Write(payload)
	chanSendErr <- err
	<-dr.enableSend
}

func (dr *Serial) receiverFrameRoutine() {
	for {
		payload, l := dr.receiverSerial()
		if l > 0 {
			dr.bus.Publish(payload)
		}
	}
}

func (dr *Serial) receiverSerial() ([]byte, uint16) {
	var n int
	var err error

	header := make([]byte, 1)
	lengthBytes := make([]byte, 2)
	var payloadBytes []byte
	var lengthPayload uint16
	crc := make([]byte, 1)
	var checkCRC byte
	data := make([]byte, 1)

	for {
		// receive Header 1 byte:
		for {
			n, err = dr.serial.Read(header)
			if err != nil {
				return nil, 0
			}
			if n < 1 {
				continue
			}
			if header[0] == PREMBEL {
				break
			}
		}

		// receive length 2 byte
		n, err = dr.serial.Read(data)
		if err != nil {
			return nil, 0
		}
		if n < 1 {
			continue
		}
		lengthBytes[0] = data[0]

		n, err = dr.serial.Read(data)
		if err != nil {
			return nil, 0
		}
		if n < 1 {
			continue
		}
		lengthBytes[1] = data[0]

		lengthPayload = (uint16(lengthBytes[0]) << 8) | uint16(lengthBytes[1])
		if lengthPayload <= 1 {
			continue
		}

		// reset checkCRC
		checkCRC = PREMBEL + lengthBytes[0] + lengthBytes[1]

		// receive payload
		payloadBytes = make([]byte, lengthPayload)
		for i := range payloadBytes {
			n, err = dr.serial.Read(data)
			if err != nil || n < 1 {
				return nil, 0
			}
			payloadBytes[i] = data[0]

			// calculate Crc
			checkCRC += data[0]
		}

		// receive CRC
		n, err = dr.serial.Read(crc)
		if err != nil {
			return nil, 0
		}
		if n < 1 {
			continue
		}

		// uncomment de su dung CRC
		// if checkCRC != crc[0] {
		// 	continue
		// }
		break
	}

	return payloadBytes, lengthPayload
}
