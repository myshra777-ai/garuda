// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package fixtures

// Good3 uses a builtin element type. Variadic strings are legitimate.
// Expected: 0 violations.
func Good3(names ...string) string {
	if len(names) > 0 {
		return names[0]
	}
	return ""
}
