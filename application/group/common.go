package group

import (
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	nw "github.com/phanvanhai/device-service-support/network"
	tc "github.com/phanvanhai/device-service-support/transceiver"
)

const (
	OnOffDr           = "LightGroup-OnOff"
	DimmingDr         = "LightGroup-Dimming"
	OnOffScheduleDr   = "LightGroup-OnOffSchedule"
	DimmingScheduleDr = "LightGroup-DimmingSchedule"
	MethodDr          = "LightGroup-Method"
	DeviceDr          = "LightGroup-Device"
	ListDeviceDr      = "LightGroup-ListDevice"
	ScenarioDr        = "LightGroup-Scenario"
)

const (
	PutMethod    = "put"
	DeleteMethod = "delete"
)

const (
	GroupLimit           = 50
	OnOffScheduleLimit   = 15
	DimmingScheduleLimit = 15
)

const (
	Name = "LightGroup"
)

var gr *LightGroup

type LightGroup struct {
	lc      logger.LoggingClient
	asyncCh chan<- *sdkModel.AsyncValues
	tc      tc.Transceiver
	nw      nw.Network
}

func NewClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transceiver) (*LightGroup, error) {
	if gr == nil {
		gr, err := initializeClient(lc, asyncCh, nw, tc)
		return gr, err
	}
	return gr, nil
}

func initializeClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transceiver) (*LightGroup, error) {
	gr := &LightGroup{
		lc:      lc,
		asyncCh: asyncCh,
		nw:      nw,
		tc:      tc,
	}
	return gr, nil
}
