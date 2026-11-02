// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDebugTokenAllowed(t *testing.T) {
	cases := []struct {
		name    string
		remote  string
		headers map[string]string
		want    bool
	}{
		{"ipv4 loopback", "127.0.0.1:5555", nil, true},
		{"ipv6 loopback", "[::1]:5555", nil, true},
		{"remote ipv4", "203.0.113.9:4444", nil, false},
		{"private lan", "192.168.1.20:4444", nil, false},
		{"loopback behind proxy (XFF)", "127.0.0.1:5555", map[string]string{"X-Forwarded-For": "203.0.113.9"}, false},
		{"loopback behind proxy (Forwarded)", "127.0.0.1:5555", map[string]string{"Forwarded": "for=203.0.113.9"}, false},
		{"loopback behind proxy (X-Real-Ip)", "127.0.0.1:5555", map[string]string{"X-Real-Ip": "203.0.113.9"}, false},
		{"garbage remote", "not-an-ip", nil, false},
		{"empty remote", "", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/debug/token", nil)
			r.RemoteAddr = tc.remote
			for k, v := range tc.headers {
				r.Header.Set(k, v)
			}
			if got := debugTokenAllowed(r); got != tc.want {
				t.Fatalf("debugTokenAllowed(%q, %v) = %v, want %v", tc.remote, tc.headers, got, tc.want)
			}
		})
	}
	if debugTokenAllowed(nil) {
		t.Fatal("nil request must be rejected")
	}

	t.Setenv("GARUDA_ALLOW_REMOTE_DEBUG_TOKEN", "1")
	r := httptest.NewRequest(http.MethodGet, "/debug/token", nil)
	r.RemoteAddr = "172.17.0.1:5555"
	if !debugTokenAllowed(r) {
		t.Fatal("opt-in env must allow docker-bridge callers")
	}
}
