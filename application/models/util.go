package models

const (
	ScheduleNilStr = "[]"
	TimeError      = 0x00000000
)

func CheckScheduleTime(t uint32) bool {
	return (t != TimeError)
}

func CreateScheuleTimeError() uint32 {
	return TimeError
}
