// Package env carries what every screen needs: the store and a way to
// open other screens without importing them.
package env

import (
	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/store"
)

// Routes opens screens. The app implements it.
type Routes interface {
	// OpenChip shows a chip's specifications.
	OpenChip(id string)
	// OpenMaker shows a manufacturer's chips of a category ("" for all),
	// by segment and generation.
	OpenMaker(id, category string)
	// OpenDomain lists the generations of a manufacturer's chips of a
	// category and segment, such as Intel's desktop CPUs.
	OpenDomain(maker, category, segment string)
	// OpenGeneration lists the chips of one of those generations.
	OpenGeneration(maker, category, segment, name string)
	// OpenCategory lists the companies making a category of chips.
	OpenCategory(id string)
	// Explore switches to the Explore tab with q applied.
	Explore(q store.Query)
	// OpenLicenses shows the open-source licenses.
	OpenLicenses()
	// Makers switches to the Manufacturers tab.
	Makers()
	// Back leaves the current screen.
	Back()
}

// Env is shared by every screen.
type Env struct {
	Store  *store.Store
	Routes Routes
	Toasts *giode.ToasterElement
}
