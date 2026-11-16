// Package home is the start screen: an overview of the database with
// shortcuts into it.
package home

import (
	"fmt"
	"strconv"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"github.com/mehran9675/silicondb/internal/format"
	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is the home tab.
type Screen struct {
	env    *env.Env
	scroll *scroll.Scroll
	search *giode.PressableElement
}

// New returns the home screen.
func New(e *env.Env) *Screen {
	s := &Screen{env: e}
	s.scroll = giode.Scroll(s.content, giode.Styles{ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}})
	s.search = searchShortcut(func() { e.Routes.Explore(store.Query{}) })
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element { return s.scroll }

func (s *Screen) content() giode.Element {
	cat := s.env.Store.Catalog.Get()
	first, last := cat.YearSpan()
	years := "—"
	if first != 0 {
		years = strconv.Itoa(first) + "–" + strconv.Itoa(last)
	}
	return components.Column(18,
		giode.Box(giode.Styles{Height: 10}),
		components.HeroCard(tr.T("app.name"), tr.T("app.tagline"), []components.HeroStat{
			{Value: format.Number(float64(len(cat.Chips)), 0), Label: tr.T("home.stat.chips")},
			{Value: fmt.Sprint(len(cat.Manufacturers)), Label: tr.T("home.stat.makers")},
			{Value: years, Label: tr.T("home.stat.years")},
		}),
		s.search,
		components.SyncBanner(s.env.Store.Status.Get()),
		categories(s.env, cat),
		recent(s.env, cat),
		latest(s.env, cat),
		fastest(s.env, cat),
		pioneers(s.env, cat),
		topMakers(s.env, cat),
		giode.Text(tr.T("home.data_updated", "date", format.Date(cat.Updated)), theme.Caption().With(giode.Styles{TextAlign: giode.TextAlignCenter, Width: -1})),
	)
}
