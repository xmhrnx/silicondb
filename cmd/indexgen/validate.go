package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/mehran9675/silicondb/internal/model"
)

var (
	idPattern   = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9.]+)*$`)
	datePattern = regexp.MustCompile(`^\d{4}(-(0[1-9]|1[0-2])(-(0[1-9]|[12]\d|3[01]))?)?$`)
)

type report struct {
	errors, warnings []string
	shards, chips    int
	// gaps, when set, is called for every chip so -report can list
	// missing specs.
	gaps     func(id, category string, year int, specs map[string]json.RawMessage, na []string)
	complete *completeness
}

func (r *report) errorf(format string, a ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, a...))
}
func (r *report) warnf(format string, a ...any) {
	r.warnings = append(r.warnings, fmt.Sprintf(format, a...))
}

// rawShard mirrors model.Shard with specs kept raw, so value types can be
// checked.
type rawShard struct {
	Manufacturer string `json:"manufacturer"`
	Category     string `json:"category"`
	Chips        []struct {
		ID       string                     `json:"id"`
		Name     string                     `json:"name"`
		Family   string                     `json:"family"`
		Gen      string                     `json:"generation"`
		Released string                     `json:"released"`
		Specs    map[string]json.RawMessage `json:"specs"`
		NA       []string                   `json:"na"`
	} `json:"chips"`
}

// validateShard checks one shard file and returns its chip count.
func validateShard(rel string, b []byte, maker, cat string, schema *model.Schema, ids map[string]string, rep *report) int {
	var s rawShard
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&s); err != nil {
		rep.errorf("%s: %v", rel, err)
		return 0
	}
	known := false
	for _, c := range schema.Categories {
		known = known || c.ID == cat
	}
	if !known {
		rep.errorf("%s: unknown category %q", rel, cat)
	}
	if s.Manufacturer != maker || s.Category != cat {
		rep.errorf("%s: manufacturer/category %q/%q don't match the path", rel, s.Manufacturer, s.Category)
	}
	for i, c := range s.Chips {
		where := fmt.Sprintf("%s: chip %d (%s)", rel, i, c.ID)
		switch {
		case !idPattern.MatchString(c.ID):
			rep.errorf("%s: invalid id", where)
		case ids[c.ID] != "":
			rep.errorf("%s: id also used in %s", where, ids[c.ID])
		default:
			ids[c.ID] = rel
		}
		if c.Name == "" {
			rep.errorf("%s: name is empty", where)
		}
		if c.Family == "" {
			rep.warnf("%s: family is empty", where)
		}
		if c.Gen == "" {
			rep.warnf("%s: generation is empty", where)
		}
		if m, ok := c.Specs["market"]; ok && !validMarket(m) {
			rep.errorf("%s: market %s isn't one of the segments in data/README.md", where, m)
		}
		if !datePattern.MatchString(c.Released) {
			rep.errorf("%s: released %q isn't YYYY, YYYY-MM or YYYY-MM-DD", where, c.Released)
		}
		for _, k := range c.NA {
			if _, ok := c.Specs[k]; ok {
				rep.errorf("%s: %q is both a spec and listed in na", where, k)
			}
		}
		if rep.gaps != nil {
			rep.gaps(c.ID, cat, year(c.Released), c.Specs, c.NA)
		}
		for k, v := range c.Specs {
			if schema.Field(k) == nil {
				rep.warnf("%s: spec %q isn't in schema.json", where, k)
			}
			if !validValue(v) {
				rep.errorf("%s: spec %q must be a number, string or array of strings", where, k)
			}
		}
	}
	return len(s.Chips)
}

func validValue(raw json.RawMessage) bool {
	var v any
	if json.Unmarshal(raw, &v) != nil {
		return false
	}
	switch v := v.(type) {
	case float64, string, bool:
		return true
	case []any:
		for _, e := range v {
			if _, ok := e.(string); !ok {
				return false
			}
		}
		return true
	}
	return false
}

// year returns the year of a YYYY[-MM[-DD]] date, or 0.
func year(date string) int {
	y := 0
	for i := 0; i < len(date) && i < 4; i++ {
		y = y*10 + int(date[i]-'0')
	}
	return y
}

// markets are the segments a chip can be sold for; the app groups each
// company's chips by them.
var markets = map[string]bool{
	"Desktop": true, "Laptop": true, "Workstation": true, "Server": true, "Data center": true,
	"Mobile": true, "Tablet": true, "Wearable": true, "Console": true, "Automotive": true,
	"Embedded": true, "Industrial": true,
}

func validMarket(raw json.RawMessage) bool {
	var s string
	return json.Unmarshal(raw, &s) == nil && markets[s]
}
