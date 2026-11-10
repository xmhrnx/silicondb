package main

import (
	"fmt"
	"regexp"
	"strings"
)

// derived fills specs that follow from values the file already records, so
// they need no download and can't be wrong: the rated memory speed is
// written into memory_type in nearly every entry ("DDR5-6400",
// "LPDDR5X-8533", "DDR5-5600 / DDR4-3200").
var memoryRating = regexp.MustCompile(`(?:DDR|LPDDR|GDDR|HBM)[A-Za-z0-9]*\s*-\s*(\d{3,5})|(\d{4,5})\s*MT/s`)

// derive returns values implied by the chip's own existing specs.
func derive(specs *node) map[string]any {
	out := map[string]any{}
	if specs == nil {
		return out
	}
	if specs.get("memory_speed_mts") == nil {
		if mt, ok := specs.get("memory_type").str(); ok {
			if v, ok := topRating(mt); ok {
				out["memory_speed_mts"] = v
			}
		}
	}
	return out
}

// topRating returns the fastest MT/s rating in a memory-type string such as
// "DDR5-6400 / DDR4-3200". Ratings below 100 are part numbers or capacities,
// not speeds.
func topRating(s string) (float64, bool) {
	best, found := 0.0, false
	for _, m := range memoryRating.FindAllStringSubmatch(s, -1) {
		raw := m[1]
		if raw == "" {
			raw = m[2]
		}
		v, _, ok := parseNumber(raw)
		if !ok || v < 100 || v > 20000 {
			continue
		}
		if v > best {
			best, found = v, true
		}
	}
	return best, found
}

// deriveAll merges values implied by the chip's own specs and reports the
// keys it added. It needs no download, so it runs even for chips whose
// sources have no page.
func (e *enricher) deriveAll(t *target) []string {
	vals := derive(t.specs)
	if len(vals) == 0 {
		return nil
	}
	added := mergeSpecs(t.specs, t.na, vals, nil)
	if len(added) == 0 {
		return nil
	}
	e.changed++
	for _, k := range added {
		e.adds[k]++
	}
	if e.verbose {
		fmt.Printf("%s: %s (derived)\n", t.id, strings.Join(added, ", "))
	}
	return added
}
