// Package theme defines the app's look: colors, type and spacing.
package theme

import "github.com/xmhrnx/giode"

// Dark is the default palette: deep graphite surfaces with a violet
// primary.
var Dark = giode.Palette{
	Primary:    "#8B5CF6",
	Secondary:  "#22D3EE",
	Background: "#08090D",
	Surface:    "#12141B",
	Text:       "#F4F5F9",
	Border:     "#232633",
	OnPrimary:  "#FFFFFF",
	Muted:      "#8A8FA3",
	Accent:     "#22D3EE",
	Success:    "#22C55E",
	Warning:    "#F59E0B",
	Danger:     "#F43F5E",
	Info:       "#38BDF8",
}

// Light is the light-mode palette.
var Light = giode.Palette{
	Primary:    "#7C3AED",
	Secondary:  "#0891B2",
	Background: "#F5F6FA",
	Surface:    "#FFFFFF",
	Text:       "#0E1018",
	Border:     "#E3E5EE",
	OnPrimary:  "#FFFFFF",
	Muted:      "#646A7E",
	Accent:     "#0891B2",
	Success:    "#16A34A",
	Warning:    "#D97706",
	Danger:     "#E11D48",
	Info:       "#0284C7",
}

// Apply installs the palettes and selects the mode.
func Apply(dark bool) {
	giode.SetPalette(giode.ModeDark, Dark)
	giode.SetPalette(giode.ModeLight, Light)
	SetDark(dark)
}

// SetDark switches between dark and light mode.
func SetDark(dark bool) {
	if dark {
		giode.SetMode(giode.ModeDark)
	} else {
		giode.SetMode(giode.ModeLight)
	}
}

// IsDark reports whether dark mode is active.
func IsDark() bool { return giode.CurrentMode() == giode.ModeDark }
