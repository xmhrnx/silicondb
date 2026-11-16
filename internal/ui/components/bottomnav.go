package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// BottomNav is the tab bar along the bottom of the screen. Create it once.
type BottomNav struct {
	tabs     []NavTab
	items    []*giode.PressableElement
	selected func() int
}

// NewBottomNav returns a bar of tabs; selected reports the active tab and
// onSelect is called with the tab tapped.
func NewBottomNav(tabs []NavTab, selected func() int, onSelect func(i int)) *BottomNav {
	n := &BottomNav{tabs: tabs, selected: selected}
	for i, t := range tabs {
		n.items = append(n.items, giode.Pressable(func(s giode.PressState) giode.Element {
			return navTabItem(t, n.selected() == i, s)
		}).Label(tr.T(t.Label)).OnClick(func(giode.ClickEvent) { onSelect(i) }))
	}
	return n
}

// View lays the bar out.
func (n *BottomNav) View() giode.Element {
	cells := make([]giode.Element, len(n.items))
	for i, it := range n.items {
		cells[i] = giode.Box(giode.Styles{FlexGrow: 1}, it)
	}
	return giode.Stack(giode.Styles{Width: -1, Background: theme.Surface},
		giode.Divider(giode.Styles{Color: theme.Border}),
		giode.Row(giode.Styles{Width: -1, MaxWidth: 640, Wrap: giode.WrapNoWrap, Padding: giode.SymmetricInset(8, 2)}, cells...),
	)
}
