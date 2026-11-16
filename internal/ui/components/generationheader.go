package components

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// GenerationHeader titles a generation's chips: its name, its years and
// how many chips it has.
func GenerationHeader(g *model.Generation) giode.Element {
	years := ""
	switch {
	case g.First != 0 && g.First == g.Last:
		years = strconv.Itoa(g.First)
	case g.First != 0:
		years = strconv.Itoa(g.First) + "–" + strconv.Itoa(g.Last)
	}
	return giode.Row(giode.Styles{
		Width: -1, Gap: 10, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
		Padding: giode.Inset{Top: 18, Bottom: 2},
	},
		giode.Box(giode.Styles{Width: 4, Height: 34, BorderRadius: theme.RadiusPill, Background: theme.Primary}),
		giode.Stack(giode.Styles{FlexGrow: 1, Gap: 2},
			giode.Text(g.Name, theme.Heading().With(giode.Styles{FontSize: 17})),
			giode.Text(years, theme.Caption()),
		),
		giode.Text(tr.Count("count.chips", len(g.Chips)), theme.Caption().With(giode.Styles{FontWeight: giode.FontWeightSemiBold})),
	)
}
