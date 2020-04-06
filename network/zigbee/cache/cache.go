package cache

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/network/zigbee/cm"
	"github.com/phanvanhai/device-service-support/support/common"

	sdk "github.com/phanvanhai/device-sdk-go"
)

var zc *zigbeeCache

type ZigbeeCache interface {
	DeviceNameByID(id string) string
	DeviceIDByName(name string) string
	DeviceNameByNetID(netID string) (string, error)
	NetIDByDeviceName(name string) (string, error)
	DeviceResourceByNetResource(name string, netResource string) (string, error)
	NetResourceByDeviceResource(resourceName string) (string, error)
	UpdateDeviceCache(device models.Device)
	DeleteDeviceCache(name string)
	GetType(name string) string
	GenerateNetGroupID() (string, error)
}

type netResourceID struct {
	netResource string
	profileID   string
}

type zigbeeCache struct {
	edgeNetDeviceID map[string]string
	netDeviceID     map[string]string
	deviceIDName    map[string]string
	deviceNameID    map[string]string
	edgeNetResource map[string]netResourceID
	netEdgeResource map[netResourceID]string
	edgeObjectType  map[string]string
	groupAddress    []int
	mutex           sync.Mutex
}

func (zc *zigbeeCache) DeviceNameByID(id string) string {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	name, _ := zc.deviceIDName[id]
	return name
}

func (zc *zigbeeCache) DeviceIDByName(name string) string {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	id, _ := zc.deviceNameID[name]
	return id
}

func (zc *zigbeeCache) DeviceNameByNetID(netID string) (string, error) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	id, ok := zc.netDeviceID[netID]
	if !ok {
		return "", fmt.Errorf("Not found device ID by network ID: %s in cache", netID)
	}
	name, _ := zc.deviceIDName[id]
	return name, nil
}

func (zc *zigbeeCache) NetIDByDeviceName(name string) (string, error) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	id, ok := zc.deviceNameID[name]
	if !ok {
		return "", fmt.Errorf("Not found device name: %s in cache", name)
	}

	netID, ok := zc.edgeNetDeviceID[id]
	if !ok {
		return "", fmt.Errorf("Not found network ID for device name: %s in cache", name)
	}
	return netID, nil
}

func (zc *zigbeeCache) DeviceResourceByNetResource(name string, netResource string) (string, error) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	svc := sdk.RunningService()
	d, err := svc.GetDeviceByName(name)
	if err != nil {
		return "", err
	}

	netResID := netResourceID{
		netResource: netResource,
		profileID:   d.Profile.Id,
	}

	rs, ok := zc.netEdgeResource[netResID]
	if !ok {
		return "", fmt.Errorf("Not found device resource by network resource: %s and device name: %s in cache", netResource, name)
	}
	return rs, nil
}

func (zc *zigbeeCache) NetResourceByDeviceResource(resourceName string) (string, error) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	netRes, ok := zc.edgeNetResource[resourceName]
	if !ok {
		return "", fmt.Errorf("Not found network resource for device resource: %s in cache", resourceName)
	}

	return netRes.netResource, nil
}

func getNetIDFromProtocols(p map[string]models.ProtocolProperties) string {
	pp, ok := p[common.GeneralProtocolNameConst]
	if !ok {
		return ""
	}
	id, ok := pp[cm.NetIDProperty]
	if !ok {
		return ""
	}
	return id
}

func getTypeFromProtocols(p map[string]models.ProtocolProperties) string {
	pp, ok := p[common.GeneralProtocolNameConst]
	if !ok {
		return ""
	}
	t, ok := pp[common.TypePropertyConst]
	if !ok {
		return ""
	}
	return t
}

func getNetResourceFromDeviceResource(dr *models.DeviceResource) string {
	netResource, _ := dr.Attributes[cm.AttributeNetResource]
	return netResource
}

func (zc *zigbeeCache) updateProfile(p *models.DeviceProfile) {
	for _, rs := range p.DeviceResources {
		netResource := getNetResourceFromDeviceResource(&rs)
		if netResource != "" {
			netResID := netResourceID{
				netResource: netResource,
				profileID:   p.Id,
			}
			zc.edgeNetResource[rs.Name] = netResID
			zc.netEdgeResource[netResID] = rs.Name
		}
	}
}

func (zc *zigbeeCache) UpdateDeviceCache(device *models.Device) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	netID := getNetIDFromProtocols(device.Protocols)
	objectType := getTypeFromProtocols(device.Protocols)
	if netID == "" {
		// loai bo device
		zc.deleteDevice(device.Name)
		return
	}

	if objectType == common.GroupTypeConst {
		// get address from netID
		i64, _ := strconv.ParseInt(netID, 16, 32)
		addr := int(i64 & 0xFFFF)
		zc.groupAddress = append(zc.groupAddress, addr)
	}

	oldName := zc.deviceIDName[device.Id]
	if oldName != "" {
		delete(zc.deviceNameID, oldName)
	}

	zc.deviceNameID[device.Name] = device.Id
	zc.deviceIDName[device.Id] = device.Name
	zc.edgeNetDeviceID[device.Id] = netID
	zc.netDeviceID[netID] = device.Id
	zc.edgeObjectType[device.Id] = objectType
	zc.updateProfile(&device.Profile)
}

func (zc *zigbeeCache) deleteGroupAddressByNetID(netID string) {
	netInt64, err := strconv.ParseUint(netID, 16, 32)
	if err == nil {
		addr := int(netInt64 & 0xFFFF)
		for i, v := range zc.groupAddress {
			if v == addr {
				zc.groupAddress[i] = zc.groupAddress[len(zc.groupAddress)-1]
				zc.groupAddress = zc.groupAddress[:len(zc.groupAddress)-1]
			}
		}
	}
}

func (zc *zigbeeCache) deleteDevice(name string) {
	deviceID, ok := zc.deviceNameID[name]
	if !ok {
		return
	}

	netDevice, ok := zc.edgeNetDeviceID[deviceID]
	if ok {
		t, _ := zc.edgeObjectType[deviceID]
		if t == common.GroupTypeConst {
			zc.deleteGroupAddressByNetID(netDevice)
		}

		delete(zc.deviceNameID, name)
		delete(zc.deviceIDName, deviceID)
		delete(zc.edgeNetDeviceID, deviceID)
		delete(zc.netDeviceID, netDevice)
		delete(zc.edgeObjectType, deviceID)
	}
}

func (zc *zigbeeCache) DeleteDeviceCache(name string) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	zc.deleteDevice(name)
}

func (zc *zigbeeCache) GenerateNetGroupID() (string, error) {
	// Don't use group adrress = 0
	var addr int
	l := len(zc.groupAddress)
	if l >= (math.MaxUint16 - 1) {
		return "", fmt.Errorf("Loi: Da su dung het so luong nhom cho phep")
	} else if l == 0 {
		addr = 1
	} else {
		sort.Ints(zc.groupAddress)
		if zc.groupAddress[l-1] < math.MaxUint16 {
			addr = zc.groupAddress[l-1] + 1
		} else {
			for i, v := range zc.groupAddress {
				if (i + 1) != v {
					addr = i + 1
				}
			}
		}
	}

	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	zc.groupAddress = append(zc.groupAddress, addr)
	addr32 := int32(cm.PrefixHexValueNetGroupID<<16 | addr)
	// straddr := strconv.FormatInt(addr32, 16)
	straddr := fmt.Sprintf("%08X", addr32)
	return straddr, nil
}

func (zc *zigbeeCache) GetType(name string) string {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	deviceID, ok := zc.deviceNameID[name]
	if !ok {
		return ""
	}
	t, _ := zc.edgeObjectType[deviceID]
	return t
}

// InitCache basic state for cache
func initCache() {
	svc := sdk.RunningService()
	ds := svc.Devices()
	prs := svc.DeviceProfiles()

	defaultIDSize := len(ds) * 2
	defaultResourceSize := len(prs) * 3
	idName := make(map[string]string, defaultIDSize)
	nameID := make(map[string]string, defaultIDSize)
	edgeNetDeviceID := make(map[string]string, defaultIDSize)
	netDeviceID := make(map[string]string, defaultIDSize)
	edgeNetResource := make(map[string]netResourceID, defaultResourceSize)
	netEdgeResource := make(map[netResourceID]string, defaultResourceSize)
	edgeObjectType := make(map[string]string, defaultIDSize)
	groups := make([]int, 0, defaultIDSize)

	zc = &zigbeeCache{
		deviceNameID:    nameID,
		deviceIDName:    idName,
		edgeNetDeviceID: edgeNetDeviceID,
		edgeNetResource: edgeNetResource,
		netDeviceID:     netDeviceID,
		netEdgeResource: netEdgeResource,
		edgeObjectType:  edgeObjectType,
		groupAddress:    groups,
	}

	for _, d := range ds {
		zc.UpdateDeviceCache(&d)
	}
}

func Cache() *zigbeeCache {
	if zc == nil {
		initCache()
	}
	return zc
}
