// Package detail shows every specification of one chip.
package detail

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/scroll"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is a chip's specification sheet.
type Screen struct {
	env     *env.Env
	id      string
	scroll  *scroll.Scroll
	back    *components.RoundButton
	compare *components.RoundButton
}

// New returns the screen for the chip with id.
func New(e *env.Env, id string) *Screen {
	s := &Screen{env: e, id: id}
	s.scroll = giode.Scroll(s.content, giode.Styles{ScrollBar: giode.ScrollBar{HiddenMaxWidth: 600}})
	s.back = components.NewRoundButton("chevron-left", tr.T("common.back"), e.Routes.Back)
	s.compare = components.NewRoundButton("compare", tr.T("detail.compare"), s.toggleCompare)
	e.Store.Viewed(id)
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element { return s.scroll }

func (s *Screen) toggleCompare() {
	if s.env.Store.ToggleCompare(s.id) {
		s.env.Toasts.Show(giode.ToastSuccess, tr.T("detail.added"))
	} else {
		s.env.Toasts.Show(giode.ToastInfo, tr.T("detail.removed"))
	}
}

func (s *Screen) content() giode.Element {
	cat := s.env.Store.Catalog.Get()
	chip := cat.Chip(s.id)
	if chip == nil {
		return components.Column(16,
			components.AppBar(s.back, tr.T("common.not_found"), ""),
			components.EmptyState("alert", tr.T("detail.missing.title"), tr.T("detail.missing.message"), nil),
		)
	}
	if s.env.Store.InCompare(s.id) {
		s.compare.Icon("check", theme.Primary)
	} else {
		s.compare.Icon("plus", "")
	}
	children := []giode.Element{
		components.AppBar(s.back, "", "", s.compare),
		hero(s.env, cat, chip),
		highlights(cat, chip),
	}
	if title, scores := benchmarks(cat, chip); len(scores) > 0 {
		color := cat.Schema.Category(chip.Category).Color
		children = append(children, components.BenchmarkSection(title, tr.T("detail.benchmarks_note"), color, scores))
	}
	for _, sec := range sections(cat.Schema, chip) {
		children = append(children, components.SpecSection(sec.title, sec.lines))
	}
	children = append(children, related(s.env, cat, chip))
	return components.Column(16, children...)
}

// related lists other chips of the same generation and manufacturer.
func related(e *env.Env, cat *model.Catalog, chip *model.Chip) giode.Element {
	gen := chip.GenerationName()
	rows := []giode.Element{components.SectionHeader(tr.T("detail.more_in", "family", gen), nil)}
	for _, c := range cat.ChipsBy(chip.Manufacturer) {
		if c == chip || c.GenerationName() != gen || len(rows) > 5 {
			continue
		}
		id := c.ID
		rows = append(rows, components.ChipRow("related", cat, c, func() { e.Routes.OpenChip(id) }))
	}
	if len(rows) == 1 {
		return nil
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10}, rows...)
}
