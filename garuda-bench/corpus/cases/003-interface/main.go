// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

type Reader interface {
	Read(p []byte) (n int, err error)
}

type CustomBuffer struct{}

func (b *CustomBuffer) Read(p []byte) (n int, err error) {
	return 0, nil
}

func main() {}
