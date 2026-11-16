package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// StatTile shows one headline figure with its label, tinted by color.
// Long values, like a span of years, get a smaller size so they fit three
// tiles to a row on a phone.
func StatTile(label, value, color string) giode.Element {
	valueStyle := theme.Value()
	if n := len([]rune(value)); n > 6 {
		valueStyle.FontSize = 15
	}
	return giode.Stack(giode.Styles{
		Width: -1, Gap: 6,
		Padding:      giode.UniformInset(14),
		Background:   theme.Tint(color, 0.08),
		BorderRadius: theme.RadiusTile,
		BorderWidth:  1,
		BorderColor:  theme.Tint(color, 0.22),
	},
		giode.Text(label, theme.Overline().With(giode.Styles{MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis})),
		giode.Text(value, valueStyle.With(giode.Styles{MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis})),
	)
}
