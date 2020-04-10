package scenario

import (
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
)

const (
	Name = "Scenario"
)

var s *Scenario

type Scenario struct {
	lc logger.LoggingClient
}

func NewClient(lc logger.LoggingClient) (*Scenario, error) {
	if s == nil {
		s, err := initializeClient(lc)
		return s, err
	}
	return s, nil
}

func initializeClient(lc logger.LoggingClient) (*Scenario, error) {
	s := &Scenario{
		lc: lc,
	}
	return s, nil
}
