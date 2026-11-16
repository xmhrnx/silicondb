package detail

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/format"
	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// hero is the chip's banner in its category color: icon, name, maker,
// family and release date.
func hero(e *env.Env, cat *model.Catalog, chip *model.Chip) giode.Element {
	c := cat.Schema.Category(chip.Category)
	maker := giode.Keyed("detail-maker", chip, func() *giode.PressableElement {
		id := chip.Manufacturer
		return giode.Pressable(func(s giode.PressState) giode.Element {
			bg := giode.Mix("#FFFFFF", c.Color, 0.18)
			if s.Pressed {
				bg = giode.Mix("#FFFFFF", c.Color, 0.30)
			}
			return giode.Row(giode.Styles{
				Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
				Padding: giode.SymmetricInset(12, 6), Background: bg, BorderRadius: theme.RadiusPill,
			},
				theme.Icon("factory", 14, "#FFFFFF"),
				giode.Text(cat.MakerName(id), giode.Styles{FontFamily: theme.Family, FontSize: 13, FontWeight: giode.FontWeightSemiBold, Color: "#FFFFFF"}),
				theme.Icon("chevron-right", 14, "#FFFFFF"),
			)
		}).Label(cat.MakerName(id)).OnClick(func(giode.ClickEvent) { e.Routes.OpenMaker(id, chip.Category) })
	})

	meta := tr.CategoryTitle(c)
	if g := chip.GenerationName(); g != "" {
		meta = g + " · " + tr.CategoryName(c)
	}
	return giode.Box(giode.Styles{
		FlexDirection:      giode.FlexDirectionColumn,
		Width:              -1,
		Gap:                14,
		Padding:            giode.UniformInset(22),
		BorderRadius:       26,
		Background:         c.Color,
		BackgroundGradient: theme.AccentGradient(c.Color),
		BoxShadow:          giode.BoxShadow{Y: 12, Blur: 30, Color: c.Color + "40"},
		Overflow:           giode.OverflowHidden,
	},
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			giode.Box(giode.Styles{
				Width: 48, Height: 48, Align: giode.AlignCenter, Justify: giode.JustifyCenter,
				Background: giode.Mix("#FFFFFF", c.Color, 0.22), BorderRadius: 15,
			}, theme.Icon(c.Icon, 26, "#FFFFFF")),
		),
		giode.Stack(giode.Styles{Gap: 4},
			giode.Text(chip.Name, giode.Styles{FontFamily: theme.Family, FontSize: 30, FontWeight: giode.FontWeightBold, Color: "#FFFFFF", LineHeight: 36}),
			giode.Text(meta, giode.Styles{FontFamily: theme.Family, FontSize: 14, Color: giode.Mix("#FFFFFF", c.Color, 0.82)}),
		),
		giode.Row(giode.Styles{Gap: 8, Align: giode.AlignCenter},
			maker,
			giode.Row(giode.Styles{
				Gap: 6, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
				Padding: giode.SymmetricInset(12, 6), Background: giode.Mix("#000000", c.Color, 0.18), BorderRadius: theme.RadiusPill,
			},
				theme.Icon("calendar", 14, "#FFFFFF"),
				giode.Text(format.Date(chip.Released), giode.Styles{FontFamily: theme.Family, FontSize: 13, FontWeight: giode.FontWeightSemiBold, Color: "#FFFFFF"}),
			),
		),
	)
}
