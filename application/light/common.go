package light

import (
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	nw "github.com/phanvanhai/device-service-support/network"
	tc "github.com/phanvanhai/device-service-support/transfer"
)

const (
	OnOffDr           = "Light_OnOff"
	DimmingDr         = "Light_Dimming"
	OnOffScheduleDr   = "Light_OnOffSchedule"
	DimmingScheduleDr = "Light_DimmingSchedule"
	LightMeasureDr    = "Light_LightMeasure"
	ReportTimeDr      = "Light_ReportTime"
	RealtimeDr        = "Light_Realtime"
	HistoricalEventDr = "Light_HistoricalEvent"
	GroupDr           = "Light_Group"
	ScenarioDr        = "Light_Scenario"
	PingDr            = "Light_Ping"
)

const (
	GroupLimit           = 50
	OnOffScheduleLimit   = 16
	DimmingScheduleLimit = 16
)

const (
	Name = "Light"
)

var l *Light

type Light struct {
	lc      logger.LoggingClient
	asyncCh chan<- *sdkModel.AsyncValues
	tc      tc.Transfer
	nw      nw.Network
}

func NewClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transfer) (*Light, error) {
	if l == nil {
		l, err := initializeClient(lc, asyncCh, nw, tc)
		return l, err
	}
	return l, nil
}

func initializeClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transfer) (*Light, error) {
	l := &Light{
		lc:      lc,
		asyncCh: asyncCh,
		nw:      nw,
		tc:      tc,
	}
	return l, nil
}
