package explore

import (
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/chip"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// header is the title, search field and filters above the results.
func (s *Screen) header(cat *model.Catalog) giode.Element {
	children := []giode.Element{
		components.AppBar(nil, tr.T("explore.title"), tr.Count("count.chips", len(s.results))),
		s.search.View(),
		s.categoryChips(cat),
		s.decadeChips(cat),
	}
	if s.query.Maker != "" {
		children = append(children, s.makerChip(cat))
	}
	children = append(children, s.sort.View(), giode.Box(giode.Styles{Height: 2}))
	return giode.Stack(giode.Styles{Width: -1, Gap: 12}, children...)
}

// categoryChips filters by category.
func (s *Screen) categoryChips(cat *model.Catalog) giode.Element {
	counts := cat.CountByCategory()
	chips := []giode.Element{s.filterChip("cat", "", tr.T("common.all"), s.query.Category == "", func() { s.query.Category = "" })}
	for _, c := range cat.Schema.Categories {
		if counts[c.ID] == 0 {
			continue
		}
		id := c.ID
		chips = append(chips, s.filterChip("cat", id, tr.CategoryName(c), s.query.Category == id, func() {
			if s.query.Category == id {
				s.query.Category = ""
			} else {
				s.query.Category = id
			}
		}))
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 8}, chips...)
}

// decadeChips filters by decade of release.
func (s *Screen) decadeChips(cat *model.Catalog) giode.Element {
	var chips []giode.Element
	for _, d := range store.Decades(cat) {
		d := d
		chips = append(chips, s.filterChip("decade", fmt.Sprint(d), tr.T("explore.decade", "decade", fmt.Sprint(d)), s.query.Decade == d, func() {
			if s.query.Decade == d {
				s.query.Decade = 0
			} else {
				s.query.Decade = d
			}
		}))
	}
	return giode.Row(giode.Styles{Width: -1, Gap: 8}, chips...)
}

// makerChip shows the manufacturer filter, removable.
func (s *Screen) makerChip(cat *model.Catalog) giode.Element {
	c := giode.Keyed("explore-maker", s.query.Maker, func() *chip.Chip {
		return giode.Chip(cat.MakerName(s.query.Maker)).
			Icon(theme.Icon("factory", 14, theme.OnPrimary)).
			Selected(true).
			OnRemove(func() { s.query.Maker = "" })
	})
	return giode.Row(giode.Styles{Gap: 8}, c)
}

// filterChip is a selectable chip, kept per (group, id).
func (s *Screen) filterChip(group, id, label string, selected bool, onClick func()) giode.Element {
	return components.FilterChip("explore-"+group, id, label, "", selected, onClick)
}
