// Package generation lists the chips of one product generation.
package generation

import (
	"strconv"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is a generation's page.
type Screen struct {
	env                            *env.Env
	maker, category, segment, name string
	back                           *components.RoundButton
	scroll                         *scroll.Scroll
}

// New returns the page of the generation called name among a
// manufacturer's chips of a category and segment.
func New(e *env.Env, maker, category, segment, name string) *Screen {
	s := &Screen{env: e, maker: maker, category: category, segment: segment, name: name}
	s.back = components.NewRoundButton("chevron-left", tr.T("common.back"), e.Routes.Back)
	s.scroll = giode.Scroll(s.content, giode.Styles{ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}})
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element { return s.scroll }

func (s *Screen) content() giode.Element {
	cat := s.env.Store.Catalog.Get()
	var gen *model.Generation
	for _, g := range model.Generations(cat.ChipsIn(s.maker, s.category, s.segment)) {
		if g.Name == s.name {
			gen = g
		}
	}
	if gen == nil {
		return components.Column(16,
			components.AppBar(s.back, s.name, ""),
			components.EmptyState("search", tr.T("explore.empty.title"), "", nil))
	}
	sub := cat.MakerName(s.maker) + " · " + tr.Or("market."+s.segment, s.segment)
	if gen.First != 0 {
		years := strconv.Itoa(gen.First)
		if gen.Last != gen.First {
			years += "–" + strconv.Itoa(gen.Last)
		}
		sub += " · " + years
	}
	children := []giode.Element{components.AppBar(s.back, s.name, sub)}
	for _, c := range gen.Chips {
		id := c.ID
		children = append(children, components.ChipRow("generation", cat, c, func() { s.env.Routes.OpenChip(id) }))
	}
	return components.Column(10, children...)
}
