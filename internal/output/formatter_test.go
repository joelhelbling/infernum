package output_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/output"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestTableFormatterCompare(t *testing.T) {
	var buf bytes.Buffer
	f := output.NewTableFormatter(&buf)

	resp := models.CompareResponse{
		Groups: []models.CompareGroup{
			{
				Hardware: &models.HardwareInfo{GPUName: "RTX 4090", CPUModel: "i9-14900K", RAMGB: 64},
				Stats:    models.SummaryStats{MeanEvalRate: 85.2, StddevEvalRate: 3.1, MeanPromptEvalRate: 320.5},
				RunCount: 5,
			},
			{
				Hardware: &models.HardwareInfo{GPUName: "RTX 3090", CPUModel: "i7-12700K", RAMGB: 32},
				Stats:    models.SummaryStats{MeanEvalRate: 42.1, StddevEvalRate: 2.0, MeanPromptEvalRate: 180.3},
				RunCount: 3,
			},
		},
	}

	if err := f.FormatComparison(resp); err != nil {
		t.Fatalf("FormatComparison failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "RTX 4090") {
		t.Error("output should contain GPU name 'RTX 4090'")
	}
	if !strings.Contains(out, "85.2") {
		t.Error("output should contain eval rate '85.2'")
	}
}

func TestJSONFormatterCompare(t *testing.T) {
	var buf bytes.Buffer
	f := output.NewJSONFormatter(&buf)

	resp := models.CompareResponse{
		Groups: []models.CompareGroup{
			{
				Hardware: &models.HardwareInfo{GPUName: "RTX 4090"},
				Stats:    models.SummaryStats{MeanEvalRate: 85.2},
				RunCount: 5,
			},
		},
	}

	if err := f.FormatComparison(resp); err != nil {
		t.Fatalf("FormatComparison failed: %v", err)
	}

	// Verify it's valid JSON
	var decoded models.CompareResponse
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if decoded.Groups[0].Stats.MeanEvalRate != 85.2 {
		t.Errorf("expected eval rate 85.2, got %f", decoded.Groups[0].Stats.MeanEvalRate)
	}
}

func TestTableFormatterRun(t *testing.T) {
	var buf bytes.Buffer
	f := output.NewTableFormatter(&buf)

	run := models.RunResponse{
		ID:        "run-123",
		ModelName: "llama3:8b",
		Hardware:  models.HardwareInfo{GPUName: "RTX 4090", RAMGB: 64},
		Results: []models.PromptResult{
			{
				PromptID: "p1", RunNumber: 1,
				Stats:   models.OllamaStats{EvalRate: 32.4, PromptEvalRate: 155.0},
				Success: true,
			},
		},
	}

	if err := f.FormatRun(run); err != nil {
		t.Fatalf("FormatRun failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "llama3:8b") {
		t.Error("output should contain model name")
	}
}
