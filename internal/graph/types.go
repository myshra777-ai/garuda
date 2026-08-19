// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package graph

// Node represents an entity in the graph.
type Node struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Label    string `json:"label"`
	Kind     string `json:"kind"`
	Package  string `json:"package"`
	File     string `json:"file"`
	Exported bool   `json:"exported"`
	Impact   int    `json:"impact,omitempty"`
}

// Edge represents a relationship between two nodes.
type Edge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Type      string `json:"type"`
	CrossRepo bool   `json:"cross_repo,omitempty"`
}
