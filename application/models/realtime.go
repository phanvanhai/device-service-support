package models

import (
	"fmt"
	"time"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
)

func parseTimeToInt64(t time.Time) uint64 {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	var result uint64
	result = (uint64(year) << 40) | (uint64(month) << 32) | (uint64(day) << 24) | (uint64(hour) << 16) | (uint64(min) << 8) | uint64(sec)
	return result
}

func UpdateRealtimeToDevice(cm NormalWriteCommand, dev *models.Device, resourceName string) error {
	request, ok := NewCommandRequest(dev.Name, resourceName)
	if !ok {
		return fmt.Errorf("khong tim thay resource")
	}

	t := time.Now()
	time64 := parseTimeToInt64(t)
	cmvlConverted, _ := sdkModel.NewUint64Value(resourceName, 0, time64)

	return cm.NormalWriteCommand(dev, request, cmvlConverted)
}
