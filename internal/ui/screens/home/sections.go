package home

import (
	"sort"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

const sectionSize = 5

// recent lists the chips the user opened lately.
func recent(e *env.Env, cat *model.Catalog) giode.Element {
	var chips []*model.Chip
	for _, id := range e.Store.Recent() {
		if c := cat.Chip(id); c != nil && len(chips) < 3 {
			chips = append(chips, c)
		}
	}
	return chipList(e, "recent", tr.T("home.recent"), cat, chips, nil)
}

// latest lists the newest chips.
func latest(e *env.Env, cat *model.Catalog) giode.Element {
	chips := cat.Chips[:min(sectionSize, len(cat.Chips))]
	return chipList(e, "latest", tr.T("home.latest"), cat, chips,
		components.TextLink("home-latest", tr.T("common.see_all"), func() { e.Routes.Explore(store.Query{}) }))
}

// fastest lists the processors with the highest benchmark score.
func fastest(e *env.Env, cat *model.Catalog) giode.Element {
	chips := store.Query{Sort: store.Fastest}.Run(cat)
	var top []*model.Chip
	for _, c := range chips {
		if _, ok := c.Number(store.SpeedKey); !ok || len(top) == sectionSize {
			break
		}
		top = append(top, c)
	}
	return chipList(e, "fastest", tr.T("home.fastest"), cat, top,
		components.TextLink("home-fastest", tr.T("common.see_all"), func() { e.Routes.Explore(store.Query{Sort: store.Fastest}) }))
}

// pioneers lists the oldest chips: where it all began.
func pioneers(e *env.Env, cat *model.Catalog) giode.Element {
	n := len(cat.Chips)
	var chips []*model.Chip
	for i := n - 1; i >= 0 && len(chips) < sectionSize; i-- {
		chips = append(chips, cat.Chips[i])
	}
	return chipList(e, "pioneers", tr.T("home.pioneers"), cat, chips,
		components.TextLink("home-pioneers", tr.T("common.see_all"), func() { e.Routes.Explore(store.Query{Sort: store.OldestFirst}) }))
}

// topMakers lists the manufacturers with the most chips.
func topMakers(e *env.Env, cat *model.Catalog) giode.Element {
	makers := append([]*model.Manufacturer(nil), cat.Manufacturers...)
	sort.SliceStable(makers, func(i, j int) bool {
		return len(cat.ChipsBy(makers[i].ID)) > len(cat.ChipsBy(makers[j].ID))
	})
	makers = makers[:min(sectionSize, len(makers))]
	rows := []giode.Element{components.SectionHeader(tr.T("home.top_makers"),
		components.TextLink("home-makers", tr.T("common.see_all"), func() { e.Routes.Makers() }))}
	for _, m := range makers {
		id := m.ID
		rows = append(rows, components.MakerRow(m, len(cat.ChipsBy(id)), func() { e.Routes.OpenMaker(id, "") }))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10}, rows...)
}
