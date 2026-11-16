package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// Card is a rounded surface with a hairline border that groups content.
func Card(st giode.Styles, children ...giode.Element) giode.Element {
	base := giode.Styles{
		FlexDirection: giode.FlexDirectionColumn,
		Width:         -1,
		Padding:       giode.UniformInset(16),
		Background:    theme.Surface,
		BorderRadius:  theme.RadiusCard,
		BorderWidth:   1,
		BorderColor:   theme.Border,
	}
	return giode.Box(base.With(st), children...)
}
