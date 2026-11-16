// Package store holds the app's state: the catalog, the sync status and
// the user's preferences.
package store

import (
	"context"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/xmhrnx/giode"

	"github.com/mehran9675/silicondb/internal/config"
	"github.com/mehran9675/silicondb/internal/data"
	"github.com/mehran9675/silicondb/internal/model"
)

// maxRecent is how many recently viewed chips are remembered.
const maxRecent = 12

// Store is the app's shared state. Catalog and Status may be read from
// the view at any time; they're updated from background goroutines.
type Store struct {
	Catalog *giode.State[*model.Catalog]
	Status  *giode.State[SyncStatus]

	repo      *data.Repository
	prefsPath string

	mu      sync.Mutex
	prefs   Prefs
	syncing bool

	// Preferences are written in the background so opening a chip, which
	// records it as recently viewed, doesn't wait on the disk. saveMu
	// orders the writes; written is the newest revision on disk, so an
	// older snapshot never overwrites a newer one.
	saveMu   sync.Mutex
	revision uint64
	written  uint64
}

// New returns a store keeping its files under dir. An empty dir keeps
// everything in memory.
func New(dir string) *Store {
	s := &Store{
		Catalog: giode.NewState[*model.Catalog](nil),
		Status:  giode.NewState(SyncStatus{}),
	}
	cacheDir := ""
	if dir != "" {
		cacheDir = filepath.Join(dir, "cache")
		s.prefsPath = filepath.Join(dir, "prefs.json")
	}
	s.prefs = loadPrefs(s.prefsPath)
	if s.prefs.LastSynced != 0 {
		s.Status.Set(SyncStatus{LastSynced: time.Unix(s.prefs.LastSynced, 0)})
	}
	s.repo = data.NewRepository(cacheDir, s.DataURL())
	return s
}

// Start shows the offline copy right away and then syncs in the
// background.
func (s *Store) Start() {
	go func() {
		if cat, err := s.repo.LoadCached(); err == nil {
			s.Catalog.Set(cat)
		}
		s.sync()
	}()
}

// Refresh syncs with the remote database unless a sync is running.
func (s *Store) Refresh() { go s.sync() }

func (s *Store) sync() {
	s.mu.Lock()
	if s.syncing {
		s.mu.Unlock()
		return
	}
	s.syncing = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.syncing = false
		s.mu.Unlock()
	}()

	last := s.Status.Get().LastSynced
	s.Status.Set(SyncStatus{Phase: Syncing, LastSynced: last})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cat, res, err := s.repo.Sync(ctx, func(done, total int) {
		s.Status.Update(func(st SyncStatus) SyncStatus {
			st.Done, st.Total = done, total
			return st
		})
	})
	if err != nil {
		s.Status.Set(SyncStatus{Phase: Failed, Err: err.Error(), LastSynced: last})
		return
	}
	now := time.Now()
	s.Catalog.Set(cat)
	s.Status.Set(SyncStatus{Phase: Synced, Files: res.Files, Bytes: res.Bytes, LastSynced: now})
	s.updatePrefs(func(p *Prefs) { p.LastSynced = now.Unix() })
}

// ClearCache deletes the offline copy and downloads everything again.
func (s *Store) ClearCache() {
	go func() {
		_ = s.repo.ClearCache()
		s.sync()
	}()
}

// DataURL returns the database source in use.
func (s *Store) DataURL() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.prefs.DataURL != "" {
		return s.prefs.DataURL
	}
	return config.DataURL
}

// SetDataURL changes the database source ("" restores the default) and
// downloads from it.
func (s *Store) SetDataURL(u string) {
	if u == config.DataURL {
		u = ""
	}
	s.updatePrefs(func(p *Prefs) { p.DataURL = u })
	s.repo.SetSource(s.DataURL())
	s.ClearCache()
}

// Dark reports whether dark mode is on.
func (s *Store) Dark() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prefs.Dark
}

// SetDark switches dark mode on or off.
func (s *Store) SetDark(on bool) { s.updatePrefs(func(p *Prefs) { p.Dark = on }) }

// Locale returns the chosen UI language, or "" to follow the system.
func (s *Store) Locale() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.prefs.Locale
}

// SetLocale chooses the UI language ("" follows the system).
func (s *Store) SetLocale(code string) { s.updatePrefs(func(p *Prefs) { p.Locale = code }) }

// Compare returns the ids of the chips in the comparison.
func (s *Store) Compare() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.prefs.Compare)
}

// InCompare reports whether id is in the comparison.
func (s *Store) InCompare(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Contains(s.prefs.Compare, id)
}

// ToggleCompare adds id to the comparison or removes it. Adding to a full
// comparison drops the oldest entry. It returns whether id is now in it.
func (s *Store) ToggleCompare(id string) bool {
	in := false
	s.updatePrefs(func(p *Prefs) {
		if i := slices.Index(p.Compare, id); i >= 0 {
			p.Compare = slices.Delete(p.Compare, i, i+1)
			return
		}
		p.Compare = append(p.Compare, id)
		if len(p.Compare) > config.MaxCompare {
			p.Compare = p.Compare[len(p.Compare)-config.MaxCompare:]
		}
		in = true
	})
	return in
}

// ClearCompare empties the comparison.
func (s *Store) ClearCompare() { s.updatePrefs(func(p *Prefs) { p.Compare = nil }) }

// Recent returns recently viewed chip ids, most recent first.
func (s *Store) Recent() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.prefs.Recent)
}

// Viewed records that the chip id was opened.
func (s *Store) Viewed(id string) {
	s.updatePrefs(func(p *Prefs) {
		if i := slices.Index(p.Recent, id); i >= 0 {
			p.Recent = slices.Delete(p.Recent, i, i+1)
		}
		p.Recent = slices.Insert(p.Recent, 0, id)
		if len(p.Recent) > maxRecent {
			p.Recent = p.Recent[:maxRecent]
		}
	})
}

// Save writes the preferences to disk and returns once they're written,
// for the moment the app goes to the background.
func (s *Store) Save() {
	s.mu.Lock()
	s.revision++
	p, rev := s.prefs.clone(), s.revision
	s.mu.Unlock()
	s.write(p, rev)
}

func (s *Store) updatePrefs(fn func(p *Prefs)) {
	s.mu.Lock()
	fn(&s.prefs)
	s.revision++
	p, rev := s.prefs.clone(), s.revision
	s.mu.Unlock()
	go s.write(p, rev)
}

// write saves p unless a newer revision was saved already.
func (s *Store) write(p Prefs, rev uint64) {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()
	if rev <= s.written {
		return
	}
	savePrefs(s.prefsPath, p)
	s.written = rev
}
