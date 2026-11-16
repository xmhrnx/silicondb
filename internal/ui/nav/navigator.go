package nav

import "github.com/xmhrnx/giode"

// entry is a pushed page with the transition it entered with.
type entry struct {
	page  Page
	enter *giode.TransitionElement
}

// Navigator shows one of several root pages, chosen by tab, with pages
// pushed above it. Back pops the top page, then returns to the first tab.
type Navigator struct {
	roots []Page
	tab   int
	stack []entry
	fade  *giode.TransitionElement

	// OnTabChange is called after the tab changes.
	OnTabChange func(tab int)
}

// New returns a navigator over the root pages, one per tab.
func New(roots ...Page) *Navigator {
	return &Navigator{roots: roots}
}

// Tab returns the selected tab.
func (n *Navigator) Tab() int { return n.tab }

// SetTab selects a tab and clears the pages pushed above it.
func (n *Navigator) SetTab(i int) {
	if i < 0 || i >= len(n.roots) {
		return
	}
	if i == n.tab && len(n.stack) == 0 {
		return
	}
	n.tab = i
	n.stack = nil
	n.fade = giode.Transition(giode.Box(giode.Styles{}), giode.Fade, giode.TransitionFast, giode.EaseOut)
	if n.OnTabChange != nil {
		n.OnTabChange(i)
	}
}

// Push shows p above the current page.
func (n *Navigator) Push(p Page) {
	n.stack = append(n.stack, entry{
		page:  p,
		enter: giode.Transition(giode.Box(giode.Styles{}), giode.Fade|giode.SlideLeft, giode.TransitionNormal, giode.EaseOut),
	})
}

// Back pops the top page, or returns to the first tab. It reports whether
// there was anything to go back from.
func (n *Navigator) Back() bool {
	if len(n.stack) > 0 {
		n.stack = n.stack[:len(n.stack)-1]
		return true
	}
	if n.tab != 0 {
		n.SetTab(0)
		return true
	}
	return false
}

// Depth returns how many pages are pushed above the root.
func (n *Navigator) Depth() int { return len(n.stack) }

// View lays out the current page and claims the Back button while there
// is somewhere to go back to.
func (n *Navigator) View() giode.Element {
	if len(n.stack) > 0 || n.tab != 0 {
		giode.HandleBack(func() { n.Back() })
	}
	if len(n.stack) > 0 {
		top := n.stack[len(n.stack)-1]
		return top.enter.Child(top.page.View())
	}
	view := n.roots[n.tab].View()
	if n.fade != nil {
		return n.fade.Child(view)
	}
	return view
}
