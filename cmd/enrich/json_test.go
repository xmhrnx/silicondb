package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestRoundTrip guards the file layout: decoding and re-encoding a shard
// must reproduce it exactly, since the tool rewrites these files in place.
func TestRoundTrip(t *testing.T) {
	files, err := filepath.Glob("../../data/manufacturers/*/*.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Skip("no data files")
	}
	for _, f := range files {
		want, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		n, err := decode(want)
		if err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		if got := encode(n); !bytes.Equal(got, want) {
			t.Errorf("%s: re-encoding changed the file (%d vs %d bytes)", f, len(got), len(want))
		}
	}
}

func TestScalarsKeepText(t *testing.T) {
	in := []byte("{\r\n  \"a\": 1.500,\r\n  \"b\": \"x\u00e9 \\\"q\\\"\",\r\n  \"c\": []\r\n}\r\n")
	n, err := decode(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := encode(n); !bytes.Equal(got, in) {
		t.Errorf("got %q, want %q", got, in)
	}
}
