package format

import (
	"strings"

	"github.com/mehran9675/silicondb/internal/model"
)

// Value formats a spec value as described by field. field may be nil for
// keys the schema doesn't know.
func Value(field *model.Field, v any) string {
	switch v := v.(type) {
	case float64:
		if field == nil {
			return Number(v, 2)
		}
		return withUnit(field, v)
	case string:
		return v
	case bool:
		if v {
			return text("common.yes", "Yes")
		}
		return text("common.no", "No")
	case []any:
		parts := make([]string, 0, len(v))
		for _, e := range v {
			if s, ok := e.(string); ok {
				parts = append(parts, s)
			}
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

func withUnit(f *model.Field, v float64) string {
	switch f.Format {
	case "mhz":
		switch {
		case v >= 1000:
			return Number(v/1000, 2) + " GHz"
		case v < 1:
			return Number(v*1000, 0) + " kHz"
		}
		return Number(v, 2) + " MHz"
	case "kb":
		return bytes(v * 1024)
	case "mb":
		return bytes(v * 1024 * 1024)
	case "nm":
		if v >= 1000 {
			return Number(v/1000, 1) + " µm"
		}
		return Number(v, 1) + " nm"
	case "count":
		return Count(v)
	case "score":
		return Number(v, 0)
	case "usd":
		return "$" + Number(v, 0)
	case "tflops":
		if v < 1 {
			return Number(v*1000, 1) + " GFLOPS"
		}
		return Number(v, 2) + " TFLOPS"
	case "pcie":
		return "PCIe " + Number(v, 1)
	}
	if f.Unit != "" {
		return Number(v, 2) + " " + f.Unit
	}
	return Number(v, 2)
}

func bytes(b float64) string {
	const k = 1024.0
	switch {
	case b >= k*k*k*k:
		return Number(b/(k*k*k*k), 2) + " TB"
	case b >= k*k*k:
		return Number(b/(k*k*k), 2) + " GB"
	case b >= k*k:
		return Number(b/(k*k), 2) + " MB"
	case b >= k:
		return Number(b/k, 2) + " KB"
	}
	return Number(b, 0) + " B"
}

// Label returns field's label, or one derived from key when the schema
// doesn't know it: "pipeline_stages" becomes "Pipeline stages".
func Label(field *model.Field, key string) string {
	if field != nil && field.Label != "" {
		return field.Label
	}
	s := strings.ReplaceAll(key, "_", " ")
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
