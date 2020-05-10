package models

import (
	"time"
)

type Realtime interface {
	WriteRealtimeToDevice(deviceName string, time uint64) error
}

func parseTimeToInt64(t time.Time) uint64 {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	var result uint64
	result = (uint64(year) << 40) | (uint64(month) << 32) | (uint64(day) << 24) | (uint64(hour) << 16) | (uint64(min) << 8) | uint64(sec)
	return result
}

func UpdateRealtimeToDevice(realtimer Realtime, devName string) error {
	t := time.Now()
	time64 := parseTimeToInt64(t)
	return realtimer.WriteRealtimeToDevice(devName, time64)
}
