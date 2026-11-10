package main

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	// firstNumber matches the first number in a cell, allowing thousands
	// separators like the "13,382" CPU-Monkey prints.
	firstNumber = regexp.MustCompile(`-?\d[\d,]*(?:\.\d+)?`)
	trailingTag = regexp.MustCompile(`\[[^\]]*\]`)
)

// parseNumber returns the first number in s and the text after it, so
// callers can read a unit. Commas are treated as thousands separators.
func parseNumber(s string) (float64, string, bool) {
	s = trailingTag.ReplaceAllString(s, "")
	loc := firstNumber.FindStringIndex(s)
	if loc == nil {
		return 0, "", false
	}
	raw := s[loc[0]:loc[1]]
	clean := ""
	for _, r := range raw {
		if r != ',' {
			clean += string(r)
		}
	}
	v, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0, "", false
	}
	return v, s[loc[1]:], true
}

// toMB converts a capacity with a unit suffix to mebibytes.
func toNumber(s string) (float64, bool) {
	v, _, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	return v, true
}

// cacheSizeKB converts a cache description such as "512 KB", "1.5 MB" or
// "96MB shared" to KiB.
func cacheSizeKB(s string) (float64, bool) {
	v, rest, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	unit := strings.ToLower(strings.TrimSpace(rest))
	switch {
	case strings.HasPrefix(unit, "gb"):
		return v * 1024 * 1024, true
	case strings.HasPrefix(unit, "mb"):
		return v * 1024, true
	case strings.HasPrefix(unit, "kb"), strings.HasPrefix(unit, "k"):
		return v, true
	case strings.HasPrefix(unit, "b"):
		return v / 1024, true
	}
	return 0, false
}

// capacityMB converts sizes such as "24 GB", "8192 MB" or "192 GiB" to MiB.
func capacityMB(s string) (float64, bool) {
	v, rest, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	unit := strings.ToLower(strings.TrimSpace(rest))
	switch {
	case strings.HasPrefix(unit, "t"):
		return v * 1024 * 1024, true
	case strings.HasPrefix(unit, "g"):
		return v * 1024, true
	case strings.HasPrefix(unit, "m"):
		return v, true
	case strings.HasPrefix(unit, "k"):
		return v / 1024, true
	}
	return 0, false
}

// clockMHz converts "4.7 GHz", "2200 MHz" and "2200MHz" to MHz.
func clockMHz(s string) (float64, bool) {
	v, rest, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	unit := strings.ToLower(strings.TrimSpace(rest))
	switch {
	case strings.HasPrefix(unit, "g"):
		return v * 1000, true
	case strings.HasPrefix(unit, "m"):
		return v, true
	case strings.HasPrefix(unit, "k"):
		return v / 1000, true
	}
	return 0, false
}

// flops converts a throughput cell to TFLOPS. Cells are usually TFLOPS but
// older tables use GFLOPS and the odd wiki table uses MFLOPS.
func flops(s string) (float64, bool) {
	v, rest, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	unit := strings.ToLower(rest)
	switch {
	case strings.Contains(unit, "pflops"), strings.HasPrefix(unit, "p"):
		return v * 1000, true
	case strings.Contains(unit, "tflops"), strings.HasPrefix(unit, "t"):
		return v, true
	case strings.Contains(unit, "gflops"), strings.HasPrefix(unit, "g"):
		return v / 1000, true
	case strings.Contains(unit, "mflops"), strings.HasPrefix(unit, "m"):
		return v / 1000000, true
	}
	return 0, false
}

// processNM converts "5 nm" to nanometres.
func processNM(s string) (float64, bool) {
	v, rest, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	unit := strings.ToLower(strings.TrimSpace(rest))
	switch {
	case strings.HasPrefix(unit, "µm"), strings.HasPrefix(unit, "μm"), strings.HasPrefix(unit, "um"),
		strings.HasPrefix(unit, "micromet"), strings.HasPrefix(unit, "micrometre"):
		return v * 1000, true
	case strings.HasPrefix(unit, "nm"), strings.HasPrefix(unit, "nanomet"):
		return v, true
	}
	return 0, false
}

// transistors converts "28 billions", "76.3 billion" or "3500000" to a
// plain count.
func transistors(s string) (float64, bool) {
	v, rest, ok := parseNumber(s)
	if !ok {
		return 0, false
	}
	unit := strings.ToLower(rest)
	switch {
	case strings.Contains(unit, "billion"):
		return v * 1e9, true
	case strings.Contains(unit, "million"):
		return v * 1e6, true
	}
	// Already a count when the source gives no unit and a large value.
	if v >= 1000 {
		return v, true
	}
	return 0, false
}

// firstOf returns the first non-empty trimmed string.
func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if t := strings.TrimSpace(s); t != "" && t != "-" {
			return t
		}
	}
	return ""
}
