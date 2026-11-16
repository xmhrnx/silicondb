package format

import (
	"strings"

	"github.com/xmhrnx/giode/i18n"
)

// text returns the translation of key, or en when no translation is
// loaded (as in tests and tools).
func text(key, en string, data ...any) string {
	if s := i18n.Intl(key, data...); s != key {
		return s
	}
	if len(data) == 0 {
		return en
	}
	return fill(en, data...)
}

// fill replaces {name} placeholders in s from name, value pairs.
func fill(s string, pairs ...any) string {
	for i := 0; i+1 < len(pairs); i += 2 {
		name, _ := pairs[i].(string)
		val, _ := pairs[i+1].(string)
		s = strings.ReplaceAll(s, "{"+name+"}", val)
	}
	return s
}

// separators returns the digit group and decimal separators of the
// current language.
func separators() (group, decimal string) {
	lang, _, _ := strings.Cut(strings.ToLower(i18n.Locale()), "-")
	lang, _, _ = strings.Cut(lang, "_")
	switch lang {
	case "de", "es", "it", "pt", "nl", "da", "el", "hr", "sl", "ro", "id", "tr":
		return ".", ","
	case "fr":
		return " ", ","
	case "pl", "sv", "nb", "no", "nn", "fi", "cs", "sk", "hu", "bg", "et", "lv", "lt", "uk", "ru", "sr":
		return " ", ","
	}
	return ",", "."
}
