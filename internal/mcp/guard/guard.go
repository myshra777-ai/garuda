// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package guard

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

var (
	ErrSessionExpired     = errors.New("GARUDA_GUARD_001: Session duration exceeded configured limit")
	ErrTokenBudgetLimit   = errors.New("GARUDA_GUARD_002: Token consumption quota exceeded")
	ErrWriteBlockedOnProd = errors.New("GARUDA_GUARD_003: Write/mutation tool call rejected on production-tagged workspace")
)

type WorkspaceMode string

const (
	ModeDevelopment WorkspaceMode = "development"
	ModeProduction  WorkspaceMode = "production"
)

type Config struct {
	MaxSessionDuration time.Duration `yaml:"max_session_duration"`
	MaxTokens          int64         `yaml:"max_tokens"`
	Mode               WorkspaceMode `yaml:"mode"`
}

type SessionGuard struct {
	cfg        Config
	startedAt  time.Time
	tokensUsed atomic.Int64
}

func NewSessionGuard(cfg Config) *SessionGuard {
	if cfg.MaxSessionDuration == 0 {
		cfg.MaxSessionDuration = 60 * time.Minute
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 200000
	}
	if cfg.Mode == "" {
		cfg.Mode = ModeDevelopment
	}

	return &SessionGuard{
		cfg:       cfg,
		startedAt: time.Now(),
	}
}

// Intercept validates session lifecycle, budget, and environmental policy before tool execution.
func (g *SessionGuard) Intercept(ctx context.Context, toolName string, isMutating bool, estimatedTokens int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// 1. Session duration check
	if time.Since(g.startedAt) > g.cfg.MaxSessionDuration {
		return fmt.Errorf("%w (limit: %v)", ErrSessionExpired, g.cfg.MaxSessionDuration)
	}

	// 2. Token quota check
	current := g.tokensUsed.Add(estimatedTokens)
	if current > g.cfg.MaxTokens {
		return fmt.Errorf("%w (consumed: %d, limit: %d)", ErrTokenBudgetLimit, current, g.cfg.MaxTokens)
	}

	// 3. Environmental mutation block
	if g.cfg.Mode == ModeProduction && isMutating {
		return fmt.Errorf("%w (tool: %s)", ErrWriteBlockedOnProd, toolName)
	}

	return nil
}

func (g *SessionGuard) RecordTokens(consumed int64) {
	g.tokensUsed.Add(consumed)
}

func (g *SessionGuard) Usage() (used int64, limit int64, uptime time.Duration) {
	return g.tokensUsed.Load(), g.cfg.MaxTokens, time.Since(g.startedAt)
}
