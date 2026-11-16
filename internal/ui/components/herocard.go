package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// HeroStat is a figure on the hero card.
type HeroStat struct {
	Value, Label string
}

// HeroCard is the gradient banner at the top of the home screen.
func HeroCard(title, tagline string, stats []HeroStat) giode.Element {
	cells := make([]giode.Element, len(stats))
	for i, s := range stats {
		cells[i] = giode.Stack(giode.Styles{Gap: 2},
			giode.Text(s.Value, giode.Styles{FontFamily: theme.Family, FontSize: 22, FontWeight: giode.FontWeightBold, Color: "#FFFFFF"}),
			giode.Text(s.Label, giode.Styles{FontFamily: theme.Family, FontSize: 12, FontWeight: giode.FontWeightMedium, Color: "#E9E3FF"}),
		)
	}
	return giode.Box(giode.Styles{
		FlexDirection:      giode.FlexDirectionColumn,
		Width:              -1,
		Gap:                18,
		Padding:            giode.UniformInset(22),
		BorderRadius:       26,
		Background:         "#7C3AED",
		BackgroundGradient: theme.BrandGradient(125),
		BoxShadow:          giode.BoxShadow{Y: 12, Blur: 32, Color: "#7C3AED40"},
		Overflow:           giode.OverflowHidden,
	},
		giode.Row(giode.Styles{Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
			giode.Box(giode.Styles{
				Width: 44, Height: 44, Align: giode.AlignCenter, Justify: giode.JustifyCenter,
				Background: giode.Mix("#FFFFFF", "#7C3AED", 0.22), BorderRadius: 14,
			}, theme.Icon("cpu", 24, "#FFFFFF")),
			giode.Stack(giode.Styles{Gap: 2},
				giode.Text(title, giode.Styles{FontFamily: theme.Family, FontSize: 26, FontWeight: giode.FontWeightBold, Color: "#FFFFFF"}),
				giode.Text(tagline, giode.Styles{FontFamily: theme.Family, FontSize: 13, Color: "#EDE9FE"}),
			),
		),
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifySpaceBetween, Wrap: giode.WrapNoWrap}, cells...),
	)
}
