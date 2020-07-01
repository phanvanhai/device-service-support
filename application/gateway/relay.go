package gateway

import "fmt"

func (g *Gateway) getRelay() bool {
	g.relay1.Input()
	status := uint8(g.relay1.Read())
	return (status != 0)
}

func (g *Gateway) setRelay(value bool) {
	g.lc.Info(fmt.Sprintf("Set state of Relay 1 = %t", value))
	g.relay1.Output()
	if value == true {
		g.relay1.High()
	} else {
		g.relay1.Low()
	}
}
