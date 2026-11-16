package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// EmptyState fills a screen that has nothing to show, with an optional
// action below the message.
func EmptyState(icon, title, message string, action giode.Element) giode.Element {
	children := []giode.Element{
		IconBadge(icon, theme.Primary, 64),
		giode.Text(title, theme.Heading().With(giode.Styles{TextAlign: giode.TextAlignCenter})),
		giode.Box(giode.Styles{MaxWidth: 300},
			giode.Text(message, theme.Caption().With(giode.Styles{FontSize: 14, TextAlign: giode.TextAlignCenter, LineHeight: 20}))),
	}
	if action != nil {
		children = append(children, giode.Box(giode.Styles{Padding: giode.Inset{Top: 8}}, action))
	}
	return giode.Stack(giode.Styles{
		Width: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 10,
		Padding: giode.SymmetricInset(24, 48),
	}, children...)
}
