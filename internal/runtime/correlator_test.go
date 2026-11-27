// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package runtime

import "testing"

func TestSpanFunctionName(t *testing.T) {
	cases := []struct {
		name      string
		obs       RuntimeObservation
		wantName  string
		wantSpace string
	}{
		{
			name:     "code.function with package",
			obs:      RuntimeObservation{Attributes: map[string]interface{}{"code.function": "payments.Charge"}},
			wantName: "Charge",
		},
		{
			name:      "pointer receiver method with namespace",
			obs:       RuntimeObservation{Attributes: map[string]interface{}{"code.function": "payments.(*Handler).Charge", "code.namespace": " github.com/acme/payments "}},
			wantName:  "Charge",
			wantSpace: "github.com/acme/payments",
		},
		{
			name:     "bare pointer type",
			obs:      RuntimeObservation{Attributes: map[string]interface{}{"code.function": "*Handler"}},
			wantName: "Handler",
		},
		{
			name:     "falls back to the operation",
			obs:      RuntimeObservation{Operation: "billing.Invoice", Attributes: map[string]interface{}{}},
			wantName: "Invoice",
		},
		{
			name:     "empty code.function falls back to the operation",
			obs:      RuntimeObservation{Operation: "Refund", Attributes: map[string]interface{}{"code.function": ""}},
			wantName: "Refund",
		},
		{
			name:     "nil attributes",
			obs:      RuntimeObservation{Operation: "Capture"},
			wantName: "Capture",
		},
		{
			name:     "non-string attribute is ignored",
			obs:      RuntimeObservation{Operation: "Capture", Attributes: map[string]interface{}{"code.function": 42}},
			wantName: "Capture",
		},
		{
			name:     "nothing to correlate",
			obs:      RuntimeObservation{},
			wantName: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotSpace := spanFunctionName(&tc.obs)
			if gotName != tc.wantName || gotSpace != tc.wantSpace {
				t.Fatalf("spanFunctionName = (%q, %q), want (%q, %q)", gotName, gotSpace, tc.wantName, tc.wantSpace)
			}
		})
	}
}
