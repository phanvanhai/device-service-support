package zigbee

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/edgexfoundry/device-service-package/network/zigbee/cache"
	"github.com/edgexfoundry/device-service-package/network/zigbee/common"
	"github.com/edgexfoundry/device-service-package/network/zigbee/models"

	sdk "github.com/edgexfoundry/device-sdk-go"
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/device-service-package/support/pubsub"
	"github.com/edgexfoundry/device-service-package/transceiver"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
)

func initialize(lc logger.LoggingClient, tc transceiver.Transceiver, config map[string]string) (*Zigbee, error) {
	zb := &Zigbee{
		logger: lc,
		tc:     tc,
		config: config,
	}
	zb.eventBus = pubsub.NewPublisher(common.TIMEPUB*time.Second, common.CHANSIZEPUB)
	cache.Cache()
	go zb.distributionEventRoutine()
	return zb, nil
}

func (zb *Zigbee) Close() error {
	zb.eventBus.Close()
	return nil
}

func (zb *Zigbee) UpdateObjectCallback(object *contract.Device) {
	cache.Cache().UpdateDeviceCache(object)
}

func (zb *Zigbee) DeleteObjectCallback(objectID string) {
	cache.Cache().DeleteDeviceCache(objectID)
}

func (zb *Zigbee) AddObject(newObject *contract.Device) (*contract.Device, error) {
	if _, err := cache.Cache().NetIDByDeviceID(newObject.Id); err != nil {
		return nil, nil
	}

	pp, ok := getNetworkProperties(newObject)
	if !ok {
		return nil, fmt.Errorf("Loi khong co thong tin Protocols.%s", common.ProtocolNameConst)
	}

	objectType, _ := pp[common.TypePropertyConst]
	if objectType == "" {
		return nil, fmt.Errorf("Loi khong biet loai doi tuong")
	}

	switch objectType {
	case common.DeviceTypeConst:
		mac, _ := pp[common.MACPropertyConst]
		lk, _ := pp[common.LinkKeyPropertyConst]
		if mac == "" {
			return nil, fmt.Errorf("Loi khong co thong tin MAC cua thiet bi")
		}

		devPacket := models.DevicePacket{
			Header: models.Header{
				Cmd: common.AddObjectCmdConst,
			},
			MAC:     mac,
			LinkKey: lk,
		}
		rawRequest, err := json.Marshal(&devPacket)
		if err != nil {
			return nil, err
		}

		rep, err := zb.sendRequestWithResponse(rawRequest, filterAddObject)
		r := models.DevicePacket{}
		err = json.Unmarshal(rep.([]byte), &r)
		if err != nil {
			return nil, err
		}
		if r.StatusCode != uint8(common.Success) {
			return nil, fmt.Errorf(r.StatusMessage)
		}

		netID := r.NetDevice
		pp[common.IDPropertyConst] = netID
		newObject.Protocols[common.ProtocolNameConst] = pp

		return newObject, nil
	case common.GroupTypeConst:
		netID, err := cache.Cache().GenerateNetGroupID()
		if err != nil {
			return nil, err
		}

		pp[common.IDPropertyConst] = netID
		newObject.Protocols[common.ProtocolNameConst] = pp
		return newObject, nil
	}

	return nil, nil
}

func filterAddObject(v interface{}) bool {
	a := models.Header{}
	// Error with unmarshaling
	if err := json.Unmarshal(v.([]byte), &a); err != nil {
		return false
	}
	if a.Cmd == common.AddObjectCmdConst {
		return true
	}
	return false
}

func getNetworkProperties(d *contract.Device) (contract.ProtocolProperties, bool) {
	pp, ok := d.Protocols[common.ProtocolNameConst]
	return pp, ok
}

// UpdateObject hien tai khong su dung
func (zb *Zigbee) UpdateObject(newObject *contract.Device) error {
	return nil
}

// DeleteObject
func (zb *Zigbee) DeleteObject(objectID string) error {
	netID, err := cache.Cache().NetIDByDeviceID(objectID)
	// err != nil -> Object not exist
	if err != nil {
		return nil
	}

	// loai Group khong can xoa, vi thong tin can xoa la groupAddress
	// se duoc xoa trong DeleteObjectCallback
	if cache.Cache().GetType(objectID) == common.DeviceTypeConst {
		devPacket := models.DevicePacket{
			Header: models.Header{
				Cmd: common.DeleteObjectCmdConst,
			},
			NetDevice: netID,
		}
		rawRequest, err := json.Marshal(&devPacket)
		if err != nil {
			return err
		}

		rep, err := zb.sendRequestWithResponse(rawRequest, filterAddObject)
		r := models.DevicePacket{}
		err = json.Unmarshal(rep.([]byte), &r)
		if err != nil {
			return err
		}
		if r.StatusCode != uint8(common.Success) {
			return fmt.Errorf(r.StatusMessage)
		}

		return nil
	}

	return nil
}

func filterDistributionEvent(v interface{}) bool {
	a := models.Header{}
	// Error with unmarshaling
	if err := json.Unmarshal(v.([]byte), &a); err != nil {
		return false
	}
	if a.Cmd == common.ReportConst {
		return true
	}
	return false
}

func (zb *Zigbee) distributionEventRoutine() {
	rawEvents := zb.tc.Listen(filterDistributionEvent)
	for payload := range rawEvents {
		zb.eventBus.Publish(payload)
	}
}

func (zb *Zigbee) sendRequestWithResponse(rawRequest []byte, responseFilter func(v interface{}) bool) (rep interface{}, err error) {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	err = zb.tc.Sender(rawRequest, common.SendRequestTimeoutConst)
	if err != nil {
		return nil, err
	}

	reper := zb.tc.Listen(responseFilter)
	defer zb.tc.CancelListen(reper)

	timeOut := time.After(time.Duration(common.ReceiverResponseTimeoutConst) * time.Millisecond)
	select {
	case <-timeOut:
		err = fmt.Errorf("Loi: Timeout")
	case rep = <-reper:
	}
	return
}

func (zb *Zigbee) sendRequestWithoutResponse(rawRequest []byte) error {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	err := zb.tc.Sender(rawRequest, common.SendRequestTimeoutConst)
	return err
}

// UpdateFirmware implement me
func (zb *Zigbee) UpdateFirmware(deviceID string, file interface{}) error {
	return nil
}

// Discovery implement me
func (zb *Zigbee) Discovery() (devices *interface{}, err error) {
	return nil, nil
}

func (zb *Zigbee) ListenEvent() chan interface{} {
	return zb.eventBus.Subscribe()
}

// DeviceIDByNetID
func (zb *Zigbee) DeviceIDByNetID(netID string) string {
	id, err := cache.Cache().DeviceIDByNetID(netID)
	if err != nil {
		return ""
	}
	return id
}

// NetIDByDeviceID
func (zb *Zigbee) NetIDByDeviceID(devID string) string {
	id, err := cache.Cache().NetIDByDeviceID(devID)
	if err != nil {
		return ""
	}
	return id
}

func (zb *Zigbee) filterCommand(id string) func(v interface{}) bool {
	return func(v interface{}) bool {
		type Alias struct {
			NetDevice string `json:"dev"`
			Cmd       uint8  `json:"cmd"`
		}
		a := Alias{}

		if err := json.Unmarshal(v.([]byte), &a); err != nil {
			return false
		}
		if a.Cmd != common.CommandCmdConst {
			return false
		}

		if zb.DeviceIDByNetID(a.NetDevice) != id {
			return false
		}
		return true
	}
}

// ConvertResourceByDevice
func (zb *Zigbee) ConvertResourceByDevice(fromDevID string, fromRs string, toDevID string) string {
	netRs, err := cache.Cache().NetResourceByDeviceResource(fromRs)
	if err != nil {
		return ""
	}

	toRs, err := cache.Cache().DeviceResourceByNetResource(toDevID, netRs)
	if err != nil {
		return ""
	}

	svc := sdk.RunningService()
	rs1, ok1 := svc.DeviceResource(fromDevID, fromRs, "")
	rs2, ok2 := svc.DeviceResource(toDevID, toRs, "")

	if !ok1 || !ok2 {
		return ""
	}

	ok := reflect.DeepEqual(rs1.Properties, rs2.Properties)
	if !ok {
		return ""
	}

	return toRs
}

// ReadCommands
func (zb *Zigbee) ReadCommands(objectID string, reqs []*sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	netEvent, err := models.CommandRequestToNetEvent(objectID, reqs)
	if err != nil {
		return nil, err
	}

	rawRequest, err := json.Marshal(netEvent)
	if err != nil {
		return nil, err
	}

	filter := zb.filterCommand(objectID)
	rep, err := zb.sendRequestWithResponse(rawRequest, filter)
	r := models.CommandPacket{}
	err = json.Unmarshal(rep.([]byte), &r)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != uint8(common.Success) {
		return nil, fmt.Errorf(r.StatusMessage)
	}

	_, cmvls, err := r.NetEvent.ToEdgeEvent()
	return cmvls, err
}

// WriteCommands
func (zb *Zigbee) WriteCommands(objectID string, reqs []*sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	netEvent, err := models.CommandValueToNetEvent(objectID, params)
	if err != nil {
		return err
	}

	rawRequest, err := json.Marshal(netEvent)
	if err != nil {
		return err
	}

	filter := zb.filterCommand(objectID)
	rep, err := zb.sendRequestWithResponse(rawRequest, filter)
	r := models.CommandPacket{}
	err = json.Unmarshal(rep.([]byte), &r)
	if err != nil {
		return err
	}
	if r.StatusCode != uint8(common.Success) {
		return fmt.Errorf(r.StatusMessage)
	}

	return nil
}

func (zb *Zigbee) CheckExist(devID string) bool {
	_, err := cache.Cache().NetIDByDeviceID(devID)
	if err != nil {
		return false
	}
	return true
}
