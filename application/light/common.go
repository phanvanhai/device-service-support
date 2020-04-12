package light

const (
	OnOffDr           = "Light-OnOff"
	DimmingDr         = "Light-Dimming"
	OnOffScheduleDr   = "Light-OnOffSchedule"
	DimmingScheduleDr = "Light-DimmingSchedule"
	MeasurePowerDr    = "Light-MeasurePower"
	ReportTimeDr      = "Light-ReportTime"
	RealtimeDr        = "Light-Realtime"
	HistoricalEventDr = "Light-HistoricalEvent"
	GroupDr           = "Light-Group"
	ScenarioDr        = "Light-Scenario"
	PingDr            = "Light-Ping"
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