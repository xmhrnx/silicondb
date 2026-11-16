package components

import "github.com/xmhrnx/giode"

// Present returns the non-nil elements, for building child lists where
// some sections may have nothing to show.
func Present(els ...giode.Element) []giode.Element {
	out := els[:0:0]
	for _, el := range els {
		if el != nil {
			out = append(out, el)
		}
	}
	return out
}
