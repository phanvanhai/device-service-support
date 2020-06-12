package zigbee

import (
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/phanvanhai/device-service-support/support/pubsub"
	"github.com/phanvanhai/device-service-support/transfer"
)

var zb *Zigbee

type Zigbee struct {
	logger   logger.LoggingClient
	config   map[string]string
	tc       transfer.Transfer
	eventBus *pubsub.Publisher
	mutex    sync.Mutex
}

func NewZigbeeClient(lc logger.LoggingClient, tc transfer.Transfer, config map[string]string) (*Zigbee, error) {
	if zb == nil {
		var err error
		zb, err = initialize(lc, tc, config)
		if err != nil {
			return nil, err
		}
	}
	return zb, nil
}
