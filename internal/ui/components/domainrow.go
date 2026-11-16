package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/model"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// DomainRow is a list row for one of a company's domains, such as its
// desktop CPUs: the segment, its category, and how many generations and
// chips it has. The click handler is set every frame.
func DomainRow(key string, cat model.Category, segment string, generations, chips int, onClick func()) giode.Element {
	p := giode.Keyed("domainrow", key, func() *giode.PressableElement { return giode.Pressable(nil) })
	title := tr.Or("market."+segment, segment)
	return p.Content(func(s giode.PressState) giode.Element {
		return listRow(s,
			IconBadge(cat.Icon, cat.Color, 44),
			giode.Stack(giode.Styles{FlexGrow: 1, Gap: 3},
				giode.Text(title, theme.Strong().With(giode.Styles{FontSize: 16})),
				giode.Text(tr.CategoryName(cat)+" · "+tr.Count("count.generations", generations),
					theme.Caption().With(giode.Styles{MaxLines: 1, TextOverflow: giode.TextOverflowEllipsis})),
			),
			countPill(chips),
		)
	}).Label(title).OnClick(func(giode.ClickEvent) { onClick() })
}
