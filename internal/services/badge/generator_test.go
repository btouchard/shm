// SPDX-License-Identifier: AGPL-3.0-or-later

package badge

import (
	"strings"
	"testing"
)

func TestNewBadge(t *testing.T) {
	badge := NewBadge("instances", "42", ColorGreen)

	if badge.Label != "instances" {
		t.Errorf("expected label 'instances', got '%s'", badge.Label)
	}
	if badge.Value != "42" {
		t.Errorf("expected value '42', got '%s'", badge.Value)
	}
	if badge.Color != ColorGreen {
		t.Errorf("expected color %s, got %s", ColorGreen, badge.Color)
	}
	if badge.LabelColor != ColorLabel {
		t.Errorf("expected label color %s, got %s", ColorLabel, badge.LabelColor)
	}
}

func TestBadgeToSVG(t *testing.T) {
	badge := NewBadge("test", "value", ColorBlue)
	svg := badge.ToSVG()

	// Check SVG structure
	if !strings.HasPrefix(svg, "<svg") {
		t.Error("SVG should start with <svg tag")
	}
	if !strings.HasSuffix(svg, "</svg>") {
		t.Error("SVG should end with </svg> tag")
	}

	// Check content
	if !strings.Contains(svg, "test") {
		t.Error("SVG should contain label text")
	}
	if !strings.Contains(svg, "value") {
		t.Error("SVG should contain value text")
	}
	if !strings.Contains(svg, ColorBlue) {
		t.Error("SVG should contain the color")
	}

	// Check accessibility
	if !strings.Contains(svg, "<title>") {
		t.Error("SVG should have a title for accessibility")
	}
	if !strings.Contains(svg, `aria-label`) {
		t.Error("SVG should have aria-label for accessibility")
	}
}

func TestBadgeCustomLabelColor(t *testing.T) {
	badge := &Badge{
		Label:      "custom",
		Value:      "test",
		Color:      ColorGreen,
		LabelColor: ColorPurple,
	}

	svg := badge.ToSVG()
	if !strings.Contains(svg, ColorPurple) {
		t.Error("SVG should contain custom label color")
	}
}
