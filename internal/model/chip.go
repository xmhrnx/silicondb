package model

import "strings"

// Chip is one processor, accelerator or other data-processing integrated
// circuit with its specifications.
type Chip struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Family string `json:"family"`
	// Generation is the product generation the chip belongs to, such as
	// "Core 14th Gen (Raptor Lake Refresh)" or "GeForce RTX 40 (Ada
	// Lovelace)": the level between a company's segment and its chips.
	Generation string         `json:"generation,omitempty"`
	Released   string         `json:"released"`
	Specs      map[string]any `json:"specs"`
	// NA lists expected specs that don't apply to this chip, such as
	// a boost clock on a 1970s CPU. Only the data tools read it.
	NA []string `json:"na,omitempty"`

	// Filled in when the shard is loaded, not stored per chip.
	Manufacturer string `json:"-"`
	Category     string `json:"-"`

	search string
}

// Year returns the release year, or 0 when it is unknown.
func (c *Chip) Year() int {
	if len(c.Released) < 4 {
		return 0
	}
	y := 0
	for _, r := range c.Released[:4] {
		if r < '0' || r > '9' {
			return 0
		}
		y = y*10 + int(r-'0')
	}
	return y
}

// Spec returns the value stored under key and whether it exists.
func (c *Chip) Spec(key string) (any, bool) {
	v, ok := c.Specs[key]
	return v, ok
}

// Number returns the numeric value of key, if it holds a number.
func (c *Chip) Number(key string) (float64, bool) {
	v, ok := c.Specs[key].(float64)
	return v, ok
}

// Matches reports whether every word of query occurs in the chip's name,
// family, manufacturer name, codename or microarchitecture.
func (c *Chip) Matches(words []string) bool {
	for _, w := range words {
		if !strings.Contains(c.search, w) {
			return false
		}
	}
	return true
}

// index precomputes the lowercase text Matches searches.
func (c *Chip) index(makerName string) {
	parts := []string{c.Name, c.Family, makerName, c.Manufacturer, c.Released}
	for _, k := range []string{"codename", "microarchitecture", "isa"} {
		if s, ok := c.Specs[k].(string); ok {
			parts = append(parts, s)
		}
	}
	c.search = strings.ToLower(strings.Join(parts, " "))
}
