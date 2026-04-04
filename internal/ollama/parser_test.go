package ollama_test

import (
	"math"
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/ollama"
)

func TestParseVerboseOutput(t *testing.T) {
	output := `The answer is 4.

total duration:       5.227891234s
load duration:        152.345ms
prompt eval count:    12 token(s)
prompt eval duration: 450.123ms
prompt eval rate:     26.66 tokens/s
eval count:           150 token(s)
eval duration:        4.625423891s
eval rate:            32.43 tokens/s`

	stats, err := ollama.ParseVerboseOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "total_duration", stats.TotalDurationS, 5.227891, 0.001)
	assertFloat(t, "load_duration", stats.LoadDurationS, 0.152345, 0.001)
	if stats.PromptEvalCount != 12 {
		t.Errorf("expected prompt_eval_count 12, got %d", stats.PromptEvalCount)
	}
	assertFloat(t, "prompt_eval_duration", stats.PromptEvalDurationS, 0.450123, 0.001)
	assertFloat(t, "prompt_eval_rate", stats.PromptEvalRate, 26.66, 0.01)
	if stats.EvalCount != 150 {
		t.Errorf("expected eval_count 150, got %d", stats.EvalCount)
	}
	assertFloat(t, "eval_duration", stats.EvalDurationS, 4.625423, 0.001)
	assertFloat(t, "eval_rate", stats.EvalRate, 32.43, 0.01)
}

func TestParseVerboseOutputMinutesDuration(t *testing.T) {
	output := `Response text here.

total duration:       1m12.423928897s
load duration:        56.38ms
prompt eval count:    364 token(s)
prompt eval duration: 2.345s
prompt eval rate:     155.22 tokens/s
eval count:           500 token(s)
eval duration:        1m9.123s
eval rate:            7.23 tokens/s`

	stats, err := ollama.ParseVerboseOutput(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertFloat(t, "total_duration", stats.TotalDurationS, 72.423928, 0.001)
	assertFloat(t, "load_duration", stats.LoadDurationS, 0.05638, 0.001)
	assertFloat(t, "eval_duration", stats.EvalDurationS, 69.123, 0.001)
}

func TestParseVerboseOutputMissingField(t *testing.T) {
	output := `Response text here.

total duration:       5.0s`

	_, err := ollama.ParseVerboseOutput(output)
	if err == nil {
		t.Fatal("expected error for incomplete output")
	}
}

func assertFloat(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Errorf("%s: expected %f, got %f (tolerance %f)", name, want, got, tolerance)
	}
}
