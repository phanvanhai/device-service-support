package application

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"

	nw "github.com/phanvanhai/device-service-support/network"
	db "github.com/phanvanhai/device-service-support/support/db"
	tc "github.com/phanvanhai/device-service-support/transceiver"

	lightApp "github.com/phanvanhai/device-service-support/application/light"
)

// Application inteface
type Application interface {
	// EventCallback duoc goi khi nhan duoc Event tu phia Device
	// Callback xu ly, lua chon co hay khong Push toi CoreData tuy theo ung dung
	EventCallback(async sdkModel.AsyncValues) error

	AddDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error
	UpdateDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error
	RemoveDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties) error

	// HandleReadCommands xu ly yeu cau GET Command
	HandleReadCommands(objectID string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error)

	// HandleWriteCommands xu ly yeu cau PUT Command
	HandleWriteCommands(objectID string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error
}

// NewApplicationClient tao 1 doi tuong Application
func NewApplicationClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transceiver, deviceName string) (Application, error) {
	profileName := db.DB().GetProfileName(deviceName)
	switch profileName {
	case lightApp.Name:
		return lightApp.NewClient(lc, asyncCh, nw, tc)
	// case Sensor:
	// 	return nil, nil
	// case LightGroup:
	// 	return nil, nil
	// case Scenario:
	// 	return nil, nil
	// case Gateway:
	// 	return nil, nil
	default:
		return nil, fmt.Errorf("unknown profile '%s' requested of object '%s' ", profileName, deviceName)
	}
}
