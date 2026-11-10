package main

import (
	"os"
	"testing"
)

// TestParseSoCSample checks the SoC parser, whose page layout differs from
// the desktop CPU one.
func TestParseSoCSample(t *testing.T) {
	raw, err := os.ReadFile("testdata/nanoreview-soc-mediatek-dimensity-9300.html")
	if err != nil {
		t.Skip("no sample page")
	}
	got := parseSoC(raw)
	want := map[string]any{
		"cores":                8.0,
		"performance_cores":    4.0,
		"efficiency_cores":     4.0,
		"boost_clock_mhz":      3250.0,
		"process_nm":           4.0,
		"transistors":          22.7e9,
		"tdp_w":                7.0,
		"shading_units":        128.0,
		"gpu_clock_mhz":        1300.0,
		"l1_cache_kb":          128.0,
		"l2_cache_kb":          1024.0,
		"l3_cache_kb":          8192.0,
		"memory_speed_mts":     4800.0,
		"memory_bandwidth_gbs": 76.8,
		"max_memory_mb":        24576.0,
	}
	for k, v := range want {
		g, ok := got[k]
		if !ok {
			t.Errorf("%s: missing, want %v", k, v)
			continue
		}
		if f, ok := g.(float64); ok && f != v {
			t.Errorf("%s: got %v, want %v", k, f, v)
		}
	}
	if s, _ := got["gpu"].(string); s != "Mali-G720 Immortalis MP12" {
		t.Errorf("gpu: got %q", s)
	}
	if s, _ := got["gpu_architecture"].(string); s != "5th gen" {
		t.Errorf("gpu_architecture: got %q, want 5th gen", s)
	}
	if s, _ := got["memory_type"].(string); s != "LPDDR5T" {
		t.Errorf("memory_type: got %q", s)
	}
	if s, _ := got["modem"].(string); s != "MediaTek T830" {
		t.Errorf("modem: got %q", s)
	}
}

// TestWikiIndex checks that a throughput table's bare numbers become TFLOPS,
// using the unit in the header above them.
func TestWikiIndex(t *testing.T) {
	raw, err := os.ReadFile("testdata/wiki-flops-table.html")
	if err != nil {
		t.Skip("no sample page")
	}
	idx := indexWiki(raw, "nvidia")
	for name, want := range map[string]float64{
		"geforce rtx 4090": 82.58,
		"geforce rtx 4060": 17.00, // "(boost)" figure, as tables list it
	} {
		vals, ok := idx[name]
		if !ok {
			t.Errorf("%s: not indexed", name)
			continue
		}
		if got, _ := vals["fp32_tflops"].(float64); got != want {
			t.Errorf("%s fp32: got %v, want %v", name, got, want)
		}
	}
	if vals := idx["geforce rtx 4090"]; vals["fp8_tensor_tflops"] != 660.60 {
		t.Errorf("geforce rtx 4090 fp8 tensor: got %v, want 660.60", vals["fp8_tensor_tflops"])
	}
	// GFLOPS columns are scaled down to TFLOPS.
	if vals := idx["geforce 205 oem"]; vals["fp32_tflops"] != 0.044864 {
		t.Errorf("geforce 205 oem fp32: got %v, want 0.044864", vals["fp32_tflops"])
	}
}

// TestDeriveMemorySpeed checks the local derivation from memory_type.
func TestDeriveMemorySpeed(t *testing.T) {
	cases := []struct {
		typ  string
		want float64
		ok   bool
	}{
		{"DDR5-6400", 6400, true},
		{"DDR5-5600 / DDR4-3200", 5600, true},
		{"LPDDR5X-8533", 8533, true},
		{"DDR5", 0, false},
		{"DDR4 (registered)", 0, false},
		{"", 0, false},
	}
	for _, c := range cases {
		got, ok := topRating(c.typ)
		if ok != c.ok || (ok && got != c.want) {
			t.Errorf("topRating(%q) = %v, %v; want %v, %v", c.typ, got, ok, c.want, c.ok)
		}
	}
}
