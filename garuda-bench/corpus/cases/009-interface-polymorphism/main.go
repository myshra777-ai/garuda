// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

type Writer interface {
	Write(p []byte) (n int, err error)
}

type Closer interface {
	Close() error
}

type FileWriter struct{}

func (f *FileWriter) Write(p []byte) (n int, err error) { return 0, nil }
func (f *FileWriter) Close() error                      { return nil }

func main() {}
