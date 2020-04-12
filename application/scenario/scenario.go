package scenario

import (
	"encoding/json"
	"fmt"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdkCommand "github.com/phanvanhai/device-sdk-go"
	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	sdk "github.com/phanvanhai/device-sdk-go/pkg/service"

	appModels "github.com/phanvanhai/device-service-support/application/models"
	"github.com/phanvanhai/device-service-support/support/common"
	"github.com/phanvanhai/device-service-support/support/db"
)

func (s *Scenario) EventCallback(async sdkModel.AsyncValues) error {
	// scenario do not have event
	return nil
}

func (s *Scenario) Initialize(dev *models.Device) error {
	err := s.updateDB(*dev)
	return err
}

func (s *Scenario) AddDeviceCallback(scenarioName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debug(fmt.Sprintf("a new Scenario is added in MetaData:%s", scenarioName))

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(scenarioName)
	if err != nil {
		return err
	}

	return s.Initialize(&dev)
}

func (s *Scenario) UpdateDeviceCallback(scenarioName string, protocols map[string]models.ProtocolProperties, adminState models.AdminState) error {
	s.lc.Debug(fmt.Sprintf("a Scenario is updated in MetaData:%s", scenarioName))

	sv := sdk.RunningService()
	dev, err := sv.GetDeviceByName(scenarioName)
	if err != nil {
		return err
	}

	return s.Initialize(&dev)
}

func (s *Scenario) RemoveDeviceCallback(scenarioName string, protocols map[string]models.ProtocolProperties) error {
	s.lc.Debug(fmt.Sprintf("a Scenario is deleted in MetaData:%s", scenarioName))

	return nil
}

func (s *Scenario) HandleReadCommands(scenarioName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	res := make([]*sdkModel.CommandValue, len(reqs))
	for i, r := range reqs {
		s.lc.Info(fmt.Sprintf("ScenarioApplication.HandleReadCommands: resource: %v, request: %v", reqs[i].DeviceResourceName, reqs[i]))

		if r.DeviceResourceName == ContentDr {
			// Lay thong tin tu Support Database va tao ket qua
			relations := db.DB().ScenarioDotElement(scenarioName)
			elementsStr, err := json.Marshal(relations)
			if err != nil {
				elementsStr = []byte(err.Error())
			}
			newCmvl := sdkModel.NewStringValue(ContentDr, 0, string(elementsStr))
			res[i] = newCmvl
		}
	}
	return res, nil
}

func (s *Scenario) HandleWriteCommands(scenarioName string, protocols map[string]models.ProtocolProperties, reqs []sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	for i, p := range params {
		s.lc.Info(fmt.Sprintf("ScenarioApplication.HandleWriteCommands: resource: %v, parameters: %v", reqs[i].DeviceResourceName, params[i]))

		switch p.DeviceResourceName {
		case TriggerDr:
			// khong quan tam gia tri Trigger

			// Lay danh sach cac element
			relations := db.DB().ScenarioDotElement(scenarioName)
			errs := make([]ElementError, len(relations))

			for i, r := range relations {
				// phan tich noi dung command-body tu relation.Content
				var ct db.ScenarioContent
				err := json.Unmarshal([]byte(r.Content), &ct)

				// gui lenh toi element va tong hop loi
				if err == nil {
					_, err = sdkCommand.CommandRequest(r.Element, ct.Command, ct.Body, appModels.SetCmdMethod, "")
				} else {
					str := fmt.Sprintf("Loi phan tich noi dung lenh cua Element:%s", r.Element)
					s.lc.Error(str)
					err = fmt.Errorf(str)
				}

				errs[i].Name = r.Element
				if err == nil {
					errs[i].Error = ""
				} else {
					errs[i].Error = err.Error()
				}
			}

			// tra ve loi neu co bat ky 1 element nao loi
			for _, e := range errs {
				if e.Error != "" {
					errStr, _ := json.Marshal(errs)
					str := fmt.Sprintf("Loi gui lenh toi cac element. Loi:%s", string(errStr))
					s.lc.Error(str)
					return fmt.Errorf(str)
				}
			}
		case ContentDr:
			sv := sdk.RunningService()
			scenario, err := sv.GetDeviceByName(scenarioName)
			if err != nil {
				return err
			}

			// phan tich noi dung yeu cau tu String -> []db.RelationContent
			value, _ := p.StringValue()

			var content []db.RelationContent
			err = json.Unmarshal([]byte(value), &content)
			if err != nil {
				str := fmt.Sprintf("Loi phan tich noi dung kich ban. Loi:%s", err.Error())
				s.lc.Error(str)
				return fmt.Errorf(str)
			}

			// cap nhap danh sach vao Database
			// luon thay Name -> ID khi luu vao Database
			pp := make(models.ProtocolProperties)
			for _, ct := range content {
				id := db.DB().NameToID(ct.Element)
				if id == "" {
					str := fmt.Sprintf("Khong to tai Element:%s", ct.Element)
					s.lc.Error(str)
					return fmt.Errorf(str)
				}
				pp[id] = ct.Content
			}
			scenario.Protocols[common.RelationProtocolNameConst] = pp
			err = sv.UpdateDevice(scenario)
			if err != nil {
				s.lc.Error(err.Error())
				return err
			}

		default:
			str := fmt.Sprintf("Khong ho tro ghi Resource:%s", p.DeviceResourceName)
			s.lc.Error(str)
			return fmt.Errorf(str)
		}
	}

	return nil
}
