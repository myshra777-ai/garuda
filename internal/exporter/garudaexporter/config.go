// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package garudaexporter

import (
	"errors"
	"time"
)

type Config struct {
	Endpoint    string        `mapstructure:"endpoint"`
	TenantID    string        `mapstructure:"tenant_id"`
	WorkspaceID string        `mapstructure:"workspace_id"`
	Timeout     time.Duration `mapstructure:"timeout"`
	MaxBatchSize int          `mapstructure:"max_batch_size"`
}

func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("garuda exporter requires an endpoint URL (e.g., http://localhost:8080/api/v1/telemetry/spans)")
	}
	if cfg.TenantID == "" {
		cfg.TenantID = "00000000-0000-0000-0000-000000000001"
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.MaxBatchSize <= 0 {
		cfg.MaxBatchSize = 1000
	}
	return nil
}
