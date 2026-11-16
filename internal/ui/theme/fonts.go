package theme

import (
	"gioui.org/font"
	"gioui.org/font/opentype"
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/assets"
)

// Family is the typeface every text uses.
const Family = "Inter"

// LoadFonts registers Inter in the weights the UI uses, falling back to
// the Go fonts if they can't be read.
func LoadFonts() {
	weights := map[string]font.Weight{
		"fonts/Inter-Regular.ttf":  font.Normal,
		"fonts/Inter-Medium.ttf":   font.Medium,
		"fonts/Inter-SemiBold.ttf": font.SemiBold,
		"fonts/Inter-Bold.ttf":     font.Bold,
	}
	var faces []font.FontFace
	for file, w := range weights {
		b, err := assets.Fonts.ReadFile(file)
		if err != nil {
			continue
		}
		face, err := opentype.Parse(b)
		if err != nil {
			continue
		}
		faces = append(faces, font.FontFace{Font: font.Font{Typeface: Family, Weight: w}, Face: face})
	}
	if len(faces) == 0 {
		giode.UseGofont()
		return
	}
	giode.RegisterFonts(faces)
}
