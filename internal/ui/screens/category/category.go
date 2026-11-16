// Package category lists the companies making one kind of chip, the first
// step of browsing Category → Company → Segment → Generation.
package category

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/i18n"

	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is a category's page.
type Screen struct {
	env    *env.Env
	id     string
	back   *components.RoundButton
	scroll *scroll.Scroll
}

// New returns the page of the category with id.
func New(e *env.Env, id string) *Screen {
	s := &Screen{env: e, id: id}
	s.back = components.NewRoundButton("chevron-left", tr.T("common.back"), e.Routes.Back)
	s.scroll = giode.Scroll(s.memo, giode.Styles{ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}})
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element { return s.scroll }

// memo replays the screen's last drawing until something it shows
// changes: the catalog or sync status (States, tracked by Memo), the
// language, or the deps passed here.
func (s *Screen) memo() giode.Element {
	return giode.Memo(s, s.content, i18n.Version())
}

func (s *Screen) content() giode.Element {
	cat := s.env.Store.Catalog.Get()
	c := cat.Schema.Category(s.id)
	makers := cat.MakersIn(s.id)
	total := cat.CountByCategory()[s.id]
	children := []giode.Element{
		components.AppBar(s.back, tr.CategoryName(c), tr.CategoryTitle(c)+" · "+tr.Count("count.chips", total)+" · "+tr.Count("count.companies", len(makers)),
			components.TextLink("category-all/"+s.id, tr.T("common.see_all"), func() { s.env.Routes.Explore(store.Query{Category: s.id}) })),
	}
	for _, m := range makers {
		id := m.ID
		n := len(cat.ChipsIn(id, s.id, ""))
		children = append(children, components.MakerRow(m, n, func() { s.env.Routes.OpenMaker(id, s.id) }))
	}
	return components.Column(10, children...)
}
