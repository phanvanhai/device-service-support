package gateway

import "fmt"

func (g *Gateway) getRelay() bool {
	return g.relay1
}

func (g *Gateway) setRelay(value bool) {
	str := fmt.Sprintf("Set state of Relay 1 = %t", value)
	g.lc.Info(str)
	g.relay1 = value
}
