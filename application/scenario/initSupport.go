package scenario

import (
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"

	"github.com/phanvanhai/device-service-support/support/common"
	"github.com/phanvanhai/device-service-support/support/db"
)

func (s *Scenario) updateDB(scenario models.Device) error {
	relations := db.DB().ScenarioDotElement(scenario.Name)
	needUpdate := false

	j := 0
	for _, r := range relations {
		if db.DB().NameToID(r.Element) != "" {
			relations[j] = r
			j++
		} else {
			str := fmt.Sprintf("Can loai bo thong tin Element:%s trong Database", r.Element)
			s.lc.Debug(str)
			needUpdate = true
		}
	}
	relations = relations[:j]

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
