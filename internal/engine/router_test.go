package engine

import (
	"context"
	"math/rand"
	"testing"
)

func TestClassifyAndRouteMarksRiskyPayloadsAsShadowExecuted(t *testing.T) {
	router := NewDynamicRouter()
	router.rng = rand.New(rand.NewSource(1))

	decision, err := router.ClassifyAndRoute(context.Background(), "Please drop the production table", 50, 0.9)
	if err != nil {
		t.Fatalf("ClassifyAndRoute returned error: %v", err)
	}

	if !decision.ConsensusRequired {
		t.Fatalf("expected consensus requirement for destructive payload")
	}

	if !decision.ShadowExecuted {
		t.Fatalf("expected risky payload to trigger shadow execution")
	}
}
