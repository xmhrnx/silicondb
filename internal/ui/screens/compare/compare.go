// Package compare shows chips side by side.
package compare

import (
	"fmt"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"github.com/mehran9675/silicondb/internal/config"
	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is the Compare tab.
type Screen struct {
	env    *env.Env
	scroll *scroll.Scroll
	table  *scroll.Scroll
	browse *giode.ButtonElement
}

// New returns the Compare screen.
func New(e *env.Env) *Screen {
	s := &Screen{env: e}
	s.scroll = giode.Scroll(s.content, giode.Styles{ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}})
	s.table = giode.Scroll(s.tableView).Axis(giode.Horizontal)
	s.browse = giode.Button(tr.T("compare.browse"), giode.Styles{BorderRadius: 14, Padding: giode.SymmetricInset(20, 12)}).
		OnClick(func() { e.Routes.Explore(store.Query{}) })
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element { return s.scroll }

func (s *Screen) chips() []*model.Chip {
	cat := s.env.Store.Catalog.Get()
	var out []*model.Chip
	for _, id := range s.env.Store.Compare() {
		if c := cat.Chip(id); c != nil {
			out = append(out, c)
		}
	}
	return out
}

func (s *Screen) content() giode.Element {
	chips := s.chips()
	sub := tr.T("compare.count", "count", fmt.Sprint(len(chips)), "max", fmt.Sprint(config.MaxCompare))
	var clear giode.Element
	if len(chips) > 0 {
		clear = components.TextLink("compare-clear", tr.T("common.clear"), s.env.Store.ClearCompare)
	}
	var trailing []giode.Element
	if clear != nil {
		trailing = append(trailing, clear)
	}
	bar := components.AppBar(nil, tr.T("compare.title"), sub, trailing...)
	if len(chips) == 0 {
		s.browse.Label(tr.T("compare.browse"))
		return components.Column(16, bar,
			components.EmptyState("compare", tr.T("compare.empty.title"), tr.T("compare.empty.message"), s.browse))
	}
	return components.Column(16, bar, s.table)
}
