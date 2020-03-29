package zigbee

import (
	"sync"

	"github.com/edgexfoundry/device-service-package/support/pubsub"
	"github.com/edgexfoundry/device-service-package/transceiver"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

var zb *Zigbee

type Zigbee struct {
	logger   logger.LoggingClient
	config   map[string]string
	tc       transceiver.Transceiver
	eventBus *pubsub.Publisher
	mutex    sync.Mutex
}

func NewZigbeeClient(lc logger.LoggingClient, tc transceiver.Transceiver, config map[string]string) (*Zigbee, error) {
	if zb == nil {
		var err error
		zb, err = initialize(lc, tc, config)
		if err != nil {
			return nil, err
		}
	}
	return zb, nil
}
