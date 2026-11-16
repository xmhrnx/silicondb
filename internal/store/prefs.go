package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
)

// Prefs are the user's settings, saved between runs.
type Prefs struct {
	Dark       bool     `json:"dark"`
	Locale     string   `json:"locale,omitempty"`
	DataURL    string   `json:"dataURL,omitempty"`
	Compare    []string `json:"compare,omitempty"`
	Recent     []string `json:"recent,omitempty"`
	LastSynced int64    `json:"lastSynced,omitempty"`
}

// clone copies p, lists included, so a snapshot being written in the
// background doesn't share them with the live preferences.
func (p Prefs) clone() Prefs {
	p.Compare = slices.Clone(p.Compare)
	p.Recent = slices.Clone(p.Recent)
	return p
}

func loadPrefs(path string) Prefs {
	p := Prefs{Dark: true}
	if path == "" {
		return p
	}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &p)
	}
	return p
}

func savePrefs(path string, p Prefs) {
	if path == "" {
		return
	}
	b, err := json.Marshal(p)
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, b, 0o644)
}
