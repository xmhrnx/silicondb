package components

import (
	"bytes"
	"image"
	"image/png"
	"sync"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/assets"
	"github.com/mehran9675/silicondb/internal/store"
	"github.com/mehran9675/silicondb/internal/ui/theme"
	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// splashBackground is the launch screen's colour (app.splash_color in
// giode.toml), so the system launch screen flows into this one.
const splashBackground = "#08090D"

var (
	logoOnce sync.Once
	logo     image.Image
)

// Splash fills the screen while the first download runs, before there is
// anything to show: the launch screen's logo, with progress below it.
func Splash(st store.SyncStatus) giode.Element {
	logoOnce.Do(func() { logo, _ = png.Decode(bytes.NewReader(assets.Logo)) })
	label := tr.T("splash.connecting")
	var progress giode.Element = giode.Spinner(giode.Styles{Width: 22, Height: 22, Color: "#A78BFA"})
	if st.Total > 0 {
		label = tr.T("splash.downloading")
		progress = giode.Box(giode.Styles{Width: 200},
			giode.Progress(float32(st.Done)/float32(st.Total), giode.Styles{Height: 4, Color: "#A78BFA", Background: "#1E2029"}))
	}
	var mark giode.Element = theme.Icon("cpu", 120, "#FFFFFF")
	if logo != nil {
		mark = giode.Image(logo, giode.Styles{Width: 200, Height: 200})
	}
	return giode.Stack(giode.Styles{
		Width: -1, Height: -1, Align: giode.AlignCenter, Justify: giode.JustifyCenter, Gap: 14,
		Background: splashBackground,
	},
		mark,
		giode.Text(tr.T("app.name"), giode.Styles{FontFamily: theme.Family, FontSize: 24, FontWeight: giode.FontWeightBold, Color: "#F4F5F9"}),
		giode.Text(label, giode.Styles{FontFamily: theme.Family, FontSize: 14, Color: "#8A8FA3"}),
		giode.Box(giode.Styles{Height: 8}),
		progress,
	)
}
