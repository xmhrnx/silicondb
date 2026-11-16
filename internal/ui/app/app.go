// Package app assembles the screens into the running application.
package app

import (
	"image"
	"time"

	"path/filepath"

	gioapp "gioui.org/app"
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/assets"
	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/components"
	"github.com/mehran9675/silicondb/internal/ui/env"
	"github.com/mehran9675/silicondb/internal/ui/nav"
	"github.com/mehran9675/silicondb/internal/ui/screens/category"
	"github.com/mehran9675/silicondb/internal/ui/screens/compare"
	"github.com/mehran9675/silicondb/internal/ui/screens/detail"
	"github.com/mehran9675/silicondb/internal/ui/screens/domain"
	"github.com/mehran9675/silicondb/internal/ui/screens/explore"
	"github.com/mehran9675/silicondb/internal/ui/screens/generation"
	"github.com/mehran9675/silicondb/internal/ui/screens/home"
	"github.com/mehran9675/silicondb/internal/ui/screens/licenses"
	"github.com/mehran9675/silicondb/internal/ui/screens/maker"
	"github.com/mehran9675/silicondb/internal/ui/screens/makers"
	"github.com/mehran9675/silicondb/internal/ui/screens/settings"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Tab indexes.
const (
	tabHome = iota
	tabExplore
	tabMakers
	tabCompare
	tabSettings
)

// App is Silicon DB.
type App struct {
	env     *env.Env
	nav     *nav.Navigator
	explore *explore.Screen
	bar     *components.BottomNav
	retry   *giode.ButtonElement
	window  *giode.App
}

// New prepares the app: fonts, icons, theme, state and screens.
func New() *App {
	theme.LoadFonts()
	theme.LoadIcons()

	dir := ""
	if d, err := gioapp.DataDir(); err == nil {
		dir = filepath.Join(d, "silicondb")
	}
	st := store.New(dir)
	theme.Apply(st.Dark())

	a := &App{}
	a.env = &env.Env{
		Store:  st,
		Routes: a,
		Toasts: giode.Toaster().Placement(giode.ToastBottomCenter).Duration(2500 * time.Millisecond).Max(2),
	}
	a.explore = explore.New(a.env)
	a.nav = nav.New(home.New(a.env), a.explore, makers.New(a.env), compare.New(a.env), settings.New(a.env))
	a.bar = components.NewBottomNav([]components.NavTab{
		{Icon: "home", Label: "tab.home"},
		{Icon: "explore", Label: "tab.explore"},
		{Icon: "factory", Label: "tab.makers"},
		{Icon: "compare", Label: "tab.compare"},
		{Icon: "settings", Label: "tab.settings"},
	}, a.nav.Tab, a.nav.SetTab)
	a.retry = giode.Button("", giode.Styles{BorderRadius: 14, Padding: giode.SymmetricInset(20, 12)}).OnClick(st.Refresh)

	a.window = giode.New(giode.Config{
		Title:        "Silicon DB",
		Size:         image.Pt(420, 880),
		Resizable:    true,
		MinWidth:     340,
		MinHeight:    560,
		Background:   theme.Background,
		Icon:         assets.AppIcon,
		OnBackground: st.Save,
		OnStartup:    st.Start,
		Intl:         &giode.IntlOptions{FS: assets.Locales, Dir: "locales", Fallback: "en"},
		Locale:       st.Locale(),
	})
	return a
}

// Run shows the window until it closes.
func (a *App) Run() error { return a.window.Run(a.view) }

func (a *App) view() giode.Element {
	page := func() giode.Element { return a.page() }
	return giode.Stack(giode.Styles{Width: -1, Height: -1},
		giode.Responsive(
			giode.Breakpoint{MaxWidth: 600, Name: "compact", View: page},
			giode.Breakpoint{Name: "wide", View: page},
		),
		a.env.Toasts,
	)
}

// page is the current screen above the tab bar, or the first-run states.
func (a *App) page() giode.Element {
	st := a.env.Store.Status.Get()
	if a.env.Store.Catalog.Get() == nil {
		if st.Phase == store.Failed {
			a.retry.Label(tr.T("common.try_again"))
			return components.EmptyState("cloud", tr.T("offline.title"), tr.T("offline.message", "error", st.Err), a.retry)
		}
		return components.Splash(st)
	}
	return giode.Stack(giode.Styles{Width: -1, Height: -1},
		giode.Box(giode.Styles{Width: -1, FlexGrow: 1}, a.nav.View()),
		a.bar.View(),
	)
}

// OpenChip implements env.Routes.
func (a *App) OpenChip(id string) { a.nav.Push(detail.New(a.env, id)) }

// OpenMaker implements env.Routes.
func (a *App) OpenMaker(id, category string) { a.nav.Push(maker.New(a.env, id, category)) }

// OpenDomain implements env.Routes.
func (a *App) OpenDomain(maker, category, segment string) {
	a.nav.Push(domain.New(a.env, maker, category, segment))
}

// OpenGeneration implements env.Routes.
func (a *App) OpenGeneration(maker, category, segment, name string) {
	a.nav.Push(generation.New(a.env, maker, category, segment, name))
}

// OpenCategory implements env.Routes.
func (a *App) OpenCategory(id string) { a.nav.Push(category.New(a.env, id)) }

// Explore implements env.Routes.
func (a *App) Explore(q store.Query) {
	a.explore.Apply(q)
	a.nav.SetTab(tabExplore)
}

// OpenLicenses implements env.Routes.
func (a *App) OpenLicenses() { a.nav.Push(licenses.New(a.env)) }

// Makers implements env.Routes.
func (a *App) Makers() { a.nav.SetTab(tabMakers) }

// Back implements env.Routes.
func (a *App) Back() { a.nav.Back() }
