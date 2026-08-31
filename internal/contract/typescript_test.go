package contract_test

import (
	"strings"
	"testing"

	"github.com/myshra777-ai/garuda/internal/contract"
)

func TestTypeExporter_GenerateTypeScript(t *testing.T) {
	src := `
package models

type PaymentRequest struct {
	ID        string            ` + "`json:\"id\"`" + `
	Amount    int64             ` + "`json:\"amount\"`" + `
	IsActive  bool              ` + "`json:\"is_active\"`" + `
	Tags      []string          ` + "`json:\"tags\"`" + `
	Metadata  map[string]string ` + "`json:\"metadata\"`" + `
	ParentID  *string           ` + "`json:\"parent_id\"`" + `
}
`
	exporter := contract.NewTypeExporter()
	ts, err := exporter.GenerateTypeScript(src)
	if err != nil {
		t.Fatalf("unexpected error parsing Go source: %v", err)
	}

	expectedSubstrings := []string{
		"export interface PaymentRequest {",
		"id: string;",
		"amount: number;",
		"is_active: boolean;",
		"tags: string[];",
		"metadata: Record<string, string>;",
		"parent_id: string | null;",
	}

	for _, substr := range expectedSubstrings {
		if !strings.Contains(ts, substr) {
			t.Errorf("missing expected TypeScript line: %q\nGenerated output:\n%s", substr, ts)
		}
	}
}
