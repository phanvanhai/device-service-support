package gateway

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
)

func (g *Gateway) updateFirmware(firmwareStr string) error {
	fileName := "firmware.bin"
	decode, err := base64.StdEncoding.DecodeString(firmwareStr)
	if err != nil {
		str := fmt.Sprintf("Loi gia ma firmware. Loi:%s", err.Error())
		g.lc.Error(str)
		return fmt.Errorf(str)
	}
	// Chmod 644 (chmod a+rwx,u-x,g-wx,o-wx) sets permissions so that:
	//  (U)ser / owner can read, can write and can't execute.
	//  (G)roup can read, can't write and can't execute.
	//  (O)thers can read, can't write and can't execute.
	err = ioutil.WriteFile(fileName, decode, 0644)
	if err != nil {
		str := fmt.Sprintf("Loi luu file firmware. Loi:%s", err.Error())
		g.lc.Error(str)
		return fmt.Errorf(str)
	}

	return g.nw.UpdateFirmware("", fileName)
}
