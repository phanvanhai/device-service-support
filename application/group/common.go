package group

const (
	OnOffDr           = "LightGroup-OnOff"
	DimmingDr         = "LightGroup-Dimming"
	OnOffScheduleDr   = "LightGroup-OnOffSchedule"
	DimmingScheduleDr = "LightGroup-DimmingSchedule"
	MethodDr          = "LightGroup-Method"
	DeviceDr          = "LightGroup-Device"
	ListDeviceDr      = "LightGroup-ListDevice"
	ScenarioDr        = "LightGroup-Scenario"
)

const (
	PutMethod    = "put"
	DeleteMethod = "delete"
)

const (
	GroupLimit           = 50
	OnOffScheduleLimit   = 15
	DimmingScheduleLimit = 15
)

const (
	ScheduleProtocolName        = "Schedule"
	OnOffSchedulePropertyName   = "OnOff"
	DimmingSchedulePropertyName = "Dimming"
)

type ElementError struct {
	Name  string
	Error string
}
