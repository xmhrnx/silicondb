// Package assets embeds the files the app ships with: fonts, icons and
// the app icon. Chip data is never embedded; it is downloaded at runtime.
package assets

import "embed"

// Fonts holds the Inter typeface in the weights the UI uses.
//
//go:embed fonts/*.ttf
var Fonts embed.FS

// Icons holds the UI's SVG icons, one file per icon.
//
//go:embed icons/*.svg
var Icons embed.FS

// AppIcon is the window and launcher icon.
//
//go:embed icon.png
var AppIcon []byte

// Locales holds the UI translations, one message file per language.
//
//go:embed locales/*.json
var Locales embed.FS

// Licenses lists the open-source licenses of the app's modules, written
// by `giode licenses`.
//
//go:embed licenses.json
var Licenses []byte

// Logo is the chip on transparency that iOS's launch screen shows; the
// in-app splash shows it the same way so launch flows into the app.
//
//go:embed launch.png
var Logo []byte
