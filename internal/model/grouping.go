package model

import "sort"

// SegmentOrder is the order segments (the "market" spec) are listed in.
var SegmentOrder = []string{
	"Desktop", "Laptop", "Workstation", "Server", "Data center", "Mobile", "Tablet",
	"Wearable", "Console", "Automotive", "Embedded", "Industrial",
}

// Segment returns the chip's segment, "" when it has none.
func (c *Chip) Segment() string {
	s, _ := c.Specs["market"].(string)
	return s
}

// GenerationName returns the chip's generation, falling back to its
// family for entries without one.
func (c *Chip) GenerationName() string {
	if c.Generation != "" {
		return c.Generation
	}
	return c.Family
}

// Segments lists the segments chips belong to, in SegmentOrder, with
// unknown ones last.
func Segments(chips []*Chip) []string {
	seen := map[string]bool{}
	for _, c := range chips {
		seen[c.Segment()] = true
	}
	var out []string
	for _, s := range SegmentOrder {
		if seen[s] {
			out = append(out, s)
			delete(seen, s)
		}
	}
	var rest []string
	for s := range seen {
		if s != "" {
			rest = append(rest, s)
		}
	}
	sort.Strings(rest)
	return append(out, rest...)
}

// Generation is a product generation and its chips, newest first.
type Generation struct {
	Name        string
	Chips       []*Chip
	First, Last int // release years
}

// Generations groups chips by generation, the most recent generation
// first and each one's chips newest first.
func Generations(chips []*Chip) []*Generation {
	byName := map[string]*Generation{}
	var out []*Generation
	for _, c := range chips {
		name := c.GenerationName()
		g := byName[name]
		if g == nil {
			g = &Generation{Name: name}
			byName[name] = g
			out = append(out, g)
		}
		g.Chips = append(g.Chips, c)
		if y := c.Year(); y != 0 {
			if g.First == 0 || y < g.First {
				g.First = y
			}
			if y > g.Last {
				g.Last = y
			}
		}
	}
	latest := func(g *Generation) string {
		l := ""
		for _, c := range g.Chips {
			if c.Released > l {
				l = c.Released
			}
		}
		return l
	}
	sort.SliceStable(out, func(i, j int) bool { return latest(out[i]) > latest(out[j]) })
	for _, g := range out {
		sort.SliceStable(g.Chips, func(i, j int) bool { return g.Chips[i].Released > g.Chips[j].Released })
	}
	return out
}

// ChipsIn returns a manufacturer's chips of a category ("" for all) and
// segment ("" for all).
func (c *Catalog) ChipsIn(maker, category, segment string) []*Chip {
	var out []*Chip
	for _, chip := range c.ChipsBy(maker) {
		if (category == "" || chip.Category == category) && (segment == "" || chip.Segment() == segment) {
			out = append(out, chip)
		}
	}
	return out
}

// MakersIn lists the manufacturers with chips in a category, those with
// the most chips first.
func (c *Catalog) MakersIn(category string) []*Manufacturer {
	n := map[string]int{}
	for _, chip := range c.Chips {
		if chip.Category == category {
			n[chip.Manufacturer]++
		}
	}
	var out []*Manufacturer
	for _, m := range c.Manufacturers {
		if n[m.ID] > 0 {
			out = append(out, m)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return n[out[i].ID] > n[out[j].ID] })
	return out
}

// Domain is a segment of one category of a company's chips, such as its
// desktop CPUs: the level between the company and its generations.
type Domain struct {
	Category string
	Segment  string
	Chips    []*Chip
}

// Domains lists a manufacturer's domains in a category ("" for all), in
// schema category order and then SegmentOrder.
func (c *Catalog) Domains(maker, category string) []*Domain {
	var out []*Domain
	for _, cat := range c.Schema.Categories {
		if category != "" && cat.ID != category {
			continue
		}
		chips := c.ChipsIn(maker, cat.ID, "")
		for _, seg := range Segments(chips) {
			out = append(out, &Domain{Category: cat.ID, Segment: seg, Chips: c.ChipsIn(maker, cat.ID, seg)})
		}
	}
	return out
}
