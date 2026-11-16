package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// SectionHeader titles a section, with an optional element (a "See all"
// link) on the right.
func SectionHeader(title string, trailing giode.Element) giode.Element {
	children := []giode.Element{giode.Text(title, theme.Heading())}
	if trailing != nil {
		children = append(children, trailing)
	}
	return giode.Row(giode.Styles{
		Width: -1, Align: giode.AlignCenter, Justify: giode.JustifySpaceBetween, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Top: 8, Bottom: 2},
	}, children...)
}
