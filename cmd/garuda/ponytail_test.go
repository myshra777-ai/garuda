// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func TestHygieneRelationships(t *testing.T) {
	tests := []struct {
		name  string
		edges []map[string]interface{}
		want  []analyzer.Relationship
	}{
		{
			name: "valid edge",
			edges: []map[string]interface{}{
				{
					"from": "caller",
					"to":   "callee",
					"type": "CALLS",
				},
			},
			want: []analyzer.Relationship{
				{
					From: "caller",
					To:   "callee",
					Type: "CALLS",
				},
			},
		},
		{
			name: "invalid rows are skipped",
			edges: []map[string]interface{}{
				{
					"from": "caller",
					"to":   "callee",
				},
				{
					"from": 42,
					"to":   "callee",
					"type": "CALLS",
				},
				{
					"from": "",
					"to":   "callee",
					"type": "CALLS",
				},
				{
					"from": "caller",
					"to":   "",
					"type": "CALLS",
				},
				{
					"from": "caller",
					"to":   "callee",
					"type": "",
				},
			},
			want: []analyzer.Relationship{},
		},
		{
			name: "input order is preserved",
			edges: []map[string]interface{}{
				{
					"from": "a",
					"to":   "b",
					"type": "CALLS",
				},
				{
					"from": "b",
					"to":   "c",
					"type": "IMPORTS",
				},
			},
			want: []analyzer.Relationship{
				{
					From: "a",
					To:   "b",
					Type: "CALLS",
				},
				{
					From: "b",
					To:   "c",
					Type: "IMPORTS",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hygieneRelationships(tt.edges)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("hygieneRelationships() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestHygieneCommandContract(t *testing.T) {
	cmd := newHygieneCommand("hygiene [path]", "")

	if cmd.Use != "hygiene [path]" {
		t.Fatalf("Use = %q, want %q", cmd.Use, "hygiene [path]")
	}
	if cmd.Deprecated != "" {
		t.Fatalf("canonical command unexpectedly deprecated: %q", cmd.Deprecated)
	}
	if cmd.Flags().Lookup("json") == nil {
		t.Fatal("missing --json flag")
	}
	if cmd.Flags().Lookup("output") == nil {
		t.Fatal("missing --output flag")
	}
}

func TestPonytailCompatibilityCommandContract(t *testing.T) {
	cmd := newHygieneCommand("ponytail [path]", "use `garuda hygiene` instead")

	if cmd.Use != "ponytail [path]" {
		t.Fatalf("Use = %q, want %q", cmd.Use, "ponytail [path]")
	}
	if cmd.Deprecated == "" {
		t.Fatal("compatibility command must be deprecated")
	}
	if !strings.Contains(cmd.Deprecated, "garuda hygiene") {
		t.Fatalf("Deprecated = %q, want reference to garuda hygiene", cmd.Deprecated)
	}
	if cmd.Flags().Lookup("json") == nil {
		t.Fatal("missing --json flag")
	}
	if cmd.Flags().Lookup("output") == nil {
		t.Fatal("missing --output flag")
	}
}
