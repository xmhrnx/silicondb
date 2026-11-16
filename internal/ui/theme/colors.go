package theme

import "github.com/xmhrnx/giode"

// Color tokens. Derived shades are mixed in sRGB each frame so they follow
// the mode.
const (
	Text       = "theme.text"
	Muted      = "theme.muted"
	Primary    = "theme.primary"
	Accent     = "theme.accent"
	Surface    = "theme.surface"
	Background = "theme.background"
	Border     = "theme.border"
	Danger     = "theme.danger"
	Success    = "theme.success"
	OnPrimary  = "theme.onprimary"
)

// Raised is a surface one step above Surface: cards inside cards, inputs.
func Raised() string {
	if IsDark() {
		return giode.Mix("#FFFFFF", string(Dark.Surface), 0.045)
	}
	return giode.Mix("#0E1018", string(Light.Surface), 0.035)
}

// Pressed is a surface while it's being pressed.
func Pressed() string {
	if IsDark() {
		return giode.Mix("#FFFFFF", string(Dark.Surface), 0.09)
	}
	return giode.Mix("#0E1018", string(Light.Surface), 0.07)
}

// Faint is a hairline or a track behind a progress bar.
func Faint() string { return giode.Mix(Text, Background, 0.08) }

// Tint returns color blended into the surface at amount, for soft
// colored backgrounds behind icons and badges.
func Tint(color string, amount float64) string { return giode.Mix(color, Surface, amount) }

// BrandGradient is the violet-to-cyan gradient of the hero and the
// primary actions.
func BrandGradient(angle float32) *giode.LinearGradient {
	return giode.Linear(angle, "#7C3AED", "#8B5CF6", "#22D3EE")
}

// AccentGradient fades a category color into a deeper shade.
func AccentGradient(color string) *giode.LinearGradient {
	return giode.Linear(135, color, giode.Mix("#000000", color, 0.35))
}
