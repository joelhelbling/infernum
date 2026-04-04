package ollama_test

import (
	"context"
	"testing"
	"time"

	"github.com/joelhelbling/ollama-bench/internal/ollama"
)

func TestRunReturnsStats(t *testing.T) {
	// This test requires ollama to be running locally.
	// Skip in CI — run manually or in integration test suite.
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := ollama.Run(ctx, "llama3.2:1b", "What is 2 + 2?")
	if err != nil {
		t.Fatalf("ollama.Run failed: %v", err)
	}

	if result.Stats.EvalRate <= 0 {
		t.Errorf("expected positive eval rate, got %f", result.Stats.EvalRate)
	}
	if result.Stats.EvalCount <= 0 {
		t.Errorf("expected positive eval count, got %d", result.Stats.EvalCount)
	}
	if result.Output == "" {
		t.Error("expected non-empty output")
	}
}

func TestRunTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	_, err := ollama.Run(ctx, "llama3.2:1b", "Write a very long essay about the history of computing")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
