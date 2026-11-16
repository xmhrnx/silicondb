package theme

import "github.com/xmhrnx/giode"

// Spacing and radii, in dp.
const (
	Gutter      = 20
	RadiusCard  = 20
	RadiusTile  = 16
	RadiusPill  = 999
	RadiusSmall = 10
)

// Text styles. Each returns fresh Styles so callers can extend them.

// Display is the largest heading, used on hero cards.
func Display() giode.Styles {
	return giode.Styles{FontFamily: Family, FontSize: 30, FontWeight: giode.FontWeightBold, Color: Text, LineHeight: 36}
}

// Title heads a screen.
func Title() giode.Styles {
	return giode.Styles{FontFamily: Family, FontSize: 26, FontWeight: giode.FontWeightBold, Color: Text}
}

// Heading heads a section.
func Heading() giode.Styles {
	return giode.Styles{FontFamily: Family, FontSize: 18, FontWeight: giode.FontWeightSemiBold, Color: Text}
}

// Body is regular text.
func Body() giode.Styles {
	return giode.Styles{FontFamily: Family, FontSize: 15, Color: Text}
}

// Strong is emphasized body text.
func Strong() giode.Styles {
	return giode.Styles{FontFamily: Family, FontSize: 15, FontWeight: giode.FontWeightSemiBold, Color: Text}
}

// Caption is small secondary text.
func Caption() giode.Styles {
	return giode.Styles{FontFamily: Family, FontSize: 12.5, Color: Muted}
}

// Overline is a small uppercase label above values.
func Overline() giode.Styles {
	return giode.Styles{FontFamily: Family, FontSize: 11, FontWeight: giode.FontWeightSemiBold, Color: Muted, TextTransform: giode.TextTransformUppercase}
}

// Value is a prominent number in stat tiles.
func Value() giode.Styles {
	return giode.Styles{FontFamily: Family, FontSize: 19, FontWeight: giode.FontWeightBold, Color: Text}
}
