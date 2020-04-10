package light

import (
	"fmt"
	"time"

	sdkModel "github.com/phanvanhai/device-sdk-go/pkg/models"
	appModels "github.com/phanvanhai/device-service-support/application/models"
)

func ParseTimeToInt64(t time.Time) uint64 {
	year, month, day := t.Date()
	hour, min, sec := t.Clock()
	var result uint64
	result = (uint64(year) << 40) | (uint64(month) << 32) | (uint64(day) << 24) | (uint64(hour) << 16) | (uint64(min) << 8) | uint64(sec)
	return result
}

func (l *Light) updateRealtime(devName string) error {
	t := time.Now()
	time64 := ParseTimeToInt64(t)

	request, ok := appModels.NewCommandRequest(devName, RealtimeDr)
	if !ok {
		l.lc.Error("khong tim thay resource")
		return fmt.Errorf("khong tim thay resource")
	}

	cmvlConverted, err := sdkModel.NewUint64Value(RealtimeDr, 0, time64)
	if err != nil {
		l.lc.Error(err.Error())
		return err
	}
	param := make([]*sdkModel.CommandValue, 0, 1)
	param = append(param, cmvlConverted)

	reqs := make([]*sdkModel.CommandRequest, 1)
	reqs[0] = request
	err = l.nw.WriteCommands(devName, reqs, param)
	if err != nil {
		l.lc.Error(err.Error())
		l.updateOpStateAndConnectdStatus(devName, false)
		return err
	}
	return nil
}
