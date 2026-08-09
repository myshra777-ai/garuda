package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/myshra777-ai/garuda/internal/types"
)

type SessionActivity struct {
	AgentID         string
	TenantID        uuid.UUID
	LastActive      time.Time
	ToolCallHistory []string
	ActiveBudget    float64
}

type WatchdogEngine struct {
	cpStore     types.CheckpointRepository
	mu          sync.RWMutex
	sessions    map[string]*SessionActivity
	idleTimeout time.Duration
}

func NewWatchdogEngine(cpStore types.CheckpointRepository) *WatchdogEngine {
	return &WatchdogEngine{
		cpStore:     cpStore,
		sessions:    make(map[string]*SessionActivity),
		idleTimeout: 60 * time.Minute,
	}
}

// RecordActivity tracks tool execution history and triggers Dead-Man's Switch on runaway loops
func (w *WatchdogEngine) RecordActivity(ctx context.Context, tenantID uuid.UUID, agentID string, toolName string, currentMerkleRoot string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	session, exists := w.sessions[agentID]
	if !exists {
		session = &SessionActivity{
			AgentID:    agentID,
			TenantID:   tenantID,
			LastActive: time.Now().UTC(),
		}
		w.sessions[agentID] = session
	}

	session.LastActive = time.Now().UTC()
	session.ToolCallHistory = append(session.ToolCallHistory, toolName)

	// Dead-Man's Switch: Detect 10 consecutive identical tool execution calls
	if len(session.ToolCallHistory) >= 10 {
		n := len(session.ToolCallHistory)
		last10 := session.ToolCallHistory[n-10:]
		allSame := true
		for _, t := range last10 {
			if t != toolName {
				allSame = false
				break
			}
		}

		if allSame {
			slog.Warn("Dead-Man's Switch Triggered: Infinite tool loop detected", "agent_id", agentID, "tool", toolName)
			stateData, _ := json.Marshal(map[string]interface{}{
				"halt_reason": "infinite_tool_loop",
				"tool":        toolName,
				"history":     last10,
			})

			if w.cpStore != nil {
				_, err := w.cpStore.CreateCheckpoint(ctx, tenantID, agentID, "chk_deadman_triggered", "Infinite tool execution loop detected", stateData, currentMerkleRoot)
				if err != nil {
					return fmt.Errorf("failed to save deadman checkpoint: %w", err)
				}
			}
			return fmt.Errorf("DEADMAN_SWITCH_TRIGGERED: Infinite tool execution loop halted")
		}
	}

	return nil
}

// StartIdleWatchdog runs a background goroutine monitoring 60-minute inactivity
func (w *WatchdogEngine) StartIdleWatchdog(ctx context.Context, currentMerkleRoot string) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				w.checkIdleSessions(ctx, currentMerkleRoot)
			}
		}
	}()
}

func (w *WatchdogEngine) checkIdleSessions(ctx context.Context, currentMerkleRoot string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now().UTC()
	for agentID, session := range w.sessions {
		if now.Sub(session.LastActive) >= w.idleTimeout {
			slog.Info("60-Minute Idle Watchdog Triggered: Snapshotting state and flushing enclave", "agent_id", agentID)
			stateData, _ := json.Marshal(map[string]interface{}{
				"status":      "idle_suspended",
				"last_active": session.LastActive,
			})

			if w.cpStore != nil {
				_, _ = w.cpStore.CreateCheckpoint(ctx, session.TenantID, agentID, "chk_auto_idle_60m", "Agent inactive for 60 consecutive minutes", stateData, currentMerkleRoot)
			}
			delete(w.sessions, agentID)
		}
	}
}
