package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// RoundButton is a circular icon button on a raised surface, for back and
// toolbar actions. Create it once; call Icon to change its glyph.
type RoundButton struct {
	*giode.PressableElement
	icon    string
	color   string
	onClick func()
}

// NewRoundButton returns a button showing icon.
func NewRoundButton(icon, label string, onClick func()) *RoundButton {
	b := &RoundButton{icon: icon, onClick: onClick}
	b.PressableElement = giode.Pressable(b.content).Label(label).OnClick(func(giode.ClickEvent) {
		if b.onClick != nil {
			b.onClick()
		}
	})
	return b
}

// Icon changes the glyph and its color ("" for the text color).
func (b *RoundButton) Icon(icon, color string) *RoundButton {
	b.icon, b.color = icon, color
	return b
}

func (b *RoundButton) content(s giode.PressState) giode.Element {
	bg := theme.Raised()
	if s.Pressed {
		bg = theme.Pressed()
	}
	color := b.color
	if color == "" {
		color = theme.Text
	}
	return giode.Box(giode.Styles{
		Width: 42, Height: 42, Align: giode.AlignCenter, Justify: giode.JustifyCenter,
		Background: bg, BorderRadius: theme.RadiusPill, BorderWidth: 1, BorderColor: theme.Border,
	}, theme.Icon(b.icon, 20, color))
}
