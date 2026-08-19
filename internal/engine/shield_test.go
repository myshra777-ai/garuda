// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package engine

import (
	"strings"
	"testing"
)

func TestShieldRedaction(t *testing.T) {
	shield := NewShieldEngine()

	tests := []struct {
		name          string
		input         string
		expectedCount int
		shouldContain string
	}{
		{
			name:          "API Key Scrub",
			input:         "Deploying with sk-12345678901234567890123456789012",
			expectedCount: 1,
			shouldContain: "[REDACTED_SECRET]",
		},
		{
			name:          "GitHub Token Scrub",
			input:         "Auth using ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
			expectedCount: 1,
			shouldContain: "[REDACTED_SECRET]",
		},
		{
			name:          "Clean Prompt",
			input:         "Execute standard database migration query",
			expectedCount: 0,
			shouldContain: "standard database migration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clean, count := shield.Redact(tt.input)
			if count != tt.expectedCount {
				t.Errorf("expected %d redactions, got %d", tt.expectedCount, count)
			}
			if count > 0 && !strings.Contains(clean, tt.shouldContain) {
				t.Errorf("expected output to contain %s, got %s", tt.shouldContain, clean)
			}
		})
	}
}
