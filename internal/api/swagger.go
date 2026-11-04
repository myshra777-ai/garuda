// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var openapiSpec []byte

// HandleOpenAPISpec serves the raw OpenAPI specification.
func (s *Server) HandleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(openapiSpec)
}
