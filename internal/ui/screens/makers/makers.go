// Package makers lists the manufacturers in the database.
package makers

import (
	"sort"
	"strings"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is the Manufacturers tab.
type Screen struct {
	env    *env.Env
	scroll *scroll.Scroll
	search *components.SearchBar
	order  *components.Segmented
}

// New returns the Manufacturers screen.
func New(e *env.Env) *Screen {
	s := &Screen{env: e}
	s.search = components.NewSearchBar("makers.search_hint", func(string) {})
	s.order = components.NewSegmented(nil, "sort.name", "sort.most_chips", "sort.oldest")
	s.scroll = giode.Scroll(s.content, giode.Styles{ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}})
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element { return s.scroll }

func (s *Screen) content() giode.Element {
	cat := s.env.Store.Catalog.Get()
	makers := s.filtered(cat)
	children := []giode.Element{
		components.AppBar(nil, tr.T("makers.title"), tr.Count("count.companies", len(cat.Manufacturers))),
		s.search.View(),
		s.order.View(),
	}
	for _, m := range makers {
		id := m.ID
		children = append(children, components.MakerRow(m, len(cat.ChipsBy(id)), func() { s.env.Routes.OpenMaker(id, "") }))
	}
	if len(makers) == 0 {
		children = append(children, components.EmptyState("factory", tr.T("makers.empty.title"), tr.T("makers.empty.message"), nil))
	}
	return components.Column(10, children...)
}

// filtered applies the search text and the order.
func (s *Screen) filtered(cat *model.Catalog) []*model.Manufacturer {
	q := strings.ToLower(strings.TrimSpace(s.search.Text()))
	var out []*model.Manufacturer
	for _, m := range cat.Manufacturers {
		if q == "" || strings.Contains(strings.ToLower(m.Name+" "+m.Country), q) {
			out = append(out, m)
		}
	}
	switch s.order.Selected() {
	case 1:
		sort.SliceStable(out, func(i, j int) bool { return len(cat.ChipsBy(out[i].ID)) > len(cat.ChipsBy(out[j].ID)) })
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
