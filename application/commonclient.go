package application

import (
	"fmt"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
)

const (
	LIGHT      = "light"
	LIGHTGROUP = "lightgroup"
)

// Application inteface
type Application interface {
	// EventCallback duoc goi khi nhan duoc Event tu phia Device
	// Callback xu ly, lua chon co hay khong Push toi CoreData tuy theo ung dung
	EventCallback(async sdkModel.AsyncValues) error

	// ConnectCallback duoc goi khi Device gui yeu cau Connect
	// Thuong duoc goi boi EventCallback do yeu cau ket noi thuong co dang 1 Event
	ConnectCallback(objectID string) error

	// DisconnectCallback duoc goi khi Device mat ket noi
	DisconnectCallback(objectID string) error

	// ObjectCallback duoc goi khi thong tin Object duoc POST-PUT-DELETE trong MetaData
	ObjectCallback(method string, objectID string) error

	// HandleReadCommands xu ly yeu cau GET Command
	HandleReadCommands(objectID string, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error)

	// HandleWriteCommands xu ly yeu cau PUT Command
	HandleWriteCommands(objectID string, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error
}

// NewApplicationClient tao 1 doi tuong Application
func NewApplicationClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, profile models.DeviceProfile) (Application, error) {
	switch lowerProfileName := strings.ToLower(profile.Name); lowerProfileName {
	case LIGHT:
		return nil, nil
	case LIGHTGROUP:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown profile '%s' requested", profile.Name)
	}
}
