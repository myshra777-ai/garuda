package telemetry

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSSEBrokerPublish(t *testing.T) {
	broker := NewSSEBroker()

	req := httptest.NewRequest("GET", "/api/v1/telemetry/stream", nil)
	ctx, cancel := context.WithTimeout(req.Context(), 100*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	go broker.ServeHTTP(w, req)

	// Allow client subscription time to establish
	time.Sleep(20 * time.Millisecond)

	broker.Publish(EventWatchdogAlert, map[string]string{
		"agent_id": "test-agent",
		"reason":   "loop_detected",
	})

	<-ctx.Done()

	body := w.Body.String()
	if len(body) == 0 {
		t.Errorf("expected non-empty SSE stream response")
	}
}
