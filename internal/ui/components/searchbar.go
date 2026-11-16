package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"

	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// SearchBar is a rounded search field with a magnifier and a clear
// button. Create it once.
type SearchBar struct {
	Field *input.Input
	clear *giode.PressableElement
	hint  string
}

// NewSearchBar returns a search bar whose placeholder is the message
// hint, calling onChange as the text changes.
func NewSearchBar(hint string, onChange func(text string)) *SearchBar {
	b := &SearchBar{hint: hint}
	b.Field = giode.Input.Text(tr.T(hint), giode.Styles{
		FontFamily: theme.Family, FontSize: 15,
		Color: theme.Text, Background: "transparent", BorderColor: "transparent",
		Padding: giode.SymmetricInset(0, 12),
	}).OnChange(onChange)
	b.clear = giode.Pressable(func(s giode.PressState) giode.Element {
		return giode.Box(giode.Styles{Padding: giode.UniformInset(6)}, theme.Icon("close", 16, theme.Muted))
	}).Label(tr.T("common.clear_search")).OnClick(func(giode.ClickEvent) {
		b.Field.ChangeText("")
	})
	return b
}

// Text returns the current query.
func (b *SearchBar) Text() string { return b.Field.Text() }

// View lays the bar out.
func (b *SearchBar) View() giode.Element {
	b.Field.Placeholder(tr.T(b.hint))
	children := []giode.Element{
		theme.Icon("search", 18, theme.Muted),
		giode.Box(giode.Styles{FlexGrow: 1}, b.Field),
	}
	if b.Field.Text() != "" {
		children = append(children, b.clear)
	}
	border := theme.Border
	if b.Field.Focused() {
		border = theme.Primary
	}
	return giode.Row(giode.Styles{
		Width: -1, Align: giode.AlignCenter, Gap: 10, Wrap: giode.WrapNoWrap,
		Padding:      giode.Inset{Left: 16, Right: 8},
		Background:   theme.Surface,
		BorderRadius: 16,
		BorderWidth:  1,
		BorderColor:  border,
	}, children...)
}
