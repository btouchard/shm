// SPDX-License-Identifier: AGPL-3.0-or-later

package badge

const (
	ColorLabel  = "#555"
	ColorGreen  = "#00D084"
	ColorBlue   = "#3B82F6"
	ColorYellow = "#F59E0B"
	ColorRed    = "#EF4444"
	ColorPurple = "#8B5CF6"
	ColorGray   = "#6B7280"
	ColorTeal   = "#14B8A6"
	ColorIndigo = "#6366F1"
)

func GetInstancesColor(count int) string {
	switch {
	case count >= 10:
		return ColorGreen
	case count >= 5:
		return ColorYellow
	default:
		return ColorRed
	}
}

func GetMetricColor(value float64) string {
	switch {
	case value >= 1000:
		return ColorGreen
	case value >= 100:
		return ColorBlue
	case value > 0:
		return ColorYellow
	default:
		return ColorGray
	}
}
