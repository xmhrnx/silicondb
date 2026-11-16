// Package nav moves between screens: a tab bar of root screens with a
// stack of pushed screens above the current one.
package nav

import "github.com/xmhrnx/giode"

// Page is a screen. It is created when shown and keeps its own state
// (scroll offsets, fields) until it is popped.
type Page interface {
	View() giode.Element
}
