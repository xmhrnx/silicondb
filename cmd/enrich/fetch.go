package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// fetcher downloads pages for the enrich tool, keeping each one under the
// cache directory so repeated runs and interrupted runs don't re-download
// everything. Only sources whose robots.txt permits it are used: Wikipedia,
// NanoReview and CPU-Monkey.
//
// Several workers share it, so the request pacing and the counters are
// guarded by a mutex: when a site asks for a pause every worker waits, which
// is what a single polite client would do.
type fetcher struct {
	dir     string
	client  *http.Client
	offline bool

	mu      sync.Mutex
	notBe   time.Time // no request before this instant
	fetched int
	cached  int

	delay time.Duration
}

const userAgent = "SiliconDB-data/1.0 (spec enrichment; +https://github.com/xmhrnx/silicondb)"

func newFetcher(dir string, offline bool) *fetcher {
	return &fetcher{
		dir:     dir,
		offline: offline,
		client:  &http.Client{Timeout: 45 * time.Second},
		delay:   250 * time.Millisecond,
	}
}

func (f *fetcher) count(fetched, cached int) {
	f.mu.Lock()
	f.fetched += fetched
	f.cached += cached
	f.mu.Unlock()
}

// wait blocks until this client may issue the next request and then claims
// the slot, spacing requests at least delay apart across all workers.
func (f *fetcher) wait() {
	f.mu.Lock()
	now := time.Now()
	at := now
	if f.notBe.After(at) {
		at = f.notBe
	}
	// Reserve the slot before unlocking so concurrent callers queue up.
	next := at.Add(f.delay)
	f.notBe = next
	f.mu.Unlock()
	if d := time.Until(at); d > 0 {
		time.Sleep(d)
	}
}

// pause stops every worker for d, used when a site reports rate limiting.
func (f *fetcher) pause(d time.Duration) {
	f.mu.Lock()
	if at := time.Now().Add(d); at.After(f.notBe) {
		f.notBe = at
	}
	f.mu.Unlock()
}

// get returns the page at url, from the cache when present.
func (f *fetcher) get(url string) ([]byte, error) {
	path := filepath.Join(f.dir, cacheName(url))
	if b, err := os.ReadFile(path); err == nil {
		f.count(0, 1)
		return b, nil
	}
	if f.offline {
		return nil, fmt.Errorf("%s: not cached", url)
	}
	for attempt := 0; ; attempt++ {
		f.wait()
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")
		resp, err := f.client.Do(req)
		if err != nil {
			if attempt < 2 {
				f.pause(2 * time.Second)
				continue
			}
			return nil, err
		}
		switch {
		case resp.StatusCode == http.StatusTooManyRequests && attempt < 3:
			io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
			resp.Body.Close()
			f.pause(time.Duration(5*(attempt+1)) * time.Second)
			continue
		case resp.StatusCode >= 500 && attempt < 2:
			io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
			resp.Body.Close()
			f.pause(2 * time.Second)
			continue
		case resp.StatusCode != http.StatusOK:
			io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
			resp.Body.Close()
			return nil, statusError{resp.StatusCode, url}
		}
		return f.save(path, resp)
	}
}

// save reads a successful response into the cache.
func (f *fetcher) save(path string, resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(f.dir, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return nil, err
	}
	f.count(1, 0)
	return b, nil
}

type statusError struct {
	code int
	url  string
}

func (e statusError) Error() string { return fmt.Sprintf("%s: HTTP %d", e.url, e.code) }

// cacheName turns a URL into a safe file name.
func cacheName(url string) string {
	var sb strings.Builder
	for _, r := range url {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			sb.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			sb.WriteRune(r + ('a' - 'A'))
		default:
			sb.WriteByte('_')
		}
	}
	name := strings.Trim(sb.String(), "_")
	if len(name) > 150 {
		name = name[len(name)-150:]
	}
	return name + ".html"
}

// skipMiss reports whether err means "this source has no page for the chip"
// rather than a real failure, so callers can try another guess quietly.
func skipMiss(err error) bool {
	if _, ok := err.(statusError); ok {
		return true
	}
	return strings.Contains(err.Error(), "not cached")
}
