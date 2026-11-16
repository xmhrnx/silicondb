package store

import (
	"path/filepath"
	"testing"
)

func TestPrefsSurviveARestart(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	s.SetLocale("de")
	s.ToggleCompare("intel-8008")
	s.Viewed("amd-ryzen-9-9950x")
	s.Save() // waits for the background writes' result

	p := loadPrefs(filepath.Join(dir, "prefs.json"))
	if p.Locale != "de" || len(p.Compare) != 1 || len(p.Recent) != 1 {
		t.Fatalf("prefs after reload = %+v", p)
	}
}
