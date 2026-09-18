// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"embed"
	"io/fs"
	"net/http"
)

// staticFS holds the dashboard's CSS and JS. Served at /static/*.
// See internal/api/static/ for the source files.
//
//go:embed static
var staticFS embed.FS

// HandleStatic serves the embedded static assets under /static/.
//
// The StripPrefix is required because the embed.FS is rooted at
// "static/" (the directory name), while the request path is
// "/static/dashboard.css". Stripping the prefix aligns the two.
func (s *Server) HandleStatic() http.Handler {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		// fs.Sub fails only if the embed directory is missing, which
		// is a build-time error, not a runtime one. If we reach this
		// line, the binary was built incorrectly.
		panic("api: static FS is missing 'static' directory: " + err.Error())
	}
	return http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
}
