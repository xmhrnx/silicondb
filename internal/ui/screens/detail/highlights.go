package detail

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// highlights shows the chip's headline specs as tiles.
func highlights(cat *model.Catalog, chip *model.Chip) giode.Element {
	keys := cat.Schema.HighlightsFor(chip, 4)
	if len(keys) == 0 {
		return nil
	}
	color := cat.Schema.Category(chip.Category).Color
	tiles := make([]giode.Element, len(keys))
	for i, k := range keys {
		f := cat.Schema.Field(k)
		tiles[i] = components.StatTile(tr.Label(f, k), tr.Value(f, k, chip.Specs[k]), color)
	}
	cols := len(keys)
	if giode.Compact() && cols == 4 {
		cols = 2
	}
	return components.Grid(cols, 10, tiles...)
}
