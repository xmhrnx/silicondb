// Package maker shows one manufacturer's chips, by category, segment and
// generation.
package maker

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/lazylist"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// row is one line of the list: the header, a generation title or a chip.
type row struct {
	gen  *model.Generation
	chip *model.Chip
}

// Screen is a manufacturer's page.
type Screen struct {
	env      *env.Env
	id       string
	back     *components.RoundButton
	list     *lazylist.List
	category string
	segment  string
	rows     []row
}

// New returns the page of the manufacturer with id, showing the category
// given ("" for all of them).
func New(e *env.Env, id, category string) *Screen {
	s := &Screen{env: e, id: id, category: category}
	s.back = components.NewRoundButton("chevron-left", tr.T("common.back"), e.Routes.Back)
	s.list = giode.LazyList(
		func() int { return len(s.rows) + 1 },
		s.key,
		s.row,
		giode.Styles{Gap: 10, ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}},
	)
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element {
	cat := s.env.Store.Catalog.Get()
	chips := cat.ChipsIn(s.id, s.category, "")
	if s.segment != "" && len(cat.ChipsIn(s.id, s.category, s.segment)) == 0 {
		s.segment = ""
	}
	if s.segment != "" {
		chips = cat.ChipsIn(s.id, s.category, s.segment)
	}
	s.rows = s.rows[:0]
	for _, g := range model.Generations(chips) {
		s.rows = append(s.rows, row{gen: g})
		for _, c := range g.Chips {
			s.rows = append(s.rows, row{chip: c})
		}
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Justify: giode.JustifyCenter},
		giode.Box(giode.Styles{Width: -1, Height: -1, MaxWidth: components.MaxContentWidth}, s.list))
}

func (s *Screen) key(i int) any {
	if i == 0 {
		return "maker-header"
	}
	if r := s.rows[i-1]; r.chip != nil {
		return r.chip
	}
	return "gen:" + s.rows[i-1].gen.Name
}

func (s *Screen) row(i int) giode.Element {
	cat := s.env.Store.Catalog.Get()
	pad := func(el giode.Element) giode.Element {
		return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(20, 0)}, el)
	}
	if i == 0 {
		m := cat.Manufacturer(s.id)
		if m == nil {
			return pad(giode.Stack(giode.Styles{Width: -1},
				components.AppBar(s.back, tr.T("common.not_found"), ""),
				components.EmptyState("factory", tr.T("maker.missing.title"), tr.T("maker.missing.message"), nil)))
		}
		return pad(s.header(cat, m))
	}
	r := s.rows[i-1]
	if r.gen != nil {
		return pad(components.GenerationHeader(r.gen))
	}
	id := r.chip.ID
	return pad(components.ChipRow("maker", cat, r.chip, func() { s.env.Routes.OpenChip(id) }))
}
