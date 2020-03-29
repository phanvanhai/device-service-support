package db

import (
	sdk "github.com/edgexfoundry/device-sdk-go"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

type GroupContent struct {
	ParentID  string
	ElementID string
}

type ScenarioContent struct {
	ParentID  string
	ElementID string
	Command   string
	Body      interface{}
}

type ObjectDB interface {
	UpdateObject(object models.Device)
	DeleteDevice(id string)

	GetObjectType(id string) string
	IDToName(id string) string
	NameToID(name string) string
	// GetResourceByTag(tag string, profileID string) string
	GroupDotElement(id string) []GroupContent
	ScenarioDotElement(id string) []ScenarioContent
	ElementDotGroups(id string) []GroupContent
	ElementDotScenario(id string) []ScenarioContent
}

type DB struct {
	sv *sdk.Service
	lc logger.LoggingClient
}

var db *DB

func initDB() {

}

func NewDBClient(sv *sdk.Service, lc logger.LoggingClient) error {
	if db == nil {
		initDB()
	}
	return db
}

func (db *DB) UpdateGroup(gr models.Device) {

}

func (db *DB) UpdateScenario(sr models.Device) {

}

func (db *DB) UpdateScenario(dv models.Device) {

}

func (db *DB) DeleteGroup(gr models.Device) {

}

func (db *DB) DeleteScenario(sr models.Device) {

}

func (db *DB) DeleteScenario(dv models.Device) {

}

func (db *DB) GroupDotElement(id string) []GroupContent {
	return nil
}

func (db *DB) ScenarioDotElement(id string) []ScenarioContent {
	return nil
}

func (db *DB) ElementDotGroups(id string) []GroupContent {
	return nil
}

func (db *DB) ElementDotScenario(id string) []ScenarioContent {
	return nil
}
