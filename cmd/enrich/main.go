// Command enrich fills in missing specifications from public sources.
//
//	go run ./cmd/enrich                 # fill every shard
//	go run ./cmd/enrich -dry            # report what would be added
//	go run ./cmd/enrich -only intel     # one manufacturer
//	go run ./cmd/enrich -sources nanoreview
//
// It only adds specs a chip lacks and removes them from its "na" list, so
// hand-written values always win. Some values need no network at all:
// memory_speed_mts is derived from the chip's own memory_type. Downloads are
// cached under .cache/enrich; pass -offline to work from the cache alone. Run
// cmd/indexgen afterwards to refresh data/index.json.
package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

func main() {
	dir := flag.String("dir", "data", "database directory")
	cacheDir := flag.String("cache", ".cache/enrich", "download cache")
	dry := flag.Bool("dry", false, "report changes without writing files")
	only := flag.String("only", "", "only this manufacturer id")
	sources := flag.String("sources", "nanoreview,wikipedia", "comma-separated sources to use (cpumonkey is opt-in)")
	category := flag.String("category", "", "only shards of this category")
	since := flag.Int("since", 2005, "skip chips released before this year")
	limit := flag.Int("limit", 0, "stop after this many chips (0 for no limit)")
	skip := flag.Int("skip", 0, "skip this many chips first (for trying one era)")
	idPrefix := flag.String("id", "", "only chips whose id starts with this")
	offline := flag.Bool("offline", false, "use the download cache only")
	workers := flag.Int("workers", 4, "pages fetched in parallel")
	verbose := flag.Bool("v", false, "print every added spec")
	flag.Parse()

	enabled := map[string]bool{}
	for _, s := range strings.Split(*sources, ",") {
		enabled[strings.TrimSpace(s)] = true
	}

	f := newFetcher(*cacheDir, *offline)
	// Wikipedia is fetched once per run, so it is always worth doing.
	tool := &enricher{
		f:       f,
		nano:    &nanoReview{f: f},
		monkey:  &cpuMonkey{f: f},
		dry:     *dry,
		only:    *only,
		cat:     *category,
		since:   *since,
		limit:   *limit,
		skip:    *skip,
		idPref:  *idPrefix,
		verbose: *verbose,
		workers: *workers,
		sources: enabled,
	}
	if err := tool.run(*dir); err != nil {
		fmt.Fprintln(os.Stderr, "enrich:", err)
		os.Exit(1)
	}
	tool.summary()
}

type enricher struct {
	f       *fetcher
	nano    *nanoReview
	monkey  *cpuMonkey
	dry     bool
	only    string
	cat     string
	since   int
	limit   int
	skip    int
	idPref  string
	verbose bool
	workers int
	sources map[string]bool

	sch  *schema
	once sync.Once
	wiki *wikiIndex

	chips, changed, srcHits, srcMiss int
	adds                             map[string]int
	files                            []string
	errs                             []string
	mu                               sync.Mutex
}

// wikiData loads the Wikipedia GPU lists once, the first time a graphics
// card needs one.
func (e *enricher) wikiData() *wikiIndex {
	e.once.Do(func() {
		idx := &wikiIndex{}
		idx.load(e.f)
		e.wiki = idx
	})
	return e.wiki
}

// target is one chip to enrich.
type target struct {
	file     string
	maker    string
	category string
	doc      *node
	specs    *node
	na       *node
	id, name string
}

func (e *enricher) run(dir string) error {
	schemaBytes, err := os.ReadFile(filepath.Join(dir, "schema.json"))
	if err != nil {
		return err
	}
	schemaDoc, err := decode(schemaBytes)
	if err != nil {
		return fmt.Errorf("schema.json: %w", err)
	}
	e.sch = newSchema(schemaDoc)
	e.adds = map[string]int{}

	makers, err := os.ReadDir(filepath.Join(dir, "manufacturers"))
	if err != nil {
		return err
	}
	var files []string
	for _, m := range makers {
		if !m.IsDir() || (e.only != "" && m.Name() != e.only) {
			continue
		}
		dirs, err := os.ReadDir(filepath.Join(dir, "manufacturers", m.Name()))
		if err != nil {
			return err
		}
		for _, f := range dirs {
			name := f.Name()
			if f.IsDir() || name == "manufacturer.json" || !strings.HasSuffix(name, ".json") {
				continue
			}
			if e.cat != "" && strings.TrimSuffix(name, ".json") != e.cat {
				continue
			}
			files = append(files, filepath.Join(dir, "manufacturers", m.Name(), name))
		}
	}
	sort.Strings(files)

	for _, file := range files {
		if err := e.shardFile(file); err != nil {
			return err
		}
	}
	sort.Strings(e.files)
	return nil
}

func (e *enricher) shardFile(file string) error {
	raw, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	doc, err := decode(raw)
	if err != nil {
		return fmt.Errorf("%s: %w", file, err)
	}
	chips := doc.get("chips")
	if chips == nil || chips.kind != 'a' {
		return fmt.Errorf("%s: no chips", file)
	}
	maker, _ := doc.get("manufacturer").str()
	category, _ := doc.get("category").str()

	// Pick the chips worth fetching first, so the worker pool only does the
	// requests keepable results depend on.
	var targets []*target
	for _, chip := range chips.arr {
		id, _ := chip.get("id").str()
		name, _ := chip.get("name").str()
		if e.idPref != "" && !strings.HasPrefix(id, e.idPref) {
			continue
		}
		if y := chipYear(chip); e.since > 0 && y > 0 && y < e.since {
			continue
		}
		if e.skip > 0 {
			e.skip--
			continue
		}
		if e.limit > 0 && e.chips >= e.limit {
			break
		}
		e.chips++
		targets = append(targets, &target{
			file: file, maker: maker, category: category, doc: doc,
			specs: chip.get("specs"), na: chip.get("na"), id: id, name: name,
		})
	}
	if len(targets) == 0 {
		return nil
	}
	results := e.fetchAll(targets)
	dirty := false
	for _, t := range targets {
		changed := len(e.deriveAll(t)) > 0
		if len(e.apply(t, results[t.id])) > 0 {
			changed = true
		}
		if changed {
			// Reordering once, after both passes, puts derived and fetched
			// keys in schema order in a single rewrite.
			e.sch.sortSpecs(t.specs)
			dirty = true
		}
	}
	if !dirty {
		return nil
	}
	rel := path.Join("manufacturers", maker, category+".json")
	if e.dry {
		e.files = append(e.files, rel+" (would change)")
		return nil
	}
	if err := os.WriteFile(file, encode(doc), 0o644); err != nil {
		return err
	}
	e.files = append(e.files, rel)
	return nil
}

// fetchAll queries every source for the targets, in parallel, and returns
// the values found per chip id.
func (e *enricher) fetchAll(targets []*target) map[string]map[string]any {
	out := make(map[string]map[string]any, len(targets))
	var mu sync.Mutex
	jobs := make(chan *target)
	var wg sync.WaitGroup
	n := e.workers
	if n < 1 {
		n = 1
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range jobs {
				vals := e.fetchOne(t)
				if len(vals) == 0 {
					continue
				}
				mu.Lock()
				out[t.id] = vals
				mu.Unlock()
			}
		}()
	}
	for _, t := range targets {
		jobs <- t
	}
	close(jobs)
	wg.Wait()
	return out
}

// fetchOne gathers what the enabled sources know about a chip.
func (e *enricher) fetchOne(t *target) map[string]any {
	vals := map[string]any{}
	if e.sources["nanoreview"] {
		got, err := e.nano.specs(t.maker, t.category, t.id, t.name)
		e.note(t.id, "nanoreview", err)
		for k, v := range got {
			if allowed := allowedKeys(t.category, "nanoreview"); allowed[k] {
				vals[k] = v
			}
		}
	}
	if e.sources["wikipedia"] && t.category == "gpu" {
		if got := e.wikiData().lookup(t.name, t.maker); got != nil {
			for k, v := range got {
				if allowedWiki()[k] {
					vals[k] = v
				}
			}
		}
	}
	if e.sources["cpumonkey"] {
		got, err := e.monkey.benchmarks(t.maker, t.category, t.id, t.name)
		e.note(t.id, "cpu-monkey", err)
		for k, v := range got {
			vals[k] = v
		}
	}
	return vals
}

func (e *enricher) note(id, source string, err error) {
	if err == nil {
		return
	}
	e.mu.Lock()
	e.errs = append(e.errs, fmt.Sprintf("%s: %s: %v", id, source, err))
	e.mu.Unlock()
}

// apply writes the fetched values into the chip and reports whether the
// shard changed.
func (e *enricher) apply(t *target, vals map[string]any) []string {
	if len(vals) == 0 || t.specs == nil {
		e.srcMiss++
		return nil
	}
	e.srcHits++
	added := mergeSpecs(t.specs, t.na, vals, nil)
	if len(added) == 0 {
		e.srcMiss++
		return nil
	}
	e.changed++
	for _, k := range added {
		e.adds[k]++
	}
	if e.verbose {
		fmt.Printf("%s: %s\n", t.id, strings.Join(added, ", "))
	}
	// The na list keeps its original order, so removing a key leaves the
	// rest of the file's diff untouched.
	e.sch.sortSpecs(t.specs)
	return added
}

// chipYear reads the release year from a chip's "released" field.
func chipYear(chip *node) int {
	s, _ := chip.get("released").str()
	if len(s) < 4 {
		return 0
	}
	y := 0
	for i := 0; i < 4; i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0
		}
		y = y*10 + int(s[i]-'0')
	}
	return y
}

func (e *enricher) summary() {
	for _, err := range e.errs {
		fmt.Fprintln(os.Stderr, "warning:", err)
	}
	fmt.Printf("%d chips, %d changed, %d pages matched, %d unchanged\n", e.chips, e.changed, e.srcHits, e.srcMiss)
	e.f.mu.Lock()
	fetched, cached := e.f.fetched, e.f.cached
	e.f.mu.Unlock()
	fmt.Printf("downloads: %d fetched, %d cached\n", fetched, cached)
	if len(e.adds) > 0 {
		keys := make([]string, 0, len(e.adds))
		for k := range e.adds {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return e.adds[keys[i]] > e.adds[keys[j]] })
		fmt.Println("added specs:")
		for _, k := range keys {
			fmt.Printf("  %-26s %d\n", k, e.adds[k])
		}
	}
	if len(e.files) > 0 {
		fmt.Printf("%d shards updated:\n", len(e.files))
		for _, f := range e.files {
			fmt.Println("  " + f)
		}
	}
}
