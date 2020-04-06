package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"

	"github.com/phanvanhai/device-service-support/application/light/cm"
)

type EdgeScheduleOnOff struct {
	Owner string `json:"owner, omitempty"`
	Time  uint32 `json:"time, omitempty"`
	Value bool   `json:"value, omitempty"`
}

type NetScheduleOnOff struct {
	Owner uint16
	Time  uint32
	Value bool
}

func ScheduleOnOffNetToEdge(net NetScheduleOnOff, owner string) EdgeScheduleOnOff {
	result := EdgeScheduleOnOff{
		Owner: owner,
		Time:  net.Time,
		Value: net.Value,
	}
	return result
}

func ScheduleOnOffEdgeToNet(edge EdgeScheduleOnOff, owner uint16) NetScheduleOnOff {
	result := NetScheduleOnOff{
		Owner: owner,
		Time:  edge.Time,
		Value: edge.Value,
	}
	return result
}

func NetScheduleOnOffToString(scheules []NetScheduleOnOff, size int) string {
	if len(scheules) < size {
		s := make([]NetScheduleOnOff, size-len(scheules))
		scheules = append(scheules, s...)
	}
	if len(scheules) > size {
		scheules = scheules[:size]
	}
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, scheules)
	scheulesByte := buf.Bytes()
	return base64.StdEncoding.EncodeToString(scheulesByte)
}

// kich thuoc bieu dien phai dung = size
func StringToNetScheduleOnOff(scheduleStr string, size int) ([]NetScheduleOnOff, error) {
	decode, err := base64.StdEncoding.DecodeString(scheduleStr)
	if err != nil {
		return nil, err
	}

	sch := make([]NetScheduleOnOff, size)
	reader := bytes.NewBuffer(decode)
	err = binary.Read(reader, binary.BigEndian, sch)
	if err != nil {
		return nil, err
	}

	result := make([]NetScheduleOnOff, 0, size)
	for i := 0; i < size; i++ {
		if cm.CheckScheduleTime(sch[i].Time) == false {
			continue
		}
		result = append(result, sch[i])
	}
	return result, nil
}
