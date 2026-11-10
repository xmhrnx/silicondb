package main

import (
	"fmt"
	"sort"
	"strings"
)

// schema describes how to lay out a shard: spec keys follow the order of
// the fields in schema.json, benchmarks come last, and keys the schema
// doesn't know keep their relative order at the end.
type schema struct {
	order   map[string]int
	group   map[string]string
	known   map[string]bool
	naOrder map[string]int
}

func newSchema(doc *node) *schema {
	s := &schema{order: map[string]int{}, group: map[string]string{}, known: map[string]bool{}, naOrder: map[string]int{}}
	i := 0
	for _, g := range doc.get("groups").arr {
		id, _ := g.get("id").str()
		for _, f := range g.get("fields").arr {
			key, _ := f.get("key").str()
			if key == "" {
				continue
			}
			if _, ok := s.order[key]; !ok {
				s.order[key] = i
				s.group[key] = id
			}
			s.known[key] = true
			i++
		}
	}
	return s
}

// isBenchmark reports whether key holds a benchmark score, which the files
// place after every other spec.
func isBenchmark(key string) bool {
	switch {
	case len(key) > 10 && key[:10] == "geekbench6":
		return true
	case len(key) > 9 && key[:9] == "cinebench":
		return true
	}
	return false
}

// mergeable decides whether a value pulled from a source may be stored.
// The sources mix units and occasionally publish something odd, so only
// positive numbers and non-empty strings are accepted, and a handful of
// keys get range checks.
func mergeable(key string, v any) bool {
	switch v := v.(type) {
	case float64:
		if v <= 0 {
			return false
		}
		switch key {
		case "process_nm":
			return v >= 0.5 && v <= 100000
		case "transistors":
			return v >= 10
		case "cores", "threads", "shading_units", "tmus", "rops", "compute_units",
			"tensor_cores", "rt_cores", "execution_units", "memory_channels",
			"pcie_lanes", "launch_price_usd":
			return v <= 1e7
		case "tdp_w":
			return v <= 100000
		case "max_memory_mb", "memory_size_mb":
			return v >= 1
		case "memory_speed_mts":
			return v >= 100 && v <= 100000
		case "fp32_tflops", "fp16_tflops", "fp64_tflops", "bf16_tflops", "fp8_tflops", "fp4_tflops":
			return v <= 1e5
		case "l1_cache_kb", "l2_cache_kb", "l3_cache_kb", "slc_kb", "l0_cache_kb":
			return v <= 1e8
		}
		return true
	case string:
		s := strings.TrimSpace(v)
		switch s {
		case "", "-", "n/a", "N/A", "none", "None", "unknown", "Unknown":
			return false
		}
		// A bare number is never a name, a codename or an API list; it means
		// the page had a stray link where the value should be.
		if _, _, ok := parseNumber(s); ok && !strings.ContainsAny(s, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
			return false
		}
		return true
	case []string:
		if len(v) == 0 {
			return false
		}
		for _, s := range v {
			if strings.TrimSpace(s) == "" {
				return false
			}
		}
		return true
	case bool:
		return true
	}
	return false
}

// toNode converts a parsed value to a JSON node.
func toNode(v any) *node {
	switch v := v.(type) {
	case float64:
		return newNumber(v)
	case string:
		return newString(v)
	case bool:
		return newBool(v)
	case []string:
		return newStrings(v)
	}
	return nil
}

// sortSpecs reorders a chip's specs: schema keys in schema order, then
// unknown keys in their current order, then benchmarks. The na list is
// sorted so diffs stay small.
func (s *schema) sortSpecs(specs *node) {
	if specs == nil || specs.kind != 'o' {
		return
	}
	var known, unknown, bench []field
	for _, f := range specs.obj {
		switch {
		case isBenchmark(f.key):
			bench = append(bench, f)
		case s.known[f.key]:
			known = append(known, f)
		default:
			unknown = append(unknown, f)
		}
	}
	sort.SliceStable(known, func(i, j int) bool {
		return s.order[known[i].key] < s.order[known[j].key]
	})
	specs.obj = append(append(known, unknown...), bench...)
}

// mergeSpecs writes the source values into the chip that aren't already
// there, and removes them from the na list. It returns the keys added.
func mergeSpecs(specs *node, na *node, vals map[string]any, allow map[string]bool) []string {
	var added []string
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := vals[k]
		if allow != nil && !allow[k] {
			continue
		}
		if specs.get(k) != nil {
			continue
		}
		if !mergeable(k, v) {
			continue
		}
		n := toNode(v)
		if n == nil {
			continue
		}
		specs.set(k, n)
		if na != nil {
			na.remove(k)
		}
		added = append(added, k)
	}
	return added
}

// allowedKeys limits which source values a category accepts, so a CPU page
// can't add graphics-card fields and the reverse.
func allowedKeys(category, source string) map[string]bool {
	cpuCommon := []string{
		"model_number", "isa", "codename", "peak_temperature_c", "unlocked_multiplier",
		"process_nm", "transistors", "die_size_mm2", "tdp_w", "socket",
		"base_clock_mhz", "boost_clock_mhz", "l1_cache_kb", "l2_cache_kb", "l3_cache_kb",
		"gpu", "gpu_clock_mhz", "gpu_boost_clock_mhz", "shading_units", "execution_units",
		"tmus", "rops",
		"max_memory_mb", "memory_channels", "memory_bandwidth_gbs", "memory_type",
		"memory_speed_mts", "ecc_memory", "pcie_version", "pcie_lanes", "instruction_sets",
		"geekbench6_single", "geekbench6_multi", "geekbench6_opencl",
		"cinebench_2024_single", "cinebench_2024_multi",
	}
	gpuKeys := []string{
		"microarchitecture", "codename", "launch_price_usd",
		"gpu_clock_mhz", "gpu_boost_clock_mhz", "shading_units", "tmus", "rops",
		"compute_units", "tensor_cores", "rt_cores", "l1_cache_kb", "l2_cache_kb",
		"pixel_rate_gpixels", "texture_rate_gtexels", "fp32_tflops",
		"interface", "tdp_w", "process_nm", "foundry", "die_size_mm2", "transistors",
		"memory_type", "memory_size_mb", "memory_clock_mhz", "memory_speed_mts",
		"memory_bus_bits", "memory_bandwidth_gbs", "ecc_memory", "graphics_api",
	}
	socKeys := []string{
		"isa", "model_number", "market", "cores", "performance_cores", "efficiency_cores",
		"core_config", "boost_clock_mhz", "l1_cache_kb", "l2_cache_kb", "l3_cache_kb",
		"process_nm", "transistors", "tdp_w", "foundry",
		"gpu", "gpu_cores", "shading_units", "compute_units", "execution_units",
		"gpu_clock_mhz", "gpu_architecture", "fp32_tflops", "graphics_api",
		"memory_type", "memory_speed_mts", "memory_bandwidth_gbs", "max_memory_mb",
		"memory_channels", "memory_bus_bits", "modem", "connectivity",
		"geekbench6_single", "geekbench6_multi",
	}
	var keys []string
	switch {
	case source == "nanoreview" && category == "gpu":
		keys = gpuKeys
	case source == "nanoreview" && category == "soc":
		keys = socKeys
	case source == "nanoreview":
		keys = cpuCommon
	}
	if keys == nil {
		return nil
	}
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}

func (s *schema) describe(key string) string {
	if g, ok := s.group[key]; ok {
		return fmt.Sprintf("%s (%s)", key, g)
	}
	return key
}
