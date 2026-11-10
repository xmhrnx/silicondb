package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// cpuMonkey supplies benchmark scores. Every processor page carries a
// schema.org Product block whose additionalProperty list holds the scores,
// and CPU-Monkey's robots.txt allows every path.
type cpuMonkey struct {
	f *fetcher
}

// cpuMonkeySlugs guesses page names for a chip: the database id with
// underscores, and the plain model name.
func cpuMonkeySlugs(id, name, maker string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(s string) {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	add(strings.ReplaceAll(id, "-", "_"))
	model := name
	if maker != "" && strings.HasPrefix(strings.ToLower(model), strings.ToLower(maker)+" ") {
		model = strings.TrimSpace(model[len(maker):])
	}
	add(slugify(maker + "_" + model))
	add(slugify(model))
	return out
}

// benchmarks fetches the page for a chip and returns its benchmark scores
// keyed by the database's spec keys. It returns nil when no page matches.
func (c *cpuMonkey) benchmarks(maker, category, id, name string) (map[string]float64, error) {
	if category == "ai" || category == "fpga" {
		return nil, nil
	}
	for _, slug := range cpuMonkeySlugs(id, name, maker) {
		page, err := c.f.get("https://www.cpu-monkey.com/en/cpu-" + url.PathEscape(slug))
		if err != nil {
			if skipMiss(err) {
				continue
			}
			return nil, err
		}
		scores := cpuMonkeyScores(page)
		if len(scores) > 0 {
			return scores, nil
		}
	}
	return nil, nil
}

var ldJSON = regexp.MustCompile(`(?s)<script[^>]+type="application/ld\+json"[^>]*>(.*?)</script>`)

// cpuMonkeyScores reads the Product block out of a CPU-Monkey page.
func cpuMonkeyScores(page []byte) map[string]float64 {
	for _, m := range ldJSON.FindAllSubmatch(page, -1) {
		var doc struct {
			Graph []struct {
				Type     string `json:"@type"`
				Property []struct {
					Name  string `json:"name"`
					Value any    `json:"value"`
				} `json:"additionalProperty"`
			} `json:"@graph"`
			Type     string `json:"@type"`
			Property []struct {
				Name  string `json:"name"`
				Value any    `json:"value"`
			} `json:"additionalProperty"`
		}
		if json.Unmarshal(m[1], &doc) != nil {
			continue
		}
		props := doc.Property
		for _, g := range doc.Graph {
			if g.Type == "Product" && len(g.Property) > 0 {
				props = g.Property
			}
		}
		if len(props) == 0 {
			continue
		}
		out := map[string]float64{}
		for _, p := range props {
			key := cpuMonkeyKey(p.Name)
			if key == "" {
				continue
			}
			s, ok := p.Value.(string)
			if !ok {
				if f, isNum := p.Value.(float64); isNum {
					out[key] = f
				}
				continue
			}
			if v, _, ok := parseNumber(s); ok {
				out[key] = v
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return nil
}

// cpuMonkeyKey maps a CPU-Monkey property name to a spec key. Only the
// benchmarks the app shows are kept.
func cpuMonkeyKey(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	switch {
	case strings.HasPrefix(n, "geekbench 6 single"):
		return "geekbench6_single"
	case strings.HasPrefix(n, "geekbench 6 multi"):
		return "geekbench6_multi"
	case strings.HasPrefix(n, "cinebench 2024 single"):
		return "cinebench_2024_single"
	case strings.HasPrefix(n, "cinebench 2024 multi"):
		return "cinebench_2024_multi"
	case strings.HasPrefix(n, "cinebench r23 single"):
		return "cinebench_r23_single"
	case strings.HasPrefix(n, "cinebench r23 multi"):
		return "cinebench_r23_multi"
	case strings.HasPrefix(n, "cinebench r15 single"):
		return "cinebench_r15_single"
	case strings.HasPrefix(n, "cinebench r15 multi"):
		return "cinebench_r15_multi"
	}
	return ""
}

// slugify lowercases s and turns runs of other characters into hyphens.
func slugify(s string) string {
	var sb strings.Builder
	dash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			sb.WriteRune(r)
			dash = false
		case r == '+':
			sb.WriteString("-plus")
			dash = false
		default:
			if !dash && sb.Len() > 0 {
				sb.WriteByte('-')
				dash = true
			}
		}
	}
	return strings.Trim(sb.String(), "-")
}

func (c *cpuMonkey) String() string { return fmt.Sprintf("cpu-monkey(%s)", c.f.dir) }
