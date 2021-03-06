package sensor

import (
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	nw "github.com/phanvanhai/device-service-support/network"
	tc "github.com/phanvanhai/device-service-support/transfer"
)

const (
	OnOffDr           = "Sensor_OnOff"
	MeasurePowerDr    = "Sensor_MeasurePower"
	ReportTimeDr      = "Sensor_ReportTime"
	RealtimeDr        = "Sensor_Realtime"
	HistoricalEventDr = "Sensor_HistoricalEvent"
	ScenarioDr        = "Sensor_Scenario"
	PingDr            = "Sensor_Ping"
)

const (
	Name = "Sensor"
)

var s *Sensor

type Sensor struct {
	lc      logger.LoggingClient
	asyncCh chan<- *sdkModel.AsyncValues
	tc      tc.Transfer
	nw      nw.Network
}

func NewClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transfer) (*Sensor, error) {
	if s == nil {
		s, err := initializeClient(lc, asyncCh, nw, tc)
		return s, err
	}
	return s, nil
}

func initializeClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transfer) (*Sensor, error) {
	s := &Sensor{
		lc:      lc,
		asyncCh: asyncCh,
		nw:      nw,
		tc:      tc,
	}
	return s, nil
}
