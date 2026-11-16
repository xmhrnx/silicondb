// Package tr translates the UI. Message files live in assets/locales;
// labels that come from the remote schema fall back to the schema's own
// English text when a language has no translation for them, so new
// fields need no app update.
package tr

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/i18n"

	"github.com/mehran9675/silicondb/internal/format"
	"github.com/mehran9675/silicondb/internal/model"
)

// T returns the translation of key, with data filling its {placeholders}
// as a map or name, value pairs; a "count" picks the plural form.
func T(key string, data ...any) string { return giode.Intl(key, data...) }

// Or returns the translation of key, or fallback when no language has
// one.
func Or(key, fallback string, data ...any) string {
	if s := giode.Intl(key, data...); s != key {
		return s
	}
	return fallback
}

// Count returns the plural message key for n: "count.chips" with 3 is
// "3 chips".
func Count(key string, n int) string {
	return T(key, "count", format.Number(float64(n), 0), "n", n)
}

// CategoryName is a category's short name, "CPU".
func CategoryName(c model.Category) string { return Or("category."+c.ID+".name", c.Name) }

// CategoryTitle is a category's long name, "Central Processing Units".
func CategoryTitle(c model.Category) string { return Or("category."+c.ID+".title", c.Title) }

// Group is a spec section's title.
func Group(g *model.Group) string { return Or("group."+g.ID, g.Label) }

// Label is a spec's label; f may be nil for keys the schema lacks.
func Label(f *model.Field, key string) string { return Or("spec."+key, format.Label(f, key)) }

// Value formats a spec value, translating the values of text specs with
// a fixed vocabulary, such as the market segment.
func Value(f *model.Field, key string, v any) string {
	if s, ok := v.(string); ok && key == "market" {
		return Or("market."+s, s)
	}
	return format.Value(f, v)
}

// Apply switches the UI to the language code, or to the system's
// language for "".
func Apply(code string) {
	if code == "" {
		code = i18n.SystemLocale()
	}
	giode.SetLocale(code)
}
