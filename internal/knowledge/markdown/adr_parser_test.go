// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package markdown

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/knowledge"
)

func writeADR(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "0001-payments.md")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

var testTenant = uuid.MustParse("00000000-0000-0000-0000-000000000001")

const paymentADR = "# ADR 0001: Payment Refund Processing\n" +
	"\n" +
	"Status: accepted. This preamble must be ignored.\n" +
	"\n" +
	"## Context\n" +
	"\n" +
	"- `RefundHandler` must call `IdempotencyStore`\n" +
	"\n" +
	"## Decision\n" +
	"\n" +
	"- `RefundHandler` must not call `DirectBankAPI`\n" +
	"- `RefundHandler` must be idempotent\n" +
	"- The HTTP server MUST use the chi router.\n" +
	"* Payment service SHOULD log every request\n" +
	"- > quoted text must be skipped\n" +
	"- | table | must | be skipped |\n" +
	"\n" +
	"```go\n" +
	"// code must be skipped\n" +
	"```\n" +
	"\n" +
	"## Alternatives considered\n" +
	"\n" +
	"- We must not extract rules from this section\n"

func TestParseADR_ExtractsClaimsFromNormativeSections(t *testing.T) {
	path := writeADR(t, paymentADR)
	claims, err := ParseADR(path, testTenant, "payments")
	if err != nil {
		t.Fatalf("ParseADR: %v", err)
	}

	type key struct {
		subject string
		mod     knowledge.Modality
		pred    knowledge.Predicate
		object  string
	}
	want := map[key]bool{
		{"RefundHandler", knowledge.ModalityMust, knowledge.PredicateCalls, "IdempotencyStore"}:    true,
		{"RefundHandler", knowledge.ModalityMustNot, knowledge.PredicateCalls, "DirectBankAPI"}:    true,
		{"RefundHandler", knowledge.ModalityMust, knowledge.PredicateIdempotent, "IdempotencyKey"}: true,
		{"HTTP server", knowledge.ModalityMust, knowledge.Predicate("use"), "the chi router"}:      true,
		{"Payment service", knowledge.ModalityShould, knowledge.Predicate("log"), "every request"}: true,
	}

	if len(claims) != len(want) {
		for _, c := range claims {
			t.Logf("claim: %q %s %s %q (section %q)", c.Subject, c.Modality, c.Predicate, c.Object, c.SectionTitle)
		}
		t.Fatalf("got %d claims, want %d", len(claims), len(want))
	}
	for _, c := range claims {
		k := key{c.Subject, c.Modality, c.Predicate, c.Object}
		if !want[k] {
			t.Errorf("unexpected claim %+v", k)
		}
		if c.DocumentTitle != "ADR 0001: Payment Refund Processing" {
			t.Errorf("document title = %q", c.DocumentTitle)
		}
		if c.DocumentType != "ADR" || c.Workspace != "payments" || c.TenantID != testTenant {
			t.Errorf("metadata not set: %+v", c)
		}
		if c.Status != knowledge.StatusUnverified {
			t.Errorf("new claims must start UNVERIFIED, got %s", c.Status)
		}
		if c.LineStart == 0 || c.LineStart != c.LineEnd {
			t.Errorf("line numbers not set: %d-%d", c.LineStart, c.LineEnd)
		}
		if c.ID == uuid.Nil {
			t.Error("claim ID not set")
		}
	}
}

func TestParseADR_ProvenanceAndConfidence(t *testing.T) {
	path := writeADR(t, paymentADR)
	claims, err := ParseADR(path, testTenant, "payments")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range claims {
		switch {
		case c.Predicate == knowledge.PredicateCalls:
			if c.Provenance != knowledge.ProvenanceExtracted || c.Confidence != 0.95 {
				t.Errorf("backticked call claim: provenance %s confidence %v", c.Provenance, c.Confidence)
			}
		case c.Predicate == knowledge.PredicateIdempotent:
			if c.Provenance != knowledge.ProvenanceExtracted || c.Confidence != 0.90 {
				t.Errorf("idempotency claim: provenance %s confidence %v", c.Provenance, c.Confidence)
			}
		default:
			if c.Provenance != knowledge.ProvenanceInferred || c.Confidence != 0.75 {
				t.Errorf("natural-language claim: provenance %s confidence %v", c.Provenance, c.Confidence)
			}
		}
	}
}

func TestParseADR_LineNumbersPointAtSource(t *testing.T) {
	path := writeADR(t, paymentADR)
	claims, err := ParseADR(path, testTenant, "payments")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range claims {
		if c.Object == "DirectBankAPI" && c.LineStart != 11 {
			t.Fatalf("DirectBankAPI rule is on line 11, parser reported %d", c.LineStart)
		}
	}
}

func TestParseADR_MissingFile(t *testing.T) {
	if _, err := ParseADR(filepath.Join(t.TempDir(), "nope.md"), testTenant, "w"); err == nil {
		t.Fatal("expected an error for a missing file")
	}
}

func TestParseADR_NoNormativeSectionsGivesNoClaims(t *testing.T) {
	path := writeADR(t, "# Notes\n\n- The service must use TLS\n\n## Background\n\n- The API must be fast\n")
	claims, err := ParseADR(path, testTenant, "w")
	if err != nil {
		t.Fatal(err)
	}
	if len(claims) != 0 {
		t.Fatalf("expected no claims outside normative sections, got %d", len(claims))
	}
}

func TestParseADRPublic_RewritesDocumentPath(t *testing.T) {
	path := writeADR(t, paymentADR)
	claims, err := ParseADRPublic(path, "docs/adr/0001.docx", testTenant, "payments")
	if err != nil {
		t.Fatal(err)
	}
	if len(claims) == 0 {
		t.Fatal("expected claims")
	}
	for _, c := range claims {
		if c.DocumentPath != "docs/adr/0001.docx" {
			t.Fatalf("document path = %q, want the original source path", c.DocumentPath)
		}
	}
}

func TestDetectModality(t *testing.T) {
	cases := map[string]knowledge.Modality{
		"X must call Y":             knowledge.ModalityMust,
		"X MUST NOT call Y":         knowledge.ModalityMustNot,
		"X shall not call Y":        knowledge.ModalityMustNot,
		"X cannot call Y":           knowledge.ModalityMustNot,
		"X never calls Y":           knowledge.ModalityMustNot,
		"X requires Y":              knowledge.ModalityMust,
		"X has to call Y":           knowledge.ModalityMust,
		"X should call Y":           knowledge.ModalityShould,
		"X is recommended":          knowledge.ModalityShould,
		"X calls Y":                 "",
		"the mustard must be fresh": knowledge.ModalityMust,
		"customers buy mustard":     "",
	}
	for line, want := range cases {
		if got := detectModality(line); got != want {
			t.Errorf("detectModality(%q) = %q, want %q", line, got, want)
		}
	}
}

func TestExtractNaturalLanguageClaim(t *testing.T) {
	cases := []struct {
		line            string
		subj, pred, obj string
		modality        knowledge.Modality
	}{
		{"The HTTP server MUST use the chi router.", "HTTP server", "use", "the chi router", knowledge.ModalityMust},
		{"The router MUST NOT use the default mux.", "router", "use", "the default mux", knowledge.ModalityMustNot},
		{"SQLModel MUST be a class in models", "SQLModel", "is_a", "class", knowledge.ModalityMust},
		{"Ledger MUST be an append-only table", "Ledger", "is_a", "append-only", knowledge.ModalityMust},
		{"Payments must always use idempotency keys", "Payments", "use", "idempotency keys", knowledge.ModalityMust},
		// Word boundaries: "requirement" is not "require", "shallow" is not "shall".
		{"The requirement handler must call the audit log", "requirement handler", "call", "the audit log", knowledge.ModalityMust},
		{"Shallow copies must not leak secrets", "Shallow copies", "leak", "secrets", knowledge.ModalityMustNot},
		// Rejected shapes.
		{"MUST use the chi router", "", "", "", knowledge.ModalityMust},            // no subject
		{"This must happen soon", "", "", "", knowledge.ModalityMust},              // deictic subject
		{"Investors, auditors must verify it", "", "", "", knowledge.ModalityMust}, // clause, not a noun phrase
		{"The service must comply", "", "", "", knowledge.ModalityMust},            // no object
		{"The service must use a", "", "", "", knowledge.ModalityMust},             // stopword object
		{"The service is fine", "", "", "", ""},                                    // no modal
	}
	for _, tc := range cases {
		s, p, o := extractNaturalLanguageClaim(tc.line, tc.modality)
		if s != tc.subj || string(p) != tc.pred || o != tc.obj {
			t.Errorf("extractNaturalLanguageClaim(%q) = (%q, %q, %q), want (%q, %q, %q)", tc.line, s, p, o, tc.subj, tc.pred, tc.obj)
		}
	}
}

func TestIsNormativeSection(t *testing.T) {
	for _, s := range []string{"Decision", "Consequences", "Design notes", "Security Policy", "Invariants", "Behaviour"} {
		if !isNormativeSection(s) {
			t.Errorf("%q should be normative", s)
		}
	}
	for _, s := range []string{"Preamble", "Alternatives considered", "Status", "References"} {
		if isNormativeSection(s) {
			t.Errorf("%q should not be normative", s)
		}
	}
}

func TestStripBulletPrefix(t *testing.T) {
	cases := map[string]string{
		"- rule":   "rule",
		"* rule":   "rule",
		"+ rule":   "rule",
		"• rule":   "rule",
		"-   rule": "rule",
		"rule":     "rule",
		"-rule":    "-rule",
		"- ":       "",
	}
	for in, want := range cases {
		if got := stripBulletPrefix(in); got != want {
			t.Errorf("stripBulletPrefix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestIsValidSubjectAndStopword(t *testing.T) {
	if !isValidSubject("Payment service") || isValidSubject("") || isValidSubject("a, b") ||
		isValidSubject("it") || isValidSubject("one two three four five six seven") {
		t.Fatal("isValidSubject returned an unexpected result")
	}
	if !isStopword(" The ") || isStopword("ledger") {
		t.Fatal("isStopword returned an unexpected result")
	}
}
