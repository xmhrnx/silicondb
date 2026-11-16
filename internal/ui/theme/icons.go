package theme

import (
	"log"

	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/styles/properties"

	"github.com/mehran9675/silicondb/assets"
)

// LoadIcons registers the SVG icons by file name ("icons/cpu.svg" is
// "cpu").
func LoadIcons() {
	if err := giode.RegisterIconsFS(assets.Icons, "icons"); err != nil {
		log.Println("icons:", err)
	}
}

// Icon renders a registered icon at size dp in color.
func Icon(name string, size int, color string) giode.Element {
	return giode.Icon(name, giode.Styles{Width: properties.Width(size), Height: properties.Height(size), Color: color})
}
