package models

import (
	"encoding/json"
	"fmt"

	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	"github.com/phanvanhai/device-service-support/network/zigbee/cache"
)

type NetEvent struct {
	NetDevice   string       `json:"dev,omitempty"` // Address = ObjectType(1B) + Endpoint(1B) + Object Address(2B) = "0100ABCD"
	NetReadings []NetReading `json:"evt,omitempty"`
	isValidated bool         // internal member used for validation check
}

// UnmarshalJSON implements the Unmarshaler interface for the NetEvent type
func (e *NetEvent) UnmarshalJSON(data []byte) error {
	var err error
	type Alias struct {
		NetDevice   *string      `json:"dev"`
		NetReadings []NetReading `json:"evt"`
	}
	a := Alias{}

	// Error with unmarshaling
	if err = json.Unmarshal(data, &a); err != nil {
		return err
	}

	// Set the fields
	if a.NetDevice != nil {
		e.NetDevice = *a.NetDevice
	}
	e.NetReadings = a.NetReadings

	e.isValidated, err = e.Validate()
	return err
}

// Validate satisfies the Validator interface
func (e NetEvent) Validate() (bool, error) {
	if !e.isValidated {
		if e.NetDevice == "" {
			return false, fmt.Errorf("source device for event not specified")
		}
	}
	return true, nil
}

// String returns a JSON encoded string representation of the model
func (e NetEvent) String() string {
	out, err := json.Marshal(e)
	if err != nil {
		return err.Error()
	}

	return string(out)
}

func (e *NetEvent) ToEdgeEvent() (async sdkModel.AsyncValues, err error) {
	devName, err := cache.Cache().DeviceNameByNetID(e.NetDevice)
	if err != nil {
		return sdkModel.AsyncValues{}, err
	}

	cvs := make([]*sdkModel.CommandValue, 0, len(e.NetReadings))
	for _, netReading := range e.NetReadings {
		cv, err := netReading.toCommandValue(devName)
		if err != nil {
			return sdkModel.AsyncValues{}, err
		}
		cvs = append(cvs, cv)
	}

	async.DeviceName = devName
	async.CommandValues = cvs
	return
}

func CommandValueToNetEvent(name string, cvs []*sdkModel.CommandValue) (e *NetEvent, err error) {
	netDevice, err := cache.Cache().NetIDByDeviceName(name)
	if err != nil {
		return nil, err
	}

	netReadings := make([]NetReading, 0, len(cvs))
	for _, cv := range cvs {
		reading, err := commandValueToNetReading(cv)
		if err != nil {
			return nil, err
		}
		netReadings = append(netReadings, *reading)
	}
	e = &NetEvent{
		NetDevice:   netDevice,
		NetReadings: netReadings,
	}

	return
}

func CommandRequestToNetEvent(name string, rqs []*sdkModel.CommandRequest) (e *NetEvent, err error) {
	netDevice, err := cache.Cache().NetIDByDeviceName(name)
	if err != nil {
		return nil, err
	}

	netReadings := make([]NetReading, 0, len(rqs))
	for _, rq := range rqs {
		reading, err := commandRequestToNetReading(rq)
		if err != nil {
			return nil, err
		}
		netReadings = append(netReadings, *reading)
	}
	e = &NetEvent{
		NetDevice:   netDevice,
		NetReadings: netReadings,
	}
	return
}
