package models

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"

	"github.com/phanvanhai/device-service-support/application/light/cm"
)

type EdgeScheduleDimming struct {
	Owner string `json:"owner, omitempty"`
	Time  uint32 `json:"time, omitempty"`
	Value uint16 `json:"value, omitempty"`
}

type NetScheduleDimming struct {
	Owner uint16
	Time  uint32
	Value uint16
}

func ScheduleDimmingNetToEdge(net NetScheduleDimming, owner string) EdgeScheduleDimming {
	result := EdgeScheduleDimming{
		Owner: owner,
		Time:  net.Time,
		Value: net.Value,
	}
	return result
}

func ScheduleDimmingEdgeToNet(edge EdgeScheduleDimming, owner uint16) NetScheduleDimming {
	result := NetScheduleDimming{
		Owner: owner,
		Time:  edge.Time,
		Value: edge.Value,
	}
	return result
}

func NetScheduleDimmingToString(scheules []NetScheduleDimming, size int) string {
	if len(scheules) < size {
		s := make([]NetScheduleDimming, size-len(scheules))
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
func StringToNetScheduleDimming(scheduleStr string, size int) ([]NetScheduleDimming, error) {
	decode, err := base64.StdEncoding.DecodeString(scheduleStr)
	if err != nil {
		return nil, err
	}

	sch := make([]NetScheduleDimming, size)
	reader := bytes.NewBuffer(decode)
	err = binary.Read(reader, binary.BigEndian, sch)
	if err != nil {
		return nil, err
	}

	result := make([]NetScheduleDimming, 0, size)
	for i := 0; i < size; i++ {
		if cm.CheckScheduleTime(sch[i].Time) == false {
			continue
		}
		result = append(result, sch[i])
	}
	return result, nil
}
