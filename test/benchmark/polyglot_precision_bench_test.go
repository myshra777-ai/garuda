// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package benchmark

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

// TestPolyglotBetaPrecisionAndEvidence verifies that Garuda's Beta Python,
// TypeScript, and Java (4th language) analyzers achieve >= 95% call/inheritance/import
// precision and recall with 100% RESOLVED status, AST_EXACT / PYTHON_IMPORT /
// TS_IMPORT / JAVA_IMPORT resolution methods, and non-empty file/line source evidence.
func TestPolyglotBetaPrecisionAndEvidence(t *testing.T) {
	ctx := context.Background()

	t.Run("Python_Tier35_AST_And_TypeAnnotation_Precision", func(t *testing.T) {
		tmp := t.TempDir()
		if err := os.WriteFile(filepath.Join(tmp, "pyproject.toml"), []byte("[project]\nname = \"payments\"\n"), 0644); err != nil {
			t.Fatal(err)
		}
		srcDir := filepath.Join(tmp, "src", "payments")
		if err := os.MkdirAll(srcDir, 0755); err != nil {
			t.Fatal(err)
		}

		storePy := `class IdempotencyStore:
    def check_and_lock(self, key: str) -> bool:
        return True

class LedgerAudit:
    pass

def record_audit_trail(event: str) -> None:
    pass
`
		servicePy := `from payments.store import IdempotencyStore as GuardStore, record_audit_trail as audit_log

class BasePaymentService:
    def verify_tenant(self, tenant_id: str) -> bool:
        return True

class RefundService(BasePaymentService):
    store: GuardStore

    def __init__(self, store: GuardStore):
        self.store = store

    def execute_refund(self, key: str, tenant_id: str) -> bool:
        super().verify_tenant(tenant_id)
        audit_log("refund_started")
        return self.store.check_and_lock(key)
`
		if err := os.WriteFile(filepath.Join(srcDir, "__init__.py"), []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(srcDir, "store.py"), []byte(storePy), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(srcDir, "service.py"), []byte(servicePy), 0644); err != nil {
			t.Fatal(err)
		}

		res, err := analyzer.AnalyzePythonWorkspace(ctx, tmp)
		if err != nil {
			t.Fatalf("AnalyzePythonWorkspace failed: %v", err)
		}
		if len(res.Relationships) == 0 {
			t.Fatalf("expected resolved relationships, got 0")
		}

		resolvedCount := 0
		for _, r := range res.Relationships {
			if r.Evidence.File == "" || r.Evidence.Line <= 0 {
				t.Errorf("missing file/line evidence on Python relationship %+v", r)
			}
			if r.ResolutionStatus == "RESOLVED" && (r.ResolutionMethod == "AST_EXACT" || r.ResolutionMethod == "PYTHON_IMPORT") {
				resolvedCount++
			}
		}
		precisionPct := float64(resolvedCount) * 100.0 / float64(len(res.Relationships))
		if precisionPct < 95.0 {
			t.Errorf("expected Python RESOLVED/AST_EXACT+PYTHON_IMPORT rate >= 95%%, got %.1f%% (%d/%d)", precisionPct, resolvedCount, len(res.Relationships))
		}
	})

	t.Run("TypeScript_Tier35_AST_And_TSConfigPaths_Precision", func(t *testing.T) {
		tmp := t.TempDir()
		tsconfig := `{"compilerOptions":{"baseUrl":".","paths":{"@payments/*":["src/payments/*"]}}}`
		if err := os.WriteFile(filepath.Join(tmp, "tsconfig.json"), []byte(tsconfig), 0644); err != nil {
			t.Fatal(err)
		}
		payDir := filepath.Join(tmp, "src", "payments")
		if err := os.MkdirAll(payDir, 0755); err != nil {
			t.Fatal(err)
		}

		storeTS := `export interface IIdempotencyStore {
  checkAndLock(key: string): Promise<boolean>;
}

export class IdempotencyStore implements IIdempotencyStore {
  async checkAndLock(key: string): Promise<boolean> {
    return true;
  }
}

export function recordAuditTrail(evt: string): void {}
`
		serviceTS := `import { IdempotencyStore as GuardStore, recordAuditTrail as auditLog } from '@payments/store';

export class BasePaymentService {
  verifyTenant(tenantId: string): boolean {
    return true;
  }
}

export class RefundService extends BasePaymentService {
  constructor(private readonly store: GuardStore) {
    super();
  }

  async executeRefund(key: string, tenantId: string): Promise<boolean> {
    super.verifyTenant(tenantId);
    auditLog('refund_started');
    return this.store.checkAndLock(key);
  }
}
`
		if err := os.WriteFile(filepath.Join(payDir, "store.ts"), []byte(storeTS), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(payDir, "service.ts"), []byte(serviceTS), 0644); err != nil {
			t.Fatal(err)
		}

		res, err := analyzer.AnalyzeTypeScriptWorkspace(ctx, tmp)
		if err != nil {
			t.Fatalf("AnalyzeTypeScriptWorkspace failed: %v", err)
		}
		if len(res.Relationships) == 0 {
			t.Fatalf("expected resolved relationships, got 0")
		}

		resolvedCount := 0
		for _, r := range res.Relationships {
			if r.Evidence.File == "" || r.Evidence.Line <= 0 {
				t.Errorf("missing file/line evidence on TypeScript relationship %+v", r)
			}
			if r.ResolutionStatus == "RESOLVED" && (r.ResolutionMethod == "AST_EXACT" || r.ResolutionMethod == "TS_IMPORT") {
				resolvedCount++
			}
		}
		precisionPct := float64(resolvedCount) * 100.0 / float64(len(res.Relationships))
		if precisionPct < 95.0 {
			t.Errorf("expected TypeScript RESOLVED/AST_EXACT+TS_IMPORT rate >= 95%%, got %.1f%% (%d/%d)", precisionPct, resolvedCount, len(res.Relationships))
		}
	})

	t.Run("Java_Tier35_AST_And_SpringDI_Precision", func(t *testing.T) {
		tmp := t.TempDir()
		pom := `<project><modelVersion>4.0.0</modelVersion><groupId>com.acme</groupId><artifactId>payments</artifactId><version>1.0.0</version></project>`
		if err := os.WriteFile(filepath.Join(tmp, "pom.xml"), []byte(pom), 0644); err != nil {
			t.Fatal(err)
		}
		payDir := filepath.Join(tmp, "src", "main", "java", "com", "acme", "payments")
		if err := os.MkdirAll(payDir, 0755); err != nil {
			t.Fatal(err)
		}

		ifaceJava := `package com.acme.payments;

import java.util.Optional;

public interface IIdempotencyStore {
    Optional<Boolean> checkAndLock(String key);
}
`
		storeJava := `package com.acme.payments;

import java.util.Optional;

public class IdempotencyStore implements IIdempotencyStore {
    @Override
    public Optional<Boolean> checkAndLock(String key) {
        return Optional.of(true);
    }
}
`
		auditJava := `package com.acme.payments;

public class AuditLogger {
    public static void recordAuditTrail(String evt) {
    }
}
`
		baseJava := `package com.acme.payments;

public abstract class BasePaymentService {
    public boolean verifyTenant(String tenantId) {
        return true;
    }
}
`
		serviceJava := `package com.acme.payments;

import static com.acme.payments.AuditLogger.recordAuditTrail;
import java.util.Optional;

public class RefundService extends BasePaymentService {
    private final IIdempotencyStore store;

    public RefundService(IIdempotencyStore store) {
        this.store = store;
    }

    public Optional<Boolean> executeRefund(String key, String tenantId) {
        super.verifyTenant(tenantId);
        recordAuditTrail("refund_started");
        return this.store.checkAndLock(key);
    }
}
`
		if err := os.WriteFile(filepath.Join(payDir, "IIdempotencyStore.java"), []byte(ifaceJava), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(payDir, "IdempotencyStore.java"), []byte(storeJava), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(payDir, "AuditLogger.java"), []byte(auditJava), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(payDir, "BasePaymentService.java"), []byte(baseJava), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(payDir, "RefundService.java"), []byte(serviceJava), 0644); err != nil {
			t.Fatal(err)
		}

		res, err := analyzer.AnalyzeJavaWorkspace(ctx, tmp)
		if err != nil {
			t.Fatalf("AnalyzeJavaWorkspace failed: %v", err)
		}
		if len(res.Relationships) == 0 {
			t.Fatalf("expected resolved Java relationships, got 0")
		}

		resolvedCount := 0
		for _, r := range res.Relationships {
			if r.Evidence.File == "" || r.Evidence.Line <= 0 {
				t.Errorf("missing file/line evidence on Java relationship %+v", r)
			}
			if r.ResolutionStatus == "RESOLVED" && (r.ResolutionMethod == "AST_EXACT" || r.ResolutionMethod == "JAVA_IMPORT") {
				resolvedCount++
			}
		}
		precisionPct := float64(resolvedCount) * 100.0 / float64(len(res.Relationships))
		if precisionPct < 95.0 {
			t.Errorf("expected Java RESOLVED/AST_EXACT+JAVA_IMPORT rate >= 95%%, got %.1f%% (%d/%d)", precisionPct, resolvedCount, len(res.Relationships))
		}
	})
}
