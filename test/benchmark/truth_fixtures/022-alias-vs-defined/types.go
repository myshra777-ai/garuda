// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package identity

type RawToken = string

type AccountID string

func (a AccountID) String() string {
	return string(a)
}

func ValidateAccount(id AccountID, token RawToken) bool {
	return id.String() != "" && len(token) > 0
}
