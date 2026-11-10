package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// The database files are written with a fixed layout: two-space indent,
// CRLF line endings and a trailing CRLF. They are also order-sensitive:
// spec keys follow schema.json order, benchmarks last, and keys the
// schema doesn't know keep their relative order. encoding/json sorts map
// keys alphabetically and escapes HTML, so it can't round-trip them. This
// file keeps objects as ordered field lists and re-encodes scalars from
// their original text, so untouched values come back byte for byte.

// node is a JSON value. Scalars keep their source text in raw; objects and
// arrays keep their children in order.
type node struct {
	kind byte // 'o' object, 'a' array, 'v' scalar
	obj  []field
	arr  []*node
	raw  string
}

type field struct {
	key string
	val *node
}

func newObject(pairs ...field) *node { return &node{kind: 'o', obj: pairs} }
func newArray(items ...*node) *node  { return &node{kind: 'a', arr: items} }
func newString(s string) *node       { return &node{kind: 'v', raw: quoteJSON(s)} }
func newNumber(f float64) *node      { return &node{kind: 'v', raw: formatNumber(f)} }
func newBool(b bool) *node           { return &node{kind: 'v', raw: boolText(b)} }
func newStrings(ss []string) *node {
	n := &node{kind: 'a'}
	for _, s := range ss {
		n.arr = append(n.arr, newString(s))
	}
	return n
}

func boolText(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func (n *node) get(key string) *node {
	if n == nil || n.kind != 'o' {
		return nil
	}
	for i := range n.obj {
		if n.obj[i].key == key {
			return n.obj[i].val
		}
	}
	return nil
}

// set replaces the value of key in place, or appends it at the end.
func (n *node) set(key string, val *node) {
	for i := range n.obj {
		if n.obj[i].key == key {
			n.obj[i].val = val
			return
		}
	}
	n.obj = append(n.obj, field{key, val})
}

// insertAt inserts a field at index i.
func (n *node) insertAt(i int, key string, val *node) {
	if i >= len(n.obj) {
		n.obj = append(n.obj, field{key, val})
		return
	}
	n.obj = append(n.obj, field{})
	copy(n.obj[i+1:], n.obj[i:])
	n.obj[i] = field{key, val}
}

// remove deletes key from an object, or the matching string from an array,
// and reports whether it was there.
func (n *node) remove(key string) bool {
	if n == nil {
		return false
	}
	switch n.kind {
	case 'a':
		for i, e := range n.arr {
			if s, ok := e.str(); ok && s == key {
				n.arr = append(n.arr[:i], n.arr[i+1:]...)
				return true
			}
		}
	case 'o':
		for i := range n.obj {
			if n.obj[i].key == key {
				n.obj = append(n.obj[:i], n.obj[i+1:]...)
				return true
			}
		}
	}
	return false
}

// keys returns the object's keys in order.
func (n *node) keys() []string {
	if n == nil || n.kind != 'o' {
		return nil
	}
	ks := make([]string, len(n.obj))
	for i := range n.obj {
		ks[i] = n.obj[i].key
	}
	return ks
}

func (n *node) str() (string, bool) {
	if n == nil || n.kind != 'v' {
		return "", false
	}
	var s string
	if json.Unmarshal([]byte(n.raw), &s) != nil {
		return "", false
	}
	return s, true
}

func (n *node) num() (float64, bool) {
	if n == nil || n.kind != 'v' {
		return 0, false
	}
	var f float64
	if json.Unmarshal([]byte(n.raw), &f) != nil {
		return 0, false
	}
	return f, true
}

// index returns the position of v in an array of strings, or -1.
func (n *node) index(s string) int {
	if n == nil || n.kind != 'a' {
		return -1
	}
	for i, e := range n.arr {
		if v, ok := e.str(); ok && v == s {
			return i
		}
	}
	return -1
}

// decode parses JSON into a node tree, keeping numbers and strings raw.
func decode(b []byte) (*node, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	n, err := decodeValue(d)
	if err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, errors.New("trailing data after JSON value")
	}
	return n, nil
}

func decodeValue(d *json.Decoder) (*node, error) {
	t, err := d.Token()
	if err != nil {
		return nil, err
	}
	switch v := t.(type) {
	case json.Delim:
		switch v {
		case '{':
			obj := &node{kind: 'o'}
			for d.More() {
				kt, err := d.Token()
				if err != nil {
					return nil, err
				}
				key, ok := kt.(string)
				if !ok {
					return nil, fmt.Errorf("object key is %T, not a string", kt)
				}
				val, err := decodeValue(d)
				if err != nil {
					return nil, err
				}
				obj.obj = append(obj.obj, field{key, val})
			}
			if _, err := d.Token(); err != nil { // closing }
				return nil, err
			}
			return obj, nil
		case '[':
			arr := &node{kind: 'a'}
			for d.More() {
				val, err := decodeValue(d)
				if err != nil {
					return nil, err
				}
				arr.arr = append(arr.arr, val)
			}
			if _, err := d.Token(); err != nil { // closing ]
				return nil, err
			}
			return arr, nil
		}
		return nil, fmt.Errorf("unexpected delimiter %q", v)
	case string:
		return newString(v), nil
	case json.Number:
		return &node{kind: 'v', raw: v.String()}, nil
	case bool:
		return &node{kind: 'v', raw: boolText(v)}, nil
	case nil:
		return &node{kind: 'v', raw: "null"}, nil
	}
	return nil, fmt.Errorf("unexpected token %T", t)
}

// encode renders n the way the database files are laid out.
func encode(n *node) []byte {
	var sb strings.Builder
	encodeValue(&sb, n, 0)
	sb.WriteString("\r\n")
	return []byte(sb.String())
}

func encodeValue(sb *strings.Builder, n *node, depth int) {
	switch n.kind {
	case 'v':
		sb.WriteString(n.raw)
	case 'a':
		if len(n.arr) == 0 {
			sb.WriteString("[]")
			return
		}
		sb.WriteString("[\r\n")
		for i, e := range n.arr {
			if i > 0 {
				sb.WriteString(",\r\n")
			}
			writeIndent(sb, depth+1)
			encodeValue(sb, e, depth+1)
		}
		sb.WriteString("\r\n")
		writeIndent(sb, depth)
		sb.WriteString("]")
	case 'o':
		if len(n.obj) == 0 {
			sb.WriteString("{}")
			return
		}
		sb.WriteString("{\r\n")
		for i, f := range n.obj {
			if i > 0 {
				sb.WriteString(",\r\n")
			}
			writeIndent(sb, depth+1)
			sb.WriteString(quoteJSON(f.key))
			sb.WriteString(": ")
			encodeValue(sb, f.val, depth+1)
		}
		sb.WriteString("\r\n")
		writeIndent(sb, depth)
		sb.WriteString("}")
	}
}

func writeIndent(sb *strings.Builder, depth int) {
	for i := 0; i < depth; i++ {
		sb.WriteString("  ")
	}
}

// quoteJSON encodes s as a JSON string without escaping non-ASCII text, the
// same way the generator that wrote these files did.
func quoteJSON(s string) string {
	var sb strings.Builder
	sb.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\b':
			sb.WriteString(`\b`)
		case '\f':
			sb.WriteString(`\f`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(&sb, `\u%04x`, r)
				continue
			}
			sb.WriteRune(r)
		}
	}
	sb.WriteByte('"')
	return sb.String()
}

// formatNumber writes a float the way the existing files do: whole numbers
// without a fraction, others as the shortest text that round-trips.
func formatNumber(f float64) string {
	if f == float64(int64(f)) && f < 1e15 && f > -1e15 {
		return fmt.Sprintf("%d", int64(f))
	}
	return strings.TrimSuffix(fmt.Sprintf("%v", f), ".0")
}
