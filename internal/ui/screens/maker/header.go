package maker

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// header is the company card with its numbers.
func header(cat *model.Catalog, m *model.Manufacturer) giode.Element {
	all := cat.ChipsBy(m.ID)
	first, last := 0, 0
	for _, c := range all {
		if y := c.Year(); y != 0 {
			if first == 0 || y < first {
				first = y
			}
			last = max(last, y)
		}
	}
	span := "—"
	if first != 0 {
		span = strconv.Itoa(first) + "–" + strconv.Itoa(last)
	}
	status := tr.T("maker.active")
	if !m.Active() {
		status = tr.T("maker.until", "year", strconv.Itoa(m.Defunct))
	}
	return components.Card(giode.Styles{Gap: 16, Padding: giode.UniformInset(20), Margin: giode.Inset{Bottom: 8}},
		giode.Row(giode.Styles{Gap: 14, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Avatar(m.Name, giode.Styles{Width: 60}),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Text(m.Name, theme.Title()),
				giode.Text(components.MakerSubtitle(m), theme.Caption().With(giode.Styles{FontSize: 13.5})),
			),
		),
		components.Grid(3, 10,
			components.StatTile(tr.T("maker.chips"), strconv.Itoa(len(all)), theme.Primary),
			components.StatTile(tr.T("maker.releases"), span, theme.Accent),
			components.StatTile(tr.T("maker.status"), status, theme.Success),
		),
	)
}
