// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package typescript

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTypeScriptCrossFileAliasAndInstanceCallsResolution(t *testing.T) {
	tmp := t.TempDir()

	tsconfigJSON := `{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    }
  }
}`
	if err := os.WriteFile(filepath.Join(tmp, "tsconfig.json"), []byte(tsconfigJSON), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "src", "db"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(tmp, "src", "api"), 0755); err != nil {
		t.Fatal(err)
	}

	baseTS := `export abstract class BaseController {
  validateToken(token: string): boolean {
    return token.length > 0;
  }
}
`
	repoTS := `export interface IOrderRepository {
  fetchOrder(orderId: string): Promise<string>;
}

export class OrderRepository implements IOrderRepository {
  async fetchOrder(orderId: string): Promise<string> {
    return orderId;
  }
}

export function executeRawQuery(sql: string): void {}
`
	barrelTS := `export * from './repo';
`
	apiTS := `import * as dbMod from '@/db';
import { OrderRepository as RepoAlias, executeRawQuery as runSQL } from '@/db';
import { BaseController } from './base';

export class CheckoutController extends BaseController {
  constructor(private readonly repo: RepoAlias) {
    super();
  }

  async getRepo(): Promise<RepoAlias> {
    return this.repo;
  }

  async handleCheckout(id: string): Promise<string> {
    super.validateToken(id);
    runSQL('SELECT 1');
    dbMod.executeRawQuery('SELECT 2');
    const localRepo = await this.getRepo();
    return localRepo.fetchOrder(id);
  }
}
`
	if err := os.WriteFile(filepath.Join(tmp, "src", "api", "base.ts"), []byte(baseTS), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "db", "repo.ts"), []byte(repoTS), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "db", "index.ts"), []byte(barrelTS), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmp, "src", "api", "checkout.ts"), []byte(apiTS), 0644); err != nil {
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

	handleID := idByName["src.api.checkout.handleCheckout"]
	fetchID := idByName["src.db.repo.fetchOrder"]
	execID := idByName["src.db.repo.executeRawQuery"]
	validateID := idByName["src.api.base.validateToken"]

	if handleID == "" || fetchID == "" || execID == "" || validateID == "" {
		t.Fatalf("missing expected entities: handle=%q fetch=%q exec=%q validate=%q (map=%v)", handleID, fetchID, execID, validateID, idByName)
	}

	foundFetch := false
	foundExec := false
	foundSuperValidate := false
	for _, rel := range res.Relationships {
		if rel.FilePath == "" || rel.Line <= 0 {
			t.Errorf("relationship %s -> %s (%s) missing FilePath or Line: file=%q line=%d", rel.FromID, rel.ToID, rel.ClaimType, rel.FilePath, rel.Line)
		}
		if rel.ResolutionStatus != "RESOLVED" {
			t.Errorf("expected ResolutionStatus=RESOLVED, got %q", rel.ResolutionStatus)
		}
		if rel.FromID == handleID && rel.ToID == fetchID && rel.ClaimType == "CALLS" {
			foundFetch = true
			if rel.Confidence < 0.95 || rel.ResolutionMethod != "AST_EXACT" {
				t.Errorf("expected confidence >= 0.95 and AST_EXACT for fetchOrder, got %f / %s", rel.Confidence, rel.ResolutionMethod)
			}
		}
		if rel.FromID == handleID && rel.ToID == execID && rel.ClaimType == "CALLS" {
			foundExec = true
			if rel.Confidence < 0.95 || rel.ResolutionMethod != "AST_EXACT" {
				t.Errorf("expected confidence >= 0.95 and AST_EXACT for executeRawQuery, got %f / %s", rel.Confidence, rel.ResolutionMethod)
			}
		}
		if rel.FromID == handleID && rel.ToID == validateID && rel.ClaimType == "CALLS" {
			foundSuperValidate = true
		}
	}

	if !foundFetch {
		t.Errorf("expected CALLS edge from handleCheckout to OrderRepository.fetchOrder via tsconfig @/db barrel + Promise<RepoAlias> return type")
	}
	if !foundExec {
		t.Errorf("expected CALLS edge from handleCheckout to executeRawQuery")
	}
	if !foundSuperValidate {
		t.Errorf("expected super.validateToken() CALLS edge from handleCheckout to BaseController.validateToken")
	}
}
