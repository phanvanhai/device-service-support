package models

import (
	"fmt"
	"strconv"

	sdkModel "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/phanvanhai/device-service-support/support/common"
)

const ReportTimeInitStringValueConst = "15"

func UpdateReportTimeConfigToDevice(cm NormalWriteCommand, dev *models.Device, resourceName string, time uint16) error {
	request, ok := NewCommandRequest(dev.Name, resourceName)
	if !ok {
		return fmt.Errorf("khong tim thay resource")
	}

	cmvlConverted, _ := sdkModel.NewUint16Value(resourceName, 0, time)

	return cm.NormalWriteCommand(dev, request, cmvlConverted)
}

func FillReportTimeToDB(dev *models.Device, value string) {
	SetProperty(dev, common.ReportTimeProtocolName, common.ReportTimePropertyName, value)
}

func GetReportTimeFromDB(dev models.Device) (time uint16) {
	time64, _ := strconv.ParseUint(ReportTimeInitStringValueConst, 10, 16)
	time = uint16(time64)
	timeStr, ok := GetProperty(&dev, common.ReportTimeProtocolName, common.ReportTimePropertyName)
	if !ok {
		return
	}

	time64, _ = strconv.ParseUint(timeStr, 10, 16)
	time = uint16(time64)
	return
}
