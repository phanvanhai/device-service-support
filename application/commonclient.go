package application

import (
	"fmt"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	nw "github.com/phanvanhai/device-service-support/network"
	db "github.com/phanvanhai/device-service-support/support/db"
	tc "github.com/phanvanhai/device-service-support/transfer"

	gatewayApp "github.com/phanvanhai/device-service-support/application/gateway"
	lightGroupApp "github.com/phanvanhai/device-service-support/application/group"
	lightApp "github.com/phanvanhai/device-service-support/application/light"
	scenarioApp "github.com/phanvanhai/device-service-support/application/scenario"
	sensorApp "github.com/phanvanhai/device-service-support/application/sensor"
)

// Application inteface
type Application interface {
	// EventCallback duoc goi khi nhan duoc Event tu phia Device
	// Callback xu ly, lua chon co hay khong Push toi CoreData tuy theo ung dung
	EventCallback(async sdkModel.AsyncValues) error

	Initialize(dev *models.Device) error

	AddDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error
	UpdateDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error
	RemoveDeviceCallback(deviceName string, protocols map[string]models.ProtocolProperties) error

	// HandleReadCommands xu ly yeu cau GET Command
	HandleReadCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error)

	// HandleWriteCommands xu ly yeu cau PUT Command
	HandleWriteCommands(deviceName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error
}

// NewApplicationClient tao 1 doi tuong Application
func NewApplicationClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transfer, deviceName string) (Application, error) {
	profileName := db.DB().GetProfileName(deviceName)
	switch profileName {
	case lightApp.Name:
		return lightApp.NewClient(lc, asyncCh, nw, tc)
	case sensorApp.Name:
		return sensorApp.NewClient(lc, asyncCh, nw, tc)
	case lightGroupApp.Name:
		return lightGroupApp.NewClient(lc, asyncCh, nw, tc)
	case scenarioApp.Name:
		return scenarioApp.NewClient(lc)
	case gatewayApp.Name:
		return gatewayApp.NewClient(lc, nw)
	default:
		return nil, fmt.Errorf("unknown profile '%s' requested of object '%s' ", profileName, deviceName)
	}
}
