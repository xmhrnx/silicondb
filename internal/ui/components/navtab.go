package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// NavTab is one destination of the bottom navigation bar; Label is a
// message key.
type NavTab struct {
	Icon, Label string
}

// navTabItem draws a tab: the icon on a pill when selected, and its label.
func navTabItem(t NavTab, selected bool, s giode.PressState) giode.Element {
	color := theme.Muted
	pill := "transparent"
	if selected {
		color = theme.Primary
		pill = theme.Tint(theme.Primary, 0.16)
	} else if s.Pressed || s.Hovered {
		color = theme.Text
	}
	weight := giode.FontWeightMedium
	if selected {
		weight = giode.FontWeightSemiBold
	}
	return giode.Stack(giode.Styles{Width: -1, Align: giode.AlignCenter, Gap: 4, Padding: giode.SymmetricInset(0, 8)},
		giode.Box(giode.Styles{
			Width: 56, Height: 30, Align: giode.AlignCenter, Justify: giode.JustifyCenter,
			Background: pill, BorderRadius: theme.RadiusPill,
		}, theme.Icon(t.Icon, 21, color)),
		giode.Text(tr.T(t.Label), giode.Styles{FontFamily: theme.Family, FontSize: 11.5, FontWeight: weight, Color: color}),
	)
}
