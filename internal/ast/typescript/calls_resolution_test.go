// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package typescript

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTypeScriptCrossFileAliasAndInstanceCallsResolution(t *testing.T) {
	tmp := t.TempDir()

	dbTS := `export class OrderRepository {
  fetchOrder(orderId: string): string {
    return orderId;
  }
}

export function executeRawQuery(sql: string): void {}
`
	apiTS := `import * as dbMod from './db';
import { OrderRepository as RepoAlias, executeRawQuery as runSQL } from './db';

export class CheckoutController {
  private repo: RepoAlias;

  constructor() {
    this.repo = new RepoAlias();
  }

  handleCheckout(id: string): string {
    runSQL('SELECT 1');
    dbMod.executeRawQuery('SELECT 2');
    const localRepo = new RepoAlias();
    localRepo.fetchOrder(id);
    return this.repo.fetchOrder(id);
  }
}
`
	if err := os.WriteFile(filepath.Join(tmp, "db.ts"), []byte(dbTS), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "api.ts"), []byte(apiTS), 0644); err != nil {
		t.Fatal(err)
	}

	res, err := AnalyzeDirectory(tmp)
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	idByName := make(map[string]string)
	for _, e := range res.Entities {
		idByName[e.Package+"."+e.Name] = e.ID
	}

	handleID := idByName["api.handleCheckout"]
	fetchID := idByName["db.fetchOrder"]
	execID := idByName["db.executeRawQuery"]

	if handleID == "" || fetchID == "" || execID == "" {
		t.Fatalf("missing expected entities: handle=%q fetch=%q exec=%q", handleID, fetchID, execID)
	}

	foundFetch := false
	foundExec := false
	for _, rel := range res.Relationships {
		if rel.FromID == handleID && rel.ToID == fetchID && rel.ClaimType == "CALLS" {
			foundFetch = true
			if rel.Confidence < 0.95 {
				t.Errorf("expected confidence >= 0.95 for cross-file instance method call, got %f", rel.Confidence)
			}
		}
		if rel.FromID == handleID && rel.ToID == execID && rel.ClaimType == "CALLS" {
			foundExec = true
			if rel.Confidence < 0.95 {
				t.Errorf("expected confidence >= 0.95 for cross-file alias call, got %f", rel.Confidence)
			}
		}
	}

	if !foundFetch {
		t.Errorf("expected CALLS edge from api.CheckoutController.handleCheckout to db.OrderRepository.fetchOrder")
	}
	if !foundExec {
		t.Errorf("expected CALLS edge from api.CheckoutController.handleCheckout to db.executeRawQuery")
	}
}
