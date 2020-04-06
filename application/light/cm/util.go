package cm

import (
	"strconv"
	"time"
)

func CheckScheduleTime(t uint32) bool {
	return (t != 0xFFFFF)
}

func ParseTimeToInt64(t time.Time) uint64 {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	var result uint64
	result = (uint64(year) << 40) | (uint64(month) << 32) | (uint64(day) << 24) | (uint64(hour) << 16) | (uint64(min) << 8) | uint64(sec)
	return result
}

func GetAddress16(rootNet string, srcNet string) uint16 {
	if srcNet == rootNet {
		return 0
	}
	addr64, _ := strconv.ParseInt(srcNet, 16, 32)
	return uint16(addr64 & 0xFFFF)
}
