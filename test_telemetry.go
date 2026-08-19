// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package main

import (
	"context"
	"time"

	"github.com/myshra777-ai/garuda/internal/telemetry"
)

func main() {
	println("Starting telemetry test...")

	cfg := &telemetry.Config{
		Enabled:       true,
		Endpoint:      "https://httpbin.org/post",
		BatchSize:     1,
		FlushInterval: 2 * time.Second,
		MaxQueueSize:  100,
		Mode:          "active",
	}

	println("Initializing telemetry...")
	if err := telemetry.InitTelemetry(cfg); err != nil {
		println("Telemetry init error:", err.Error())
		return
	}
	println("Telemetry initialized successfully")

	println("Recording decision event...")
	telemetry.RecordDecisionProposedWithModel(
		"gpt-4o",
		"openai",
		"canonical",
		"infra",
		"db",
		0.9,
		100,
		0,
		0,
	)
	println("Event recorded, waiting for flush...")

	time.Sleep(5 * time.Second)

	println("Shutting down...")
	telemetry.ShutdownTelemetry(context.Background())
	println("Done!")
}
