package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// MaxContentWidth keeps lines readable on tablets and desktops.
const MaxContentWidth = 760

// Column is a screen's content: a vertical stack at most MaxContentWidth
// wide, centered, with the side gutter. Nil children are skipped, so
// sections with nothing to show take no space.
func Column(gap int, children ...giode.Element) giode.Element {
	children = Present(children...)
	return giode.Box(giode.Styles{Width: -1, Justify: giode.JustifyCenter},
		giode.Stack(giode.Styles{
			Width: -1, MaxWidth: MaxContentWidth, Gap: properties.Gap(gap),
			Padding: giode.Inset{Left: theme.Gutter, Right: theme.Gutter, Bottom: 28},
		}, children...))
}
