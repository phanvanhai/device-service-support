package scenario

import (
	"fmt"

	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	"github.com/phanvanhai/device-service-support/support/common"
	"github.com/phanvanhai/device-service-support/support/db"
)

func (s *Scenario) updateDB(scenario models.Device) error {
	relations := db.DB().ScenarioDotElement(scenario.Name)

	needUpdate := false
	oldpp, ok := scenario.Protocols[common.RelationProtocolNameConst]
	if !ok {
		needUpdate = true
	} else {
		if len(oldpp) != len(relations) {
			needUpdate = true
		}
	}

	if needUpdate {
		str := fmt.Sprintf("Cap nhap lai Database cua Scenario:%s", scenario.Name)
		s.lc.Debug(str)
		pp := make(models.ProtocolProperties)
		for _, r := range relations {
			id := db.DB().NameToID(r.Element)
			pp[id] = r.Content
		}
		scenario.Protocols[common.RelationProtocolNameConst] = pp
		sv := sdk.RunningService()
		return sv.UpdateDevice(scenario)
	}

	return nil
}
