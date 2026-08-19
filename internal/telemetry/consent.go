// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const consentFile = ".garuda/consent.json"

type Consent struct {
	TelemetryConsent bool      `json:"telemetry_consent"`
	ConsentGrantedAt time.Time `json:"consent_granted_at"`
	ConsentVersion   string    `json:"consent_version"`
}

func parseBoolEnv(value string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "1", "yes", "y":
		return true, true
	case "false", "0", "no", "n":
		return false, true
	default:
		return false, false
	}
}

// IsConsentGiven returns true if user has explicitly consented.
func IsConsentGiven() bool {
	if enabledValue, ok := parseBoolEnv(os.Getenv("GARUDA_TELEMETRY_ENABLED")); ok && !enabledValue {
		return false
	}

	if consentValue, ok := parseBoolEnv(os.Getenv("GARUDA_TELEMETRY_CONSENT")); ok {
		return consentValue
	}

	if consentFileEnv := os.Getenv("GARUDA_TELEMETRY_CONSENT_FILE"); consentFileEnv != "" {
		if consentValue, ok := parseBoolEnv(consentFileEnv); ok {
			return consentValue
		}
		if data, err := os.ReadFile(consentFileEnv); err == nil {
			if consentValue, ok := parseBoolEnv(strings.TrimSpace(string(data))); ok {
				return consentValue
			}
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	path := filepath.Join(home, consentFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var c Consent
	if err := json.Unmarshal(data, &c); err != nil {
		return false
	}
	return c.TelemetryConsent
}

// SaveConsent writes the consent decision to disk.
func SaveConsent(consent bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".garuda")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	c := Consent{
		TelemetryConsent: consent,
		ConsentGrantedAt: time.Now().UTC(),
		ConsentVersion:   "1.0",
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "consent.json"), data, 0644)
}

// ShowConsentPrompt displays the consent message (called on first run).
func ShowConsentPrompt() bool {
	// In a real CLI, this would be interactive.
	// For now, we simulate with environment variable.
	if os.Getenv("GARUDA_TELEMETRY_CONSENT") == "true" {
		_ = SaveConsent(true)
		return true
	}
	// If no consent stored, default to false (opt‑out by default).
	return false
}
