package main

import (
	"os"
	"testing"
)

// TestParseCPUSample checks the NanoReview parser against a saved page.
func TestParseCPUSample(t *testing.T) {
	raw, err := os.ReadFile("testdata/nanoreview-cpu-amd-ryzen-9-9950x.html")
	if err != nil {
		t.Skip("no sample page")
	}
	got := parseCPU(raw)
	want := map[string]any{
		"cores":            16.0,
		"threads":          32.0,
		"process_nm":       4.0,
		"l3_cache_kb":      65536.0,
		"memory_channels":  2.0,
		"memory_speed_mts": 5600.0,
		"pcie_lanes":       28.0,
	}
	for k, v := range want {
		if g, ok := got[k]; !ok {
			t.Errorf("%s: missing, want %v", k, v)
		} else if g != v {
			t.Errorf("%s: got %v, want %v", k, g, v)
		}
	}
	if s, _ := got["pcie_version"].(string); s != "5.0" {
		t.Errorf("pcie_version: got %q, want 5.0", s)
	}
	if g, ok := got["instruction_sets"].([]string); !ok || len(g) == 0 {
		t.Errorf("instruction_sets: got %v, want a non-empty list", got["instruction_sets"])
	}
	if g, ok := got["memory_type"].(string); !ok || g != "DDR5-5600" {
		t.Errorf("memory_type: got %v, want DDR5-5600", got["memory_type"])
	}
}

// TestParseGPUSample checks the NanoReview graphics parser.
func TestParseGPUSample(t *testing.T) {
	raw, err := os.ReadFile("testdata/nanoreview-gpu-geforce-rtx-4090.html")
	if err != nil {
		t.Skip("no sample page")
	}
	got := parseGPU(raw)
	want := map[string]any{
		"shading_units":        16384.0,
		"tmus":                 512.0,
		"rops":                 176.0,
		"compute_units":        128.0,
		"tensor_cores":         512.0,
		"rt_cores":             128.0,
		"gpu_clock_mhz":        2235.0,
		"gpu_boost_clock_mhz":  2520.0,
		"memory_size_mb":       24576.0,
		"memory_bus_bits":      384.0,
		"memory_bandwidth_gbs": 1010.0,
		"fp32_tflops":          82.6,
		"l2_cache_kb":          73728.0,
	}
	for k, v := range want {
		if g, ok := got[k]; !ok {
			t.Errorf("%s: missing, want %v", k, v)
		} else if g != v {
			t.Errorf("%s: got %v, want %v", k, g, v)
		}
	}
	if s, _ := got["codename"].(string); s != "AD102" {
		t.Errorf("codename: got %q, want AD102", s)
	}
}

// TestParseCPUSample3D guards the label that a score heading could shadow:
// "Integrated GPU" is the spec row, "Integrated Graphics" is a rating.
func TestParseCPUSample3D(t *testing.T) {
	raw, err := os.ReadFile("testdata/nanoreview-cpu-amd-ryzen-9-9950x3d.html")
	if err != nil {
		t.Skip("no sample page")
	}
	got := parseCPU(raw)
	if s, _ := got["gpu"].(string); s != "Radeon Graphics" {
		t.Errorf("gpu: got %v, want Radeon Graphics", got["gpu"])
	}
	if v, ok := got["ecc_memory"].(bool); !ok || !v {
		t.Errorf("ecc_memory: got %v, want true", got["ecc_memory"])
	}
	if v, ok := got["unlocked_multiplier"].(bool); !ok || !v {
		t.Errorf("unlocked_multiplier: got %v, want true", got["unlocked_multiplier"])
	}
}

// TestParseMonkeySample checks the benchmark extraction.
func TestParseMonkeySample(t *testing.T) {
	raw, err := os.ReadFile("testdata/cpumonkey-cpu-intel_core_i9_14900k.html")
	if err != nil {
		t.Skip("no sample page")
	}
	got := cpuMonkeyScores(raw)
	want := map[string]float64{
		"geekbench6_single":     3289,
		"geekbench6_multi":      23051,
		"cinebench_r23_single":  2358,
		"cinebench_r23_multi":   38712,
		"cinebench_2024_single": 139,
		"cinebench_2024_multi":  2177,
	}
	for k, v := range want {
		if g, ok := got[k]; !ok {
			t.Errorf("%s: missing, want %v", k, v)
		} else if g != v {
			t.Errorf("%s: got %v, want %v", k, g, v)
		}
	}
}
