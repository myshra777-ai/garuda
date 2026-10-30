// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDashboardTalonHubUnavailableWithoutPostgres(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/talon-hub?workspace=default", nil)
	rec := httptest.NewRecorder()

	s.HandleDashboardTalonHub(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
	}
	if strings.Contains(strings.ToUpper(rec.Body.String()), "ALLOW") {
		t.Fatalf("unavailable hub must not contain a fabricated ALLOW: %s", rec.Body.String())
	}
}

func TestDashboardTalonCheckUnavailableWithoutPostgres(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/dashboard/talon-check?workspace=default",
		strings.NewReader(`{"workspace":"default","action":"run_guard"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.HandleDashboardTalonCheck(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusServiceUnavailable, rec.Body.String())
	}
	if strings.Contains(strings.ToUpper(rec.Body.String()), "ALLOW") {
		t.Fatalf("unavailable check must not contain a fabricated ALLOW: %s", rec.Body.String())
	}
}

func TestDashboardTalonCheckRejectsMalformedJSONBeforeStoreLookup(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboard/talon-check", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.HandleDashboardTalonCheck(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestDashboardTalonCheckRejectsUnsupportedAction(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/dashboard/talon-check",
		strings.NewReader(`{"action":"pretend_clean"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	s.HandleDashboardTalonCheck(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
