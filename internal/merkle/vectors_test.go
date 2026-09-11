// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package merkle

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// TestVectorsV1 asserts that the canonical encoder reproduces every
// frozen vector in testdata/vectors_v1.json byte-for-byte.
//
// A failure here means the canonical encoding has changed. That is a
// breaking change to the Merkle layer: every historical hash becomes
// unverifiable. The fix is NOT to update the vector file — it is to
// revert the change or to bump the version byte and open a new ADR.
func TestVectorsV1(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "vectors_v1.json"))
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var file struct {
		Vectors []struct {
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input,omitempty"`
			Hex   string          `json:"hex"`
		} `json:"vectors"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}

	for _, v := range file.Vectors {
		t.Run(v.Name, func(t *testing.T) {
			var got []byte
			var err error

			switch v.Name {
			case "empty":
				enc := NewEncoder()
				got, err = enc.Finish()

			case "single_uuid":
				enc := NewEncoder()
				enc.UUID(fieldDecisionID, uuid.Nil)
				got, err = enc.Finish()

			case "single_string":
				enc := NewEncoder()
				enc.String(fieldDecisionTitle, "hello")
				got, err = enc.Finish()

			case "unicode_string":
				enc := NewEncoder()
				enc.String(fieldDecisionTitle, "日本語")
				got, err = enc.Finish()

			case "empty_string_slice":
				enc := NewEncoder()
				enc.StringSlice(fieldDecisionEvidence, nil)
				got, err = enc.Finish()

			case "multi_string_slice":
				enc := NewEncoder()
				enc.StringSlice(fieldDecisionEvidence, []string{"zeta", "alpha", "mu"})
				got, err = enc.Finish()

			case "full_decision":
				var in struct {
					ID          string   `json:"id"`
					Title       string   `json:"title"`
					Status      string   `json:"status"`
					ScopeDomain string   `json:"scope_domain"`
					ScopeSystem string   `json:"scope_system"`
					Owner       string   `json:"owner"`
					EvidenceIDs []string `json:"evidence_ids"`
				}
				if err := json.Unmarshal(v.Input, &in); err != nil {
					t.Fatalf("parse input: %v", err)
				}
				enc := NewEncoder()
				enc.UUID(fieldDecisionID, uuid.MustParse(in.ID))
				enc.String(fieldDecisionTitle, in.Title)
				enc.String(fieldDecisionStatus, in.Status)
				enc.String(fieldDecisionScopeDo, in.ScopeDomain)
				enc.String(fieldDecisionScopeSys, in.ScopeSystem)
				enc.String(fieldDecisionOwner, in.Owner)
				enc.StringSlice(fieldDecisionEvidence, in.EvidenceIDs)
				got, err = enc.Finish()

			default:
				t.Fatalf("unknown vector %q", v.Name)
			}

			if err != nil {
				t.Fatalf("encode: %v", err)
			}

			want, err := hex.DecodeString(v.Hex)
			if err != nil {
				t.Fatalf("decode expected hex: %v", err)
			}

			if hex.EncodeToString(got) != v.Hex {
				t.Errorf("vector %q mismatch:\n  got  = %s\n  want = %s",
					v.Name, hex.EncodeToString(got), v.Hex)
			}
			_ = want
		})
	}
}
