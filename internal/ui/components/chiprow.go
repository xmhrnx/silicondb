package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/format"
	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// ChipRow is a list row for a chip: its category icon, name, maker and
// year, and its headline spec. Tapping it calls onClick. scope keeps the
// press state of a chip apart when it appears in two lists at once.
func ChipRow(scope string, cat *model.Catalog, chip *model.Chip, onClick func()) giode.Element {
	return giode.Keyed("chiprow/"+scope, chip, func() *giode.PressableElement {
		return giode.Pressable(func(s giode.PressState) giode.Element {
			return chipRowContent(cat, chip, s)
		}).Label(chip.Name).OnClick(func(giode.ClickEvent) { onClick() })
	})
}

func chipRowContent(cat *model.Catalog, chip *model.Chip, s giode.PressState) giode.Element {
	c := cat.Schema.Category(chip.Category)
	bg := theme.Surface
	if s.Pressed {
		bg = theme.Pressed()
	} else if s.Hovered {
		bg = theme.Raised()
	}
	sub := cat.MakerName(chip.Manufacturer)
	if g := chip.GenerationName(); g != "" && g != chip.Name {
		sub += " · " + g
	}

	var trailing giode.Element = giode.Box(giode.Styles{})
	if keys := cat.Schema.HighlightsFor(chip, 1); len(keys) > 0 {
		f := cat.Schema.Field(keys[0])
		trailing = giode.Stack(giode.Styles{Align: giode.AlignEnd, Gap: 2},
			giode.Text(tr.Value(f, keys[0], chip.Specs[keys[0]]), giode.Styles{
				FontFamily: theme.Family, FontSize: 13.5, FontWeight: giode.FontWeightSemiBold, Color: theme.Text,
				MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis,
			}),
			giode.Text(tr.Label(f, keys[0]), giode.Styles{FontFamily: theme.Family, FontSize: 11, Color: theme.Muted, MaxLines: 1}),
		)
	}

	return giode.Row(giode.Styles{
		Width: -1, Align: giode.AlignCenter, Gap: 14, Wrap: giode.WrapNoWrap,
		Padding:      giode.SymmetricInset(14, 12),
		Background:   bg,
		BorderRadius: 18,
		BorderWidth:  1,
		BorderColor:  theme.Border,
	},
		IconBadge(c.Icon, c.Color, 44),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
			giode.Text(chip.Name, giode.Styles{
				FontFamily: theme.Family, FontSize: 15.5, FontWeight: giode.FontWeightSemiBold, Color: theme.Text,
				MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis,
			}),
			giode.Text(sub, giode.Styles{
				FontFamily: theme.Family, FontSize: 12.5, Color: theme.Muted,
				MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis,
			}),
			giode.Row(giode.Styles{Gap: 6, Align: giode.AlignCenter, Padding: giode.Inset{Top: 3}},
				CategoryBadge(c),
				giode.Text(format.Date(chip.Released), giode.Styles{FontFamily: theme.Family, FontSize: 11.5, Color: theme.Muted}),
			),
		),
		giode.Box(giode.Styles{MaxWidth: 120}, trailing),
	)
}
