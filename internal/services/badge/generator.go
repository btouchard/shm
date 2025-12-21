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

// ToSVG generates an SVG badge in shields.io flat-square style.
func (b *Badge) ToSVG() string {
	if b.LabelColor == "" {
		b.LabelColor = ColorLabel
	}

	labelWidth := len(b.Label)*7 + 10
	valueWidth := len(b.Value)*7 + 10
	totalWidth := labelWidth + valueWidth

	var svg strings.Builder

	svg.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">`,
		totalWidth, b.Label, b.Value))
	svg.WriteString(fmt.Sprintf(`<title>%s: %s</title>`, b.Label, b.Value))
	svg.WriteString(fmt.Sprintf(`<rect width="%d" height="20" fill="%s"/>`, labelWidth, b.LabelColor))
	svg.WriteString(fmt.Sprintf(`<rect x="%d" width="%d" height="20" fill="%s"/>`, labelWidth, valueWidth, b.Color))
	svg.WriteString(`<g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" text-rendering="geometricPrecision" font-size="11">`)

	labelX := labelWidth / 2
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>`, labelX, b.Label))
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="14">%s</text>`, labelX, b.Label))

	valueX := labelWidth + valueWidth/2
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="15" fill="#010101" fill-opacity=".3">%s</text>`, valueX, b.Value))
	svg.WriteString(fmt.Sprintf(`<text x="%d" y="14">%s</text>`, valueX, b.Value))

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
