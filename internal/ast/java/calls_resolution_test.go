// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package java

import (
	"os"
	"path/filepath"
	"testing"
)

func TestJavaCrossPackageSpringDIAndCallsResolution(t *testing.T) {
	tmp := t.TempDir()

	pomXML := `<project><modelVersion>4.0.0</modelVersion><groupId>com.acme</groupId><artifactId>payments</artifactId><version>1.0.0</version></project>`
	if err := os.WriteFile(filepath.Join(tmp, "pom.xml"), []byte(pomXML), 0644); err != nil {
		t.Fatal(err)
	}

	dbDir := filepath.Join(tmp, "src", "main", "java", "com", "acme", "db")
	apiDir := filepath.Join(tmp, "src", "main", "java", "com", "acme", "api")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		t.Fatal(err)
	}

	repoIfaceJava := `package com.acme.db;

import java.util.Optional;

public interface IOrderRepository {
    Optional<String> fetchOrder(String orderId);
}
`
	repoImplJava := `package com.acme.db;

import java.util.Optional;

public class OrderRepository implements IOrderRepository {
    @Override
    public Optional<String> fetchOrder(String orderId) {
        return Optional.of(orderId);
    }
}
`
	auditLoggerJava := `package com.acme.db;

public class AuditLogger {
    public static void recordAuditTrail(String event) {
    }
}
`
	baseCtrlJava := `package com.acme.api;

public abstract class BaseController {
    public boolean validateToken(String token) {
        return token != null && !token.isEmpty();
    }
}
`
	checkoutCtrlJava := `package com.acme.api;

import com.acme.db.IOrderRepository;
import static com.acme.db.AuditLogger.recordAuditTrail;
import java.util.Optional;

public class CheckoutController extends BaseController {
    private final IOrderRepository repo;

    public CheckoutController(IOrderRepository repo) {
        this.repo = repo;
    }

    public Optional<IOrderRepository> getRepo() {
        return Optional.of(this.repo);
    }

    public Optional<String> handleCheckout(String id, String token) {
        super.validateToken(token);
        recordAuditTrail("checkout_started");
        var resolvedRepo = this.getRepo();
        return resolvedRepo.fetchOrder(id);
    }
}
`
	if err := os.WriteFile(filepath.Join(dbDir, "IOrderRepository.java"), []byte(repoIfaceJava), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "OrderRepository.java"), []byte(repoImplJava), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dbDir, "AuditLogger.java"), []byte(auditLoggerJava), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(apiDir, "BaseController.java"), []byte(baseCtrlJava), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(apiDir, "CheckoutController.java"), []byte(checkoutCtrlJava), 0644); err != nil {
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

	handleID := idByName["com.acme.api.CheckoutController.handleCheckout"]
	fetchID := idByName["com.acme.db.OrderRepository.fetchOrder"]
	auditID := idByName["com.acme.db.AuditLogger.recordAuditTrail"]
	validateID := idByName["com.acme.api.BaseController.validateToken"]

	if handleID == "" || fetchID == "" || auditID == "" || validateID == "" {
		t.Fatalf("missing expected Java entities: handle=%q fetch=%q audit=%q validate=%q (map=%v)", handleID, fetchID, auditID, validateID, idByName)
	}

	foundFetch := false
	foundAudit := false
	foundSuperValidate := false
	for _, rel := range res.Relationships {
		if rel.FilePath == "" || rel.Line <= 0 {
			t.Errorf("relationship %s -> %s (%s) missing FilePath or Line", rel.FromID, rel.ToID, rel.ClaimType)
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
		if rel.FromID == handleID && rel.ToID == auditID && rel.ClaimType == "CALLS" {
			foundAudit = true
			if rel.Confidence < 0.95 || rel.ResolutionMethod != "AST_EXACT" {
				t.Errorf("expected confidence >= 0.95 and AST_EXACT for recordAuditTrail, got %f / %s", rel.Confidence, rel.ResolutionMethod)
			}
		}
		if rel.FromID == handleID && rel.ToID == validateID && rel.ClaimType == "CALLS" {
			foundSuperValidate = true
		}
	}

	if !foundFetch {
		t.Errorf("expected CALLS edge from CheckoutController.handleCheckout to OrderRepository.fetchOrder via IOrderRepository DI + Optional<IOrderRepository> return type")
	}
	if !foundAudit {
		t.Errorf("expected CALLS edge from CheckoutController.handleCheckout to AuditLogger.recordAuditTrail via static import")
	}
	if !foundSuperValidate {
		t.Errorf("expected super.validateToken() CALLS edge from CheckoutController.handleCheckout to BaseController.validateToken")
	}
}
