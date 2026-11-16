// Package explore is the searchable, filterable list of every chip.
package explore

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/lazylist"
	"github.com/xmhrnx/giode/i18n"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// headerKey identifies the filter header, the list's first row.
const headerKey = "explore-header"

// Screen is the Explore tab.
type Screen struct {
	env    *env.Env
	search *components.SearchBar
	sort   *components.Segmented
	list   *lazylist.List

	query   store.Query
	results []*model.Chip
	ranFor  store.Query
	ranOn   *model.Catalog
}

// New returns the Explore screen.
func New(e *env.Env) *Screen {
	s := &Screen{env: e}
	s.search = components.NewSearchBar("explore.search_hint", func(string) {})
	s.sort = components.NewSegmented(func(i int) { s.query.Sort = store.SortOrder(i) }, store.SortKeys...)
	s.list = giode.LazyList(
		func() int { return len(s.results) + 2 },
		s.key,
		s.row,
		giode.Styles{Gap: 10, ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}},
	).Version(s.version)
	return s
}

// Apply replaces the filters with q and scrolls to the top.
func (s *Screen) Apply(q store.Query) {
	s.query = q
	s.search.Field.SetText(q.Text)
	s.sort.SetSelected(int(q.Sort))
	s.list.ScrollToKey(headerKey, giode.LazyAlignStart)
}

// View lays the screen out.
func (s *Screen) View() giode.Element {
	s.query.Text = s.search.Text()
	cat := s.env.Store.Catalog.Get()
	if s.ranOn != cat || s.ranFor != s.query {
		s.results = s.query.Run(cat)
		s.ranOn, s.ranFor = cat, s.query
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Justify: giode.JustifyCenter},
		giode.Box(giode.Styles{Width: -1, Height: -1, MaxWidth: components.MaxContentWidth}, s.list))
}

// rowVersion changes when a row's drawing would; rows replay their last
// drawing otherwise.
type rowVersion struct {
	query  store.Query
	chip   *model.Chip
	n      int
	locale uint64
}

func (s *Screen) version(i int) any {
	switch {
	case i == 0:
		return rowVersion{query: s.query, n: len(s.results), locale: i18n.Version()}
	case i > len(s.results):
		return rowVersion{n: len(s.results), locale: i18n.Version()}
	}
	return rowVersion{chip: s.results[i-1], locale: i18n.Version()}
}

func (s *Screen) key(i int) any {
	switch {
	case i == 0:
		return headerKey
	case i > len(s.results):
		return "explore-footer"
	}
	return s.results[i-1]
}

func (s *Screen) row(i int) giode.Element {
	pad := func(el giode.Element) giode.Element {
		return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(20, 0)}, el)
	}
	cat := s.env.Store.Catalog.Get()
	switch {
	case i == 0:
		return pad(s.header(cat))
	case i > len(s.results):
		if len(s.results) == 0 {
			return pad(components.EmptyState("search", tr.T("explore.empty.title"), tr.T("explore.empty.message"), nil))
		}
		return giode.Box(giode.Styles{Height: 24})
	}
	chip := s.results[i-1]
	id := chip.ID
	return pad(components.ChipRow("explore", cat, chip, func() { s.env.Routes.OpenChip(id) }))
}
