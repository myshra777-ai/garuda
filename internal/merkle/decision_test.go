// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/google/uuid"
)

func TestCanonicalDecision_Deterministic(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	a := CanonicalDecision(id, "Title", "active", "sec", "auth", "owner", []string{"e1", "e2"})
	b := CanonicalDecision(id, "Title", "active", "sec", "auth", "owner", []string{"e1", "e2"})
	if !bytes.Equal(a, b) {
		t.Errorf("decision hash not deterministic:\n  a=%x\n  b=%x", a, b)
	}
}

func TestCanonicalDecision_EvidenceOrderIndependent(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	a := CanonicalDecision(id, "Title", "active", "sec", "auth", "owner", []string{"beta", "alpha", "gamma"})
	b := CanonicalDecision(id, "Title", "active", "sec", "auth", "owner", []string{"gamma", "alpha", "beta"})
	if !bytes.Equal(a, b) {
		t.Errorf("decision hash depends on evidence order:\n  a=%x\n  b=%x", a, b)
	}
}

func TestCanonicalDecision_FieldSensitivity(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	base := CanonicalDecision(id, "Title", "active", "sec", "auth", "owner", nil)

	cases := []struct {
		name string
		hash []byte
	}{
		{"title", CanonicalDecision(id, "TitlE", "active", "sec", "auth", "owner", nil)},
		{"status", CanonicalDecision(id, "Title", "inactive", "sec", "auth", "owner", nil)},
		{"scope_domain", CanonicalDecision(id, "Title", "active", "SEC", "auth", "owner", nil)},
		{"scope_system", CanonicalDecision(id, "Title", "active", "sec", "AUTH", "owner", nil)},
		{"owner", CanonicalDecision(id, "Title", "active", "sec", "auth", "OWNER", nil)},
		{"id", CanonicalDecision(uuid.New(), "Title", "active", "sec", "auth", "owner", nil)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if bytes.Equal(base, tc.hash) {
				t.Errorf("changing %s did not change the hash", tc.name)
			}
		})
	}
}

func TestCanonicalDecision_LengthPrefixingDisambiguates(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	a := CanonicalDecision(id, "ab", "c", "sec", "auth", "owner", nil)
	b := CanonicalDecision(id, "a", "bc", "sec", "auth", "owner", nil)
	if bytes.Equal(a, b) {
		t.Errorf("length-prefixing failed: [\"ab\",\"c\"] and [\"a\",\"bc\"] collided")
	}
}

func TestCanonicalDecision_Length(t *testing.T) {
	id := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	h := CanonicalDecision(id, "t", "s", "d", "sys", "own", nil)
	if len(h) != 32 {
		t.Errorf("decision hash length = %d, want 32", len(h))
	}
}

// TestCanonicalDecision_PrintVectors prints decision hashes for fixed
// inputs. Run with -v and freeze into testdata/decision_vectors_v1.json
// in a follow-up commit.
func TestCanonicalDecision_PrintVectors(t *testing.T) {
	type vec struct {
		name     string
		id       string
		title    string
		status   string
		scopeDo  string
		scopeSys string
		owner    string
		evidence []string
	}

	vectors := []vec{
		{
			name: "empty_decision",
			id:   "00000000-0000-0000-0000-000000000000",
		},
		{
			name:     "minimal_decision",
			id:       "11111111-2222-3333-4444-555555555555",
			title:    "t",
			status:   "s",
			scopeDo:  "d",
			scopeSys: "sys",
			owner:    "own",
		},
		{
			name:     "decision_with_evidence",
			id:       "11111111-2222-3333-4444-555555555555",
			title:    "Test decision",
			status:   "active",
			scopeDo:  "security",
			scopeSys: "auth",
			owner:    "rohit",
			evidence: []string{"e1", "e2", "e3"},
		},
		{
			name:     "unicode_decision",
			id:       "11111111-2222-3333-4444-555555555555",
			title:    "日本語",
			status:   "active",
			scopeDo:  "sec",
			scopeSys: "auth",
			owner:    "rohit",
		},
	}

	for _, v := range vectors {
		id := uuid.MustParse(v.id)
		hexHash := hex.EncodeToString(CanonicalDecision(
			id, v.title, v.status, v.scopeDo, v.scopeSys, v.owner, v.evidence,
		))
		t.Logf("%-25s %s", v.name, hexHash)
	}
}
