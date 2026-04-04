package models_test

import (
	"encoding/json"
	"testing"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestPromptResultJSON(t *testing.T) {
	pr := models.PromptResult{
		PromptID:  "prompt-1",
		RunNumber: 1,
		Stats: models.OllamaStats{
			TotalDurationS:      5.23,
			LoadDurationS:       0.15,
			PromptEvalCount:     12,
			PromptEvalDurationS: 0.45,
			PromptEvalRate:      26.67,
			EvalCount:           150,
			EvalDurationS:       4.63,
			EvalRate:            32.4,
		},
		Success:      true,
		ErrorMessage: "",
	}

	data, err := json.Marshal(pr)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded models.PromptResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Stats.EvalRate != 32.4 {
		t.Errorf("expected eval rate 32.4, got %f", decoded.Stats.EvalRate)
	}
	if !decoded.Success {
		t.Error("expected success to be true")
	}
}

func TestPublishRequestJSON(t *testing.T) {
	req := models.PublishRequest{
		SuiteID:      "suite-1",
		SuiteVersion: "1.0.0",
		ModelName:    "llama3:8b",
		Hardware: models.HardwareInfo{
			OSName:   "linux",
			CPUModel: "AMD Ryzen 9 7950X",
			CPUCores: 32,
			RAMGB:    64,
			GPUName:  "NVIDIA RTX 4090",
			VRAMGB:   24,
		},
		Results: []models.PromptResult{
			{
				PromptID:  "prompt-1",
				RunNumber: 1,
				Stats: models.OllamaStats{
					EvalRate: 32.4,
				},
				Success: true,
			},
		},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded models.PublishRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.ModelName != "llama3:8b" {
		t.Errorf("expected model 'llama3:8b', got %q", decoded.ModelName)
	}
}
