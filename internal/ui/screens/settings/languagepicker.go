package settings

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/dropdown"
	"github.com/xmhrnx/giode/i18n"

	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// languagePicker chooses the UI language: the system's, or one of the
// translations, each named in its own language. It is remade when the
// language changes, since "System default" is translated.
type languagePicker struct {
	store   *store.Store
	sel     *dropdown.Select
	version uint64
}

func newLanguagePicker(st *store.Store) *languagePicker { return &languagePicker{store: st} }

func (l *languagePicker) View() giode.Element {
	if v := i18n.Version(); l.sel == nil || v != l.version {
		options := []string{tr.T("settings.language.system")}
		selected := 0
		for i, lang := range tr.Languages {
			options = append(options, lang.Name)
			if lang.Code == l.store.Locale() {
				selected = i + 1
			}
		}
		l.sel = giode.Input.Select(giode.Styles{
			Width: -1, FontFamily: theme.Family, Color: theme.Text, Background: theme.Raised(),
			BorderColor: theme.Border, BorderRadius: 12,
		}, options...).Sheet(giode.SheetCompact).OnChange(func(i int) {
			code := ""
			if i > 0 {
				code = tr.Languages[i-1].Code
			}
			l.store.SetLocale(code)
			tr.Apply(code)
		})
		l.sel.SetSelected(selected)
		l.version = v
	}
	return l.sel
}
