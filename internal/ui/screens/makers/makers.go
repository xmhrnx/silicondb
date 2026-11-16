// Package makers lists the manufacturers in the database.
package makers

import (
	"sort"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/lazylist"
	"github.com/xmhrnx/giode/i18n"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

const headerKey = "makers-header"

// Screen is the Manufacturers tab. Its list is a LazyList: only rows near
// the viewport are built, and unchanged rows replay their drawing.
type Screen struct {
	env    *env.Env
	list   *lazylist.List
	search *components.SearchBar
	order  *components.Segmented

	// The filtered, ordered makers, and what they were worked out for.
	makers []*model.Manufacturer
	ranFor filter
}

// filter is what the list shows: it's recomputed when one of these
// changes, not every frame.
type filter struct {
	cat   *model.Catalog
	text  string
	order int
}

// rowVersion changes when a row's drawing would: its data or the
// language.
type rowVersion struct {
	maker  *model.Manufacturer
	count  int
	locale uint64
}

// New returns the Manufacturers screen.
func New(e *env.Env) *Screen {
	s := &Screen{env: e}
	s.search = components.NewSearchBar("makers.search_hint", func(string) {})
	s.order = components.NewSegmented(nil, "sort.name", "sort.most_chips", "sort.oldest")
	s.list = giode.LazyList(
		func() int { return len(s.makers) + 2 },
		s.key,
		s.row,
		giode.Styles{Gap: 10, ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}},
	).Version(s.version)
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element {
	f := filter{cat: s.env.Store.Catalog.Get(), text: strings.ToLower(strings.TrimSpace(s.search.Text())), order: s.order.Selected()}
	if f != s.ranFor {
		s.makers = filtered(f)
		s.ranFor = f
	}
	return giode.Box(giode.Styles{Width: -1, Height: -1, Justify: giode.JustifyCenter},
		giode.Box(giode.Styles{Width: -1, Height: -1, MaxWidth: components.MaxContentWidth}, s.list))
}

func (s *Screen) key(i int) any {
	switch {
	case i == 0:
		return headerKey
	case i > len(s.makers):
		return "makers-footer"
	}
	return s.makers[i-1]
}

// version changes when a row's drawing would. Rows replay their last
// drawing otherwise; the search field and sort control in the header
// still respond, since a cached row rebuilds when a control in it changes
// its look or holds focus.
func (s *Screen) version(i int) any {
	switch {
	case i == 0:
		return rowVersion{count: len(s.ranFor.cat.Manufacturers), locale: i18n.Version()}
	case i > len(s.makers):
		return rowVersion{count: len(s.makers), locale: i18n.Version()}
	}
	m := s.makers[i-1]
	return rowVersion{maker: m, count: len(s.ranFor.cat.ChipsBy(m.ID)), locale: i18n.Version()}
}

func (s *Screen) row(i int) giode.Element {
	pad := func(el giode.Element) giode.Element {
		return giode.Box(giode.Styles{Width: -1, Padding: giode.SymmetricInset(20, 0)}, el)
	}
	cat := s.ranFor.cat
	switch {
	case i == 0:
		return pad(giode.Stack(giode.Styles{Width: -1, Gap: 10},
			components.AppBar(nil, tr.T("makers.title"), tr.Count("count.companies", len(cat.Manufacturers))),
			s.search.View(),
			s.order.View(),
		))
	case i > len(s.makers):
		if len(s.makers) == 0 {
			return pad(components.EmptyState("factory", tr.T("makers.empty.title"), tr.T("makers.empty.message"), nil))
		}
		return giode.Box(giode.Styles{Height: 24})
	}
	m := s.makers[i-1]
	id := m.ID
	return pad(components.MakerRow(m, len(cat.ChipsBy(id)), func() { s.env.Routes.OpenMaker(id, "") }))
}

// filtered applies the search text and the order.
func filtered(f filter) []*model.Manufacturer {
	var out []*model.Manufacturer
	for _, m := range f.cat.Manufacturers {
		if f.text == "" || strings.Contains(strings.ToLower(m.Name+" "+m.Country), f.text) {
			out = append(out, m)
		}
	}
	switch f.order {
	case 1:
		sort.SliceStable(out, func(i, j int) bool { return len(f.cat.ChipsBy(out[i].ID)) > len(f.cat.ChipsBy(out[j].ID)) })
	case 2:
		sort.SliceStable(out, func(i, j int) bool { return founded(out[i]) < founded(out[j]) })
	}
	return out
}

func founded(m *model.Manufacturer) int {
	if m.Founded == 0 {
		return 1 << 30
	}
	return m.Founded
}
