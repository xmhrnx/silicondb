package data

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/mehran9675/silicondb/internal/model"
)

// remote is an in-memory database server that counts requests per path.
type remote struct {
	mu    sync.Mutex
	files map[string][]byte
	hits  map[string]int
}

func (r *remote) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	path := req.URL.Path[1:]
	r.hits[path]++
	b, ok := r.files[path]
	if !ok {
		http.NotFound(w, req)
		return
	}
	w.Write(b)
}

func (r *remote) set(path string, v any) {
	b, _ := json.Marshal(v)
	r.mu.Lock()
	r.files[path] = b
	r.mu.Unlock()
}

// publish rewrites the index from the current files, as indexgen does.
func (r *remote) publish() {
	ix := model.Index{Format: 1, Schema: model.FileRef{Path: "schema.json", Rev: Rev(r.files["schema.json"])}}
	ix.Manufacturers = []model.IndexEntry{{
		Manufacturer: model.Manufacturer{ID: "intel", Name: "Intel"},
		Shards: []model.ShardRef{
			{Path: "manufacturers/intel/cpu.json", Category: "cpu", Count: 1, Rev: Rev(r.files["manufacturers/intel/cpu.json"])},
			{Path: "manufacturers/intel/mcu.json", Category: "mcu", Count: 1, Rev: Rev(r.files["manufacturers/intel/mcu.json"])},
		},
	}}
	r.set(IndexPath, ix)
}

func shard(cat, id, name string) model.Shard {
	return model.Shard{Manufacturer: "intel", Category: cat, Chips: []*model.Chip{{ID: id, Name: name, Family: "x", Released: "1972"}}}
}

func (r *remote) downloads() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for p, h := range r.hits {
		if p != IndexPath {
			n += h
		}
	}
	r.hits = map[string]int{}
	return n
}

func TestSyncDownloadsOnlyChangedFiles(t *testing.T) {
	r := &remote{files: map[string][]byte{}, hits: map[string]int{}}
	r.set("schema.json", model.Schema{Format: 1, Categories: []model.Category{{ID: "cpu", Name: "CPU"}}})
	r.set("manufacturers/intel/cpu.json", shard("cpu", "intel-8008", "8008"))
	r.set("manufacturers/intel/mcu.json", shard("mcu", "intel-8048", "8048"))
	r.publish()
	srv := httptest.NewServer(r)
	defer srv.Close()

	dir := t.TempDir()
	repo := NewRepository(dir, srv.URL)
	if _, err := repo.LoadCached(); err != ErrNoCache {
		t.Fatalf("LoadCached on an empty cache: %v, want ErrNoCache", err)
	}
	cat, res, err := repo.Sync(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Files != 3 || r.downloads() != 3 || len(cat.Chips) != 2 {
		t.Fatalf("first sync: %d files, %d chips; want 3 files, 2 chips", res.Files, len(cat.Chips))
	}

	if _, res, err = repo.Sync(context.Background(), nil); err != nil || res.Files != 0 || r.downloads() != 0 {
		t.Fatalf("unchanged sync downloaded %d files (err %v), want 0", res.Files, err)
	}

	r.set("manufacturers/intel/cpu.json", shard("cpu", "intel-8008", "8008 (rev B)"))
	r.publish()
	cat, res, err = repo.Sync(context.Background(), nil)
	if err != nil || res.Files != 1 || r.downloads() != 1 {
		t.Fatalf("one changed shard: downloaded %d files (err %v), want 1", res.Files, err)
	}
	if got := cat.Chip("intel-8008").Name; got != "8008 (rev B)" {
		t.Fatalf("updated name = %q", got)
	}

	// A fresh repository on the same directory works offline.
	cat, err = NewRepository(dir, "http://127.0.0.1:1").LoadCached()
	if err != nil || len(cat.Chips) != 2 {
		t.Fatalf("offline load: %v, %d chips", err, len(cat.Chips))
	}
}

func TestRevIgnoresLineEndings(t *testing.T) {
	if Rev([]byte("{\n  \"a\": 1\n}\n")) != Rev([]byte("{\r\n  \"a\": 1\r\n}\r\n")) {
		t.Fatal("a CRLF copy of a file must have the LF copy's revision")
	}
}
