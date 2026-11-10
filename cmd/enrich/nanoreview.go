package main

import (
	"net/url"
	"regexp"
	"strings"
)

// nanoReview supplies technical specs for CPUs, APUs, SoCs and GPUs. Its
// robots.txt allows these paths; only the statistics page is disallowed and
// the tool never reads it.
type nanoReview struct {
	f *fetcher
}

// nanoSlugs guesses NanoReview page names. CPU pages use the database id
// as-is; GPU pages drop the manufacturer prefix ("nvidia-geforce-rtx-4090"
// is "geforce-rtx-4090").
func nanoSlugs(id, name, maker, category string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(s string) {
		s = strings.Trim(s, "-")
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	switch category {
	case "gpu":
		add(strings.TrimPrefix(id, maker+"-"))
		add(slugify(name))
		add(slugify(maker + "-" + name))
	default:
		add(id)
		add(slugify(maker + "-" + name))
		add(slugify(name))
	}
	return out
}

// specs fetches the page for a chip and returns its specs. kind picks the
// NanoReview section: graphics cards live under /gpu/, everything else
// under /cpu/.
func (n *nanoReview) specs(maker, category, id, name string) (map[string]any, error) {
	kind := "cpu"
	switch category {
	case "gpu":
		kind = "gpu"
	case "soc":
		// Phone and tablet chips have their own section.
		kind = "soc"
	}
	// Page names differ from this database's ids in small ways, so try each
	// guess and keep whichever page yielded the most specs. Ambiguous pages
	// can belong to a sibling model and carry fewer rows.
	var best map[string]any
	for _, slug := range nanoSlugs(id, name, maker, category) {
		page, err := n.f.get("https://nanoreview.net/en/" + kind + "/" + url.PathEscape(slug))
		if err != nil {
			if skipMiss(err) {
				continue
			}
			return nil, err
		}
		var specs map[string]any
		switch kind {
		case "gpu":
			specs = parseGPU(page)
		case "soc":
			specs = parseSoC(page)
		default:
			specs = parseCPU(page)
		}
		if len(specs) > len(best) {
			best = specs
		}
	}
	return best, nil
}

// nanoValue finds the text that follows label in a flattened NanoReview
// page, e.g. "Base Frequency (P)|4.3 GHz|Turbo Boost Frequency (P)|5.7 GHz".
func nanoValue(t, label string) string {
	needle := "|" + label + "|"
	if i := strings.Index(t, needle); i >= 0 {
		rest := t[i+len(needle):]
		if j := strings.IndexByte(rest, '|'); j >= 0 {
			return strings.TrimSpace(rest[:j])
		}
		return strings.TrimSpace(rest)
	}
	// The first label in a block has no leading separator.
	if strings.HasPrefix(t, label+"|") {
		rest := t[len(label)+1:]
		if j := strings.IndexByte(rest, '|'); j >= 0 {
			return strings.TrimSpace(rest[:j])
		}
	}
	return ""
}

var listItems = regexp.MustCompile(`-\s*([A-Za-z0-9][^|]*)`)

// parseCPU extracts spec values from a NanoReview CPU page.
func parseCPU(page []byte) map[string]any {
	body := text(page)
	out := map[string]any{}
	put := func(key string, v any) {
		if _, ok := out[key]; !ok {
			out[key] = v
		}
	}

	if v := nanoValue(body, "Model number"); v != "" && v != "-" {
		put("model_number", v)
	}
	if v := nanoValue(body, "Instruction Set"); v != "" && v != "-" {
		put("isa", v)
	}
	if v := nanoValue(body, "Codename"); v != "" && v != "-" {
		put("codename", stripParens(v))
	}
	if v := nanoValue(body, "Peak Temperature"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("peak_temperature_c", c)
		}
	}
	if v := nanoValue(body, "Unlocked Multiplier"); v != "" {
		switch strings.ToLower(v) {
		case "yes":
			put("unlocked_multiplier", true)
		case "no":
			put("unlocked_multiplier", false)
		}
	}
	if v := nanoValue(body, "Fabrication Process"); v != "" {
		if nm, ok := processNM(v); ok {
			put("process_nm", nm)
		}
	}
	if v := nanoValue(body, "Transistors"); v != "" {
		if c, ok := transistors(v); ok {
			put("transistors", c)
		}
	}
	if v := nanoValue(body, "Die Size"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("die_size_mm2", c)
		}
	}
	if v := nanoValue(body, "TDP (PL1)"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("tdp_w", c)
		}
	}
	if v := nanoValue(body, "Socket"); v != "" && v != "-" {
		put("socket", v)
	}

	// Clocks: the P-core figures apply to the whole chip for x86 parts.
	base := firstNonEmpty(nanoValue(body, "Base Frequency (P)"), nanoValue(body, "Base Frequency (E)"))
	if c, ok := clockMHz(base); ok {
		put("base_clock_mhz", c)
	}
	boost := firstNonEmpty(nanoValue(body, "Turbo Boost Frequency (P)"), nanoValue(body, "Turbo Frequency (P)"))
	if c, ok := clockMHz(boost); ok {
		put("boost_clock_mhz", c)
	}

	if v := nanoValue(body, "Total Cores"); v != "" {
		if c, ok := toNumber(v); ok {
			put("cores", c)
		}
	}
	if v := nanoValue(body, "Total Threads"); v != "" {
		if c, ok := toNumber(v); ok {
			put("threads", c)
		}
	}
	if v := nanoValue(body, "P-Cores"); v != "" {
		if c, ok := toNumber(v); ok {
			put("performance_cores", c)
		}
	}
	if v := nanoValue(body, "E-Cores"); v != "" {
		if c, ok := toNumber(v); ok {
			put("efficiency_cores", c)
		}
	}

	// Caches are given per core for L1/L2 and shared for L3.
	cores, _ := out["cores"].(float64)
	if kb, ok := nanoValueKB(body, "L1 Cache", "per core", cores); ok {
		put("l1_cache_kb", kb)
	}
	if kb, ok := nanoValueKB(body, "L2 Cache", "per core", cores); ok {
		put("l2_cache_kb", kb)
	}
	if kb, ok := nanoValueKB(body, "L3 Cache", "", 1); ok {
		put("l3_cache_kb", kb)
	}

	// Integrated graphics. The specification table labels the row "Integrated
	// GPU"; the "Integrated Graphics" heading higher up is a score, so it must
	// not be used.
	// "No" means the chip has no integrated graphics, which the spec should
	// simply omit rather than show as the word "No".
	if v := nanoValue(body, "Integrated GPU"); v != "" && v != "-" && !strings.EqualFold(v, "no") {
		put("gpu", v)
	}
	if v := nanoValue(body, "GPU Boost Clock"); v != "" {
		if c, ok := clockMHz(v); ok {
			put("gpu_boost_clock_mhz", c)
		}
	}
	if v := nanoValue(body, "GPU Base Clock"); v != "" {
		if c, ok := clockMHz(v); ok {
			put("gpu_clock_mhz", c)
		}
	}
	if v := nanoValue(body, "Shading Units"); v != "" {
		if c, ok := toNumber(v); ok {
			put("shading_units", c)
		}
	}
	if v := nanoValue(body, "Execution Units"); v != "" {
		if c, ok := toNumber(v); ok {
			put("execution_units", c)
		}
	}
	if v := nanoValue(body, "TMUs"); v != "" {
		if c, ok := toNumber(v); ok {
			put("tmus", c)
		}
	}
	if v := nanoValue(body, "ROPs"); v != "" {
		if c, ok := toNumber(v); ok {
			put("rops", c)
		}
	}
	if v := nanoValue(body, "iGPU FLOPS"); v != "" {
		if f, ok := flops(v); ok {
			put("fp32_tflops", f)
		}
	}

	// Memory support.
	if v := nanoValue(body, "Max. Memory Size"); v != "" {
		if mb, ok := capacityMB(v); ok {
			put("max_memory_mb", mb)
		}
	}
	if v := nanoValue(body, "Memory Channels"); v != "" {
		if c, ok := toNumber(v); ok {
			put("memory_channels", c)
		}
	}
	if v := nanoValue(body, "Max. Memory Bandwidth"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("memory_bandwidth_gbs", c)
		}
	}
	if v := nanoValue(body, "ECC Support"); v != "" {
		switch strings.ToLower(v) {
		case "yes":
			put("ecc_memory", true)
		case "no":
			put("ecc_memory", false)
		}
	}
	if v := memoryTypes(body); len(v) > 0 {
		put("memory_type", strings.Join(v, " / "))
		if mts, ok := topMemorySpeed(v); ok {
			put("memory_speed_mts", mts)
		}
	}

	// PCI Express.
	if v := nanoValue(body, "PCI Express Version"); v != "" && v != "-" {
		put("pcie_version", v)
	}
	if v := nanoValue(body, "PCI Express Lanes"); v != "" {
		if c, ok := toNumber(v); ok {
			put("pcie_lanes", c)
		}
	}

	// Benchmarks. NanoReview reports Cinebench 2024 and Geekbench 6, which
	// are the scores this database shows.
	for key, label := range map[string]string{
		"geekbench6_single":     "Geekbench 6 (Single-Core)",
		"geekbench6_multi":      "Geekbench 6 (Multi-Core)",
		"cinebench_2024_single": "Cinebench 2024 (Single)",
		"cinebench_2024_multi":  "Cinebench 2024 (Multi)",
	} {
		if v := nanoValue(body, label); v != "" {
			if f, _, ok := parseNumber(v); ok {
				put(key, f)
			}
		}
	}

	// Technologies become the instruction-set extensions list.
	if v := technologies(body); len(v) > 0 {
		put("instruction_sets", v)
	}
	return out
}

// parseGPU extracts spec values from a NanoReview graphics card page.
func parseGPU(page []byte) map[string]any {
	body := text(page)
	out := map[string]any{}
	put := func(key string, v any) {
		if _, ok := out[key]; !ok {
			out[key] = v
		}
	}

	if v := nanoValue(body, "Architecture"); v != "" && v != "-" {
		put("microarchitecture", v)
	}
	if v := nanoValue(body, "GPU Codename"); v != "" && v != "-" {
		put("codename", v)
	}
	if v := nanoValue(body, "Launch price (MSRP)"); v != "" && v != "-" {
		if c, _, ok := parseNumber(v); ok {
			put("launch_price_usd", c)
		}
	}
	if v := nanoValue(body, "Base Clock"); v != "" {
		if c, ok := clockMHz(v); ok {
			put("gpu_clock_mhz", c)
		}
	}
	if v := nanoValue(body, "Boost Clock"); v != "" {
		if c, ok := clockMHz(v); ok {
			put("gpu_boost_clock_mhz", c)
		}
	}
	if v := nanoValue(body, "Shading Units"); v != "" {
		if c, ok := toNumber(v); ok {
			put("shading_units", c)
		}
	}
	if v := nanoValue(body, "Texture Mapping Units (TMUs)"); v != "" {
		if c, ok := toNumber(v); ok {
			put("tmus", c)
		}
	}
	if v := nanoValue(body, "Render Output Units (ROPs)"); v != "" {
		if c, ok := toNumber(v); ok {
			put("rops", c)
		}
	}
	if v := nanoValue(body, "Compute Units (Pipelines)"); v != "" {
		if c, ok := toNumber(v); ok {
			put("compute_units", c)
		}
	}
	if v := nanoValue(body, "Tensor Cores"); v != "" {
		if c, ok := toNumber(v); ok {
			put("tensor_cores", c)
		}
	}
	if v := nanoValue(body, "Ray-tracing Cores"); v != "" {
		if c, ok := toNumber(v); ok {
			put("rt_cores", c)
		}
	}
	if kb, ok := nanoValueKB(body, "L1 Cache", "", 1); ok {
		put("l1_cache_kb", kb)
	}
	if kb, ok := nanoValueKB(body, "L2 Cache", "", 1); ok {
		put("l2_cache_kb", kb)
	}
	if v := nanoValue(body, "Pixel Fill Rate"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("pixel_rate_gpixels", c)
		}
	}
	if v := nanoValue(body, "Texture Fill Rate"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("texture_rate_gtexels", c)
		}
	}
	if v := nanoValue(body, "FLOPS (FP32)"); v != "" {
		if f, ok := flops(v); ok {
			put("fp32_tflops", f)
		}
	}
	if v := nanoValue(body, "Interface"); v != "" && v != "-" {
		put("interface", v)
	}
	if v := nanoValue(body, "TGP"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("tdp_w", c)
		}
	}
	if v := nanoValue(body, "Fabrication Process"); v != "" {
		if nm, ok := processNM(v); ok {
			put("process_nm", nm)
		}
	}
	if v := nanoValue(body, "Manufacturing"); v != "" && v != "-" {
		put("foundry", v)
	}
	if v := nanoValue(body, "Die Size"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("die_size_mm2", c)
		}
	}
	if v := nanoValue(body, "Transistor Count"); v != "" {
		if c, ok := transistors(v); ok {
			put("transistors", c)
		}
	}
	if v := nanoValue(body, "Memory Type"); v != "" && v != "-" {
		put("memory_type", v)
	}
	if v := nanoValue(body, "Memory Size"); v != "" {
		if mb, ok := capacityMB(v); ok {
			put("memory_size_mb", mb)
		}
	}
	if v := nanoValue(body, "Memory Clock"); v != "" {
		if c, ok := clockMHz(v); ok {
			put("memory_clock_mhz", c)
		}
	}
	if v := nanoValue(body, "Effective Memory Speed"); v != "" {
		if c, ok := toNumber(v); ok {
			put("memory_speed_mts", c)
		}
	}
	if v := nanoValue(body, "Bus"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("memory_bus_bits", c)
		}
	}
	if v := nanoValue(body, "Memory Bandwidth"); v != "" {
		if c, _, ok := parseNumber(v); ok {
			put("memory_bandwidth_gbs", c)
		}
	}
	if v := nanoValue(body, "ECC"); v != "" {
		switch strings.ToLower(v) {
		case "yes":
			put("ecc_memory", true)
		case "no":
			put("ecc_memory", false)
		}
	}

	// APIs are spread over separate rows.
	var apis []string
	for _, a := range []struct{ label, name string }{
		{"DirectX", "DirectX"}, {"Vulkan", "Vulkan"}, {"OpenGL", "OpenGL"}, {"OpenCL", "OpenCL"},
	} {
		if v := nanoValue(body, a.label); v != "" && v != "-" {
			apis = append(apis, a.name+" "+v)
		}
	}
	if len(apis) > 0 {
		put("graphics_api", apis)
	}
	return out
}

// nanoValueKB reads a cache field, scaling a per-core figure by cores.
func nanoValueKB(body, label, perCore string, cores float64) (float64, bool) {
	v := nanoValue(body, label)
	if v == "" {
		return 0, false
	}
	kb, ok := cacheSizeKB(v)
	if !ok {
		return 0, false
	}
	if perCore != "" && strings.Contains(strings.ToLower(v), perCore) && cores > 1 {
		kb *= cores
	}
	return kb, true
}

// memoryTypes reads the bulleted memory list that follows "Memory Types".
func memoryTypes(body string) []string {
	i := strings.Index(body, "Memory Types|")
	if i < 0 {
		return nil
	}
	seg := body[i+len("Memory Types|"):]
	if j := strings.Index(seg, "Max. Memory"); j >= 0 {
		seg = seg[:j]
	}
	var out []string
	for _, m := range listItems.FindAllStringSubmatch(seg, -1) {
		s := strings.TrimSpace(m[1])
		if s != "" && s != "-" {
			out = append(out, s)
		}
	}
	return out
}

// topMemorySpeed returns the highest MT/s figure in a memory-type list such
// as "DDR5-5600 / DDR4-3200".
func topMemorySpeed(types []string) (float64, bool) {
	best, found := 0.0, false
	for _, t := range types {
		loc := regexp.MustCompile(`-(\d{3,5})`).FindStringSubmatch(t)
		if loc == nil {
			continue
		}
		if v, _, ok := parseNumber(loc[1]); ok && v > best {
			best, found = v, true
		}
	}
	return best, found
}

// technologies joins the "Technologies" bullet list into an extension list,
// dropping the category prefixes ("Compute:", "Virtualization:") so the
// rows read as plain extension names.
func technologies(body string) []string {
	i := strings.Index(body, "Technologies|")
	if i < 0 {
		return nil
	}
	seg := body[i+len("Technologies|"):]
	if j := strings.Index(seg, "Competitors"); j >= 0 {
		seg = seg[:j]
	}
	var out []string
	seen := map[string]bool{}
	for _, m := range listItems.FindAllStringSubmatch(seg, -1) {
		s := strings.TrimSpace(m[1])
		if k := strings.LastIndex(s, ":"); k >= 0 {
			s = strings.TrimSpace(s[k+1:])
		}
		for _, part := range strings.Split(s, ",") {
			p := strings.TrimSpace(part)
			if p == "" || p == "-" || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// stripParens drops a trailing parenthetical, so "Zen 5 (Granite Ridge)"
// becomes "Zen 5".
func stripParens(s string) string {
	if i := strings.Index(s, "("); i > 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}
