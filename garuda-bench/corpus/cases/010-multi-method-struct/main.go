// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

type Service interface {
	Start() error
	Stop() error
}

type HTTPService struct{}
func (h *HTTPService) Start() error { return nil }
func (h *HTTPService) Stop() error  { return nil }

type GRPCService struct{}
func (g *GRPCService) Start() error { return nil }
func (g *GRPCService) Stop() error  { return nil }

func main() {}
