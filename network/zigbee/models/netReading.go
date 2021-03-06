package models

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/phanvanhai/device-service-support/network/zigbee/cache"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
)

type NetReading struct {
	NetOrigin    int64              `json:"o,omitempty"` // When the reading was created
	NetResource  string             `json:"r,omitempty"` // = AttributeID(2B) + ClusterID(2B) + ProfileID(2B)
	NetValueType sdkModel.ValueType `json:"t"`
	NetValue     string             `json:"v,omitempty"` // interface{} --> base64 --> string
	isValidated  bool               // internal member used for validation check
}

// UnmarshalJSON implements the Unmarshaler interface for the Reading type
func (r *NetReading) UnmarshalJSON(data []byte) error {
	var err error
	type Alias struct {
		NetOrigin    int64              `json:"o"`
		NetResource  *string            `json:"r"`
		NetValueType sdkModel.ValueType `json:"t"`
		NetValue     *string            `json:"v"`
	}
	a := Alias{}

	// Error with unmarshaling
	if err = json.Unmarshal(data, &a); err != nil {
		return err
	}

	// Set the fields
	if a.NetResource != nil {
		r.NetResource = *a.NetResource
	}

	if a.NetValue != nil {
		r.NetValue = *a.NetValue
	}

	r.NetOrigin = a.NetOrigin
	r.NetValueType = a.NetValueType

	r.isValidated, err = r.Validate()
	return err
}

// Validate satisfies the Validator interface
func (r NetReading) Validate() (bool, error) {
	if !r.isValidated {
		if r.NetResource == "" {
			return false, fmt.Errorf("name for network reading's value descriptor not specified")
		}
		if r.NetValue == "" {
			return false, fmt.Errorf("network reading has no value")
		}
	}
	return true, nil
}

// String returns a JSON encoded string representation of the model
func (r NetReading) String() string {
	out, err := json.Marshal(r)
	if err != nil {
		return err.Error()
	}

	return string(out)
}

func (r *NetReading) toCommandValue(devName string) (cv *sdkModel.CommandValue, err error) {
	resourceName, err := cache.Cache().DeviceResourceByNetResource(devName, r.NetResource)
	if err != nil {
		return
	}

	svc := sdk.RunningService()
	dr, ok := svc.DeviceResource(devName, resourceName, "")
	if !ok {
		return
	}
	r.NetValueType = sdkModel.ParseValueType(dr.Properties.Value.Type)

	switch r.NetValueType {
	case sdkModel.Binary:
		value, err := base64.StdEncoding.DecodeString(r.NetValue)
		if err != nil {
			return nil, err
		}
		cv, err = sdkModel.NewBinaryValue(resourceName, r.NetOrigin, value)
	case sdkModel.String:
		cv = sdkModel.NewStringValue(resourceName, r.NetOrigin, r.NetValue)
	default:
		value, err := base64.StdEncoding.DecodeString(r.NetValue)
		if err != nil {
			return nil, err
		}
		cv = &sdkModel.CommandValue{
			DeviceResourceName: resourceName,
			Origin:             r.NetOrigin,
			Type:               r.NetValueType,
			NumericValue:       value,
		}
	}
	return
}

func commandValueToNetReading(cv *sdkModel.CommandValue) (netReading *NetReading, err error) {
	netResourceName, err := cache.Cache().NetResourceByDeviceResource(cv.DeviceResourceName)
	if err != nil {
		return
	}

	netReading = &NetReading{
		NetResource:  netResourceName,
		NetOrigin:    cv.Origin,
		NetValueType: cv.Type,
	}
	// encodeValue --> string base64
	switch cv.Type {
	case sdkModel.Binary:
		netReading.NetValue = base64.StdEncoding.EncodeToString(cv.BinValue)
	case sdkModel.String:
		value, _ := cv.StringValue()
		// netReading.NetValue = base64.StdEncoding.EncodeToString([]byte(value))
		netReading.NetValue = value
	default:
		netReading.NetValue = base64.StdEncoding.EncodeToString(cv.NumericValue)
	}

	return
}

func commandRequestToNetReading(rq *sdkModel.CommandRequest) (netReading *NetReading, err error) {
	netResourceName, err := cache.Cache().NetResourceByDeviceResource(rq.DeviceResourceName)
	if err != nil {
		return
	}

	netReading = &NetReading{
		NetResource:  netResourceName,
		NetValueType: rq.Type,
	}
	return
}
