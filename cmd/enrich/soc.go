package main

import (
	"regexp"
	"strings"
)

// parseSoC extracts specs from a NanoReview phone/tablet SoC page. Those
// pages have their own wording ("CPU Architecture", "Frequency", "Process",
// "Total shaders") and report Geekbench 6 as a "single / multi" pair.
func parseSoC(page []byte) map[string]any {
	body := text(page)
	out := map[string]any{}
	put := func(key string, v any) {
		if _, ok := out[key]; !ok {
			out[key] = v
		}
	}

	if v := nanoValue(body, "Instruction set"); v != "" && v != "-" {
		put("isa", v)
	}
	if v := nanoValue(body, "Model number"); v != "" && v != "-" {
		put("model_number", v)
	}
	if v := nanoValue(body, "Announced"); v != "" && v != "-" {
		put("announced", v)
	}
	if v := nanoValue(body, "Class"); v != "" && v != "-" {
		put("market_class", v)
	}
	if v := nanoValue(body, "Cores"); v != "" {
		if c, ok := toNumber(v); ok {
			put("cores", c)
		}
	}
	if v := nanoValue(body, "Frequency"); v != "" {
		if c, ok := clockMHz(v); ok {
			put("boost_clock_mhz", c)
		}
	}
	// "1x 3.25 GHz – Cortex-X4" lines give the core layout and the split
	// between performance and efficiency cores. On these pages the three
	// clusters run together without separators, so the raw text is searched.
	if cfg, cores, p, e := socCoreConfig(text(page)); cfg != "" {
		put("core_config", cfg)
		if cores > 0 {
			put("cores", cores)
		}
		if p > 0 {
			put("performance_cores", p)
		}
		if e > 0 {
			put("efficiency_cores", e)
		}
	}
	if v := nanoValue(body, "L1 cache"); v != "" {
		if kb, ok := cacheSizeKB(v); ok {
			put("l1_cache_kb", kb)
		}
	}
	if v := nanoValue(body, "L2 cache"); v != "" {
		if kb, ok := cacheSizeKB(v); ok {
			put("l2_cache_kb", kb)
		}
	}
	if v := nanoValue(body, "L3 cache"); v != "" {
		if kb, ok := cacheSizeKB(v); ok {
			put("l3_cache_kb", kb)
		}
	}
	if v := nanoValue(body, "Process"); v != "" {
		if nm, ok := processNM(v); ok {
			put("process_nm", nm)
		}
	}
	if v := nanoValue(body, "Transistor count"); v != "" {
		if c, ok := transistors(v); ok {
			put("transistors", c)
		}
	}
	if v := nanoValue(body, "TDP (Sustained Power Limit)"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("tdp_w", c)
		}
	}
	if v := nanoValue(body, "Manufacturing"); v != "" && v != "-" {
		put("foundry", v)
	}

	// Graphics.
	if v := nanoValue(body, "GPU name"); v != "" && v != "-" {
		put("gpu", v)
	}
	// The GPU row is labelled just "Architecture", but the CPU core layout
	// also mentions the word, so only take the value inside the graphics
	// block, which starts at "GPU name".
	if v := nanoValue(graphicsSection(body), "Architecture"); v != "" && v != "-" {
		put("gpu_architecture", v)
	}
	if v := nanoValue(body, "GPU frequency"); v != "" {
		if c, ok := clockMHz(v); ok {
			put("gpu_clock_mhz", c)
		}
	}
	if v := nanoValue(body, "Pipelines"); v != "" {
		if c, ok := toNumber(v); ok {
			put("compute_units", c)
			put("gpu_cores", c)
		}
	}
	if v := nanoValue(body, "Shading units"); v != "" {
		if c, ok := toNumber(v); ok {
			put("shading_units", c)
		}
	}
	if v := nanoValue(body, "Total shaders"); v != "" {
		if c, ok := toNumber(v); ok {
			put("execution_units", c)
		}
	}
	if v := nanoValue(body, "FLOPS"); v != "" {
		if f, ok := flops(v); ok {
			put("fp32_tflops", f)
		}
	}
	var apis []string
	for _, a := range []struct{ label, name string }{
		{"Vulkan version", "Vulkan"}, {"OpenCL version", "OpenCL"}, {"DirectX version", "DirectX"},
	} {
		if v := nanoValue(body, a.label); v != "" && v != "-" {
			apis = append(apis, a.name+" "+v)
		}
	}
	if len(apis) > 0 {
		put("graphics_api", apis)
	}

	// Memory.
	if v := nanoValue(body, "Memory type"); v != "" && v != "-" {
		put("memory_type", v)
	}
	if v := nanoValue(body, "Memory frequency"); v != "" {
		if c, ok := clockMHz(v); ok {
			put("memory_speed_mts", c)
		}
	}
	// NanoReview labels this "Gb/s" but means GB/s, the same figure the
	// rest of the database uses.
	if v := nanoValue(body, "Max bandwidth"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("memory_bandwidth_gbs", c)
		}
	}
	if v := nanoValue(body, "Max size"); v != "" {
		if mb, ok := capacityMB(v); ok {
			put("max_memory_mb", mb)
		}
	}
	if v := nanoValue(body, "Bus"); v != "" {
		// "4x 16 Bit" means four 16-bit channels.
		if ch, _, ok := parseNumber(v); ok && ch > 0 {
			put("memory_channels", ch)
			if bits, _, ok := parseNumber(strings.TrimSpace(afterX(v))); ok {
				put("memory_bus_bits", ch*bits)
			}
		}
	}

	// Wireless and storage.
	if v := nanoValue(body, "Modem"); v != "" && v != "-" {
		put("modem", v)
	}
	var conn []string
	for _, label := range []string{"Wi-Fi", "Bluetooth"} {
		if v := nanoValue(body, label); v != "" && v != "-" {
			conn = append(conn, label+" "+v)
		}
	}
	if v := nanoValue(body, "5G support"); strings.EqualFold(v, "yes") {
		conn = append(conn, "5G")
	}
	if v := nanoValue(body, "Storage type"); v != "" && v != "-" {
		conn = append(conn, v+" storage")
	}
	if len(conn) > 0 {
		put("connectivity", conn)
	}

	// Benchmarks: "Geekbench 6 CPU 2186 / 7422".
	if single, multi, ok := socGeekbench(body); ok {
		put("geekbench6_single", single)
		put("geekbench6_multi", multi)
	}
	return out
}

var (
	// coreLine matches a layout line such as "1x 3.25 GHz – Cortex-X4".
	coreLine = regexp.MustCompile(`(\d+)\s*x\s*([0-9.]+\s*(?:GHz|MHz))`)
	// gb6Pair matches "Geekbench 6 CPU 2186 / 7422".
	gb6Pair = regexp.MustCompile(`Geekbench 6 CPU (\d[\d,]*)\s*/\s*(\d[\d,]*)`)
)

// socCoreConfig summarises the core layout and returns the counts of the
// fastest (performance) and remaining (efficiency) cores.
func socCoreConfig(raw string) (cfg string, cores, perf, eff int) {
	i := strings.Index(raw, "Detailed specifications")
	if i < 0 {
		i = 0
	}
	soc := raw[i:]
	// The flattened text separates each token, so the label reads
	// "CPU|Architecture".
	j := strings.Index(soc, "CPU|Architecture|")
	if j < 0 {
		return "", 0, 0, 0
	}
	seg := soc[j+len("CPU|Architecture|"):]
	if k := strings.Index(seg, "Cores"); k >= 0 {
		seg = seg[:k]
	}
	var parts []string
	best := 0.0
	clusters := [][2]int{} // count and clock
	for _, m := range coreLine.FindAllStringSubmatch(seg, -1) {
		n := atoiDefault(m[1], 0)
		if n <= 0 {
			continue
		}
		ghz, _, _ := parseNumber(m[2])
		cores += n
		clusters = append(clusters, [2]int{n, int(ghz * 1000)})
		parts = append(parts, strings.TrimSpace(m[0]))
		if ghz > best {
			best = ghz
		}
	}
	if len(parts) == 0 {
		return "", 0, 0, 0
	}
	// Only the fastest cluster counts as performance cores; the rest are
	// efficiency cores, the same split the desktop chips record.
	for _, c := range clusters {
		if float64(c[1]) >= best*1000-1 {
			perf += c[0]
		} else {
			eff += c[0]
		}
	}
	return strings.Join(parts, " + "), cores, perf, eff
}

// graphicsSection returns the part of the page describing the GPU, so labels
// that also appear in the CPU block ("Architecture") are read from there.
func graphicsSection(body string) string {
	i := strings.Index(body, "GPU name|")
	if i < 0 {
		return body
	}
	seg := body[i:]
	if j := strings.Index(seg, "AI Accelerator"); j >= 0 {
		seg = seg[:j]
	}
	return seg
}

// socGeekbench reads the Geekbench 6 pair NanoReview shows for an SoC.
func socGeekbench(body string) (single, multi float64, ok bool) {
	m := gb6Pair.FindStringSubmatch(body)
	if m == nil {
		return 0, 0, false
	}
	s, _, ok1 := parseNumber(m[1])
	mu, _, ok2 := parseNumber(m[2])
	if !ok1 || !ok2 {
		return 0, 0, false
	}
	return s, mu, true
}

// afterX returns the text after the first "x" in a channel count such as
// "4x 16 Bit".
func afterX(s string) string {
	if i := strings.IndexByte(s, 'x'); i >= 0 {
		return s[i+1:]
	}
	return ""
}
