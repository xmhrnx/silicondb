package settings

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/toggle"
	"github.com/xmhrnx/giode/i18n"

	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// darkSwitch turns dark mode on and off. Its label is set when the switch
// is made, so it is remade when the language changes.
type darkSwitch struct {
	store   *store.Store
	sw      *toggle.Switch
	version uint64
}

func newDarkSwitch(st *store.Store) *darkSwitch { return &darkSwitch{store: st} }

func (d *darkSwitch) View() giode.Element {
	if v := i18n.Version(); d.sw == nil || v != d.version {
		d.sw = giode.Input.Switch(tr.T("settings.dark_mode"), giode.Styles{Width: -1, FontFamily: theme.Family, FontSize: 15}).
			OnChange(func(on bool) {
				d.store.SetDark(on)
				theme.SetDark(on)
			})
		d.sw.SetChecked(d.store.Dark())
		d.version = v
	}
	return d.sw
}
