// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package diff

import (
	"testing"

	"github.com/myshra777-ai/garuda/internal/analyzer"
)

func TestParseUnifiedDiff_MultiFileAndHunks(t *testing.T) {
	sample := `diff --git a/internal/api/handler.go b/internal/api/handler.go
index 1234567..89abcdef 100644
--- a/internal/api/handler.go
+++ b/internal/api/handler.go
@@ -42,0 +43,5 @@ func (s *Server) HandleCheckout() {
+	res := db.Query("SELECT * FROM orders")
+	_ = res
+}
@@ -108,3 +113,2 @@ func (s *Server) HandleHealth() {
-	oldCall()
+	newCall()
diff --git a/internal/store/orders.go b/internal/store/orders.go
new file mode 100644
index 0000000..1111111
--- /dev/null
+++ b/internal/store/orders.go
@@ -0,0 +1,20 @@
+package store
`

	deltas := ParseUnifiedDiff(sample)
	if len(deltas) != 2 {
		t.Fatalf("expected 2 file deltas, got %d", len(deltas))
	}

	d0 := deltas[0]
	if d0.FilePath != "internal/api/handler.go" {
		t.Errorf("expected internal/api/handler.go, got %s", d0.FilePath)
	}
	if len(d0.Hunks) != 2 {
		t.Fatalf("expected 2 hunks in handler.go, got %d", len(d0.Hunks))
	}
	if d0.Hunks[0].Start != 43 || d0.Hunks[0].End != 47 {
		t.Errorf("unexpected hunk[0]: %+v", d0.Hunks[0])
	}
	if d0.Hunks[1].Start != 113 || d0.Hunks[1].End != 114 {
		t.Errorf("unexpected hunk[1]: %+v", d0.Hunks[1])
	}

	d1 := deltas[1]
	if d1.FilePath != "internal/store/orders.go" || d1.Status != "A" {
		t.Errorf("unexpected second delta: %+v", d1)
	}

	entities := []analyzer.Entity{
		{
			ID:        "ent-1",
			Name:      "HandleCheckout",
			Kind:      analyzer.KindMethod,
			Package:   "api",
			File:      "/repo/internal/api/handler.go",
			LineStart: 35,
			LineEnd:   60,
		},
		{
			ID:        "ent-2",
			Name:      "UntouchedHelper",
			Kind:      analyzer.KindFunction,
			Package:   "api",
			File:      "/repo/internal/api/handler.go",
			LineStart: 70,
			LineEnd:   95,
		},
		{
			ID:        "ent-3",
			Name:      "HandleHealth",
			Kind:      analyzer.KindMethod,
			Package:   "api",
			File:      "/repo/internal/api/handler.go",
			LineStart: 105,
			LineEnd:   120,
		},
	}

	matched := MatchAnalyzerEntities(entities, "/repo", deltas)
	if len(matched) != 2 {
		t.Fatalf("expected 2 matched entities (HandleCheckout, HandleHealth), got %d", len(matched))
	}
	if matched[0].Name != "HandleCheckout" || matched[1].Name != "HandleHealth" {
		t.Errorf("unexpected matched entities: %+v", matched)
	}
}
