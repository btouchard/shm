// SPDX-License-Identifier: AGPL-3.0-or-later

package badge

// Color palette for badges - custom and original
const (
	// Label background (left side)
	ColorLabel = "#404040"

	// Status colors (right side)
	ColorGreen   = "#00D084" // Success, high numbers
	ColorBlue    = "#3B82F6" // Info, neutral
	ColorYellow  = "#F59E0B" // Warning, medium numbers
	ColorRed     = "#EF4444" // Error, low numbers
	ColorPurple  = "#8B5CF6" // Special, version info
	ColorGray    = "#6B7280" // Inactive, no data
	ColorTeal    = "#14B8A6" // Custom metrics
	ColorIndigo  = "#6366F1" // Combined stats
)

// GetInstancesColor returns the appropriate color based on instance count.
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

// GetMetricColor returns the color for a custom metric based on value.
// This is a simple heuristic - adjust based on your needs.
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
