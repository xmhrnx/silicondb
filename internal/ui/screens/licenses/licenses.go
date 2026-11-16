// Package licenses shows the open-source licenses of the code the app is
// built from.
package licenses

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/assets"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen lists the licenses.
type Screen struct {
	back *components.RoundButton
	list giode.Element
}

// New returns the licenses screen.
func New(e *env.Env) *Screen {
	return &Screen{
		back: components.NewRoundButton("chevron-left", tr.T("common.back"), e.Routes.Back),
		list: giode.LicensesView(assets.Licenses),
	}
}

// View lays the screen out.
func (s *Screen) View() giode.Element {
	return giode.Box(giode.Styles{Width: -1, Height: -1, Justify: giode.JustifyCenter},
		giode.Stack(giode.Styles{Width: -1, Height: -1, MaxWidth: components.MaxContentWidth, Padding: giode.SymmetricInset(20, 0)},
			components.AppBar(s.back, tr.T("settings.licenses"), ""),
			giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, s.list),
		))
}
