// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package garudaexporter

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"
)

const (
	TypeStr   = "garuda"
	Stability = component.StabilityLevelAlpha
)

func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		component.MustNewType(TypeStr),
		createDefaultConfig,
		exporter.WithTraces(createTracesExporter, Stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Endpoint:     "http://localhost:8080/api/v1/telemetry/spans",
		TenantID:     "00000000-0000-0000-0000-000000000001",
		WorkspaceID:  "532a8e33-975d-48a3-8f88-221cef52fec4",
		Timeout:      5 * time.Second,
		MaxBatchSize: 1000,
	}
}

func createTracesExporter(
	ctx context.Context,
	set exporter.Settings,
	cfg component.Config,
) (exporter.Traces, error) {
	gCfg := cfg.(*Config)
	if err := gCfg.Validate(); err != nil {
		return nil, err
	}

	exp := newGarudaExporter(gCfg, set.Logger)
	return exporterhelper.NewTracesExporter(
		ctx,
		set,
		cfg,
		exp.pushTraces,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithTimeout(exporterhelper.TimeoutConfig{Timeout: gCfg.Timeout}),
		exporterhelper.WithRetry(exporterhelper.NewDefaultRetryConfig()),
	)
}
