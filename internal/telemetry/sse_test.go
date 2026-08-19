// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package telemetry

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type safeResponseWriter struct {
	*httptest.ResponseRecorder
	mu sync.Mutex
}

func newSafeResponseWriter() *safeResponseWriter {
	return &safeResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
	}
}

func (w *safeResponseWriter) Write(buf []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ResponseRecorder.Write(buf)
}

func (w *safeResponseWriter) BodyString() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.ResponseRecorder.Body.String()
}

func TestSSEBrokerPublish(t *testing.T) {
	broker := NewSSEBroker()
	w := newSafeResponseWriter()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "/telemetry/events", nil)

	handlerDone := make(chan struct{})
	go func() {
		defer close(handlerDone)
		broker.ServeHTTP(w, req)
	}()

	// Allow client goroutine time to register subscriber
	time.Sleep(50 * time.Millisecond)

	// Publish with EventType and payload interface{}
	broker.Publish(EventType("TEST_EVENT"), map[string]interface{}{"status": "ok"})

	time.Sleep(50 * time.Millisecond)
	cancel() // Disconnect client to gracefully close stream
	<-handlerDone

	output := w.BodyString()
	if !bytes.Contains([]byte(output), []byte("TEST_EVENT")) {
		t.Fatalf("expected SSE event payload in response, got: %s", output)
	}
}
