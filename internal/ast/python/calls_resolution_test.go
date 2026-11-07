// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package python

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPythonCrossFileAliasAndInstanceCallsResolution(t *testing.T) {
	tmp := t.TempDir()

	baseSrc := `class BaseController:
    def validate_session(self, token: str) -> bool:
        return True
`
	dbSrc := `class OrderRepository:
    def fetch_order(self, order_id: str) -> dict:
        return {"id": order_id}

def execute_raw_sql(query: str) -> list:
    return []

def build_repository() -> Optional[OrderRepository]:
    return OrderRepository()
`
	apiSrc := `import db as database_layer
from db import OrderRepository as RepoAlias, execute_raw_sql as run_sql, build_repository
from base import BaseController

class CheckoutHandler(BaseController):
    def __init__(self):
        self.repo = RepoAlias()

    def get_repo(self) -> RepoAlias:
        return self.repo

    def handle_post(self, order_id: str):
        super().validate_session(order_id)
        run_sql("SELECT 1")
        database_layer.execute_raw_sql("SELECT 2")
        local_repo = build_repository()
        local_repo.fetch_order(order_id)
        self.get_repo().fetch_order(order_id)
        return self.repo.fetch_order(order_id)
`
	if err := os.WriteFile(filepath.Join(tmp, "base.py"), []byte(baseSrc), 0644); err != nil {
		t.Fatal(err)
	}
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
	validateID := idByName["base.validate_session"]
	buildRepoID := idByName["db.build_repository"]

	if handlePostID == "" || fetchOrderID == "" || execSQLID == "" || validateID == "" || buildRepoID == "" {
		t.Fatalf("missing expected entities: handle_post=%q fetch_order=%q execute_raw_sql=%q validate=%q build_repo=%q",
			handlePostID, fetchOrderID, execSQLID, validateID, buildRepoID)
	}

	foundFetchOrder := false
	foundExecSQL := false
	foundSuperValidate := false
	foundBuildRepo := false
	for _, rel := range res.Relationships {
		if rel.FilePath == "" || rel.Line <= 0 {
			t.Errorf("relationship %s -> %s (%s) missing FilePath or Line: file=%q line=%d", rel.FromID, rel.ToID, rel.ClaimType, rel.FilePath, rel.Line)
		}
		if rel.ResolutionStatus != "RESOLVED" {
			t.Errorf("expected ResolutionStatus=RESOLVED, got %q", rel.ResolutionStatus)
		}
		if rel.FromID == handlePostID && rel.ToID == fetchOrderID && rel.ClaimType == "CALLS" {
			foundFetchOrder = true
			if rel.Confidence < 0.95 || rel.ResolutionMethod != "AST_EXACT" {
				t.Errorf("expected confidence>=0.95 and AST_EXACT for fetch_order, got %f / %s", rel.Confidence, rel.ResolutionMethod)
			}
		}
		if rel.FromID == handlePostID && rel.ToID == execSQLID && rel.ClaimType == "CALLS" {
			foundExecSQL = true
			if rel.Confidence < 0.95 || rel.ResolutionMethod != "AST_EXACT" {
				t.Errorf("expected confidence>=0.95 and AST_EXACT for execute_raw_sql, got %f / %s", rel.Confidence, rel.ResolutionMethod)
			}
		}
		if rel.FromID == handlePostID && rel.ToID == validateID && rel.ClaimType == "CALLS" {
			foundSuperValidate = true
		}
		if rel.FromID == handlePostID && rel.ToID == buildRepoID && rel.ClaimType == "CALLS" {
			foundBuildRepo = true
		}
	}

	if !foundFetchOrder {
		t.Errorf("expected CALLS edge from api.CheckoutHandler.handle_post to db.OrderRepository.fetch_order")
	}
	if !foundExecSQL {
		t.Errorf("expected CALLS edge from api.CheckoutHandler.handle_post to db.execute_raw_sql")
	}
	if !foundSuperValidate {
		t.Errorf("expected super().validate_session() CALLS edge from api.CheckoutHandler.handle_post to base.BaseController.validate_session")
	}
	if !foundBuildRepo {
		t.Errorf("expected CALLS edge from api.CheckoutHandler.handle_post to db.build_repository")
	}
}
