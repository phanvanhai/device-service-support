package cache

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"sync"

	"github.com/edgexfoundry/device-service-package/network/zigbee/common"
	"github.com/edgexfoundry/go-mod-core-contracts/models"

	sdk "github.com/edgexfoundry/device-sdk-go"
)

var zc *zigbeeCache

type ZigbeeCache interface {
	DeviceIDByNetID(netID string) (string, error)
	NetIDByDeviceID(deviceID string) (string, error)
	DeviceResourceByNetResource(deviceID string, netResource string) (string, error)
	NetResourceByDeviceResource(resourceName string) (string, error)
	UpdateDeviceCache(device models.Device)
	DeleteDeviceCache(deviceID string)
	GetType(deviceID string) string
	GenerateNetGroupID() (string, error)
}

type netResourceID struct {
	netResource string
	profileID   string
}

type zigbeeCache struct {
	edgeNetDeviceID map[string]string
	netDeviceID     map[string]string
	edgeNetResource map[string]netResourceID
	netEdgeResource map[netResourceID]string
	edgeObjectType  map[string]string
	groupAddress    []int
	mutex           sync.Mutex
}

func (zc *zigbeeCache) DeviceIDByNetID(netID string) (string, error) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	id, ok := zc.netDeviceID[netID]
	if !ok {
		return "", fmt.Errorf("Not found device ID by network ID: %s in cache", netID)
	}
	return id, nil
}

func (zc *zigbeeCache) NetIDByDeviceID(deviceID string) (string, error) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	id, ok := zc.edgeNetDeviceID[deviceID]
	if !ok {
		return "", fmt.Errorf("Not found network ID for device ID: %s in cache", deviceID)
	}
	return id, nil
}

func (zc *zigbeeCache) DeviceResourceByNetResource(deviceID string, netResource string) (string, error) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	svc := sdk.RunningService()
	d, err := svc.GetDeviceByID(deviceID)
	if err != nil {
		return "", err
	}

	netResID := netResourceID{
		netResource: netResource,
		profileID:   d.Profile.Id,
	}
	rs, ok := zc.netEdgeResource[netResID]
	if !ok {
		return "", fmt.Errorf("Not found device resource by network resource: %s and device ID: %s in cache", netResource, deviceID)
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
	pp, ok := p[common.ProtocolNameConst]
	if !ok {
		return ""
	}
	id, ok := pp[common.IDPropertyConst]
	if !ok {
		return ""
	}
	return id
}

func getTypeFromProtocols(p map[string]models.ProtocolProperties) string {
	pp, ok := p[common.ProtocolNameConst]
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
	netResource, _ := dr.Attributes[common.AttributeNetResource]
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
		zc.deleteDevice(device.Id)
		return
	}

	if objectType == common.GroupTypeConst {
		// get address from netID
		i64, _ := strconv.ParseInt(netID, 16, 32)
		addr := int(i64 & 0xFFFF)
		zc.groupAddress = append(zc.groupAddress, addr)
	}

	zc.edgeNetDeviceID[device.Id] = netID
	zc.netDeviceID[netID] = device.Id
	zc.edgeObjectType[device.Id] = objectType
	zc.updateProfile(&device.Profile)
}

func (zc *zigbeeCache) deleteGroupAddressByNetID(netID string) {
	netInt64, err := strconv.ParseUint(netID, 16, 32)
	if err == nil {
		prefixNet := netInt64 & 0xFFFF0000
		if prefixNet == common.PrefixHexValueNetGroupIDConst {
			addr := int(netInt64 & 0xFFFFF)
			for i, v := range zc.groupAddress {
				if v == addr {
					zc.groupAddress[i] = zc.groupAddress[len(zc.groupAddress)-1]
					zc.groupAddress = zc.groupAddress[:len(zc.groupAddress)-1]
				}
			}
		}
	}
}

func (zc *zigbeeCache) deleteDevice(deviceID string) {
	netDevice, ok := zc.edgeNetDeviceID[deviceID]
	if ok {
		zc.deleteGroupAddressByNetID(netDevice)
		delete(zc.edgeNetDeviceID, deviceID)
		delete(zc.netDeviceID, netDevice)
		delete(zc.edgeObjectType, deviceID)
	}
}

func (zc *zigbeeCache) DeleteDeviceCache(deviceID string) {
	zc.mutex.Lock()
	defer zc.mutex.Unlock()

	zc.deleteDevice(deviceID)
}

func (zc *zigbeeCache) GenerateNetGroupID() (string, error) {
	var addr int
	l := len(zc.groupAddress)
	if l >= math.MaxUint16 {
		return "", fmt.Errorf("Loi: Da su dung het so luong nhom cho phep")
	} else if l == 0 {
		addr = 0
	} else {
		sort.Ints(zc.groupAddress)
		if zc.groupAddress[l-1] < math.MaxUint16 {
			addr = zc.groupAddress[l-1] + 1
		} else {
			for i, v := range zc.groupAddress {
				if i != v {
					addr = i
				}
			}
		}
	}
	zc.groupAddress = append(zc.groupAddress, addr)
	addr64 := int64(common.PrefixHexValueNetGroupIDConst<<16 | addr)
	straddr := strconv.FormatInt(addr64, 16)
	return straddr, nil
}

func (zc *zigbeeCache) GetType(deviceID string) string {
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
	edgeNetDeviceID := make(map[string]string, defaultIDSize)
	netDeviceID := make(map[string]string, defaultIDSize)
	edgeNetResource := make(map[string]netResourceID, defaultResourceSize)
	netEdgeResource := make(map[netResourceID]string, defaultResourceSize)
	edgeObjectType := make(map[string]string, defaultIDSize)
	groups := make([]int, 0, defaultIDSize)

	zc = &zigbeeCache{
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
