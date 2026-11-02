// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"strings"
)

// resolveBindAddr returns the listen address for garuda-api.
//
// Default is loopback-only (127.0.0.1:<port>). The previous default, ":"+port,
// exposed the API (including unauthenticated debug endpoints) on every network
// interface. To expose it deliberately, set GARUDA_BIND to a full address such as
// "0.0.0.0:8080" (the container images do this; publish the port to loopback on
// the host unless you intend public exposure).
func resolveBindAddr(port string) string {
	if v := strings.TrimSpace(os.Getenv("GARUDA_BIND")); v != "" {
		return v
	}
	port = strings.TrimSpace(port)
	if port == "" {
		port = "8080"
	}
	return "127.0.0.1:" + port
}
