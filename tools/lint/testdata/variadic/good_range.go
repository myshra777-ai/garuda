// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

type Opts3 struct {
	D bool
}

// Good2 ranges over the slice.
// Expected: 0 violations.
func Good2(opts ...Opts3) bool {
	for _, o := range opts {
		if o.D {
			return true
		}
	}
	return false
}
