package db

import (
	"sync"

	"github.com/phanvanhai/device-service-support/support/common"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	sdk "github.com/phanvanhai/device-sdk-go"
)

type Relation struct {
	Parent  string `json:"Parent, omitempty"`
	Element string `json:"Element, omitempty"`
}

type RelationContent struct {
	Relation
	Content string `json:"Content, omitempty"`
}
type ObjectInfo struct {
	Name        string
	Connected   bool
	Type        string
	ProfileName string
	Protocols   map[string]models.ProtocolProperties
}

type ObjectDB interface {
	UpdateObject(object *models.Device)
	DeleteDevice(name string)

	GetProfileName(name string) string
	GetObjectType(name string) string
	GetConnectedStatus(name string) bool
	SetConnectedStatus(name string, status bool)
	GetProtocols(name string) map[string]models.ProtocolProperties
	UpdateProtocols(name string, protocols map[string]models.ProtocolProperties)
	GetProperty(name string, key string) models.ProtocolProperties
	UpdateProperty(name string, key string, value models.ProtocolProperties)
	DeleteProperty(name string, key string)

	IDToName(id string) string
	NameToID(name string) string
	GroupDotElement(name string) []RelationContent
	ScenarioDotElement(name string) []RelationContent
	ElementDotGroups(name string) []RelationContent
	ElementDotScenario(name string) []RelationContent
}
type database struct {
	objectIDName     map[string]ObjectInfo
	objectNameID     map[string]string
	relationScenario map[Relation]string
	relationGroup    map[Relation]string
	mutex            sync.Mutex
}

var db *database

func initDB() {
	svc := sdk.RunningService()
	ds := svc.Devices()

	defaultIDSize := len(ds) * 2

	objectIDName := make(map[string]ObjectInfo, defaultIDSize)
	objectNameID := make(map[string]string, defaultIDSize)
	relationScenario := make(map[Relation]string, defaultIDSize)
	relationGroup := make(map[Relation]string, defaultIDSize)

	db = &database{
		objectIDName:     objectIDName,
		objectNameID:     objectNameID,
		relationGroup:    relationGroup,
		relationScenario: relationScenario,
	}

	for _, d := range ds {
		db.UpdateObject(&d)
	}
}

func DB() *database {
	if db == nil {
		initDB()
	}
	return db
}

func (db *database) UpdateObject(object *models.Device) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	protocolGeneral, ok := object.Protocols[common.GeneralProtocolNameConst]
	if !ok {
		return
	}

	objectType, ok := protocolGeneral[common.TypePropertyConst]
	if !ok {
		return
	}

	mapRelation, relationExist := object.Protocols[common.RelationProtocolNameConst]

	id := object.Id
	name := object.Name
	connected := false

	var protocols map[string]models.ProtocolProperties
	oldInfo, ok := db.objectIDName[id]
	if ok {
		connected = oldInfo.Connected
		oldName := oldInfo.Name
		protocols = oldInfo.Protocols
		if oldName != name {
			delete(db.objectNameID, oldName)
		}
	} else {
		protocols = make(map[string]models.ProtocolProperties)
	}

	db.objectIDName[id] = ObjectInfo{
		Name:        name,
		Connected:   connected,
		Type:        objectType,
		ProfileName: object.Profile.Name,
		Protocols:   protocols,
	}

	db.objectNameID[name] = id

	if relationExist {

		mr := db.getRelationMapByType(objectType)
		if mr != nil {
			deleteAllRelationByParent(mr, id)
			for elementID, content := range mapRelation {
				rl := Relation{
					Parent:  id,
					Element: elementID,
				}
				mr[rl] = content
			}
		}
	}
}

func (db *database) DeleteDevice(name string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return
	}

	info := db.objectIDName[id]
	p := info.Protocols
	for k := range p {
		delete(p, k)
	}
	delete(db.objectIDName, id)

	delete(db.objectNameID, name)
	db.deleteAllRelationByID(id)
}

func (db *database) GetProfileName(name string) string {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return ""
	}

	info, _ := db.objectIDName[id]
	return info.ProfileName
}

func (db *database) GetObjectType(name string) string {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.getObjectType(name)
}

func (db *database) GetConnectedStatus(name string) bool {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return false
	}

	info, _ := db.objectIDName[id]
	return info.Connected
}

func (db *database) SetConnectedStatus(name string, status bool) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return
	}

	info, _ := db.objectIDName[id]
	info.Connected = status
	db.objectIDName[id] = info
}

func (db *database) GetProtocols(name string) map[string]models.ProtocolProperties {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return nil
	}

	info, _ := db.objectIDName[id]
	return info.Protocols
}

func (db *database) UpdateProtocols(name string, protocols map[string]models.ProtocolProperties) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return
	}

	info, _ := db.objectIDName[id]
	p := info.Protocols
	for k := range p {
		delete(p, k)
	}
	info.Protocols = protocols
	db.objectIDName[id] = info
}

func (db *database) GetProperty(name string, key string) (models.ProtocolProperties, bool) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return nil, false
	}

	info, _ := db.objectIDName[id]
	pp, ok := info.Protocols[key]
	return pp, ok
}

func (db *database) UpdateProperty(name string, key string, value models.ProtocolProperties) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return
	}

	info, _ := db.objectIDName[id]
	info.Protocols[key] = value
}

func (db *database) DeleteProperty(name string, key string) {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	id, existID := db.objectNameID[name]
	if !existID {
		return
	}

	info, _ := db.objectIDName[id]
	delete(info.Protocols, key)
}

func (db *database) IDToName(id string) string {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.getNameByID(id)
}

func (db *database) NameToID(name string) string {
	db.mutex.Lock()
	defer db.mutex.Unlock()

	return db.getNameByID(name)
}

func (db *database) GroupDotElement(name string) []RelationContent {
	if db.getObjectType(name) != common.GroupTypeConst {
		return nil
	}

	id := db.getIDByName(name)
	var result []RelationContent
	for r, ct := range db.relationGroup {
		if r.Parent == id {
			e, _ := db.objectIDName[r.Element]
			content := RelationContent{
				Relation: Relation{
					Parent:  name,
					Element: e.Name,
				},
				Content: ct,
			}
			result = append(result, content)
		}
	}
	return result
}

func (db *database) ScenarioDotElement(name string) []RelationContent {
	if db.getObjectType(name) != common.ScenarioTypeConst {
		return nil
	}

	id := db.getIDByName(name)
	var result []RelationContent
	for r, ct := range db.relationScenario {
		if r.Parent == id {
			e, _ := db.objectIDName[r.Element]
			content := RelationContent{
				Relation: Relation{
					Parent:  name,
					Element: e.Name,
				},
				Content: ct,
			}
			result = append(result, content)
		}
	}
	return result
}

func (db *database) ElementDotGroups(name string) []RelationContent {
	id := db.getIDByName(name)
	if id == "" {
		return nil
	}

	var result []RelationContent
	for r, ct := range db.relationGroup {
		if r.Element == id {
			p, _ := db.objectIDName[r.Parent]
			content := RelationContent{
				Relation: Relation{
					Parent:  p.Name,
					Element: name,
				},
				Content: ct,
			}
			result = append(result, content)
		}
	}
	return result
}

func (db *database) ElementDotScenario(name string) []RelationContent {
	id := db.getIDByName(name)
	if id == "" {
		return nil
	}

	var result []RelationContent
	for r, ct := range db.relationScenario {
		if r.Element == id {
			p, _ := db.objectIDName[r.Parent]
			content := RelationContent{
				Relation: Relation{
					Parent:  p.Name,
					Element: name,
				},
				Content: ct,
			}
			result = append(result, content)
		}
	}
	return result
}

func (db *database) getObjectType(name string) string {
	id, existID := db.objectNameID[name]
	if !existID {
		return ""
	}

	info, _ := db.objectIDName[id]
	return info.Type
}

func (db *database) getNameByID(id string) string {
	name, exist := db.objectIDName[id]
	if !exist {
		return ""
	}
	return name.Name
}

func (db *database) getIDByName(name string) string {
	id, exist := db.objectNameID[name]
	if !exist {
		return ""
	}
	return id
}

func (db *database) getRelationMapByType(t string) map[Relation]string {
	var mr map[Relation]string
	if t == common.GroupTypeConst {
		mr = db.relationGroup
	} else if t == common.ScenarioTypeConst {
		mr = db.relationScenario
	}
	return mr
}

func deleteAllRelationByParent(m map[Relation]string, parent string) {
	for r := range m {
		if r.Parent == parent {
			delete(m, r)
		}
	}
}

func deleteAllRelationByElement(m map[Relation]string, element string) {
	for r := range m {
		if r.Element == element {
			delete(m, r)
		}
	}
}

func (db *database) deleteAllRelationByID(id string) {
	for r := range db.relationGroup {
		if r.Element == id || r.Parent == id {
			delete(db.relationGroup, r)
		}
	}

	for r := range db.relationScenario {
		if r.Element == id || r.Parent == id {
			delete(db.relationScenario, r)
		}
	}
}
