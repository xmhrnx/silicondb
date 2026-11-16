package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// SpecRow is one label and value line of a spec section. last omits the
// divider below it.
func SpecRow(label, value string, last bool) giode.Element {
	row := giode.Row(giode.Styles{
		Width: -1, Gap: 16, Justify: giode.JustifySpaceBetween, Align: giode.AlignStart, Wrap: giode.WrapNoWrap,
		Padding: giode.SymmetricInset(0, 11),
	},
		giode.Box(giode.Styles{MaxWidth: 170, FlexGrow: 0},
			giode.Text(label, giode.Styles{FontFamily: theme.Family, FontSize: 14, Color: theme.Muted})),
		giode.Box(giode.Styles{FlexGrow: 1, Justify: giode.JustifyEnd},
			giode.Text(value, giode.Styles{
				FontFamily: theme.Family, FontSize: 14, FontWeight: giode.FontWeightMedium, Color: theme.Text,
				TextAlign: giode.TextAlignEnd,
			})),
	)
	if last {
		return row
	}
	return giode.Stack(giode.Styles{Width: -1}, row, giode.Divider(giode.Styles{Color: theme.Border}))
}
