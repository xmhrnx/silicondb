package home

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// searchShortcut looks like a search field and opens Explore.
func searchShortcut(onClick func()) *giode.PressableElement {
	return giode.Pressable(func(s giode.PressState) giode.Element {
		bg := theme.Surface
		if s.Pressed {
			bg = theme.Pressed()
		}
		return giode.Row(giode.Styles{
			Width: -1, Align: giode.AlignCenter, Gap: 10, Wrap: giode.WrapNoWrap,
			Padding:      giode.SymmetricInset(16, 15),
			Background:   bg,
			BorderRadius: 16,
			BorderWidth:  1,
			BorderColor:  theme.Border,
		},
			theme.Icon("search", 18, theme.Muted),
			giode.Text(tr.T("home.search_hint"), theme.Body().With(giode.Styles{Color: theme.Muted})),
		)
	}).Label(tr.T("common.search")).OnClick(func(giode.ClickEvent) { onClick() })
}
