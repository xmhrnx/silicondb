package components

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// listRow is the card shared by the drill-down rows: content, then a
// chevron, shaded while pressed.
func listRow(s giode.PressState, children ...giode.Element) giode.Element {
	bg := theme.Surface
	if s.Pressed {
		bg = theme.Pressed()
	} else if s.Hovered {
		bg = theme.Raised()
	}
	children = append(children, theme.Icon("chevron-right", 18, theme.Muted))
	return giode.Row(giode.Styles{
		Width: -1, Align: giode.AlignCenter, Gap: 14, Wrap: giode.WrapNoWrap,
		Padding:      giode.SymmetricInset(14, 14),
		Background:   bg,
		BorderRadius: 18,
		BorderWidth:  1,
		BorderColor:  theme.Border,
	}, children...)
}

// countPill shows a number of chips on a raised pill.
func countPill(n int) giode.Element {
	return giode.Box(giode.Styles{
		Padding: giode.SymmetricInset(10, 4), Background: theme.Raised(), BorderRadius: theme.RadiusPill,
	}, giode.Text(strconv.Itoa(n), giode.Styles{FontFamily: theme.Family, FontSize: 12.5, FontWeight: giode.FontWeightSemiBold, Color: theme.Text}))
}
