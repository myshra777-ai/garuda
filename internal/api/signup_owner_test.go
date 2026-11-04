// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import "testing"

func TestIsCanonicalBootstrapOwnerTrustsNoDefaultAddress(t *testing.T) {
	t.Setenv("GARUDA_OWNER_EMAIL", "")
	for _, email := range []string{"myshra777@gmail.com", "admin@local", "someone@example.com", ""} {
		if isCanonicalBootstrapOwner(email) {
			t.Fatalf("%q must not be treated as the owner when GARUDA_OWNER_EMAIL is unset", email)
		}
	}
}

func TestIsCanonicalBootstrapOwnerMatchesConfiguredAddressOnly(t *testing.T) {
	t.Setenv("GARUDA_OWNER_EMAIL", "Owner@Example.com")
	if !isCanonicalBootstrapOwner("  owner@example.com ") {
		t.Fatal("the configured owner address should match, ignoring case and spaces")
	}
	if isCanonicalBootstrapOwner("myshra777@gmail.com") || isCanonicalBootstrapOwner("other@example.com") {
		t.Fatal("only the configured owner address should match")
	}
}
