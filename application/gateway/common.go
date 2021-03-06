package gateway

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	nw "github.com/phanvanhai/device-service-support/network"
	"github.com/stianeikeland/go-rpio"
)

const (
	OnOffRelay1Dr          = "Gateway_OnOffRelay1"
	EventDr                = "Gateway_Event"
	UpdateDeviceFirmwareDr = "Gateway_UpdateDeviceFirmware"
)

const (
	Name = "Gateway"
)

var g *Gateway

type Gateway struct {
	lc     logger.LoggingClient
	nw     nw.Network
	relay1 rpio.Pin
}

func NewClient(lc logger.LoggingClient, nw nw.Network) (*Gateway, error) {
	if g == nil {
		g, err := initializeClient(lc, nw)
		return g, err
	}
	return g, nil
}

func initializeClient(lc logger.LoggingClient, nw nw.Network) (*Gateway, error) {
	g := &Gateway{
		lc: lc,
		nw: nw,
	}
	return g, nil
}
