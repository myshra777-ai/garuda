// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestGenerateInitPassword(t *testing.T) {
	a, err := generateInitPassword()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := generateInitPassword()
	if len(a) != 24 || len(a) < 8 {
		t.Fatalf("want 24 chars (>= 8 required by init), got %d: %q", len(a), a)
	}
	if a == b {
		t.Fatal("two generated passwords must differ")
	}
	if a == "garuda2026" {
		t.Fatal("must not equal the legacy default")
	}
}

func TestGenerateControlToken(t *testing.T) {
	a, err := generateControlToken()
	if err != nil {
		t.Fatal(err)
	}
	b, _ := generateControlToken()
	if len(a) != 52 || a[:4] != "gct_" {
		t.Fatalf("want gct_ + 48 hex chars (52 total), got %d: %q", len(a), a)
	}
	if a == b {
		t.Fatal("two generated tokens must differ")
	}
	if a == "devtoken-changeme-please" {
		t.Fatal("must not equal the legacy default")
	}
}
