package components

import (
	"github.com/xmhrnx/giode"
)

// Grid lays items out in rows of cols equal-width cells, gap dp apart.
func Grid(cols, gap int, items ...giode.Element) giode.Element {
	half := gap / 2
	rows := make([]giode.Element, 0, (len(items)+cols-1)/cols)
	for i := 0; i < len(items); i += cols {
		cells := make([]giode.Element, 0, cols)
		for j := 0; j < cols; j++ {
			var item giode.Element = giode.Box(giode.Styles{})
			if i+j < len(items) {
				item = items[i+j]
			}
			cells = append(cells, giode.Box(giode.Styles{
				FlexGrow: 1,
				Padding:  giode.Inset{Left: half, Right: half, Top: half, Bottom: half},
			}, item))
		}
		rows = append(rows, giode.Row(giode.Styles{Width: -1, Wrap: giode.WrapNoWrap}, cells...))
	}
	return giode.Stack(giode.Styles{Width: -1, Margin: giode.Inset{Left: -half, Right: -half}}, rows...)
}
