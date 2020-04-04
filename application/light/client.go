package light

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"

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
		l, err := initialize(lc, asyncCh, nw, tc)
		return l, err
	}
	return l, nil
}
