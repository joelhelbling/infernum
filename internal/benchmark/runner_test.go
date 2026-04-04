package benchmark_test

import (
	"context"
	"testing"
	"time"

	"github.com/joelhelbling/ollama-bench/internal/benchmark"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

// FakeOllamaRunner simulates ollama execution for testing
type FakeOllamaRunner struct {
	CallCount int
}

func (f *FakeOllamaRunner) Run(ctx context.Context, model, prompt string) (models.OllamaStats, error) {
	f.CallCount++
	return models.OllamaStats{
		TotalDurationS:  2.5,
		EvalCount:       100,
		EvalDurationS:   2.0,
		EvalRate:        50.0,
		PromptEvalRate:  200.0,
		PromptEvalCount: 10,
	}, nil
}

func TestRunnerExecutesSuite(t *testing.T) {
	fake := &FakeOllamaRunner{}
	runner := benchmark.NewRunner(fake)

	suite := models.Suite{
		ID:      "suite-1",
		Slug:    "default",
		Version: "1.0.0",
		Prompts: []models.Prompt{
			{ID: "p1", Name: "test1", Text: "Hello"},
			{ID: "p2", Name: "test2", Text: "World"},
		},
		Parameters: models.SuiteParameters{
			RunsPerPrompt:   2,
			TimeoutSeconds:  30,
			CooldownSeconds: 0,
		},
	}

	results, err := runner.Execute(context.Background(), suite, []string{"llama3:8b"}, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// 2 prompts x 2 runs x 1 model = 4 calls
	if fake.CallCount != 4 {
		t.Errorf("expected 4 ollama calls, got %d", fake.CallCount)
	}

	// Should have 1 model result set
	if len(results) != 1 {
		t.Fatalf("expected 1 model result, got %d", len(results))
	}

	if results[0].ModelName != "llama3:8b" {
		t.Errorf("expected model 'llama3:8b', got %q", results[0].ModelName)
	}

	// 2 prompts x 2 runs = 4 prompt results
	if len(results[0].Results) != 4 {
		t.Errorf("expected 4 prompt results, got %d", len(results[0].Results))
	}
}

func TestRunnerMultipleModels(t *testing.T) {
	fake := &FakeOllamaRunner{}
	runner := benchmark.NewRunner(fake)

	suite := models.Suite{
		Prompts: []models.Prompt{
			{ID: "p1", Name: "test1", Text: "Hello"},
		},
		Parameters: models.SuiteParameters{RunsPerPrompt: 1, TimeoutSeconds: 30},
	}

	results, err := runner.Execute(context.Background(), suite, []string{"model-a", "model-b"}, nil)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 model results, got %d", len(results))
	}
	if fake.CallCount != 2 {
		t.Errorf("expected 2 calls, got %d", fake.CallCount)
	}
}

func TestRunnerCallsProgressCallback(t *testing.T) {
	fake := &FakeOllamaRunner{}
	runner := benchmark.NewRunner(fake)

	suite := models.Suite{
		Prompts:    []models.Prompt{{ID: "p1", Name: "test1", Text: "Hello"}},
		Parameters: models.SuiteParameters{RunsPerPrompt: 1, TimeoutSeconds: 30},
	}

	var progressCalls []benchmark.Progress
	cb := func(p benchmark.Progress) {
		progressCalls = append(progressCalls, p)
	}

	runner.Execute(context.Background(), suite, []string{"llama3:8b"}, cb)

	if len(progressCalls) == 0 {
		t.Error("expected progress callbacks")
	}

	last := progressCalls[len(progressCalls)-1]
	if last.Completed != last.Total {
		t.Errorf("final progress should show completion: %d/%d", last.Completed, last.Total)
	}
}

// Verify cooldown is respected (it should add delay)
func TestRunnerCooldown(t *testing.T) {
	fake := &FakeOllamaRunner{}
	runner := benchmark.NewRunner(fake)

	suite := models.Suite{
		Prompts: []models.Prompt{
			{ID: "p1", Name: "test1", Text: "Hello"},
		},
		Parameters: models.SuiteParameters{
			RunsPerPrompt:   2,
			TimeoutSeconds:  30,
			CooldownSeconds: 1,
		},
	}

	start := time.Now()
	runner.Execute(context.Background(), suite, []string{"llama3:8b"}, nil)
	elapsed := time.Since(start)

	// With 2 runs and 1s cooldown between them, should take at least 1s
	if elapsed < 900*time.Millisecond {
		t.Errorf("expected at least ~1s for cooldown, took %v", elapsed)
	}
}
