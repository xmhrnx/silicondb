package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// CategoryBadge is a small pill naming a chip's category in its color.
func CategoryBadge(c model.Category) giode.Element {
	return giode.Box(giode.Styles{
		Padding:      giode.SymmetricInset(9, 3),
		Background:   theme.Tint(c.Color, 0.18),
		BorderRadius: theme.RadiusPill,
	}, giode.Text(tr.CategoryName(c), giode.Styles{
		FontFamily: theme.Family, FontSize: 11, FontWeight: giode.FontWeightSemiBold, Color: c.Color,
	}))
}
