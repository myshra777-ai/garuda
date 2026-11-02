// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"crypto/rand"
	"encoding/hex"
)

// generateInitPassword returns a random 24-hex-character password for the owner
// account created by `garuda init` when the operator supplies none. It replaces
// the previous well-known default so a fresh install never ships a guessable login.
func generateInitPassword() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generateControlToken returns a random bearer token for the /admin Control Plane
// ("gct_" + 48 hex characters). It replaces the previous well-known default
// "devtoken-changeme-please" so a fresh install never ships a guessable operator token.
func generateControlToken() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "gct_" + hex.EncodeToString(b), nil
}
