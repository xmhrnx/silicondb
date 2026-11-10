package main

import (
	"regexp"
	"strings"
)

// wikipedia supplies the FP16/FP32/FP64 (and on recent architectures FP8 and
// FP4) throughput that graphics-card pages don't state. Its GPU list
// articles carry these columns for every generation, and Wikipedia permits
// automated reads through the API.
//
// Each list is downloaded once per run and indexed by normalised model name,
// because the tables name cards the way the articles title them ("GeForce
// RTX 4090"), which is also how this database names them.
type wikiTable struct {
	url  string
	kind string // "nvidia", "amd" or "intel"
}

var wikiLists = []wikiTable{
	{"https://en.wikipedia.org/wiki/List_of_Nvidia_graphics_processing_units", "nvidia"},
	{"https://en.wikipedia.org/wiki/List_of_AMD_graphics_processing_units", "amd"},
	{"https://en.wikipedia.org/wiki/List_of_Intel_graphics_processing_units", "intel"},
}

// wikiIndex maps normalised model names to the FP figures a list gives.
type wikiIndex map[string]map[string]any

func (w *wikiIndex) merge(other wikiIndex) {
	for name, vals := range other {
		cur, ok := (*w)[name]
		if !ok {
			(*w)[name] = vals
			continue
		}
		for k, v := range vals {
			if _, exists := cur[k]; !exists {
				cur[k] = v
			}
		}
	}
}

// load downloads every list once. Missing lists are skipped, so the tool
// still works when one article is unavailable.
func (w *wikiIndex) load(f *fetcher) {
	for _, list := range wikiLists {
		doc, err := f.get(list.url)
		if err != nil {
			continue
		}
		w.merge(indexWiki(doc, list.kind))
	}
}

// lookup finds a card by name, trying the name as written and with common
// suffixes trimmed.
func (w wikiIndex) lookup(name string, maker string) map[string]any {
	candidates := []string{normaliseName(name)}
	bare := normaliseName(strings.TrimPrefix(strings.ToLower(name), maker+" "))
	candidates = append(candidates, bare)
	for _, c := range candidates {
		if vals, ok := w[c]; ok {
			return vals
		}
	}
	return nil
}

var (
	refTag      = regexp.MustCompile(`\[[^\]]*\]`)
	parentheses = regexp.MustCompile(`\([^)]*\)`)
	nonAlnum    = regexp.MustCompile(`[^a-z0-9+]+`)
)

// normaliseName reduces a model name to lower-case words, so "GeForce RTX
// 4090", "GeForce\xa0RTX 4090[13]" and "geforce rtx 4090 (24gb)" agree.
func normaliseName(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = refTag.ReplaceAllString(s, " ")
	s = parentheses.ReplaceAllString(s, " ")
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "+", " plus ")
	s = nonAlnum.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}

// indexWiki reads the FP columns out of a GPU list article.
func indexWiki(doc []byte, kind string) wikiIndex {
	out := wikiIndex{}
	for _, t := range tables(doc) {
		if len(t.rows) < 3 {
			continue
		}
		// The data starts after the header rows. Headers can span two or
		// three rows because group labels are merged vertically, so the
		// column names are stitched together from each row.
		hdrEnd := headerRows(t.rows)
		if hdrEnd == 0 || len(t.rows) <= hdrEnd {
			continue
		}
		header := combineHeaderRows(t.rows[:hdrEnd])
		data := t.rows[hdrEnd:]
		super := strings.ToLower(strings.Join(header, " "))
		modelCol := -1
		cols := map[string]flopCol{}
		for i, h := range header {
			hl := strings.ToLower(h)
			unit := headerUnit(hl, super)
			switch {
			case modelCol < 0 && strings.Contains(hl, "model"):
				modelCol = i
			case strings.Contains(hl, "half precision") && !strings.Contains(hl, "tensor"):
				setCol(cols, "fp16_tflops", i, unit)
			case strings.Contains(hl, "single precision"):
				setCol(cols, "fp32_tflops", i, unit)
			case strings.Contains(hl, "double precision") && !strings.Contains(hl, "tensor"):
				setCol(cols, "fp64_tflops", i, unit)
			case strings.Contains(hl, "tensor") && strings.Contains(hl, "fp8") && strings.Contains(hl, "dense"):
				// FP8 exists only on tensor cores, and the database stores
				// that dense figure in fp8_tflops as well (as it does for
				// H100 and the other accelerators).
				setCol(cols, "fp8_tflops", i, unit)
				setCol(cols, "fp8_tensor_tflops", i, unit)
			case strings.Contains(hl, "tensor") && strings.Contains(hl, "fp32"):
				setCol(cols, "fp32_tensor_tflops", i, unit)
			case strings.Contains(hl, "tensor") && strings.Contains(hl, "bf16"):
				setCol(cols, "bf16_tflops", i, unit)
				setCol(cols, "bf16_tensor_tflops", i, unit)
			case strings.Contains(hl, "tensor") && strings.Contains(hl, "fp16") && strings.Contains(hl, "dense"):
				setCol(cols, "fp16_tensor_tflops", i, unit)
			case strings.Contains(hl, "tensor") && strings.Contains(hl, "fp4") && strings.Contains(hl, "dense"):
				setCol(cols, "fp4_tflops", i, unit)
				setCol(cols, "fp4_tensor_tflops", i, unit)
			}
		}
		if modelCol < 0 || len(cols) == 0 {
			continue
		}
		for _, row := range data {
			if modelCol >= len(row) {
				continue
			}
			name := normaliseName(row[modelCol])
			if name == "" || name == "model" || name == "model name" {
				continue
			}
			vals := out[name]
			if vals == nil {
				vals = map[string]any{}
				out[name] = vals
			}
			for key, col := range cols {
				if col.index >= len(row) {
					continue
				}
				if _, have := vals[key]; have {
					continue
				}
				// Tables give base and boost as "628.74 (751.62)"; the boost
				// figure is the one they list for comparison.
				cell := row[col.index]
				if open := strings.IndexByte(cell, '('); open >= 0 {
					if close := strings.IndexByte(cell[open:], ')'); close > 0 {
						cell = cell[open+1 : open+close]
					}
				}
				// Cells are bare numbers; the unit lives in the header.
				if v, _, ok := parseNumber(cell); ok {
					vals[key] = col.scale * v
				}
			}
			if len(vals) == 0 {
				delete(out, name)
			}
		}
	}
	return out
}

// looksLikeHeader reports whether a row names columns rather than holding
// data, judged by the words that only appear in headers.
func looksLikeHeader(row []string) bool {
	for _, c := range row {
		l := strings.ToLower(c)
		switch {
		case strings.Contains(l, "precision"),
			strings.Contains(l, "tflops"),
			strings.Contains(l, "gflops"),
			strings.Contains(l, "model name"):
			return true
		}
	}
	return false
}

// looksLikeData reports whether a row holds a card rather than headers,
// judged by a launch date in the second or third column.
func looksLikeData(row []string) bool {
	for i, c := range row {
		if i > 3 {
			break
		}
		l := strings.ToLower(c)
		if strings.Contains(l, "20") && (strings.Contains(l, " 20") || strings.Contains(l, "20")) && !strings.Contains(l, "clock") {
			if len(l) > 6 && (strings.Contains(l, "january") || strings.Contains(l, "february") ||
				strings.Contains(l, "march") || strings.Contains(l, "april") || strings.Contains(l, "may") ||
				strings.Contains(l, "june") || strings.Contains(l, "july") || strings.Contains(l, "august") ||
				strings.Contains(l, "september") || strings.Contains(l, "october") || strings.Contains(l, "november") ||
				strings.Contains(l, "december")) {
				return true
			}
		}
	}
	return false
}

// combineHeaderRows joins each column's text across the header rows, so a
// merged "Processing Power (TFLOPS)" label above "Tensor (AI) FP8 Dense"
// yields one string naming both.
func combineHeaderRows(rows [][]string) []string {
	width := 0
	for _, r := range rows {
		if len(r) > width {
			width = len(r)
		}
	}
	out := make([]string, width)
	for i := range out {
		var parts []string
		for _, r := range rows {
			if i < len(r) && r[i] != "" {
				parts = append(parts, r[i])
			}
		}
		out[i] = strings.Join(parts, " ")
	}
	return out
}

// headerRows returns how many leading rows are headers.
func headerRows(rows [][]string) int {
	n := 0
	for _, r := range rows {
		if !looksLikeHeader(r) && looksLikeData(r) {
			break
		}
		if !looksLikeHeader(r) && n > 0 {
			break
		}
		if !looksLikeHeader(r) && n == 0 {
			// A table with no recognisable header row.
			return 0
		}
		n++
		if n > 3 {
			break
		}
	}
	return n
}

// flopCol is a throughput column: where it is and how to scale its bare
// numbers to TFLOPS.
type flopCol struct {
	index int
	scale float64
}

func setCol(cols map[string]flopCol, key string, i int, unit string) {
	if _, ok := cols[key]; !ok {
		cols[key] = flopCol{i, scaleFor(unit)}
	}
}

// headerUnit returns the unit text for a column: its own cell if it names
// one, otherwise the merged group header above it.
func headerUnit(col, super string) string {
	if i := strings.Index(col, "tflops"); i >= 0 {
		return "tflops"
	}
	if i := strings.Index(col, "gflops"); i >= 0 {
		return "gflops"
	}
	if i := strings.Index(col, "mflops"); i >= 0 {
		return "mflops"
	}
	return super
}

// scaleFor converts a unit to a factor that turns its numbers into TFLOPS.
func scaleFor(unit string) float64 {
	switch {
	case strings.Contains(unit, "tflops"):
		return 1
	case strings.Contains(unit, "gflops"):
		return 0.001
	case strings.Contains(unit, "mflops"):
		return 1e-6
	}
	return 1
}

// allowedWiki limits Wikipedia's figures to the performance keys, since its
// tables also carry clocks and sizes this database already has from better
// sources.
func allowedWiki() map[string]bool {
	return map[string]bool{
		"fp16_tflops":        true,
		"fp32_tflops":        true,
		"fp64_tflops":        true,
		"bf16_tflops":        true,
		"fp4_tflops":         true,
		"fp32_tensor_tflops": true,
		"fp16_tensor_tflops": true,
		"fp8_tflops":         true,
		"fp8_tensor_tflops":  true,
		"fp4_tensor_tflops":  true,
		"bf16_tensor_tflops": true,
	}
}
