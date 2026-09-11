// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestCanonical_Determinism asserts that encoding the same input twice
// produces byte-identical output.
func TestCanonical_Determinism(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	encodeOnce := func() []byte {
		enc := NewEncoder()
		enc.UUID(fieldDecisionID, id)
		enc.String(fieldDecisionTitle, "Test decision")
		enc.String(fieldDecisionStatus, "active")
		enc.String(fieldDecisionScopeDo, "security")
		enc.String(fieldDecisionScopeSys, "auth")
		enc.String(fieldDecisionOwner, "rohit")
		enc.StringSlice(fieldDecisionEvidence, []string{"e1", "e2"})
		out, err := enc.Finish()
		if err != nil {
			t.Fatalf("Finish: %v", err)
		}
		return out
	}

	a := encodeOnce()
	b := encodeOnce()

	if !bytes.Equal(a, b) {
		t.Errorf("canonical encoding is not deterministic:\n  a=%x\n  b=%x", a, b)
	}
}

// TestCanonical_FieldOrderEnforced asserts that writing fields out of
// strictly increasing ID order is rejected.
func TestCanonical_FieldOrderEnforced(t *testing.T) {
	enc := NewEncoder()
	enc.String(fieldDecisionTitle, "title") // 0x02
	enc.String(fieldDecisionID, "id")       // 0x01 — out of order

	_, err := enc.Finish()
	if err == nil {
		t.Fatal("expected error for out-of-order fields, got nil")
	}
	if !errors.Is(err, ErrCanonicalFieldOrder) {
		t.Errorf("expected ErrCanonicalFieldOrder, got %v", err)
	}
}

// TestCanonical_LengthPrefixingDisambiguates asserts that the encoding
// distinguishes ["ab", "c"] from ["a", "bc"].
func TestCanonical_LengthPrefixingDisambiguates(t *testing.T) {
	encodeTwoStrings := func(a, b string) []byte {
		enc := NewEncoder()
		enc.String(fieldDecisionTitle, a)
		enc.String(fieldDecisionStatus, b)
		out, _ := enc.Finish()
		return out
	}

	left := encodeTwoStrings("ab", "c")
	right := encodeTwoStrings("a", "bc")

	if bytes.Equal(left, right) {
		t.Errorf("length-prefixing failed: [\"ab\",\"c\"] and [\"a\",\"bc\"] produced identical bytes:\n  %x", left)
	}
}

// TestCanonical_StringSliceOrderIndependent asserts that the canonical
// form of a string slice does not depend on input order.
func TestCanonical_StringSliceOrderIndependent(t *testing.T) {
	encodeSlice := func(ss []string) []byte {
		enc := NewEncoder()
		enc.StringSlice(fieldDecisionEvidence, ss)
		out, _ := enc.Finish()
		return out
	}

	a := encodeSlice([]string{"alpha", "beta", "gamma"})
	b := encodeSlice([]string{"gamma", "alpha", "beta"})
	c := encodeSlice([]string{"beta", "gamma", "alpha"})

	if !bytes.Equal(a, b) || !bytes.Equal(b, c) {
		t.Errorf("string slice encoding depends on input order:\n  a=%x\n  b=%x\n  c=%x", a, b, c)
	}
}

// TestCanonical_VersionByte asserts the first byte is the version constant.
func TestCanonical_VersionByte(t *testing.T) {
	enc := NewEncoder()
	enc.String(fieldDecisionTitle, "x")
	out, _ := enc.Finish()

	if len(out) == 0 {
		t.Fatal("empty output")
	}
	if out[0] != CanonicalVersionV1 {
		t.Errorf("version byte = 0x%02x, want 0x%02x", out[0], CanonicalVersionV1)
	}
}

// TestCanonical_GoldenVectors prints canonical bytes and hashes for a
// fixed set of inputs. Run with -v to see the output. These values will
// be frozen into testdata/vectors_v1.json in a follow-up commit.
//
// If the printed values differ from a previously frozen set, the
// canonical encoding has changed and the version byte must be bumped.
func TestCanonical_GoldenVectors(t *testing.T) {
	type vector struct {
		name  string
		build func() ([]byte, error)
	}

	vectors := []vector{
		{
			name: "empty",
			build: func() ([]byte, error) {
				enc := NewEncoder()
				return enc.Finish()
			},
		},
		{
			name: "single_uuid",
			build: func() ([]byte, error) {
				enc := NewEncoder()
				enc.UUID(fieldDecisionID, uuid.MustParse("00000000-0000-0000-0000-000000000000"))
				return enc.Finish()
			},
		},
		{
			name: "single_string",
			build: func() ([]byte, error) {
				enc := NewEncoder()
				enc.String(fieldDecisionTitle, "hello")
				return enc.Finish()
			},
		},
		{
			name: "unicode_string",
			build: func() ([]byte, error) {
				enc := NewEncoder()
				enc.String(fieldDecisionTitle, "日本語")
				return enc.Finish()
			},
		},
		{
			name: "empty_string_slice",
			build: func() ([]byte, error) {
				enc := NewEncoder()
				enc.StringSlice(fieldDecisionEvidence, nil)
				return enc.Finish()
			},
		},
		{
			name: "multi_string_slice",
			build: func() ([]byte, error) {
				enc := NewEncoder()
				enc.StringSlice(fieldDecisionEvidence, []string{"zeta", "alpha", "mu"})
				return enc.Finish()
			},
		},
		{
			name: "full_decision",
			build: func() ([]byte, error) {
				enc := NewEncoder()
				enc.UUID(fieldDecisionID, uuid.MustParse("11111111-2222-3333-4444-555555555555"))
				enc.String(fieldDecisionTitle, "Test decision")
				enc.String(fieldDecisionStatus, "active")
				enc.String(fieldDecisionScopeDo, "security")
				enc.String(fieldDecisionScopeSys, "auth")
				enc.String(fieldDecisionOwner, "rohit")
				enc.StringSlice(fieldDecisionEvidence, []string{"e1", "e2", "e3"})
				return enc.Finish()
			},
		},
	}

	for _, v := range vectors {
		bytes, err := v.build()
		if err != nil {
			t.Errorf("%s: %v", v.name, err)
			continue
		}
		t.Logf("%s:\n  bytes = %s\n  hex   = %s",
			v.name, string(bytes), hex.EncodeToString(bytes))
	}
}
