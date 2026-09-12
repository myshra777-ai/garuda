// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"context"
	"net/http"
	"strings"
)

type sessionContextKeyType struct{}

var sessionContextKey = sessionContextKeyType{}

// RequireSession wraps an http.HandlerFunc with session authentication.
//
// Used in two ways:
//
//   - In RegisterRoutes, as middleware on a subrouter: sess.Use(...)
//   - In cmd/garuda/dev_cmd.go's raw http.ServeMux, as a per-handler
//     wrapper: mux.HandleFunc("/dashboard", server.RequireSession(server.HandleDashboard))
//
// HTML requests are redirected to /login. API requests receive a 401
// JSON response. The distinction is based on the request path and the
// Accept header.
func (s *Server) RequireSession(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(SessionCookieName)
		if err != nil {
			s.sessionUnauthorized(w, r)
			return
		}
		sess, ok := s.sessions.Get(cookie.Value)
		if !ok {
			s.sessionUnauthorized(w, r)
			return
		}
		ctx := context.WithValue(r.Context(), sessionContextKey, sess)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) sessionUnauthorized(w http.ResponseWriter, r *http.Request) {
	if wantsHTML(r) {
		target := "/login"
		if r.URL.Path != "/" && r.URL.Path != "/dashboard" {
			target += "?next=" + r.URL.Path
		}
		http.Redirect(w, r, target, http.StatusSeeOther)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"authentication required"}`))
}

// wantsHTML reports whether the response should be HTML.
//
// /dashboard* always wants HTML. Any other path wants HTML only if the
// client sent Accept: text/html.
func wantsHTML(r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/dashboard") {
		return true
	}
	return strings.Contains(r.Header.Get("Accept"), "text/html")
}

// SessionFromContext retrieves the authenticated session, if any.
// Returns (nil, false) for public requests that reached the handler
// without session middleware.
func SessionFromContext(ctx context.Context) (*Session, bool) {
	sess, ok := ctx.Value(sessionContextKey).(*Session)
	return sess, ok
}
