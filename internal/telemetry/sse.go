// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package telemetry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type EventType string

const (
	EventDecisionCreated    EventType = "decision_created"
	EventDecisionQuarantine EventType = "decision_quarantined"
	EventCheckpointCreated  EventType = "checkpoint_created"
	EventWatchdogAlert      EventType = "watchdog_alert"
	EventRouterDispatch     EventType = "router_dispatch"
)

type SSEEvent struct {
	Type      EventType   `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

type SSEBroker struct {
	clients   map[chan SSEEvent]bool
	newClient chan chan SSEEvent
	defunct   chan chan SSEEvent
	events    chan SSEEvent
	mu        sync.RWMutex
}

func NewSSEBroker() *SSEBroker {
	broker := &SSEBroker{
		clients:   make(map[chan SSEEvent]bool),
		newClient: make(chan chan SSEEvent),
		defunct:   make(chan chan SSEEvent),
		events:    make(chan SSEEvent, 100),
	}
	go broker.start()
	return broker
}

func (b *SSEBroker) start() {
	for {
		select {
		case s := <-b.newClient:
			b.mu.Lock()
			b.clients[s] = true
			b.mu.Unlock()
		case s := <-b.defunct:
			b.mu.Lock()
			if _, ok := b.clients[s]; ok {
				delete(b.clients, s)
				close(s)
			}
			b.mu.Unlock()
		case event := <-b.events:
			b.mu.RLock()
			for client := range b.clients {
				select {
				case client <- event:
				default:
					// Drop event if client buffer is blocked
				}
			}
			b.mu.RUnlock()
		}
	}
}

func (b *SSEBroker) Publish(eventType EventType, payload interface{}) {
	event := SSEEvent{
		Type:      eventType,
		Payload:   payload,
		Timestamp: time.Now().UTC(),
	}
	b.events <- event
}

func (b *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	clientChan := make(chan SSEEvent, 10)
	b.newClient <- clientChan

	defer func() {
		b.defunct <- clientChan
	}()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case event := <-clientChan:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
			flusher.Flush()
		}
	}
}
