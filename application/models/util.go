package models

func CheckScheduleTime(t uint32) bool {
	return (t != 0xFFFFFF)
}
