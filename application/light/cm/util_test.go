package cm

import (
	"testing"
)

func TestGetAddress16FromNetDevice(t *testing.T) {
	var tests = []struct {
		input    string
		expected uint16
	}{
		{"01001234", 0x1234},
		{"0000AbcD", 0xabcd},
		{"0010GbcD", 0},
	}
	for _, tt := range tests {
		actual := GetAddress16FromNetDevice(tt.input)
		if actual != tt.expected {
			t.Errorf("error: actual=%x & expected=%x", actual, tt.expected)
		}
	}
}
