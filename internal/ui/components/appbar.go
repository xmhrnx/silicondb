package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// AppBar is a screen's title row: an optional leading element (a back
// button), the title with an optional subtitle, and trailing actions.
func AppBar(leading giode.Element, title, subtitle string, trailing ...giode.Element) giode.Element {
	head := []giode.Element{giode.Text(title, theme.Title().With(giode.Styles{MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis}))}
	if subtitle != "" {
		head = append(head, giode.Text(subtitle, theme.Caption().With(giode.Styles{FontSize: 13.5})))
	}
	var children []giode.Element
	if leading != nil {
		children = append(children, leading)
	}
	children = append(children, giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2}, head...))
	if len(trailing) > 0 {
		children = append(children, giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap}, trailing...))
	}
	return giode.Row(giode.Styles{
		Width: -1, Align: giode.AlignCenter, Gap: 12, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Top: 14, Bottom: 10},
	}, children...)
}
