// Package settings holds appearance, language, database and about
// options.
package settings

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/input"
	"github.com/xmhrnx/giode/components/scroll"

	"github.com/mehran9675/silicondb/internal/config"
	"github.com/mehran9675/silicondb/internal/format"
	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Screen is the Settings tab.
type Screen struct {
	env      *env.Env
	scroll   *scroll.Scroll
	dark     *darkSwitch
	language *languagePicker
	source   *input.Input
	refresh  *giode.ButtonElement
	redo     *giode.ButtonElement
	save     *giode.ButtonElement
	reset    *giode.ButtonElement
	licenses *giode.PressableElement
}

// New returns the Settings screen.
func New(e *env.Env) *Screen {
	s := &Screen{env: e}
	s.scroll = giode.Scroll(s.content)
	s.dark = newDarkSwitch(e.Store)
	s.language = newLanguagePicker(e.Store)
	s.source = giode.Input.Text("https://…/data/", giode.Styles{
		FontFamily: theme.Family, FontSize: 13.5, Color: theme.Text, Background: theme.Raised(),
		BorderColor: theme.Border, BorderRadius: 12,
	}).Keyboard(giode.KeyboardURL)
	s.source.SetText(e.Store.DataURL())
	s.refresh = giode.Button("", giode.Styles{BorderRadius: 12}).
		Icon(theme.Icon("refresh", 16, "")).OnClick(e.Store.Refresh)
	s.redo = giode.Button("", giode.Styles{BorderRadius: 12, Background: theme.Primary, Color: theme.Primary}).
		Variant(giode.VariantOutlined).OnClick(e.Store.ClearCache)
	s.save = giode.Button("", giode.Styles{BorderRadius: 12}).OnClick(func() {
		e.Store.SetDataURL(s.source.Text())
		e.Toasts.Show(giode.ToastInfo, tr.T("settings.source.switching"))
	})
	s.reset = giode.Button("", giode.Styles{BorderRadius: 12, Background: theme.Primary, Color: theme.Primary}).
		Variant(giode.VariantGhost).OnClick(func() {
		s.source.SetText(config.DataURL)
		e.Store.SetDataURL("")
	})
	s.licenses = giode.Pressable(nil)
	return s
}

// View lays the screen out.
func (s *Screen) View() giode.Element { return s.scroll }

func (s *Screen) content() giode.Element {
	cat := s.env.Store.Catalog.Get()
	st := s.env.Store.Status.Get()
	syncing := st.Phase == store.Syncing
	s.refresh.Label(tr.T("settings.check")).Loading(syncing)
	s.redo.Label(tr.T("settings.redownload")).Disabled(syncing)
	s.save.Label(tr.T("settings.source.use"))
	s.reset.Label(tr.T("settings.source.default"))

	return components.Column(16,
		components.AppBar(nil, tr.T("settings.title"), ""),
		components.Card(giode.Styles{Padding: giode.SymmetricInset(16, 8)},
			settingRow("moon", s.dark.View()),
			giode.Divider(giode.Styles{Color: theme.Border}),
			settingRow("globe", giode.Stack(giode.Styles{Width: -1, Gap: 6, Padding: giode.SymmetricInset(0, 8)},
				giode.Text(tr.T("settings.language"), theme.Body()),
				s.language.View(),
			)),
		),
		components.SpecSection(tr.T("settings.database"), []components.SpecLine{
			{Label: tr.T("settings.chips"), Value: format.Number(float64(len(cat.Chips)), 0)},
			{Label: tr.T("settings.makers"), Value: format.Number(float64(len(cat.Manufacturers)), 0)},
			{Label: tr.T("settings.data_updated"), Value: format.Date(cat.Updated)},
			{Label: tr.T("settings.last_checked"), Value: ago(st.LastSynced)},
			{Label: tr.T("settings.status"), Value: status(st)},
		}),
		giode.Row(giode.Styles{Gap: 10}, s.refresh, s.redo),
		components.Card(giode.Styles{Gap: 12},
			giode.Text(tr.T("settings.source"), theme.Overline().With(giode.Styles{Color: theme.Primary})),
			giode.Text(tr.T("settings.source.help"), theme.Caption().With(giode.Styles{FontSize: 13, LineHeight: 19})),
			s.source,
			giode.Row(giode.Styles{Gap: 10}, s.save, s.reset),
		),
		components.SpecSection(tr.T("settings.about"), []components.SpecLine{
			{Label: tr.T("settings.version"), Value: version()},
			{Label: tr.T("settings.typeface"), Value: "Inter (SIL OFL)"},
			{Label: tr.T("settings.built_with"), Value: "giode"},
		}),
		s.licensesRow(),
	)
}

// licensesRow opens the open-source licenses.
func (s *Screen) licensesRow() giode.Element {
	return s.licenses.Content(func(p giode.PressState) giode.Element {
		bg := theme.Surface
		if p.Pressed {
			bg = theme.Pressed()
		}
		return giode.Row(giode.Styles{
			Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap,
			Padding: giode.SymmetricInset(16, 14), Background: bg, BorderRadius: theme.RadiusCard,
			BorderWidth: 1, BorderColor: theme.Border,
		},
			components.IconBadge("info", theme.Primary, 36),
			giode.Box(giode.Styles{FlexGrow: 1}, giode.Text(tr.T("settings.licenses"), theme.Body())),
			theme.Icon("chevron-right", 18, theme.Muted),
		)
	}).Label(tr.T("settings.licenses")).OnClick(func(giode.ClickEvent) { s.env.Routes.OpenLicenses() })
}

// settingRow puts an icon before a setting's control.
func settingRow(icon string, control giode.Element) giode.Element {
	return giode.Row(giode.Styles{Width: -1, Gap: 12, Align: giode.AlignCenter, Wrap: giode.WrapNoWrap},
		components.IconBadge(icon, theme.Primary, 36),
		giode.Box(giode.Styles{FlexGrow: 1}, control),
	)
}

// version is the release the app was built as, set by the giode CLI.
func version() string {
	if giode.AppVersion == "" {
		return config.Version
	}
	if giode.AppBuild != "" {
		return giode.AppVersion + " (" + giode.AppBuild + ")"
	}
	return giode.AppVersion
}
