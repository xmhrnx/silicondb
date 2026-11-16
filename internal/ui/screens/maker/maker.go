// Package maker shows a manufacturer and its domains: the segments of
// each kind of chip it makes, such as desktop CPUs or server CPUs.
package maker

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"
	"github.com/xmhrnx/giode/i18n"

	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is a manufacturer's page.
type Screen struct {
	env      *env.Env
	id       string
	category string
	back     *components.RoundButton
	scroll   *scroll.Scroll
}

// New returns the page of the manufacturer with id, listing the domains
// of one category ("" for all of them).
func New(e *env.Env, id, category string) *Screen {
	s := &Screen{env: e, id: id, category: category}
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
	m := cat.Manufacturer(s.id)
	if m == nil {
		return components.Column(16,
			components.AppBar(s.back, tr.T("common.not_found"), ""),
			components.EmptyState("factory", tr.T("maker.missing.title"), tr.T("maker.missing.message"), nil))
	}
	children := []giode.Element{components.AppBar(s.back, "", ""), header(cat, m)}
	for _, d := range cat.Domains(s.id, s.category) {
		d := d
		key := s.id + "/" + d.Category + "/" + d.Segment
		children = append(children, components.DomainRow(key, cat.Schema.Category(d.Category), d.Segment,
			len(cat.GenerationsIn(s.id, d.Category, d.Segment)), len(d.Chips),
			func() { s.env.Routes.OpenDomain(s.id, d.Category, d.Segment) }))
	}
	return components.Column(10, children...)
}
