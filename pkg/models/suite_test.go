package models_test

import (
	"encoding/json"
	"testing"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestSuiteJSON(t *testing.T) {
	suite := models.Suite{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Slug:        "default",
		Name:        "Default Benchmark Suite",
		Description: "Standard benchmark suite for Ollama models",
		Version:     "1.0.0",
		Prompts: []models.Prompt{
			{
				ID:                    "prompt-1",
				Name:                  "simple_math",
				Category:              "reasoning",
				Text:                  "What is 2 + 2?",
				ExpectedTokenRangeMin: intPtr(1),
				ExpectedTokenRangeMax: intPtr(50),
				SortOrder:             1,
			},
		},
		Parameters: models.SuiteParameters{
			RunsPerPrompt:   3,
			TimeoutSeconds:  300,
			CooldownSeconds: 5,
		},
	}

	data, err := json.Marshal(suite)
	if err != nil {
		t.Fatalf("failed to marshal suite: %v", err)
	}

	var decoded models.Suite
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal suite: %v", err)
	}

	if decoded.Slug != "default" {
		t.Errorf("expected slug 'default', got %q", decoded.Slug)
	}
	if len(decoded.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(decoded.Prompts))
	}
	if decoded.Prompts[0].Name != "simple_math" {
		t.Errorf("expected prompt name 'simple_math', got %q", decoded.Prompts[0].Name)
	}
	if decoded.Parameters.RunsPerPrompt != 3 {
		t.Errorf("expected 3 runs per prompt, got %d", decoded.Parameters.RunsPerPrompt)
	}
}

func intPtr(i int) *int {
	return &i
}
