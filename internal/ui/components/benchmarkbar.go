package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// BenchmarkBar shows a benchmark score with a bar scaled to the best
// score in the database.
func BenchmarkBar(label, score string, fraction float32, color string) giode.Element {
	if fraction < 0.02 {
		fraction = 0.02
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 7, Padding: giode.SymmetricInset(0, 9)},
		giode.Row(giode.Styles{Width: -1, Justify: giode.JustifySpaceBetween, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap, Gap: 12},
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(label, giode.Styles{FontFamily: theme.Family, FontSize: 14, Color: theme.Muted})),
			giode.Text(score, giode.Styles{FontFamily: theme.Family, FontSize: 14, FontWeight: giode.FontWeightBold, Color: theme.Text}),
		),
		giode.Progress(fraction, giode.Styles{Height: 8, Color: color, Background: theme.Faint()}),
	)
}
