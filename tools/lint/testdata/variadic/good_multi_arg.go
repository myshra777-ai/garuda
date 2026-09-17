// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

type Opts2 struct {
	D bool
}

// Good reads opts[0] and opts[1], so the variadic is used as a slice.
// Expected: 0 violations.
func Good(opts ...Opts2) bool {
	if len(opts) > 1 {
		return opts[0].D && opts[1].D
	}
	return false
}
