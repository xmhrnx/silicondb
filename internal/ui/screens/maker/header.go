package maker

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// header is the company card, its numbers, and the category and segment
// filters.
func (s *Screen) header(cat *model.Catalog, m *model.Manufacturer) giode.Element {
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
	children := []giode.Element{
		components.AppBar(s.back, "", ""),
		components.Card(giode.Styles{Gap: 16, Padding: giode.UniformInset(20)},
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
		),
	}
	if row := s.categoryChips(cat, all); row != nil {
		children = append(children, row)
	}
	if row := s.segmentChips(cat); row != nil {
		children = append(children, row)
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 12, Padding: giode.Inset{Bottom: 2}}, children...)
}

// categoryChips picks a category of the company's chips; nil when it
// makes only one kind.
func (s *Screen) categoryChips(cat *model.Catalog, all []*model.Chip) giode.Element {
	counts := map[string]int{}
	for _, c := range all {
		counts[c.Category]++
	}
	if len(counts) < 2 {
		return nil
	}
	chips := []giode.Element{s.chip("cat", "", tr.T("common.all"), len(all), s.category == "", func() { s.category = "" })}
	for _, c := range cat.Schema.Categories {
		if counts[c.ID] == 0 {
			continue
		}
		id := c.ID
		chips = append(chips, s.chip("cat", id, tr.CategoryName(c), counts[id], s.category == id, func() { s.category = id }))
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 8}, chips...)
}

// segmentChips picks a segment (Desktop, Laptop, Server, ...) within the
// category; nil when there is only one.
func (s *Screen) segmentChips(cat *model.Catalog) giode.Element {
	chips := cat.ChipsIn(s.id, s.category, "")
	segments := model.Segments(chips)
	if len(segments) < 2 {
		return nil
	}
	row := []giode.Element{s.chip("seg", "", tr.T("common.all"), len(chips), s.segment == "", func() { s.segment = "" })}
	for _, seg := range segments {
		seg := seg
		n := len(cat.ChipsIn(s.id, s.category, seg))
		row = append(row, s.chip("seg", seg, tr.Or("market."+seg, seg), n, s.segment == seg, func() { s.segment = seg }))
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 8}, row...)
}

func (s *Screen) chip(group, id, label string, n int, selected bool, onClick func()) giode.Element {
	return components.FilterChip("maker/"+s.id+"/"+group, id, label, strconv.Itoa(n), selected, onClick)
}
