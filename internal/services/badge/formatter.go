// SPDX-License-Identifier: AGPL-3.0-or-later

package badge

import (
	"fmt"
	"math"
)

// FormatNumber formats a number with k/M/B suffixes for compact display.
func FormatNumber(n float64) string {
	if n < 0 {
		return "0"
	}

	abs := math.Abs(n)

	switch {
	case abs >= 1_000_000_000:
		return fmt.Sprintf("%.1fB", n/1_000_000_000)
	case abs >= 1_000_000:
		return fmt.Sprintf("%.1fM", n/1_000_000)
	case abs >= 1_000:
		return fmt.Sprintf("%.1fk", n/1_000)
	default:
		// For integers, don't show decimals
		if n == math.Floor(n) {
			return fmt.Sprintf("%.0f", n)
		}
		return fmt.Sprintf("%.1f", n)
	}
}

// FormatCompact creates a compact representation like "1.2k / 42".
func FormatCompact(value1 float64, label1 string, value2 int, label2 string) string {
	return fmt.Sprintf("%s %s / %d %s", FormatNumber(value1), label1, value2, label2)
}
