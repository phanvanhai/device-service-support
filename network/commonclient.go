package network

import (
	"fmt"
	"strings"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/network/zigbee"
	"github.com/phanvanhai/device-service-support/transceiver"
)

const NetworkTypeConfigConst = "NetworkType"
const (
	ZIGBEE = "zigbee"
)

// Network interface
type Network interface {
	Close() error

	// UpdateObjectCallback duoc dung de cap nhap thong tin Object vao DataBase cua mang
	// Thuong duoc goi dau tien trong ObjectCallback o lop Application
	UpdateObjectCallback(object *models.Device)

	// DeleteObjectCallback duoc dung de cap nhap thong tin Object vao DataBase cua mang
	// Thuong duoc goi sau DeleteObject() trong ObjectCallback o lop Application
	DeleteObjectCallback(name string)

	// AddObject dung de them moi 1 Object vao mang
	// Thuong duoc goi sau UpdateObjectCallback() trong ObjectCallback o lop Application
	AddObject(newObject *models.Device) (*models.Device, error)

	// UpdateObject dung de thay doi thong tin mang Object
	// Can than su dung UpdateObject vi khi duoc goi trong ObjectCallback co the dan den lap vo han
	UpdateObject(newObject *models.Device) error

	// DeleteObject dung de xoa Object ra khoi mang
	// Thuong duoc goi dau tien trong ObjectCallback o lop Application
	DeleteObject(name string, protocols map[string]models.ProtocolProperties) error

	ReadCommands(name string, reqs []*sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error)
	WriteCommands(name string, reqs []*sdkModel.CommandRequest, params []*sdkModel.CommandValue) error

	// UpdateFirmware dung de update firmwart cho Device
	// Thuong duoc su dung nhu 1 lenh cua Gateway Application
	UpdateFirmware(name string, file interface{}) error

	// Discovery tim kiem thong cac Device moi
	// Thuong duoc su dung boi Gateway Application hoac ProtocolDiscovery()
	Discovery() (devices *interface{}, err error)

	// ListenEvent lang nghe cac Event tu cac Device
	// Thuong duoc su dung boi mot goroutine o lop Application
	// internal: network se chay 1 distribution goroutine chi de nghe Event, sau do publish toi cac subscriber
	// return sdkModel.AsyncValue
	ListenEvent() chan interface{}

	// ConvertResourceByDevice convert from Resource A to Resource B of Device DevID
	ConvertResourceByDevice(fromDevName string, rsFrom string, toDevName string) string

	// DeviceIDByNetID
	DeviceNameByNetID(netID string) string

	// NetIDByDeviceID
	NetIDByDeviceName(name string) string

	// CheckExist
	CheckExist(name string) bool
}

func NewNetworkClient(networkType string, tc transceiver.Transceiver, lc logger.LoggingClient, config map[string]string) (Network, error) {
	switch nw := strings.ToLower(networkType); nw {
	case ZIGBEE:
		return zigbee.NewZigbeeClient(lc, tc, config)
	default:
		return nil, fmt.Errorf("unknown network type '%s' requested", networkType)
	}
}
