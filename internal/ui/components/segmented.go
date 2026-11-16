package components

import (
	"github.com/xmhrnx/giode"
	"github.com/xmhrnx/giode/components/segmented"
	"github.com/xmhrnx/giode/i18n"

	"github.com/mehran9675/silicondb/internal/ui/tr"
)

// Segmented is a segmented control whose options are message keys. It
// rebuilds its control, keeping the selection, when the language
// changes, and themes it every frame.
type Segmented struct {
	keys     []string
	selected int
	onChange func(i int)
	ctrl     *segmented.Control
	version  uint64
}

// NewSegmented returns a control choosing among the options named by
// keys; onChange runs with the index picked.
func NewSegmented(onChange func(i int), keys ...string) *Segmented {
	return &Segmented{keys: keys, onChange: onChange}
}

// Selected returns the selected index.
func (s *Segmented) Selected() int { return s.selected }

// SetSelected changes the selection without calling onChange.
func (s *Segmented) SetSelected(i int) {
	s.selected = i
	if s.ctrl != nil {
		s.ctrl.SetSelected(i)
	}
}

// View lays the control out.
func (s *Segmented) View() giode.Element {
	if v := i18n.Version(); s.ctrl == nil || v != s.version {
		labels := make([]string, len(s.keys))
		for i, k := range s.keys {
			labels[i] = tr.T(k)
		}
		s.ctrl = giode.Input.Segmented(labels...).OnChange(func(i int) {
			s.selected = i
			if s.onChange != nil {
				s.onChange(i)
			}
		})
		s.ctrl.SetSelected(s.selected)
		s.version = v
	}
	s.ctrl.Styles(SegmentedStyles())
	return s.ctrl
}
