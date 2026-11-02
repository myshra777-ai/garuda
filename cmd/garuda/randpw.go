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
