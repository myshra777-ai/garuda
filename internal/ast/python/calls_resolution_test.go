// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package python

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonCrossFileAliasAndInstanceCallsResolution(t *testing.T) {
	tmp := t.TempDir()

	dbSrc := `class OrderRepository:
    def fetch_order(self, order_id: str) -> dict:
        return {"id": order_id}

def execute_raw_sql(query: str) -> list:
    return []
`
	apiSrc := `import db as database_layer
from db import OrderRepository as RepoAlias, execute_raw_sql as run_sql

class CheckoutHandler:
    def __init__(self):
        self.repo = RepoAlias()

    def handle_post(self, order_id: str):
        run_sql("SELECT 1")
        database_layer.execute_raw_sql("SELECT 2")
        local_repo = RepoAlias()
        local_repo.fetch_order(order_id)
        return self.repo.fetch_order(order_id)
`
	if err := os.WriteFile(filepath.Join(tmp, "db.py"), []byte(dbSrc), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "api.py"), []byte(apiSrc), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := AnalyzeDirectory(tmp, "repo-1", "ws-1")
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	idByName := make(map[string]string)
	for _, e := range res.Entities {
		idByName[e.Package+"."+e.Name] = e.ID
	}

	handlePostID := idByName["api.handle_post"]
	fetchOrderID := idByName["db.fetch_order"]
	execSQLID := idByName["db.execute_raw_sql"]

	if handlePostID == "" || fetchOrderID == "" || execSQLID == "" {
		t.Fatalf("missing expected entities: handle_post=%q fetch_order=%q execute_raw_sql=%q",
			handlePostID, fetchOrderID, execSQLID)
	}

	foundFetchOrder := false
	foundExecSQL := false
	for _, rel := range res.Relationships {
		if rel.FromID == handlePostID && rel.ToID == fetchOrderID && rel.ClaimType == "CALLS" {
			foundFetchOrder = true
			if rel.Confidence < 0.95 {
				t.Errorf("expected confidence >= 0.95 for cross-file instance method call, got %f", rel.Confidence)
			}
		}
		if rel.FromID == handlePostID && rel.ToID == execSQLID && rel.ClaimType == "CALLS" {
			foundExecSQL = true
			if rel.Confidence < 0.95 {
				t.Errorf("expected confidence >= 0.95 for cross-file alias call, got %f", rel.Confidence)
			}
		}
	}

	if !foundFetchOrder {
		t.Errorf("expected CALLS edge from api.CheckoutHandler.handle_post to db.OrderRepository.fetch_order")
	}
	if !foundExecSQL {
		t.Errorf("expected CALLS edge from api.CheckoutHandler.handle_post to db.execute_raw_sql")
	}
}
