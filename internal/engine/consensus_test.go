package engine

import (
	"context"
	"testing"
)

func TestConsensusEvaluation(t *testing.T) {
	engine := NewConsensusEngine()
	ctx := context.Background()

	models := []string{"claude-3-5-sonnet", "deepseek-v3", "gpt-4o"}
	result, err := engine.EvaluateConsensus(ctx, "DROP TABLE test;", models)

	if err != nil {
		t.Fatalf("unexpected error during consensus evaluation: %v", err)
	}

	if !result.Matches {
		t.Errorf("expected models to reach consensus match")
	}

	if len(result.Participating) != 3 {
		t.Errorf("expected 3 participating models, got %d", len(result.Participating))
	}
}
