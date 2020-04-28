package light

import (
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"

	nw "github.com/phanvanhai/device-service-support/network"
	tc "github.com/phanvanhai/device-service-support/transceiver"
)

const (
	Name = "Light"
)

var l *Light

type Light struct {
	lc      logger.LoggingClient
	asyncCh chan<- *sdkModel.AsyncValues
	tc      tc.Transceiver
	nw      nw.Network
}

func NewClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transceiver) (*Light, error) {
	if l == nil {
		l, err := initializeClient(lc, asyncCh, nw, tc)
		return l, err
	}
	return l, nil
}

func initializeClient(lc logger.LoggingClient, asyncCh chan<- *sdkModel.AsyncValues, nw nw.Network, tc tc.Transceiver) (*Light, error) {
	l := &Light{
		lc:      lc,
		asyncCh: asyncCh,
		nw:      nw,
		tc:      tc,
	}
	return l, nil
}
