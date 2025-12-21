// SPDX-License-Identifier: AGPL-3.0-or-later

package badge

import (
	"fmt"
	"strings"
)

// Badge represents a shield-style badge with label and value.
type Badge struct {
	Label      string
	Value      string
	Color      string // Hex color for value side
	LabelColor string // Hex color for label side (default: ColorLabel)
}

// ToSVG generates an SVG representation of the badge.
func (b *Badge) ToSVG() string {
	if b.LabelColor == "" {
		b.LabelColor = ColorLabel
	}

	// Calculate widths based on text length (rough approximation: 6px per char)
	labelWidth := len(b.Label)*6 + 10
	valueWidth := len(b.Value)*6 + 10
	totalWidth := labelWidth + valueWidth

	// Build SVG
	var svg strings.Builder

	svg.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">`,
		totalWidth, b.Label, b.Value))

	// Title for accessibility
	svg.WriteString(fmt.Sprintf(`<title>%s: %s</title>`, b.Label, b.Value))

	// Linear gradient for smooth appearance
	svg.WriteString(`<linearGradient id="s" x2="0" y2="100%">`)
	svg.WriteString(`<stop offset="0" stop-color="#bbb" stop-opacity=".1"/>`)
	svg.WriteString(`<stop offset="1" stop-opacity=".1"/>`)
	svg.WriteString(`</linearGradient>`)

	// Clipping path for rounded corners
	svg.WriteString(fmt.Sprintf(`<clipPath id="r"><rect width="%d" height="20" rx="3" fill="#fff"/></clipPath>`, totalWidth))

	// Main group with clip
	svg.WriteString(`<g clip-path="url(#r)">`)

	// Label background
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="20" fill="%s"/>`, labelWidth, b.LabelColor))

	// Value background
	svg.WriteString(fmt.Sprintf(`<rect x="%d" width="%d" height="20" fill="%s"/>`, labelWidth, valueWidth, b.Color))

	// Gradient overlay
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="20" fill="url(#s)"/>`, totalWidth))

	svg.WriteString(`</g>`)

	// Text group
	svg.WriteString(`<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="110">`)

	// Label text
	labelX := labelWidth / 2
	svg.WriteString(fmt.Sprintf(`<text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>`,
		labelX*10, (labelWidth-10)*10, b.Label))
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>`,
		labelX*10, (labelWidth-10)*10, b.Label))

	// Value text
	valueX := labelWidth + valueWidth/2
	svg.WriteString(fmt.Sprintf(`<text aria-hidden="true" x="%d" y="150" fill="#010101" fill-opacity=".3" transform="scale(.1)" textLength="%d">%s</text>`,
		valueX*10, (valueWidth-10)*10, b.Value))
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="140" transform="scale(.1)" fill="#fff" textLength="%d">%s</text>`,
		valueX*10, (valueWidth-10)*10, b.Value))

	svg.WriteString(`</g>`)
	svg.WriteString(`</svg>`)

	return svg.String()
}

// NewBadge creates a new badge with the given parameters.
func NewBadge(label, value, color string) *Badge {
	return &Badge{
		Label:      label,
		Value:      value,
		Color:      color,
		LabelColor: ColorLabel,
	}
}
