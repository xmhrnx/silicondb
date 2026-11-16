package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// SegmentedStyles themes a segmented control: an opaque track with a
// raised selected segment. Call it every frame so it follows the mode.
func SegmentedStyles() giode.Styles {
	return giode.Styles{
		Width: -1, FontFamily: theme.Family, FontSize: 14,
		Background: theme.Surface, BorderColor: theme.Pressed(), Color: theme.Text,
	}
}
