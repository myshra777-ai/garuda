// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

// Command garuda-talon is the standalone Garuda Talon Active Execution Harness binary.
// It delegates directly to `garuda talon <subcommand>` so users and CI pipelines can
// invoke either `garuda talon ...` or `garuda-talon ...` interchangeably.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	self, err := os.Executable()
	garudaBin := "garuda"
	if err == nil {
		candidate := filepath.Join(filepath.Dir(self), "garuda")
		if _, statErr := os.Stat(candidate); statErr == nil {
			garudaBin = candidate
		}
	}

	args := append([]string{"talon"}, os.Args[1:]...)
	cmd := exec.Command(garudaBin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "garuda-talon error: %v\n", err)
		os.Exit(1)
	}
}
