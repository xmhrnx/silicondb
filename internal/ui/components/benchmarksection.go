package components

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/ui/theme"
)

// Benchmark is one score ready to show.
type Benchmark struct {
	Label, Score string
	// Fraction is the score divided by the best score in the database.
	Fraction float32
}

// BenchmarkSection is a card of benchmark bars in the chip's color, with
// a note on what the scores mean.
func BenchmarkSection(title, note, color string, scores []Benchmark) giode.Element {
	rows := []giode.Element{giode.Box(giode.Styles{Padding: giode.Inset{Bottom: 2}},
		giode.Text(title, theme.Overline().With(giode.Styles{Color: theme.Primary})))}
	for _, b := range scores {
		rows = append(rows, BenchmarkBar(b.Label, b.Score, b.Fraction, color))
	}
	rows = append(rows, giode.Box(giode.Styles{Padding: giode.Inset{Top: 6}},
		giode.Text(note, theme.Caption().With(giode.Styles{LineHeight: 17}))))
	return Card(giode.Styles{Padding: giode.Inset{Top: 16, Bottom: 14, Left: 16, Right: 16}}, rows...)
}
