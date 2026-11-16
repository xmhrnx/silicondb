package home

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
)

// chipList is a titled section of chip rows; trailing goes next to the
// title.
func chipList(e *env.Env, scope, title string, cat *model.Catalog, chips []*model.Chip, trailing giode.Element) giode.Element {
	if len(chips) == 0 {
		return nil
	}
	rows := []giode.Element{components.SectionHeader(title, trailing)}
	for _, c := range chips {
		id := c.ID
		rows = append(rows, components.ChipRow(scope, cat, c, func() { e.Routes.OpenChip(id) }))
	}
	return giode.Stack(giode.Styles{Width: -1, Gap: 10}, rows...)
}
