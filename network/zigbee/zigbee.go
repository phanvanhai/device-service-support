package zigbee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/phanvanhai/device-service-support/support/common"

	"github.com/phanvanhai/device-service-support/network/zigbee/cache"
	"github.com/phanvanhai/device-service-support/network/zigbee/cm"

	"github.com/phanvanhai/device-service-support/network/zigbee/models"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	sdk "github.com/edgexfoundry/device-sdk-go/pkg/service"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/support/pubsub"
	"github.com/phanvanhai/device-service-support/transceiver"
)

var (
	requestTimeout    int64
	responseTimeout   int64
	responseAddObject int64
)

func initialize(lc logger.LoggingClient, tc transceiver.Transceiver, config map[string]string) (*Zigbee, error) {
	zb := &Zigbee{
		logger: lc,
		tc:     tc,
		config: config,
	}

	timePub := cm.EventPublishTimeDefault
	timePubStr, ok := config[cm.EventPublishTimeConfigName]
	if ok {
		v, err := strconv.ParseInt(timePubStr, 10, 64)
		if err == nil {
			timePub = v
		}
	}

	bufferSize := cm.EventBufferSizeDefault
	bufferSizeStr, ok := config[cm.EventBufferSizeConfigName]
	if ok {
		v, err := strconv.ParseInt(bufferSizeStr, 10, 32)
		if err == nil {
			bufferSize = int(v)
		}
	}

	requestTimeout = cm.RequestTimeoutDefault
	requestTimeoutStr, ok := config[cm.RequestTimeoutConfigName]
	if ok {
		v, err := strconv.ParseInt(requestTimeoutStr, 10, 64)
		if err == nil {
			requestTimeout = v
		}
	}

	responseTimeout = cm.ResponseTimoutDefault
	responseTimeoutStr, ok := config[cm.ResponseTimoutConfigName]
	if ok {
		v, err := strconv.ParseInt(responseTimeoutStr, 10, 64)
		if err == nil {
			responseTimeout = v
		}
	}

	responseAddObject = cm.ResponseAddObjectTimeoutDefault
	responseAddObjectStr, ok := config[cm.ResponseAddObjectTimeoutConfigName]
	if ok {
		v, err := strconv.ParseInt(responseAddObjectStr, 10, 64)
		if err == nil {
			responseAddObject = v
		}
	}

	zb.eventBus = pubsub.NewPublisher(time.Duration(timePub)*time.Millisecond, int(bufferSize))
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
	if _, err := cache.Cache().NetIDByDeviceName(newObject.Name); err == nil {
		return nil, nil
	}

	pp, ok := getNetworkProperties(newObject)
	if !ok {
		return nil, fmt.Errorf("Loi khong co thong tin Protocols.%s", common.GeneralProtocolNameConst)
	}

	objectType, _ := pp[common.TypePropertyConst]
	if objectType == "" {
		return nil, fmt.Errorf("Loi khong biet loai doi tuong")
	}

	switch objectType {
	case common.DeviceTypeConst:
		mac, _ := pp[cm.MACProperty]
		lk, _ := pp[cm.LinkKeyProperty]
		if mac == "" {
			return nil, fmt.Errorf("Loi khong co thong tin MAC cua thiet bi")
		}

		devPacket := models.DevicePacket{
			Header: models.Header{
				Cmd: cm.AddObjectCmdConst,
			},
			MAC:     mac,
			LinkKey: lk,
		}
		rawRequest, err := json.Marshal(&devPacket)
		if err != nil {
			return nil, err
		}

		rep, err := zb.sendRequestWithResponse(rawRequest, filterAddObject, responseAddObject)
		if err != nil {
			return nil, err
		}

		r := models.DevicePacket{}
		err = json.Unmarshal(rep.([]byte), &r)
		if err != nil {
			return nil, err
		}
		if r.StatusCode != uint8(cm.Success) {
			return nil, fmt.Errorf(r.StatusMessage)
		}

		netID := r.NetDevice
		pp[cm.NetIDProperty] = netID
		newObject.Protocols[common.GeneralProtocolNameConst] = pp

		return newObject, nil
	case common.GroupTypeConst:
		netID, err := cache.Cache().GenerateNetGroupID()
		if err != nil {
			return nil, err
		}

		pp[cm.NetIDProperty] = netID
		newObject.Protocols[common.GeneralProtocolNameConst] = pp
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
	if a.Cmd == cm.AddObjectCmdConst {
		return true
	}
	return false
}

func filterDeleteObject(v interface{}) bool {
	a := models.Header{}
	// Error with unmarshaling
	if err := json.Unmarshal(v.([]byte), &a); err != nil {
		return false
	}
	if a.Cmd == cm.DeleteObjectCmdConst {
		return true
	}
	return false
}

func getNetworkProperties(d *contract.Device) (contract.ProtocolProperties, bool) {
	pp, ok := d.Protocols[common.GeneralProtocolNameConst]
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
		pp, ok := protocols[common.GeneralProtocolNameConst]
		if !ok {
			return fmt.Errorf("Loi khong co thong tin Protocols.%s", common.GeneralProtocolNameConst)
		}
		mac, _ := pp[cm.MACProperty]
		if mac == "" {
			return fmt.Errorf("Loi khong co thong tin MAC cua thiet bi")
		}

		devPacket := models.DevicePacket{
			Header: models.Header{
				Cmd: cm.DeleteObjectCmdConst,
			},
			NetDevice: netID,
			MAC:       mac,
		}
		rawRequest, err := json.Marshal(&devPacket)
		if err != nil {
			return err
		}

		rep, err := zb.sendRequestWithResponse(rawRequest, filterDeleteObject, responseTimeout)
		if err != nil {
			return err
		}

		r := models.DevicePacket{}
		err = json.Unmarshal(rep.([]byte), &r)
		if err != nil {
			return err
		}
		if r.StatusCode != uint8(cm.Success) {
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
	if a.Cmd == cm.ReportConst {
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

func (zb *Zigbee) sendRequestWithResponse(rawRequest []byte, responseFilter func(v interface{}) bool, responseTimeout int64) (rep interface{}, err error) {
	reper := zb.tc.Listen(responseFilter)
	defer zb.tc.CancelListen(reper)

	zb.logger.Debug(fmt.Sprintln("Send request:", string(rawRequest)))
	err = zb.tc.Sender(rawRequest, requestTimeout)
	if err != nil {
		return nil, err
	}

	timeOut := time.After(time.Duration(responseTimeout) * time.Millisecond)
	select {
	case <-timeOut:
		err = fmt.Errorf("Loi: Timeout")
	case rep = <-reper:
	}
	return
}

func (zb *Zigbee) sendRequestWithoutResponse(rawRequest []byte) error {
	err := zb.tc.Sender(rawRequest, requestTimeout)
	return err
}

// UpdateFirmware implement me
func (zb *Zigbee) UpdateFirmware(deviceName string, file interface{}) error {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	cmd := exec.Command("../commander/commander", "flash", file.(string), "--address", "0x80000")
	var out bytes.Buffer
	multi := io.MultiWriter(os.Stdout, &out)
	cmd.Stdout = multi
	cmd.Stderr = multi

	if err := cmd.Run(); err != nil {
		zb.logger.Error(err.Error())
		return err
	}

	zb.logger.Info(out.String())
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

func (zb *Zigbee) filterCommand(devName string, cmd uint8) func(v interface{}) bool {
	return func(v interface{}) bool {
		type Alias struct {
			NetDevice string `json:"dev"`
			Cmd       uint8  `json:"cmd"`
		}
		a := Alias{}

		if err := json.Unmarshal(v.([]byte), &a); err != nil {
			return false
		}
		if a.Cmd != cmd {
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
		str := fmt.Sprintf("Khong tim thay NetResourece theo Resource:%s - Device:%s", fromRs, fromDevName)
		zb.logger.Error(str)
		return ""
	}

	toRs, err := cache.Cache().DeviceResourceByNetResource(toDevName, netRs)
	if err != nil {
		str := fmt.Sprintf("Khong tim thay Resourece theo NetResource:%s - Device:%s", netRs, toDevName)
		zb.logger.Error(str)
		return ""
	}

	svc := sdk.RunningService()
	_, ok := svc.DeviceResource(toDevName, toRs, "")

	if !ok {
		str := fmt.Sprintf("Khong tim thay Resource tu Cache DS")
		zb.logger.Error(str)
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

	rq := models.CommandPacket{
		Header: models.Header{
			Cmd: cm.GetCommandCmdConst,
		},
		NetEvent: *netEvent,
	}
	rawRequest, err := json.Marshal(rq)
	if err != nil {
		return nil, err
	}

	t := cache.Cache().GetType(name)
	switch t {
	case common.GroupTypeConst:
		filter := zb.filterCommand(name, cm.GetCommandCmdConst)
		rep, err := zb.sendRequestWithResponse(rawRequest, filter, responseTimeout)
		if err != nil {
			return nil, err
		}

		r := models.Response{}
		err = json.Unmarshal(rep.([]byte), &r)
		if err != nil {
			return nil, err
		}
		if r.StatusCode != uint8(cm.Success) {
			return nil, fmt.Errorf(r.StatusMessage)
		}
	case common.DeviceTypeConst:
		filter := zb.filterCommand(name, cm.GetCommandCmdConst)
		rep, err := zb.sendRequestWithResponse(rawRequest, filter, responseTimeout)
		if err != nil {
			return nil, err
		}

		rp := models.Response{}
		err = json.Unmarshal(rep.([]byte), &rp)
		if err != nil {
			return nil, err
		}
		if rp.StatusCode != uint8(cm.Success) {
			return nil, fmt.Errorf(rp.StatusMessage)
		}

		r := models.CommandPacket{}
		err = json.Unmarshal(rep.([]byte), &r)
		if err != nil {
			return nil, err
		}

		async, err := r.NetEvent.ToEdgeEvent()
		if err != nil {
			return nil, err
		}

		return async.CommandValues, nil
	default:
		return nil, fmt.Errorf("Khong ho tro doi tuong khong phai loai Device hoac Group")
	}

	return nil, nil
}

// WriteCommands
func (zb *Zigbee) WriteCommands(name string, reqs []*sdkModel.CommandRequest, params []*sdkModel.CommandValue) error {
	zb.mutex.Lock()
	defer zb.mutex.Unlock()

	netEvent, err := models.CommandValueToNetEvent(name, params)
	if err != nil {
		return err
	}

	rq := models.CommandPacket{
		Header: models.Header{
			Cmd: cm.PutCommandCmdConst,
		},
		NetEvent: *netEvent,
	}
	rawRequest, err := json.Marshal(rq)
	if err != nil {
		return err
	}

	filter := zb.filterCommand(name, cm.PutCommandCmdConst)
	rep, err := zb.sendRequestWithResponse(rawRequest, filter, responseTimeout)
	if err != nil {
		return err
	}

	r := models.Response{}
	err = json.Unmarshal(rep.([]byte), &r)
	if err != nil {
		return err
	}
	zb.logger.Debug(fmt.Sprintln(r))
	if r.StatusCode != uint8(cm.Success) {
		zb.logger.Error(fmt.Sprintf("Loi status code:%d--%s", r.StatusCode, r.StatusMessage))
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
