// Package components holds the app's reusable UI pieces, one per file.
package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// IconBadge is an icon centered on a softly tinted rounded square.
func IconBadge(icon, color string, size int) giode.Element {
	return giode.Box(giode.Styles{
		Width:        properties.Width(size),
		Height:       properties.Height(size),
		Align:        giode.AlignCenter,
		Justify:      giode.JustifyCenter,
		Background:   theme.Tint(color, 0.16),
		BorderRadius: properties.BorderRadius(size * 3 / 10),
	}, theme.Icon(icon, size*11/20, color))
}
