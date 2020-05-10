package gateway

import "fmt"

func (g *Gateway) getRelay() bool {
	return g.relay1
}

func (g *Gateway) setRelay(value bool) {
	g.lc.Info(fmt.Sprintf("Set state of Relay 1 = %t", value))
	g.relay1 = value
}
