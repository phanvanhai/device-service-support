package zigbee

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/phanvanhai/device-service-support/network/zigbee/cache"
	"github.com/phanvanhai/device-service-support/network/zigbee/common"
	"github.com/phanvanhai/device-service-support/network/zigbee/models"

	sdk "github.com/edgexfoundry/device-sdk-go"
	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/support/pubsub"
	"github.com/phanvanhai/device-service-support/transceiver"
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
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	cache.Cache().UpdateDeviceCache(object)
}

func (zb *Zigbee) DeleteObjectCallback(name string) {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	cache.Cache().DeleteDeviceCache(name)
}

func (zb *Zigbee) AddObject(newObject *contract.Device) (*contract.Device, error) {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	// neu Object da co trong Cache (da duoc cap phep) --> thoat
	if _, err := cache.Cache().NetIDByDeviceName(newObject.Name); err != nil {
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
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	return nil
}

// DeleteObject
func (zb *Zigbee) DeleteObject(name string, protocols map[string]contract.ProtocolProperties) error {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	netID, err := cache.Cache().NetIDByDeviceName(name)
	// err != nil -> Object chua duoc cap phep --> khong can xoa
	if err != nil {
		return nil
	}

	// loai Group khong can xoa, vi thong tin can xoa la groupAddress
	// se duoc xoa trong DeleteObjectCallback
	if cache.Cache().GetType(name) == common.DeviceTypeConst {
		pp, ok := protocols[common.ProtocolNameConst]
		if !ok {
			return fmt.Errorf("Loi khong co thong tin Protocols.%s", common.ProtocolNameConst)
		}
		mac, _ := pp[common.MACPropertyConst]
		if mac == "" {
			return fmt.Errorf("Loi khong co thong tin MAC cua thiet bi")
		}

		devPacket := models.DevicePacket{
			Header: models.Header{
				Cmd: common.DeleteObjectCmdConst,
			},
			NetDevice: netID,
			MAC:       mac,
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
		evt := models.NetEvent{}
		if err := json.Unmarshal(payload.([]byte), &evt); err != nil {
			continue
		}

		zb.mutex.Lock()
		async, err := evt.ToEdgeEvent()
		zb.mutex.Unlock()

		if err != nil {
			continue
		}

		zb.eventBus.Publish(async)
	}
}

func (zb *Zigbee) sendRequestWithResponse(rawRequest []byte, responseFilter func(v interface{}) bool) (rep interface{}, err error) {
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
	err := zb.tc.Sender(rawRequest, common.SendRequestTimeoutConst)
	return err
}

// UpdateFirmware implement me
func (zb *Zigbee) UpdateFirmware(deviceName string, file interface{}) error {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	return nil
}

// Discovery implement me
func (zb *Zigbee) Discovery() (devices *interface{}, err error) {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	return nil, nil
}

func (zb *Zigbee) ListenEvent() chan interface{} {
	return zb.eventBus.Subscribe()
}

// DeviceNameByNetID
func (zb *Zigbee) DeviceNameByNetID(netID string) string {
	name, err := cache.Cache().DeviceNameByNetID(netID)
	if err != nil {
		return ""
	}
	return name
}

// NetIDByDeviceName
func (zb *Zigbee) NetIDByDeviceName(name string) string {
	id, err := cache.Cache().NetIDByDeviceName(name)
	if err != nil {
		return ""
	}
	return id
}

func (zb *Zigbee) filterCommand(devName string) func(v interface{}) bool {
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

		if zb.DeviceNameByNetID(a.NetDevice) != devName {
			return false
		}
		return true
	}
}

// ConvertResourceByDevice
func (zb *Zigbee) ConvertResourceByDevice(fromDevName string, fromRs string, toDevName string) string {
	netRs, err := cache.Cache().NetResourceByDeviceResource(fromRs)
	if err != nil {
		return ""
	}

	toRs, err := cache.Cache().DeviceResourceByNetResource(toDevName, netRs)
	if err != nil {
		return ""
	}

	svc := sdk.RunningService()
	rs1, ok1 := svc.DeviceResource(fromDevName, fromRs, "")
	rs2, ok2 := svc.DeviceResource(fromDevName, toRs, "")

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
func (zb *Zigbee) ReadCommands(name string, reqs []*sdkModel.CommandRequest) ([]*sdkModel.CommandValue, error) {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	netEvent, err := models.CommandRequestToNetEvent(name, reqs)
	if err != nil {
		return nil, err
	}

	rawRequest, err := json.Marshal(netEvent)
	if err != nil {
		return nil, err
	}

	filter := zb.filterCommand(name)
	rep, err := zb.sendRequestWithResponse(rawRequest, filter)
	r := models.CommandPacket{}
	err = json.Unmarshal(rep.([]byte), &r)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != uint8(common.Success) {
		return nil, fmt.Errorf(r.StatusMessage)
	}

	async, err := r.NetEvent.ToEdgeEvent()
	if err != nil {
		return nil, err
	}

	return async.CommandValues, nil
}

// WriteCommands
func (zb *Zigbee) WriteCommands(name string, reqs []*sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	netEvent, err := models.CommandValueToNetEvent(name, params)
	if err != nil {
		return err
	}

	rawRequest, err := json.Marshal(netEvent)
	if err != nil {
		return err
	}

	filter := zb.filterCommand(name)
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

func (zb *Zigbee) CheckExist(name string) bool {
	_, err := cache.Cache().NetIDByDeviceName(name)
	if err != nil {
		return false
	}
	return true
}
