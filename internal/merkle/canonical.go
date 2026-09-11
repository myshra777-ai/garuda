// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Package merkle implements the Garuda cryptographic integrity layer.
//
// This file (canonical.go) defines the canonical byte encoding used as
// input to every hash in the Merkle layer. It is the single source of
// truth for "how do we serialize a structured value into bytes that
// two independent implementations will agree on."
//
// Design rule: never trust a language-native serializer. Go's
// encoding/json, Python's json.dumps, and JavaScript's JSON.stringify
// all produce different bytes for the same logical content. Map key
// iteration order, field ordering, whitespace, and escape behavior
// are implementation details, not specifications. A hash computed over
// such output is not stable across languages, versions, or library
// upgrades.
//
// The canonical encoding defined here is:
//
//   [version:u8] [field_count:u16_be]
//   For each field in order:
//     [field_id:u8] [field_len:u32_be] [field_bytes]
//
// Properties:
//   - Fixed byte order (big-endian lengths).
//   - Every variable-length field is length-prefixed, so concatenation
//     is unambiguous: ["ab", "c"] and ["a", "bc"] produce different bytes.
//   - Field IDs make reordering detectable: a valid encoding must have
//     fields in strictly increasing ID order.
//   - Version byte allows future format changes without breaking old
//     verification.
//
// See docs/adr/0002-merkle-integrity-layer.md for the design rationale.

package merkle

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// CanonicalVersionV1 is the version byte for the current canonical
// encoding. Any change to the encoding rules requires a new version byte
// and a new set of golden test vectors.
const CanonicalVersionV1 byte = 0x01

// Field IDs used in decision hashing. The values are stable across
// versions of Garuda; changing them invalidates every historical hash.
const (
	fieldDecisionID       byte = 0x01
	fieldDecisionTitle    byte = 0x02
	fieldDecisionStatus   byte = 0x03
	fieldDecisionScopeDo  byte = 0x04
	fieldDecisionScopeSys byte = 0x05
	fieldDecisionOwner    byte = 0x06
	fieldDecisionEvidence byte = 0x07
)

// Field IDs used in epoch root hashing.
const (
	fieldEpochStaticRoot  byte = 0x10
	fieldEpochRuntimeRoot byte = 0x11
	fieldEpochParentRoot  byte = 0x12
	fieldEpochBlockHeight byte = 0x13
)

// ErrCanonicalFieldOrder indicates fields were written out of ID order.
// This is a programming error, not a runtime condition — the encoder
// detects it in tests, not in production paths.
var ErrCanonicalFieldOrder = errors.New("canonical: fields must be written in strictly increasing ID order")

// ErrCanonicalLengthOverflow indicates a single field exceeded 4GiB.
// This should never occur; it is a safety check against a corrupted caller.
var ErrCanonicalLengthOverflow = errors.New("canonical: field length exceeds uint32 maximum")

// Encoder produces a canonical byte sequence for hashing.
//
// Usage:
//
//	enc := NewEncoder()
//	enc.UUID(fieldID, id)
//	enc.String(fieldTitle, title)
//	bytes := enc.Finish()
//
// The encoder enforces that fields are written in strictly increasing
// ID order. Writing fields out of order causes Finish() to return an
// error — this protects against accidental reordering during a future
// refactor.
type Encoder struct {
	buf      []byte
	lastID   byte
	fieldCnt uint16
	finished bool
	err      error
}

// NewEncoder creates a fresh canonical encoder for the current version.
func NewEncoder() *Encoder {
	// Reserve space for version byte and field count; they are written
	// in Finish() once we know how many fields were emitted.
	buf := make([]byte, 0, 256)
	buf = append(buf, CanonicalVersionV1)
	buf = append(buf, 0, 0) // placeholder for field_count:u16
	return &Encoder{
		buf:      buf,
		lastID:   0,
		fieldCnt: 0,
	}
}

// writeField appends a length-prefixed field with the given ID.
func (e *Encoder) writeField(id byte, data []byte) {
	if e.err != nil {
		return
	}
	if id <= e.lastID {
		e.err = fmt.Errorf("%w: id 0x%02x after 0x%02x", ErrCanonicalFieldOrder, id, e.lastID)
		return
	}
	if len(data) > 0xFFFFFFFF {
		e.err = fmt.Errorf("%w: field 0x%02x has %d bytes", ErrCanonicalLengthOverflow, id, len(data))
		return
	}
	e.buf = append(e.buf, id)
	var lenBuf [4]byte
	binary.BigEndian.PutUint32(lenBuf[:], uint32(len(data)))
	e.buf = append(e.buf, lenBuf[:]...)
	e.buf = append(e.buf, data...)
	e.lastID = id
	e.fieldCnt++
}

// UUID writes a 16-byte UUID as a fixed-length field.
// The canonical representation is the raw 16 bytes, not the hyphenated string.
func (e *Encoder) UUID(id byte, u uuid.UUID) {
	e.writeField(id, u[:])
}

// Bytes writes a length-prefixed byte slice.
func (e *Encoder) Bytes(id byte, b []byte) {
	e.writeField(id, b)
}

// String writes a length-prefixed UTF-8 string.
// The bytes are the raw UTF-8 encoding; no normalization is applied.
// Callers who need Unicode normalization must normalize before calling.
func (e *Encoder) String(id byte, s string) {
	e.writeField(id, []byte(s))
}

// StringSlice writes a length-prefixed sequence of strings.
// The slice is sorted lexicographically before encoding, so the
// canonical form is independent of input order.
//
// Encoding:
//
//	[element_count:u32_be]
//	For each element: [element_len:u32_be][element_bytes]
func (e *Encoder) StringSlice(id byte, ss []string) {
	if e.err != nil {
		return
	}

	// Copy and sort to avoid mutating caller's slice.
	sorted := make([]string, len(ss))
	copy(sorted, ss)
	sortStringsInPlace(sorted)

	// Build the nested encoding in a temp buffer, then write as one field.
	var nested []byte
	var cntBuf [4]byte
	binary.BigEndian.PutUint32(cntBuf[:], uint32(len(sorted)))
	nested = append(nested, cntBuf[:]...)
	for _, s := range sorted {
		var lenBuf [4]byte
		binary.BigEndian.PutUint32(lenBuf[:], uint32(len(s)))
		nested = append(nested, lenBuf[:]...)
		nested = append(nested, []byte(s)...)
	}

	e.writeField(id, nested)
}

// Uint64 writes a fixed-length 8-byte big-endian field.
func (e *Encoder) Uint64(id byte, v uint64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], v)
	e.writeField(id, buf[:])
}

// Finish returns the canonical byte sequence.
// It returns an error if any field was written out of order or exceeded
// the length limit. The encoder is not reusable after Finish.
func (e *Encoder) Finish() ([]byte, error) {
	if e.finished {
		return nil, errors.New("canonical: Finish called twice")
	}
	e.finished = true
	if e.err != nil {
		return nil, e.err
	}
	// Patch the field_count placeholder now that we know the total.
	binary.BigEndian.PutUint16(e.buf[1:3], e.fieldCnt)
	return e.buf, nil
}

// sortStringsInPlace sorts a []string lexicographically (byte order).
// Defined here rather than importing "sort" to keep this file's
// dependencies minimal and auditable.
func sortStringsInPlace(s []string) {
	// Insertion sort — fine for the small slices we deal with here
	// (evidence IDs, typically < 100). Not for large data.
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j-1] > s[j]; j-- {
			s[j-1], s[j] = s[j], s[j-1]
		}
	}
}
