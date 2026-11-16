package data

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
)

// Rev returns the revision of a file's content: the first 16 hex digits of
// its SHA-256, with Windows line endings read as Unix ones. The index
// generator computes revisions the same way, so an index generated from a
// Windows checkout (CRLF) matches the files the server sends (LF), and the
// app doesn't download unchanged files again.
func Rev(content []byte) string {
	sum := sha256.Sum256(bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n")))
	return hex.EncodeToString(sum[:8])
}
