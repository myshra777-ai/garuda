// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net"
	"net/http"
	"os"
	"strings"
)

// debugTokenAllowed reports whether the unauthenticated /debug/token endpoint may
// serve this request. It is allowed only for direct loopback connections: the TCP
// peer must be a loopback IP and the request must not carry proxy-forwarding
// headers (a reverse proxy on the same host would otherwise make every remote
// caller look like 127.0.0.1).
//
// Docker-published ports reach the container from the bridge gateway, not loopback;
// for local dev setups like that, set GARUDA_ALLOW_REMOTE_DEBUG_TOKEN=1. The handler
// still refuses everything when GARUDA_ENV=production.
func debugTokenAllowed(r *http.Request) bool {
	if r == nil {
		return false
	}
	if os.Getenv("GARUDA_ALLOW_REMOTE_DEBUG_TOKEN") == "1" {
		return true
	}
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip", "Forwarded", "X-Forwarded-Host"} {
		if strings.TrimSpace(r.Header.Get(h)) != "" {
			return false
		}
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err != nil {
		host = strings.TrimSpace(r.RemoteAddr)
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
