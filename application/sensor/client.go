package sensor

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"

	nw "github.com/phanvanhai/device-service-support/network"
	tc "github.com/phanvanhai/device-service-support/transceiver"
)

const (
	Name = "Sensor"
)

var s *Sensor

type Sensor struct {
	lc      logger.LoggingClient
	asyncCh chan<- *sdkModel.AsyncValues
	tc      tc.Transceiver
	nw      nw.Network
}

func NewClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transceiver) (*Sensor, error) {
	if s == nil {
		s, err := initializeClient(lc, asyncCh, nw, tc)
		return s, err
	}
	return s, nil
}

func initializeClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transceiver) (*Sensor, error) {
	s := &Sensor{
		lc:      lc,
		asyncCh: asyncCh,
		nw:      nw,
		tc:      tc,
	}
	return s, nil
}
