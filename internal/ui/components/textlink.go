package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// TextLink is tappable accent text such as "See all". It holds press
// state, so it is looked up by key.
func TextLink(key, label string, onClick func()) giode.Element {
	p := giode.Keyed("textlink", key, func() *giode.PressableElement { return giode.Pressable(nil) })
	return p.Content(func(s giode.PressState) giode.Element {
		color := theme.Primary
		if s.Pressed || s.Hovered {
			color = theme.Accent
		}
		return giode.Text(label, giode.Styles{FontFamily: theme.Family, FontSize: 14, FontWeight: giode.FontWeightSemiBold, Color: color})
	}).Label(label).OnClick(func(giode.ClickEvent) { onClick() })
}
