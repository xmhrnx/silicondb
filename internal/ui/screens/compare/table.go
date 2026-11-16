package compare

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/silicondb/internal/format"
	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

const (
	labelWidth = 128
	cellWidth  = 156
)

// lowerIsBetter lists numeric specs where the smallest value wins.
var lowerIsBetter = map[string]bool{
	"process_nm": true, "tdp_w": true, "launch_price_usd": true, "die_size_mm2": true, "voltage_v": true,
}

// neutral lists numeric specs where neither direction is better.
var neutral = map[string]bool{"address_width_bits": true}

// tableView is the comparison grid: a header of chips, then one row per
// spec any of them has.
func (s *Screen) tableView() giode.Element {
	cat := s.env.Store.Catalog.Get()
	chips := s.chips()
	rows := []giode.Element{s.headerRow(cat, chips)}
	rows = append(rows, s.textRow(tr.T("compare.category"), chips, func(c *model.Chip) string { return tr.CategoryName(cat.Schema.Category(c.Category)) }))
	rows = append(rows, s.textRow(tr.T("compare.manufacturer"), chips, func(c *model.Chip) string { return cat.MakerName(c.Manufacturer) }))
	rows = append(rows, s.textRow(tr.T("compare.released"), chips, func(c *model.Chip) string { return format.Date(c.Released) }))
	for gi := range cat.Schema.Groups {
		g := &cat.Schema.Groups[gi]
		var groupRows []giode.Element
		for i := range g.Fields {
			f := &g.Fields[i]
			if r := s.specRow(f, chips); r != nil {
				groupRows = append(groupRows, r)
			}
		}
		if len(groupRows) > 0 {
			rows = append(rows, giode.Box(giode.Styles{Padding: giode.Inset{Top: 18, Bottom: 6}},
				giode.Text(tr.Group(g), theme.Overline().With(giode.Styles{Color: theme.Primary}))))
			rows = append(rows, groupRows...)
		}
	}
	return giode.Stack(giode.Styles{}, rows...)
}

func (s *Screen) headerRow(cat *model.Catalog, chips []*model.Chip) giode.Element {
	cells := []giode.Element{giode.Box(giode.Styles{Width: labelWidth})}
	for _, c := range chips {
		cat := cat.Schema.Category(c.Category)
		id := c.ID
		remove := giode.Keyed("compare-remove", id, func() *giode.PressableElement {
			return giode.Pressable(func(p giode.PressState) giode.Element {
				return giode.Box(giode.Styles{Padding: giode.UniformInset(4)}, theme.Icon("close", 16, theme.Muted))
			}).Label(tr.T("common.remove")).OnClick(func(giode.ClickEvent) { s.env.Store.ToggleCompare(id) })
		})
		open := giode.Keyed("compare-open", c, func() *giode.PressableElement {
			return giode.Pressable(func(p giode.PressState) giode.Element {
				return giode.Text(c.Name, theme.Strong().With(giode.Styles{MaxLines: 2, TextOverflow: giode.TextOverflowEllipsis}))
			}).Label(c.Name).OnClick(func(giode.ClickEvent) { s.env.Routes.OpenChip(id) })
		})
		cells = append(cells, giode.Box(giode.Styles{Width: cellWidth, Padding: giode.Inset{Right: 8}},
			components.Card(giode.Styles{Gap: 10, Padding: giode.UniformInset(12), BorderRadius: 16},
				giode.Row(giode.Styles{Width: -1, Justify: giode.JustifySpaceBetween, Align: giode.AlignStart, Wrap: giode.WrapNoWrap},
					components.IconBadge(cat.Icon, cat.Color, 34), remove),
				open,
			)))
	}
	return giode.Row(giode.Styles{Wrap: giode.WrapNoWrap, Padding: giode.Inset{Bottom: 6}}, cells...)
}

func (s *Screen) textRow(label string, chips []*model.Chip, value func(*model.Chip) string) giode.Element {
	vals := make([]string, len(chips))
	for i, c := range chips {
		vals[i] = value(c)
	}
	return row(label, vals, -1)
}

// specRow returns the row for field, or nil if no chip has it. The best
// numeric value is highlighted.
func (s *Screen) specRow(f *model.Field, chips []*model.Chip) giode.Element {
	vals := make([]string, len(chips))
	found, best, bestVal := false, -1, 0.0
	numbers := 0
	for i, c := range chips {
		v, ok := c.Specs[f.Key]
		if !ok {
			vals[i] = "—"
			continue
		}
		found = true
		vals[i] = tr.Value(f, f.Key, v)
		if n, ok := v.(float64); ok && !neutral[f.Key] {
			numbers++
			better := n > bestVal
			if lowerIsBetter[f.Key] {
				better = n < bestVal
			}
			if best < 0 || better {
				best, bestVal = i, n
			}
		}
	}
	if !found {
		return nil
	}
	if numbers < 2 {
		best = -1
	}
	return row(tr.Label(f, f.Key), vals, best)
}

// row draws a label and one cell per chip; the cell at best is
// highlighted.
func row(label string, vals []string, best int) giode.Element {
	cells := []giode.Element{giode.Box(giode.Styles{Width: labelWidth, Padding: giode.Inset{Right: 8}},
		giode.Text(label, giode.Styles{FontFamily: theme.Family, FontSize: 13, Color: theme.Muted}))}
	for i, v := range vals {
		st := giode.Styles{FontFamily: theme.Family, FontSize: 13.5, FontWeight: giode.FontWeightMedium, Color: theme.Text}
		box := giode.Styles{Width: properties.Width(cellWidth - 8), Padding: giode.SymmetricInset(8, 4), BorderRadius: 8}
		if i == best {
			st.Color = theme.Primary
			st.FontWeight = giode.FontWeightBold
			box.Background = theme.Tint(theme.Primary, 0.12)
		}
		cells = append(cells, giode.Box(giode.Styles{Width: cellWidth, Padding: giode.Inset{Right: 8}}, giode.Box(box, giode.Text(v, st))))
	}
	return giode.Stack(giode.Styles{},
		giode.Row(giode.Styles{Wrap: giode.WrapNoWrap, Align: giode.AlignCenter, Padding: giode.SymmetricInset(0, 6)}, cells...),
		giode.Box(giode.Styles{Width: properties.Width(labelWidth + cellWidth*(len(vals))), Height: 1, Background: theme.Border}),
	)
}
