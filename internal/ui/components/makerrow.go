package components

import (
	"fmt"
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// MakerRow is a list row for a manufacturer with its chip count. Its
// press state is kept per manufacturer; the content and click handler are
// set every frame, so the same company can open different pages from
// different lists.
func MakerRow(m *model.Manufacturer, count int, onClick func()) giode.Element {
	p := giode.Keyed("makerrow", m, func() *giode.PressableElement { return giode.Pressable(nil) })
	return p.Content(func(s giode.PressState) giode.Element {
		bg := theme.Surface
		if s.Pressed {
			bg = theme.Pressed()
		} else if s.Hovered {
			bg = theme.Raised()
		}
		return giode.Row(giode.Styles{
			Width: -1, Align: giode.AlignCenter, Gap: 14, Wrap: giode.WrapNoWrap,
			Padding:      giode.SymmetricInset(14, 12),
			Background:   bg,
			BorderRadius: 18,
			BorderWidth:  1,
			BorderColor:  theme.Border,
		},
			giode.Avatar(m.Name, giode.Styles{Width: 44}),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Text(m.Name, theme.Strong().With(giode.Styles{MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis})),
				giode.Text(MakerSubtitle(m), theme.Caption().With(giode.Styles{MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis})),
			),
			giode.Box(giode.Styles{
				Padding: giode.SymmetricInset(10, 4), Background: theme.Raised(), BorderRadius: theme.RadiusPill,
			}, giode.Text(fmt.Sprint(count), giode.Styles{FontFamily: theme.Family, FontSize: 12.5, FontWeight: giode.FontWeightSemiBold, Color: theme.Text})),
			theme.Icon("chevron-right", 18, theme.Muted),
		)
	}).Label(m.Name).OnClick(func(giode.ClickEvent) { onClick() })
}

// MakerSubtitle is "Country · 1968–today".
func MakerSubtitle(m *model.Manufacturer) string {
	s := m.Country
	if m.Founded != 0 {
		end := tr.T("makers.today")
		if m.Defunct != 0 {
			end = strconv.Itoa(m.Defunct)
		}
		if s != "" {
			s += " · "
		}
		s += tr.T("makers.founded", "from", strconv.Itoa(m.Founded), "to", end)
	}
	return s
}
