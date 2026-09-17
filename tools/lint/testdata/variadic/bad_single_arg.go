// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

type Opts struct {
	D bool
}

// Bad reads only opts[0]. A caller passing two options has the second
// silently ignored.
// Expected: 1 variadic-single-arg violation.
func Bad(opts ...Opts) bool {
	if len(opts) > 0 {
		return opts[0].D
	}
	return false
}
