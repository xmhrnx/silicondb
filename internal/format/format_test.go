package format

import (
	"testing"

	"github.com/mehran9675/giode/i18n"

	"github.com/mehran9675/silicondb/internal/model"
)

func TestValue(t *testing.T) {
	cases := []struct {
		format, unit string
		v            any
		want         string
	}{
		{"mhz", "", 0.5, "500 kHz"},
		{"mhz", "", 740.0, "740 MHz"},
		{"mhz", "", 3200.0, "3.2 GHz"},
		{"kb", "", 512.0, "512 KB"},
		{"kb", "", 36864.0, "36 MB"},
		{"mb", "", 24576.0, "24 GB"},
		{"nm", "", 10000.0, "10 µm"},
		{"nm", "", 3.0, "3 nm"},
		{"count", "", 3500.0, "3,500"},
		{"count", "", 92e9, "92 billion"},
		{"usd", "", 1599.0, "$1,599"},
		{"tflops", "", 82.58, "82.58 TFLOPS"},
		{"pcie", "", 5.0, "PCIe 5"},
		{"", "W", 125.0, "125 W"},
		{"", "", []any{"MMX", "SSE"}, "MMX, SSE"},
	}
	for _, c := range cases {
		f := &model.Field{Format: c.format, Unit: c.unit}
		if got := Value(f, c.v); got != c.want {
			t.Errorf("Value(%s %v) = %q, want %q", c.format, c.v, got, c.want)
		}
	}
}

func TestDate(t *testing.T) {
	for in, want := range map[string]string{"1972": "1972", "1972-04": "Apr 1972", "2024-10-24": "Oct 24, 2024"} {
		if got := Date(in); got != want {
			t.Errorf("Date(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGerman(t *testing.T) {
	i18n.Add("de", map[string]any{"month.4": "Apr.", "date.month_year": "{month} {year}", "number.billion": "Milliarden"})
	i18n.SetLocale("de")
	defer i18n.SetLocale("en")
	cases := []struct {
		got, want string
	}{
		{Value(&model.Field{Format: "mhz"}, 3200.0), "3,2 GHz"},
		{Value(&model.Field{Format: "count"}, 16384.0), "16.384"},
		{Value(&model.Field{Format: "count"}, 92e9), "92 Milliarden"},
		{Date("1972-04"), "Apr. 1972"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}
