package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// FilterChip is a selectable pill for filters, with an optional count.
// It holds press state, so it is looked up by (scope, id).
func FilterChip(scope, id, label, count string, selected bool, onClick func()) giode.Element {
	p := giode.Keyed("filterchip/"+scope, id, func() *giode.PressableElement {
		return giode.Pressable(nil)
	})
	return p.Content(func(s giode.PressState) giode.Element {
		bg, fg, border := theme.Surface, theme.Text, theme.Border
		switch {
		case selected:
			bg, fg, border = theme.Primary, theme.OnPrimary, theme.Primary
		case s.Pressed:
			bg = theme.Pressed()
		case s.Hovered:
			bg = theme.Raised()
		}
		kids := []giode.Element{giode.Text(label, giode.Styles{FontFamily: theme.Family, FontSize: 13.5, FontWeight: giode.FontWeightMedium, Color: fg})}
		if count != "" {
			kids = append(kids, giode.Text(count, giode.Styles{FontFamily: theme.Family, FontSize: 12, Color: fg, Opacity: 0.65}))
		}
		return giode.Row(giode.Styles{
			Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(14, 8), Background: bg, BorderRadius: theme.RadiusPill,
			BorderWidth: 1, BorderColor: border,
		}, kids...)
	}).Label(label).OnClick(func(giode.ClickEvent) { onClick() })
}
