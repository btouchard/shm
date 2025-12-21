// SPDX-License-Identifier: AGPL-3.0-or-later

package badge

import "testing"

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "0"},
		{5, "5"},
		{42, "42"},
		{999, "999"},
		{1000, "1.0k"},
		{1234, "1.2k"},
		{1500, "1.5k"},
		{10000, "10.0k"},
		{999999, "1000.0k"},
		{1000000, "1.0M"},
		{1234567, "1.2M"},
		{1000000000, "1.0B"},
		{1234567890, "1.2B"},
		{-100, "0"},
	}

	for _, tt := range tests {
		result := FormatNumber(tt.input)
		if result != tt.expected {
			t.Errorf("FormatNumber(%.0f) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestFormatCompact(t *testing.T) {
	result := FormatCompact(1234, "users", 42, "inst")
	expected := "1.2k users / 42 inst"
	if result != expected {
		t.Errorf("FormatCompact() = %s, expected %s", result, expected)
	}
}
