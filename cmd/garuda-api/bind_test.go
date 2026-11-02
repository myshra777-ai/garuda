// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package main

import "testing"

func TestResolveBindAddr(t *testing.T) {
	t.Setenv("GARUDA_BIND", "")
	if got := resolveBindAddr("9090"); got != "127.0.0.1:9090" {
		t.Fatalf("default must be loopback, got %q", got)
	}
	if got := resolveBindAddr(""); got != "127.0.0.1:8080" {
		t.Fatalf("empty port must default to 8080, got %q", got)
	}
	t.Setenv("GARUDA_BIND", "0.0.0.0:8080")
	if got := resolveBindAddr("9090"); got != "0.0.0.0:8080" {
		t.Fatalf("GARUDA_BIND must override, got %q", got)
	}
	t.Setenv("GARUDA_BIND", "  ")
	if got := resolveBindAddr("9090"); got != "127.0.0.1:9090" {
		t.Fatalf("blank GARUDA_BIND must be ignored, got %q", got)
	}
}
