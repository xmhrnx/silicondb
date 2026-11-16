package components

import (
	"strconv"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// GenerationRow is a list row for a product generation: its name, its
// years and how many chips it has. The click handler is set every frame.
func GenerationRow(key string, g *model.Generation, color string, onClick func()) giode.Element {
	p := giode.Keyed("generationrow", key, func() *giode.PressableElement { return giode.Pressable(nil) })
	years := ""
	switch {
	case g.First != 0 && g.First == g.Last:
		years = strconv.Itoa(g.First)
	case g.First != 0:
		years = strconv.Itoa(g.First) + "–" + strconv.Itoa(g.Last)
	}
	return p.Content(func(s giode.PressState) giode.Element {
		return listRow(s,
			giode.Box(giode.Styles{Width: 4, Height: 40, BorderRadius: theme.RadiusPill, Background: color}),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Text(g.Name, theme.Strong().With(giode.Styles{FontSize: 16})),
				giode.Text(years, theme.Caption()),
			),
			countPill(len(g.Chips)),
		)
	}).Label(g.Name).OnClick(func(giode.ClickEvent) { onClick() })
}
