// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package contract

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// APIContract represents an extracted API contract
type APIContract struct {
	ServiceName    string                 `json:"service_name"`
	Endpoint       string                 `json:"endpoint"`
	Method         string                 `json:"method"`
	RequestSchema  map[string]interface{} `json:"request_schema,omitempty"`
	ResponseSchema map[string]interface{} `json:"response_schema,omitempty"`
	ContractHash   string                 `json:"contract_hash"`
	Evidence       Evidence               `json:"evidence"`
}

// Evidence tracks source evidence for a contract
type Evidence struct {
	File   string `json:"file"`
	Line   int    `json:"line"`
	Commit string `json:"commit"`
}

// Consumer represents a service/entity that calls an API
type Consumer struct {
	EntityID string   `json:"entity_id"`
	RepoID   string   `json:"repo_id"`
	Evidence Evidence `json:"evidence"`
}

// ExtractContracts extracts API contracts from a Go repository
func ExtractContracts(repoPath, commitSHA string) ([]APIContract, []Consumer, error) {
	var contracts []APIContract
	var consumers []Consumer

	// 1. Find HTTP handlers (common patterns)
	err := filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse file
		fset := token.NewFileSet()
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		fileAst, err := parser.ParseFile(fset, path, content, parser.AllErrors)
		if err != nil {
			return nil
		}

		// Extract contracts from this file
		ast.Inspect(fileAst, func(n ast.Node) bool {
			switch x := n.(type) {
			case *ast.FuncDecl:
				// Look for HTTP handler patterns
				contract := detectHTTPContract(x, fset, path, commitSHA)
				if contract != nil {
					contracts = append(contracts, *contract)
				}
			case *ast.CallExpr:
				// Look for HTTP client calls
				consumer := detectHTTPCall(x, fset, path, commitSHA)
				if consumer != nil {
					consumers = append(consumers, *consumer)
				}
			}
			return true
		})
		return nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("failed to walk repo: %w", err)
	}

	return contracts, consumers, nil
}

// detectHTTPContract detects HTTP handlers (e.g., gin, echo, net/http)
func detectHTTPContract(fn *ast.FuncDecl, fset *token.FileSet, filePath, commitSHA string) *APIContract {
	// Check if this looks like an HTTP handler
	if !strings.Contains(fn.Name.Name, "Handler") &&
		!strings.Contains(fn.Name.Name, "Controller") &&
		!strings.Contains(fn.Name.Name, "Endpoint") {
		return nil
	}

	// Extract basic info (simplified – production version would be more sophisticated)
	return &APIContract{
		ServiceName:  extractServiceName(filePath),
		Endpoint:     "/" + strings.ToLower(fn.Name.Name),
		Method:       "POST", // Default, would be detected from router registrations
		ContractHash: generateContractHash(fn),
		Evidence: Evidence{
			File:   filePath,
			Line:   fset.Position(fn.Pos()).Line,
			Commit: commitSHA,
		},
	}
}

// detectHTTPCall detects outgoing HTTP calls
func detectHTTPCall(call *ast.CallExpr, fset *token.FileSet, filePath, commitSHA string) *Consumer {
	// Look for HTTP client calls (e.g., http.Get, client.Do, etc.)
	var called string
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		called = fun.Sel.Name
	case *ast.Ident:
		called = fun.Name
	default:
		return nil
	}

	if called == "Get" || called == "Post" || called == "Do" || called == "Request" {
		return &Consumer{
			EntityID: "", // Will be resolved by the store
			RepoID:   "", // Will be resolved by the store
			Evidence: Evidence{
				File:   filePath,
				Line:   fset.Position(call.Pos()).Line,
				Commit: commitSHA,
			},
		}
	}
	return nil
}

// Helper: extract service name from file path
func extractServiceName(filePath string) string {
	parts := strings.Split(filePath, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if strings.Contains(parts[i], "service") ||
			strings.Contains(parts[i], "api") ||
			strings.Contains(parts[i], "handler") {
			return parts[i]
		}
	}
	return "unknown"
}

// Helper: generate contract hash from function
func generateContractHash(fn *ast.FuncDecl) string {
	// In production, this would create a structural hash
	return fmt.Sprintf("%x", sha256.Sum256([]byte(fn.Name.Name)))
}
