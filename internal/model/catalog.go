package model

import (
	"sort"
	"sync"
)

// Catalog is the whole database in memory, assembled from the cached and
// downloaded files.
type Catalog struct {
	Schema        *Schema
	Updated       string
	Manufacturers []*Manufacturer
	Chips         []*Chip

	byID      map[string]*Chip
	makerByID map[string]*Manufacturer
	makerOf   map[string][]*Chip

	maxMu sync.Mutex
	max   map[string]float64

	// cache holds values derived from the catalog; see derived.
	cache sync.Map
}

// NewCatalog builds a catalog from the schema, the manufacturers and their
// shards. Chips are sorted by release date, newest first.
func NewCatalog(schema *Schema, updated string, makers []*Manufacturer, shards []*Shard) *Catalog {
	c := &Catalog{
		Schema:    schema,
		Updated:   updated,
		byID:      map[string]*Chip{},
		makerByID: map[string]*Manufacturer{},
		makerOf:   map[string][]*Chip{},
	}
	for _, m := range makers {
		c.makerByID[m.ID] = m
	}
	for _, s := range shards {
		maker := c.makerByID[s.Manufacturer]
		name := s.Manufacturer
		if maker != nil {
			name = maker.Name
		}
		for _, chip := range s.Chips {
			if chip == nil || chip.ID == "" || c.byID[chip.ID] != nil {
				continue
			}
			chip.Manufacturer = s.Manufacturer
			chip.Category = s.Category
			if chip.Specs == nil {
				chip.Specs = map[string]any{}
			}
			chip.index(name)
			c.byID[chip.ID] = chip
			c.Chips = append(c.Chips, chip)
		}
	}
	sort.SliceStable(c.Chips, func(i, j int) bool {
		if c.Chips[i].Released != c.Chips[j].Released {
			return c.Chips[i].Released > c.Chips[j].Released
		}
		return c.Chips[i].Name < c.Chips[j].Name
	})
	for _, chip := range c.Chips {
		c.makerOf[chip.Manufacturer] = append(c.makerOf[chip.Manufacturer], chip)
	}
	for _, m := range makers {
		if len(c.makerOf[m.ID]) > 0 {
			c.Manufacturers = append(c.Manufacturers, m)
		}
	}
	sort.Slice(c.Manufacturers, func(i, j int) bool {
		return c.Manufacturers[i].Name < c.Manufacturers[j].Name
	})
	return c
}

// Chip returns the chip with id, or nil.
func (c *Catalog) Chip(id string) *Chip { return c.byID[id] }

// Manufacturer returns the manufacturer with id, or nil.
func (c *Catalog) Manufacturer(id string) *Manufacturer { return c.makerByID[id] }

// MakerName returns the display name of a manufacturer id.
func (c *Catalog) MakerName(id string) string {
	if m := c.makerByID[id]; m != nil {
		return m.Name
	}
	return id
}

// ChipsBy returns a manufacturer's chips, newest first.
func (c *Catalog) ChipsBy(maker string) []*Chip { return c.makerOf[maker] }

// CountByCategory returns how many chips each category holds. The map is
// shared: don't modify it.
func (c *Catalog) CountByCategory() map[string]int {
	return derived(c, "countByCategory", func() map[string]int {
		n := map[string]int{}
		for _, chip := range c.Chips {
			n[chip.Category]++
		}
		return n
	})
}

// YearSpan returns the earliest and latest release years.
func (c *Catalog) YearSpan() (first, last int) {
	span := derived(c, "yearSpan", func() [2]int { f, l := c.yearSpan(); return [2]int{f, l} })
	return span[0], span[1]
}

func (c *Catalog) yearSpan() (first, last int) {
	for _, chip := range c.Chips {
		y := chip.Year()
		if y == 0 {
			continue
		}
		if first == 0 || y < first {
			first = y
		}
		if y > last {
			last = y
		}
	}
	return first, last
}

// Max returns the highest numeric value of a spec across the catalog, for
// scaling benchmark bars.
func (c *Catalog) Max(key string) float64 {
	c.maxMu.Lock()
	defer c.maxMu.Unlock()
	if v, ok := c.max[key]; ok {
		return v
	}
	if c.max == nil {
		c.max = map[string]float64{}
	}
	m := 0.0
	for _, chip := range c.Chips {
		if v, ok := chip.Number(key); ok && v > m {
			m = v
		}
	}
	c.max[key] = m
	return m
}
