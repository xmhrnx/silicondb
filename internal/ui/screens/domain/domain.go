// Package domain lists the generations of one of a company's domains,
// such as Intel's desktop CPUs.
package domain

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is a domain's page.
type Screen struct {
	env                      *env.Env
	maker, category, segment string
	back                     *components.RoundButton
	scroll                   *scroll.Scroll
}

// New returns the page of a manufacturer's chips of a category and
// segment.
func New(e *env.Env, maker, category, segment string) *Screen {
	s := &Screen{env: e, maker: maker, category: category, segment: segment}
	s.back = components.NewRoundButton("chevron-left", tr.T("common.back"), e.Routes.Back)
	s.scroll = giode.Scroll(s.content, giode.Styles{ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}})
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element { return s.scroll }

func (s *Screen) content() giode.Element {
	cat := s.env.Store.Catalog.Get()
	c := cat.Schema.Category(s.category)
	gens := model.Generations(cat.ChipsIn(s.maker, s.category, s.segment))
	title := tr.Or("market."+s.segment, s.segment)
	sub := cat.MakerName(s.maker) + " · " + tr.CategoryName(c) + " · " + tr.Count("count.generations", len(gens))
	children := []giode.Element{components.AppBar(s.back, title, sub)}
	for _, g := range gens {
		name := g.Name
		key := s.maker + "/" + s.category + "/" + s.segment + "/" + name
		children = append(children, components.GenerationRow(key, g, c.Color, func() {
			s.env.Routes.OpenGeneration(s.maker, s.category, s.segment, name)
		}))
	}
	if len(gens) == 0 {
		children = append(children, components.EmptyState("search", tr.T("explore.empty.title"), "", nil))
	}
	return components.Column(10, children...)
}
