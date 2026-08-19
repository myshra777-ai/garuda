// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package pool

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// KeySlot represents a single API key with its provider and status.
type KeySlot struct {
	ID        string           `json:"id"`
	Provider  string           `json:"provider"`
	APIKey    string           `json:"-"` // never serialized
	Role      string           `json:"role"`
	Priority  int              `json:"priority"`
	Status    string           `json:"status"` // active, throttled, dead
	LastCheck time.Time        `json:"last_check"`
	LatencyMs int64            `json:"latency_ms"`
	RateLimit RateLimitTracker `json:"rate_limit"`
}

// RateLimitTracker tracks usage against quotas.
type RateLimitTracker struct {
	RequestsPerMinute int           `json:"requests_per_minute"`
	TokensPerMinute   int64         `json:"tokens_per_minute"`
	LastReset         time.Time     `json:"last_reset"`
	ResetInterval     time.Duration `json:"reset_interval"`
}

// KeyPool manages all API keys with failover.
type KeyPool struct {
	mu        sync.RWMutex
	Slots     []*KeySlot `json:"slots"`
	Active    *KeySlot   `json:"active"`
	clients   map[string]ProviderClient
	healthCtx context.Context
	cancel    context.CancelFunc
}

// NewKeyPool creates a new key pool.
func NewKeyPool() *KeyPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &KeyPool{
		Slots:     []*KeySlot{},
		clients:   make(map[string]ProviderClient),
		healthCtx: ctx,
		cancel:    cancel,
	}
}

// AddSlot adds a new key slot to the pool.
func (p *KeyPool) AddSlot(provider, apiKey, role string, priority int) *KeySlot {
	p.mu.Lock()
	defer p.mu.Unlock()

	slot := &KeySlot{
		ID:       fmt.Sprintf("key-%d", len(p.Slots)+1),
		Provider: provider,
		APIKey:   apiKey,
		Role:     role,
		Priority: priority,
		Status:   "active",
		RateLimit: RateLimitTracker{
			RequestsPerMinute: 60,
			TokensPerMinute:   1000000,
			LastReset:         time.Now().UTC(),
			ResetInterval:     60 * time.Second,
		},
	}
	p.Slots = append(p.Slots, slot)

	// Cache the client
	client, _ := NewProviderClient(provider)
	if client != nil {
		p.clients[provider] = client
	}

	return slot
}

// LoadFromEnv loads key slots from environment variables.
func (p *KeyPool) LoadFromEnv() {
	providers := []string{"GEMINI", "DEEPSEEK", "OPENAI"}
	roles := []string{"PRIMARY", "SECONDARY", "TERTIARY", "QUATERNARY"}

	for _, provider := range providers {
		for _, role := range roles {
			envKey := fmt.Sprintf("%s_API_KEY_%s", provider, role)
			key := os.Getenv(envKey)
			if key != "" {
				p.AddSlot(
					provider,
					key,
					role,
					getPriority(role),
				)
			}
		}
	}
}

func getPriority(role string) int {
	switch role {
	case "PRIMARY":
		return 1
	case "SECONDARY":
		return 2
	case "TERTIARY":
		return 3
	case "QUATERNARY":
		return 4
	default:
		return 5
	}
}

// HealthCheckAll runs health checks on all slots.
func (p *KeyPool) HealthCheckAll(ctx context.Context) {
	p.mu.RLock()
	slots := p.Slots
	p.mu.RUnlock()

	for _, slot := range slots {
		p.healthCheckSlot(ctx, slot)
	}
}

// healthCheckSlot checks a single slot.
func (p *KeyPool) healthCheckSlot(ctx context.Context, slot *KeySlot) {
	client, ok := p.clients[slot.Provider]
	if !ok {
		return
	}
	result, err := client.HealthCheck(ctx, slot.APIKey)
	if err != nil {
		slot.Status = "dead"
		slot.LastCheck = time.Now().UTC()
		return
	}
	if !result.Valid {
		slot.Status = "dead"
	} else if result.Error == "rate limited" {
		slot.Status = "throttled"
	} else {
		slot.Status = "active"
	}
	slot.LatencyMs = result.LatencyMs
	slot.LastCheck = time.Now().UTC()
	if result.Quota != nil {
		slot.RateLimit.RequestsPerMinute = result.Quota.RequestsPerMinute
	}
}

// GetActiveSlot returns the highest-priority active slot.
func (p *KeyPool) GetActiveSlot() *KeySlot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, slot := range p.Slots {
		if slot.Status == "active" {
			return slot
		}
	}
	return nil
}

// RotateToNextActive switches to the next available active slot.
func (p *KeyPool) RotateToNextActive() *KeySlot {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, slot := range p.Slots {
		if slot.Status == "active" && slot != p.Active {
			p.Active = slot
			return slot
		}
	}
	return nil
}

// GetStatus returns a summary of all slots.
func (p *KeyPool) GetStatus() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	slots := []map[string]interface{}{}
	activeCount := 0
	throttledCount := 0
	deadCount := 0

	for _, s := range p.Slots {
		status := map[string]interface{}{
			"id":         s.ID,
			"provider":   s.Provider,
			"role":       s.Role,
			"status":     s.Status,
			"latency_ms": s.LatencyMs,
			"last_check": s.LastCheck,
		}
		slots = append(slots, status)
		switch s.Status {
		case "active":
			activeCount++
		case "throttled":
			throttledCount++
		case "dead":
			deadCount++
		}
	}

	return map[string]interface{}{
		"slots":           slots,
		"active_count":    activeCount,
		"throttled_count": throttledCount,
		"dead_count":      deadCount,
		"active_slot":     p.Active,
	}
}

// Close shuts down the key pool.
func (p *KeyPool) Close() {
	if p.cancel != nil {
		p.cancel()
	}
}
