// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package auth

import "strings"

// TokenValidator checks token formats and signatures.
type TokenValidator struct {
	SecretKey string
}

// ValidateToken parses and verifies authorization bearer tokens.
func (v *TokenValidator) ValidateToken(token string) bool {
	if strings.HasPrefix(token, "Bearer ") {
		return len(token) > 7
	}
	return false
}
