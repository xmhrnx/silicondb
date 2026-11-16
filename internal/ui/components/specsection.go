package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// SpecLine is a formatted spec ready to show.
type SpecLine struct {
	Label, Value string
}

// SpecSection is a card with a group title and its spec lines.
func SpecSection(title string, lines []SpecLine) giode.Element {
	rows := make([]giode.Element, 0, len(lines)+1)
	rows = append(rows, giode.Box(giode.Styles{Padding: giode.Inset{Bottom: 4}},
		giode.Text(title, theme.Overline().With(giode.Styles{Color: theme.Primary}))))
	for i, l := range lines {
		rows = append(rows, SpecRow(l.Label, l.Value, i == len(lines)-1))
	}
	return Card(giode.Styles{Padding: giode.Inset{Top: 16, Bottom: 6, Left: 16, Right: 16}}, rows...)
}
