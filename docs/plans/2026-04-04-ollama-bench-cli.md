# Ollama Bench CLI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a Go CLI tool that runs Ollama benchmarks, publishes results to a web service, and queries comparison reports.

**Architecture:** Cobra-based CLI with shared models in `pkg/` (importable by the web backend), internal logic in `internal/`. The CLI fetches benchmark suites from the backend API, runs them locally against Ollama, publishes results, and can query/display comparison reports. Outputs in ASCII table or JSON format.

**Tech Stack:** Go 1.22+, Cobra (CLI framework), `pgx`-compatible types in shared models, `crypto/sha256` for hardware fingerprinting, `tablewriter` for ASCII tables, standard `net/http` for API client.

**Spec:** `docs/superpowers/specs/2026-04-04-ollama-bench-platform-design.md`

---

## File Structure

```
ollama-bench/
├── cmd/
│   └── ollama-bench/
│       └── main.go                  # Entry point, root cobra command
├── internal/
│   ├── cli/
│   │   ├── root.go                  # Root command setup
│   │   ├── run.go                   # `run` command
│   │   ├── compare.go              # `compare` command
│   │   ├── results.go              # `results` command
│   │   ├── suites.go               # `suites` and `suite` commands
│   │   └── publish.go             # `publish` command
│   ├── benchmark/
│   │   └── runner.go               # Orchestrates benchmark execution
│   ├── ollama/
│   │   ├── runner.go               # Executes ollama subprocess
│   │   └── parser.go              # Parses ollama verbose output
│   ├── hardware/
│   │   └── detect.go              # Cross-platform hardware detection
│   ├── cache/
│   │   └── suite.go               # Suite caching logic
│   ├── config/
│   │   └── config.go              # Config file management
│   ├── output/
│   │   ├── formatter.go           # Formatter interface
│   │   ├── table.go               # ASCII table output
│   │   └── json.go                # JSON output
│   └── pending/
│       └── store.go               # Pending results storage for offline
├── pkg/
│   ├── models/
│   │   ├── suite.go               # Suite, Prompt types
│   │   ├── hardware.go            # HardwareInfo, fingerprint
│   │   ├── results.go             # RunResult, PromptResult, OllamaStats
│   │   └── api.go                 # API request/response types
│   └── apiclient/
│       └── client.go              # HTTP client for backend API
├── go.mod
├── go.sum
└── README.md
```

---

### Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `cmd/ollama-bench/main.go`
- Create: `internal/cli/root.go`

- [ ] **Step 1: Initialize Go module**

```bash
mkdir ollama-bench && cd ollama-bench
go mod init github.com/joelhelbling/ollama-bench
```

- [ ] **Step 2: Install dependencies**

```bash
go get github.com/spf13/cobra@latest
go get github.com/olekukonez/tablewriter@latest
```

- [ ] **Step 3: Create root command**

Create `internal/cli/root.go`:

```go
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	formatFlag string
	Version    = "dev"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ollama-bench",
		Short: "Benchmark Ollama models and share results",
		Long:  "Run benchmarks against local Ollama models, publish results, and compare performance across models and hardware.",
	}

	cmd.PersistentFlags().StringVar(&formatFlag, "format", "table", "Output format: table or json")

	return cmd
}

func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Create main entry point**

Create `cmd/ollama-bench/main.go`:

```go
package main

import "github.com/joelhelbling/ollama-bench/internal/cli"

func main() {
	cli.Execute()
}
```

- [ ] **Step 5: Verify it builds and runs**

```bash
go build ./cmd/ollama-bench
./ollama-bench --help
```

Expected: help output showing "Benchmark Ollama models and share results"

- [ ] **Step 6: Commit**

```bash
git init
git add .
git commit -m "feat: project scaffold with cobra root command"
```

---

### Task 2: Shared Data Models — Suite

**Files:**
- Create: `pkg/models/suite.go`
- Create: `pkg/models/suite_test.go`

- [ ] **Step 1: Write the test**

Create `pkg/models/suite_test.go`:

```go
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
				ID:       "prompt-1",
				Name:     "simple_math",
				Category: "reasoning",
				Text:     "What is 2 + 2?",
				ExpectedTokenRangeMin: intPtr(1),
				ExpectedTokenRangeMax: intPtr(50),
				SortOrder: 1,
			},
		},
		Parameters: models.SuiteParameters{
			RunsPerPrompt:  3,
			TimeoutSeconds: 300,
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/models/ -v -run TestSuiteJSON
```

Expected: FAIL — `models.Suite` not defined

- [ ] **Step 3: Write the implementation**

Create `pkg/models/suite.go`:

```go
package models

type Suite struct {
	ID          string          `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Prompts     []Prompt        `json:"prompts"`
	Parameters  SuiteParameters `json:"parameters"`
}

type Prompt struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Category              string `json:"category"`
	Text                  string `json:"text"`
	ExpectedTokenRangeMin *int   `json:"expected_token_range_min,omitempty"`
	ExpectedTokenRangeMax *int   `json:"expected_token_range_max,omitempty"`
	SortOrder             int    `json:"sort_order"`
}

type SuiteParameters struct {
	RunsPerPrompt   int `json:"runs_per_prompt"`
	TimeoutSeconds  int `json:"timeout_seconds"`
	CooldownSeconds int `json:"cooldown_seconds"`
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./pkg/models/ -v -run TestSuiteJSON
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/models/suite.go pkg/models/suite_test.go
git commit -m "feat: add Suite shared data model"
```

---

### Task 3: Shared Data Models — Hardware

**Files:**
- Create: `pkg/models/hardware.go`
- Create: `pkg/models/hardware_test.go`

- [ ] **Step 1: Write the test**

Create `pkg/models/hardware_test.go`:

```go
package models_test

import (
	"testing"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestHardwareFingerprint(t *testing.T) {
	hw := models.HardwareInfo{
		OSName:       "linux",
		OSVersion:    "6.17.0",
		Architecture: "x86_64",
		CPUModel:     "AMD Ryzen 9 7950X",
		CPUCores:     32,
		RAMGB:        63.9, // should round to 64
		GPUName:      "NVIDIA RTX 4090",
		VRAMGB:       24.0,
	}

	fp1 := hw.Fingerprint()
	if fp1 == "" {
		t.Fatal("fingerprint should not be empty")
	}

	// Same specs, slightly different RAM reporting — should produce same fingerprint
	hw2 := hw
	hw2.RAMGB = 64.1
	fp2 := hw2.Fingerprint()
	if fp1 != fp2 {
		t.Errorf("fingerprints should match after rounding: %q != %q", fp1, fp2)
	}

	// Different GPU — should produce different fingerprint
	hw3 := hw
	hw3.GPUName = "NVIDIA RTX 3090"
	fp3 := hw3.Fingerprint()
	if fp1 == fp3 {
		t.Error("fingerprints should differ for different GPUs")
	}
}

func TestHardwareFingerprintCaseInsensitive(t *testing.T) {
	hw1 := models.HardwareInfo{
		OSName:       "Linux",
		Architecture: "x86_64",
		CPUModel:     "AMD Ryzen 9 7950X",
		CPUCores:     32,
		RAMGB:        64,
	}
	hw2 := hw1
	hw2.OSName = "linux"

	if hw1.Fingerprint() != hw2.Fingerprint() {
		t.Error("fingerprint should be case-insensitive")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/models/ -v -run TestHardware
```

Expected: FAIL — `models.HardwareInfo` not defined

- [ ] **Step 3: Write the implementation**

Create `pkg/models/hardware.go`:

```go
package models

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
)

type HardwareInfo struct {
	OSName       string  `json:"os_name"`
	OSVersion    string  `json:"os_version"`
	Architecture string  `json:"architecture"`
	CPUModel     string  `json:"cpu_model"`
	CPUCores     int     `json:"cpu_cores"`
	RAMGB        float64 `json:"ram_gb"`
	GPUName      string  `json:"gpu_name,omitempty"`
	VRAMGB       float64 `json:"vram_gb,omitempty"`
}

// Fingerprint computes a SHA-256 hash of normalized hardware fields.
// RAM and VRAM are rounded to the nearest GB to avoid OS reporting variance.
func (h HardwareInfo) Fingerprint() string {
	ramRounded := int(math.Round(h.RAMGB))
	vramRounded := int(math.Round(h.VRAMGB))

	parts := []string{
		strings.ToLower(h.OSName),
		strings.ToLower(h.OSVersion),
		strings.ToLower(h.Architecture),
		strings.ToLower(h.CPUModel),
		fmt.Sprintf("%d", h.CPUCores),
		fmt.Sprintf("%d", ramRounded),
		strings.ToLower(h.GPUName),
		fmt.Sprintf("%d", vramRounded),
	}

	input := strings.Join(parts, "|")
	hash := sha256.Sum256([]byte(input))
	return fmt.Sprintf("%x", hash)
}
```

- [ ] **Step 4: Run test to verify it passes**

```bash
go test ./pkg/models/ -v -run TestHardware
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/models/hardware.go pkg/models/hardware_test.go
git commit -m "feat: add HardwareInfo model with fingerprinting"
```

---

### Task 4: Shared Data Models — Results and API Types

**Files:**
- Create: `pkg/models/results.go`
- Create: `pkg/models/api.go`
- Create: `pkg/models/results_test.go`

- [ ] **Step 1: Write the test**

Create `pkg/models/results_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/models/ -v -run "TestPromptResult|TestPublishRequest"
```

Expected: FAIL — types not defined

- [ ] **Step 3: Write results model**

Create `pkg/models/results.go`:

```go
package models

type OllamaStats struct {
	TotalDurationS      float64 `json:"total_duration_s"`
	LoadDurationS       float64 `json:"load_duration_s"`
	PromptEvalCount     int     `json:"prompt_eval_count"`
	PromptEvalDurationS float64 `json:"prompt_eval_duration_s"`
	PromptEvalRate      float64 `json:"prompt_eval_rate"`
	EvalCount           int     `json:"eval_count"`
	EvalDurationS       float64 `json:"eval_duration_s"`
	EvalRate            float64 `json:"eval_rate"`
}

type PromptResult struct {
	PromptID     string      `json:"prompt_id"`
	RunNumber    int         `json:"run_number"`
	Stats        OllamaStats `json:"stats"`
	Success      bool        `json:"success"`
	ErrorMessage string      `json:"error_message,omitempty"`
}
```

- [ ] **Step 4: Write API types**

Create `pkg/models/api.go`:

```go
package models

import "time"

// --- Requests ---

type PublishRequest struct {
	SuiteID      string         `json:"suite_id"`
	SuiteVersion string         `json:"suite_version"`
	ModelName    string         `json:"model_name"`
	Hardware     HardwareInfo   `json:"hardware"`
	Results      []PromptResult `json:"results"`
	Token        string         `json:"token,omitempty"`
	StartedAt    time.Time      `json:"started_at"`
	CompletedAt  time.Time      `json:"completed_at"`
}

// --- Responses ---

type PublishResponse struct {
	RunID string `json:"run_id"`
	Token string `json:"token"`
	URL   string `json:"url"`
}

type RunResponse struct {
	ID           string         `json:"id"`
	SuiteID      string         `json:"suite_id"`
	SuiteVersion string         `json:"suite_version"`
	ModelName    string         `json:"model_name"`
	Hardware     HardwareInfo   `json:"hardware"`
	Results      []PromptResult `json:"results"`
	StartedAt    time.Time      `json:"started_at"`
	CompletedAt  time.Time      `json:"completed_at"`
}

type CompareResponse struct {
	SuiteID      string         `json:"suite_id"`
	SuiteVersion string         `json:"suite_version"`
	Groups       []CompareGroup `json:"groups"`
}

type CompareGroup struct {
	ModelName string        `json:"model_name,omitempty"`
	Hardware  *HardwareInfo `json:"hardware,omitempty"`
	Stats     SummaryStats  `json:"stats"`
	RunCount  int           `json:"run_count"`
	Runs      []RunResponse `json:"runs,omitempty"`
}

type SummaryStats struct {
	MeanEvalRate       float64 `json:"mean_eval_rate"`
	StddevEvalRate     float64 `json:"stddev_eval_rate"`
	MeanPromptEvalRate float64 `json:"mean_prompt_eval_rate"`
	StddevPromptEvalRate float64 `json:"stddev_prompt_eval_rate"`
}

type HardwareConfigSummary struct {
	ID       string       `json:"id"`
	Hardware HardwareInfo `json:"hardware"`
	RunCount int          `json:"run_count"`
}

type ModelSummary struct {
	Name     string `json:"name"`
	RunCount int    `json:"run_count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go test ./pkg/models/ -v
```

Expected: all tests PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/models/results.go pkg/models/api.go pkg/models/results_test.go
git commit -m "feat: add Results, OllamaStats, and API request/response models"
```

---

### Task 5: Ollama Output Parser

**Files:**
- Create: `internal/ollama/parser.go`
- Create: `internal/ollama/parser_test.go`

- [ ] **Step 1: Write the test**

Create `internal/ollama/parser_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ollama/ -v -run TestParse
```

Expected: FAIL — `ollama.ParseVerboseOutput` not defined

- [ ] **Step 3: Write the implementation**

Create `internal/ollama/parser.go`:

```go
package ollama

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

var (
	totalDurationRe      = regexp.MustCompile(`total duration:\s+(.+)`)
	loadDurationRe       = regexp.MustCompile(`load duration:\s+(.+)`)
	promptEvalCountRe    = regexp.MustCompile(`prompt eval count:\s+(\d+)`)
	promptEvalDurationRe = regexp.MustCompile(`prompt eval duration:\s+(.+)`)
	promptEvalRateRe     = regexp.MustCompile(`prompt eval rate:\s+([\d.]+)`)
	// Negative lookbehind to avoid matching "prompt eval count/duration/rate"
	evalCountRe    = regexp.MustCompile(`(?:^|\n)[^\n]*(?:^|[^t] )eval count:\s+(\d+)`)
	evalDurationRe = regexp.MustCompile(`(?:^|\n)[^\n]*(?:^|[^t] )eval duration:\s+(.+)`)
	evalRateRe     = regexp.MustCompile(`(?:^|\n)[^\n]*(?:^|[^t] )eval rate:\s+([\d.]+)`)
)

func ParseVerboseOutput(output string) (models.OllamaStats, error) {
	var stats models.OllamaStats
	var err error

	stats.TotalDurationS, err = parseDurationField(totalDurationRe, output, "total duration")
	if err != nil {
		return stats, err
	}

	stats.LoadDurationS, err = parseDurationField(loadDurationRe, output, "load duration")
	if err != nil {
		return stats, err
	}

	stats.PromptEvalCount, err = parseIntField(promptEvalCountRe, output, "prompt eval count")
	if err != nil {
		return stats, err
	}

	stats.PromptEvalDurationS, err = parseDurationField(promptEvalDurationRe, output, "prompt eval duration")
	if err != nil {
		return stats, err
	}

	stats.PromptEvalRate, err = parseFloatField(promptEvalRateRe, output, "prompt eval rate")
	if err != nil {
		return stats, err
	}

	stats.EvalCount, err = parseIntField(evalCountRe, output, "eval count")
	if err != nil {
		return stats, err
	}

	stats.EvalDurationS, err = parseDurationField(evalDurationRe, output, "eval duration")
	if err != nil {
		return stats, err
	}

	stats.EvalRate, err = parseFloatField(evalRateRe, output, "eval rate")
	if err != nil {
		return stats, err
	}

	return stats, nil
}

func parseDurationField(re *regexp.Regexp, output, name string) (float64, error) {
	match := re.FindStringSubmatch(output)
	if match == nil {
		return 0, fmt.Errorf("field %q not found in output", name)
	}
	return parseDuration(strings.TrimSpace(match[1]))
}

func parseIntField(re *regexp.Regexp, output, name string) (int, error) {
	match := re.FindStringSubmatch(output)
	if match == nil {
		return 0, fmt.Errorf("field %q not found in output", name)
	}
	return strconv.Atoi(strings.TrimSpace(match[1]))
}

func parseFloatField(re *regexp.Regexp, output, name string) (float64, error) {
	match := re.FindStringSubmatch(output)
	if match == nil {
		return 0, fmt.Errorf("field %q not found in output", name)
	}
	return strconv.ParseFloat(strings.TrimSpace(match[1]), 64)
}

// parseDuration handles Go-style durations: "5.227891234s", "152.345ms", "1m12.423928897s"
func parseDuration(s string) (float64, error) {
	s = strings.TrimSpace(s)

	var totalSeconds float64

	// Handle minutes component: "1m12.423s" -> minutes=1, remainder="12.423s"
	if idx := strings.Index(s, "m"); idx > 0 && !strings.HasSuffix(s[:idx+1], "ms") {
		minutes, err := strconv.ParseFloat(s[:idx], 64)
		if err != nil {
			return 0, fmt.Errorf("invalid minutes in duration %q: %w", s, err)
		}
		totalSeconds += minutes * 60
		s = s[idx+1:]
	}

	if strings.HasSuffix(s, "ms") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(s, "ms"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid ms duration %q: %w", s, err)
		}
		totalSeconds += val / 1000
	} else if strings.HasSuffix(s, "s") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(s, "s"), 64)
		if err != nil {
			return 0, fmt.Errorf("invalid seconds duration %q: %w", s, err)
		}
		totalSeconds += val
	} else if s != "" {
		return 0, fmt.Errorf("unknown duration format: %q", s)
	}

	return totalSeconds, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/ollama/ -v -run TestParse
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ollama/parser.go internal/ollama/parser_test.go
git commit -m "feat: add ollama verbose output parser"
```

---

### Task 6: Ollama Runner

**Files:**
- Create: `internal/ollama/runner.go`
- Create: `internal/ollama/runner_test.go`

- [ ] **Step 1: Write the test**

Create `internal/ollama/runner_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/ollama/ -v -run TestRun -short
```

Expected: SKIP (short mode) — but confirms compilation. Without `-short`, FAIL — `ollama.Run` not defined.

- [ ] **Step 3: Write the implementation**

Create `internal/ollama/runner.go`:

```go
package ollama

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type RunResult struct {
	Output string
	Stats  models.OllamaStats
}

func Run(ctx context.Context, model, prompt string) (RunResult, error) {
	cmd := exec.CommandContext(ctx, "ollama", "run", "--verbose", model, prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return RunResult{}, fmt.Errorf("ollama run failed: %w\nstderr: %s", err, stderr.String())
	}

	// Ollama writes the response to stdout and stats to stderr
	combined := stdout.String() + "\n" + stderr.String()

	stats, err := ParseVerboseOutput(combined)
	if err != nil {
		return RunResult{}, fmt.Errorf("failed to parse ollama output: %w\noutput: %s", err, combined)
	}

	return RunResult{
		Output: strings.TrimSpace(stdout.String()),
		Stats:  stats,
	}, nil
}

func ListModels(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "ollama", "list")

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ollama list failed: %w", err)
	}

	var modelNames []string
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			modelNames = append(modelNames, fields[0])
		}
	}
	return modelNames, nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/ollama/ -v -short
```

Expected: PASS (integration tests skipped)

- [ ] **Step 5: Commit**

```bash
git add internal/ollama/runner.go internal/ollama/runner_test.go
git commit -m "feat: add ollama subprocess runner"
```

---

### Task 7: Hardware Detection

**Files:**
- Create: `internal/hardware/detect.go`
- Create: `internal/hardware/detect_test.go`

- [ ] **Step 1: Write the test**

Create `internal/hardware/detect_test.go`:

```go
package hardware_test

import (
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/hardware"
)

func TestDetectReturnsPopulatedFields(t *testing.T) {
	info, err := hardware.Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if info.OSName == "" {
		t.Error("OSName should not be empty")
	}
	if info.Architecture == "" {
		t.Error("Architecture should not be empty")
	}
	if info.CPUModel == "" {
		t.Error("CPUModel should not be empty")
	}
	if info.CPUCores <= 0 {
		t.Errorf("CPUCores should be positive, got %d", info.CPUCores)
	}
	if info.RAMGB <= 0 {
		t.Errorf("RAMGB should be positive, got %f", info.RAMGB)
	}
}

func TestDetectFingerprint(t *testing.T) {
	info, err := hardware.Detect()
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	fp := info.Fingerprint()
	if fp == "" {
		t.Error("Fingerprint should not be empty")
	}

	// Running twice should produce the same fingerprint
	info2, _ := hardware.Detect()
	if info.Fingerprint() != info2.Fingerprint() {
		t.Error("Fingerprint should be stable across calls")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/hardware/ -v
```

Expected: FAIL — `hardware.Detect` not defined

- [ ] **Step 3: Write the implementation**

Create `internal/hardware/detect.go`:

```go
package hardware

import (
	"math"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func Detect() (models.HardwareInfo, error) {
	info := models.HardwareInfo{
		Architecture: runtime.GOARCH,
		CPUCores:     runtime.NumCPU(),
	}

	detectOS(&info)
	detectCPU(&info)
	detectRAM(&info)
	detectGPU(&info)

	return info, nil
}

func detectOS(info *models.HardwareInfo) {
	switch runtime.GOOS {
	case "linux":
		info.OSName = "linux"
		if out, err := exec.Command("uname", "-r").Output(); err == nil {
			info.OSVersion = strings.TrimSpace(string(out))
		}
	case "darwin":
		info.OSName = "macos"
		if out, err := exec.Command("sw_vers", "-productVersion").Output(); err == nil {
			info.OSVersion = strings.TrimSpace(string(out))
		}
	case "windows":
		info.OSName = "windows"
		if out, err := exec.Command("cmd", "/c", "ver").Output(); err == nil {
			info.OSVersion = strings.TrimSpace(string(out))
		}
	default:
		info.OSName = runtime.GOOS
	}
}

func detectCPU(info *models.HardwareInfo) {
	switch runtime.GOOS {
	case "linux":
		if out, err := exec.Command("sh", "-c", `grep -m1 'model name' /proc/cpuinfo | cut -d: -f2`).Output(); err == nil {
			info.CPUModel = strings.TrimSpace(string(out))
		}
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "machdep.cpu.brand_string").Output(); err == nil {
			info.CPUModel = strings.TrimSpace(string(out))
		}
	}
	if info.CPUModel == "" {
		info.CPUModel = "unknown"
	}
}

func detectRAM(info *models.HardwareInfo) {
	switch runtime.GOOS {
	case "linux":
		if out, err := exec.Command("sh", "-c", `grep MemTotal /proc/meminfo | awk '{print $2}'`).Output(); err == nil {
			if kb, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64); err == nil {
				info.RAMGB = math.Round(kb/1024/1024*10) / 10
			}
		}
	case "darwin":
		if out, err := exec.Command("sysctl", "-n", "hw.memsize").Output(); err == nil {
			if bytes, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64); err == nil {
				info.RAMGB = math.Round(bytes/1024/1024/1024*10) / 10
			}
		}
	}
}

func detectGPU(info *models.HardwareInfo) {
	// Try NVIDIA first
	if out, err := exec.Command("nvidia-smi", "--query-gpu=name", "--format=csv,noheader,nounits").Output(); err == nil {
		name := strings.TrimSpace(strings.Split(string(out), "\n")[0])
		if name != "" {
			info.GPUName = name
			detectNvidiaVRAM(info)
			return
		}
	}

	// Try Apple Silicon
	if runtime.GOOS == "darwin" && runtime.GOARCH == "arm64" {
		info.GPUName = "Apple Silicon (integrated)"
		// On Apple Silicon, GPU shares system memory — VRAM is not separately reportable
		return
	}

	// Try AMD via rocm-smi
	if out, err := exec.Command("rocm-smi", "--showproductname").Output(); err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Card") {
				fields := strings.Fields(line)
				if len(fields) > 2 {
					info.GPUName = strings.Join(fields[2:], " ")
					break
				}
			}
		}
	}
}

func detectNvidiaVRAM(info *models.HardwareInfo) {
	if out, err := exec.Command("nvidia-smi", "--query-gpu=memory.total", "--format=csv,noheader,nounits").Output(); err == nil {
		if mb, err := strconv.ParseFloat(strings.TrimSpace(strings.Split(string(out), "\n")[0]), 64); err == nil {
			info.VRAMGB = math.Round(mb/1024*10) / 10
		}
	}
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/hardware/ -v
```

Expected: PASS (specific GPU fields may be empty depending on test machine — that's OK)

- [ ] **Step 5: Commit**

```bash
git add internal/hardware/detect.go internal/hardware/detect_test.go
git commit -m "feat: add cross-platform hardware detection"
```

---

### Task 8: API Client

**Files:**
- Create: `pkg/apiclient/client.go`
- Create: `pkg/apiclient/client_test.go`

- [ ] **Step 1: Write the test**

Create `pkg/apiclient/client_test.go`:

```go
package apiclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joelhelbling/ollama-bench/pkg/apiclient"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestGetSuite(t *testing.T) {
	suite := models.Suite{
		ID:      "suite-1",
		Slug:    "default",
		Name:    "Default Suite",
		Version: "1.0.0",
		Prompts: []models.Prompt{
			{ID: "p1", Name: "test", Category: "general", Text: "Hello"},
		},
		Parameters: models.SuiteParameters{RunsPerPrompt: 3, TimeoutSeconds: 300, CooldownSeconds: 5},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/suites/default" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(suite)
	}))
	defer server.Close()

	client := apiclient.New(server.URL, "test-version")
	got, err := client.GetSuite(context.Background(), "default")
	if err != nil {
		t.Fatalf("GetSuite failed: %v", err)
	}

	if got.Slug != "default" {
		t.Errorf("expected slug 'default', got %q", got.Slug)
	}
	if len(got.Prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(got.Prompts))
	}
}

func TestPublishResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/results" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-OllamaBench-Version") != "test-version" {
			t.Errorf("missing version header")
		}

		var req models.PublishRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.ModelName != "llama3:8b" {
			t.Errorf("expected model 'llama3:8b', got %q", req.ModelName)
		}

		json.NewEncoder(w).Encode(models.PublishResponse{
			RunID: "run-123",
			Token: "tok-abc",
			URL:   "https://bench.example.com/results/run-123",
		})
	}))
	defer server.Close()

	client := apiclient.New(server.URL, "test-version")
	resp, err := client.PublishResults(context.Background(), models.PublishRequest{
		ModelName: "llama3:8b",
	})
	if err != nil {
		t.Fatalf("PublishResults failed: %v", err)
	}

	if resp.RunID != "run-123" {
		t.Errorf("expected run ID 'run-123', got %q", resp.RunID)
	}
	if resp.Token != "tok-abc" {
		t.Errorf("expected token 'tok-abc', got %q", resp.Token)
	}
}

func TestGetComparison(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/compare" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("model") != "llama3:8b" {
			t.Errorf("expected model query param")
		}

		json.NewEncoder(w).Encode(models.CompareResponse{
			Groups: []models.CompareGroup{
				{
					Hardware: &models.HardwareInfo{GPUName: "RTX 4090"},
					Stats:    models.SummaryStats{MeanEvalRate: 85.2},
					RunCount: 5,
				},
			},
		})
	}))
	defer server.Close()

	client := apiclient.New(server.URL, "test-version")
	params := apiclient.CompareParams{Model: "llama3:8b"}
	resp, err := client.GetComparison(context.Background(), params)
	if err != nil {
		t.Fatalf("GetComparison failed: %v", err)
	}

	if len(resp.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(resp.Groups))
	}
	if resp.Groups[0].Stats.MeanEvalRate != 85.2 {
		t.Errorf("expected mean eval rate 85.2, got %f", resp.Groups[0].Stats.MeanEvalRate)
	}
}

func TestClientHandlesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal error"})
	}))
	defer server.Close()

	client := apiclient.New(server.URL, "test-version")
	_, err := client.GetSuite(context.Background(), "default")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./pkg/apiclient/ -v
```

Expected: FAIL — package not defined

- [ ] **Step 3: Write the implementation**

Create `pkg/apiclient/client.go`:

```go
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type Client struct {
	baseURL    string
	version    string
	httpClient *http.Client
}

func New(baseURL, version string) *Client {
	return &Client{
		baseURL: baseURL,
		version: version,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetSuite(ctx context.Context, slugOrID string) (models.Suite, error) {
	var suite models.Suite
	err := c.get(ctx, fmt.Sprintf("/api/v1/suites/%s", slugOrID), &suite)
	return suite, err
}

func (c *Client) ListSuites(ctx context.Context) ([]models.Suite, error) {
	var suites []models.Suite
	err := c.get(ctx, "/api/v1/suites", &suites)
	return suites, err
}

func (c *Client) PublishResults(ctx context.Context, req models.PublishRequest) (models.PublishResponse, error) {
	var resp models.PublishResponse
	err := c.post(ctx, "/api/v1/results", req, &resp)
	return resp, err
}

func (c *Client) GetRun(ctx context.Context, runID string) (models.RunResponse, error) {
	var resp models.RunResponse
	err := c.get(ctx, fmt.Sprintf("/api/v1/results/%s", runID), &resp)
	return resp, err
}

type CompareParams struct {
	Model      string
	HardwareID string
	GPU        string
	CPU        string
	RAMMin     float64
	RAMMax     float64
	VRAMMin    float64
	VRAMMax    float64
	OS         string
	Arch       string
	SuiteID    string
	Expand     string
}

func (c *Client) GetComparison(ctx context.Context, params CompareParams) (models.CompareResponse, error) {
	query := url.Values{}
	if params.Model != "" {
		query.Set("model", params.Model)
	}
	if params.HardwareID != "" {
		query.Set("hardware_id", params.HardwareID)
	}
	if params.GPU != "" {
		query.Set("gpu", params.GPU)
	}
	if params.CPU != "" {
		query.Set("cpu", params.CPU)
	}
	if params.RAMMin > 0 {
		query.Set("ram_min", fmt.Sprintf("%.0f", params.RAMMin))
	}
	if params.RAMMax > 0 {
		query.Set("ram_max", fmt.Sprintf("%.0f", params.RAMMax))
	}
	if params.VRAMMin > 0 {
		query.Set("vram_min", fmt.Sprintf("%.0f", params.VRAMMin))
	}
	if params.VRAMMax > 0 {
		query.Set("vram_max", fmt.Sprintf("%.0f", params.VRAMMax))
	}
	if params.OS != "" {
		query.Set("os", params.OS)
	}
	if params.Arch != "" {
		query.Set("arch", params.Arch)
	}
	if params.SuiteID != "" {
		query.Set("suite_id", params.SuiteID)
	}
	if params.Expand != "" {
		query.Set("expand", params.Expand)
	}

	path := "/api/v1/compare"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp models.CompareResponse
	err := c.get(ctx, path, &resp)
	return resp, err
}

func (c *Client) GetHardwareConfigs(ctx context.Context) ([]models.HardwareConfigSummary, error) {
	var configs []models.HardwareConfigSummary
	err := c.get(ctx, "/api/v1/hardware-configs", &configs)
	return configs, err
}

func (c *Client) GetModels(ctx context.Context) ([]models.ModelSummary, error) {
	var modelList []models.ModelSummary
	err := c.get(ctx, "/api/v1/models", &modelList)
	return modelList, err
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-OllamaBench-Version", c.version)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) post(ctx context.Context, path string, body any, result any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-OllamaBench-Version", c.version)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./pkg/apiclient/ -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add pkg/apiclient/client.go pkg/apiclient/client_test.go
git commit -m "feat: add HTTP API client for backend communication"
```

---

### Task 9: Suite Caching

**Files:**
- Create: `internal/cache/suite.go`
- Create: `internal/cache/suite_test.go`

- [ ] **Step 1: Write the test**

Create `internal/cache/suite_test.go`:

```go
package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/cache"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestSaveThenLoad(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewSuiteCache(dir)

	suite := models.Suite{
		ID:      "suite-1",
		Slug:    "default",
		Version: "1.0.0",
		Prompts: []models.Prompt{
			{ID: "p1", Name: "test", Category: "general", Text: "Hello"},
		},
		Parameters: models.SuiteParameters{RunsPerPrompt: 3},
	}

	if err := c.Save(suite); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(dir, "default-1.0.0.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected cache file at %s", expectedPath)
	}

	loaded, err := c.Load("default")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", loaded.Version)
	}
	if len(loaded.Prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(loaded.Prompts))
	}
}

func TestLoadNotCached(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewSuiteCache(dir)

	_, err := c.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing cache")
	}
}

func TestSaveOverwritesOlderVersion(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewSuiteCache(dir)

	suite1 := models.Suite{Slug: "default", Version: "1.0.0"}
	suite2 := models.Suite{Slug: "default", Version: "1.1.0"}

	c.Save(suite1)
	c.Save(suite2)

	loaded, err := c.Load("default")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Version != "1.1.0" {
		t.Errorf("expected version '1.1.0', got %q", loaded.Version)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/cache/ -v
```

Expected: FAIL — `cache.NewSuiteCache` not defined

- [ ] **Step 3: Write the implementation**

Create `internal/cache/suite.go`:

```go
package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type SuiteCache struct {
	dir string
}

func NewSuiteCache(dir string) *SuiteCache {
	return &SuiteCache{dir: dir}
}

func (c *SuiteCache) Save(suite models.Suite) error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	// Remove older versions of the same suite
	entries, _ := os.ReadDir(c.dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), suite.Slug+"-") && strings.HasSuffix(e.Name(), ".json") {
			os.Remove(filepath.Join(c.dir, e.Name()))
		}
	}

	filename := fmt.Sprintf("%s-%s.json", suite.Slug, suite.Version)
	path := filepath.Join(c.dir, filename)

	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling suite: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (c *SuiteCache) Load(slug string) (models.Suite, error) {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return models.Suite{}, fmt.Errorf("reading cache dir: %w", err)
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), slug+"-") && strings.HasSuffix(e.Name(), ".json") {
			path := filepath.Join(c.dir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return models.Suite{}, fmt.Errorf("reading cached suite: %w", err)
			}
			var suite models.Suite
			if err := json.Unmarshal(data, &suite); err != nil {
				return models.Suite{}, fmt.Errorf("parsing cached suite: %w", err)
			}
			return suite, nil
		}
	}

	return models.Suite{}, fmt.Errorf("suite %q not found in cache", slug)
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/cache/ -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/cache/suite.go internal/cache/suite_test.go
git commit -m "feat: add suite caching for offline support"
```

---

### Task 10: Config Management

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

- [ ] **Step 1: Write the test**

Create `internal/config/config_test.go`:

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Default()
	if cfg.APIBaseURL == "" {
		t.Error("default API base URL should not be empty")
	}
}

func TestLoadCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "ollama-bench")

	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.APIBaseURL == "" {
		t.Error("loaded config should have default API URL")
	}
}

func TestTokenPersistence(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "ollama-bench")

	if err := config.SaveToken(configDir, "tok-abc-123"); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	token, err := config.LoadToken(configDir)
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}

	if token != "tok-abc-123" {
		t.Errorf("expected token 'tok-abc-123', got %q", token)
	}
}

func TestLoadTokenMissing(t *testing.T) {
	dir := t.TempDir()

	token, err := config.LoadToken(dir)
	if err != nil {
		t.Fatalf("LoadToken should not error for missing token: %v", err)
	}
	if token != "" {
		t.Errorf("expected empty token, got %q", token)
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "ollama-bench")

	cfg := config.Config{APIBaseURL: "https://custom.example.com"}
	if err := config.Save(configDir, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.APIBaseURL != "https://custom.example.com" {
		t.Errorf("expected custom URL, got %q", loaded.APIBaseURL)
	}

	// Verify file was created
	if _, err := os.Stat(filepath.Join(configDir, "config.yaml")); os.IsNotExist(err) {
		t.Fatal("config file should exist")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/config/ -v
```

Expected: FAIL — `config.Config` not defined

- [ ] **Step 3: Install YAML dependency and write the implementation**

```bash
go get gopkg.in/yaml.v3
```

Create `internal/config/config.go`:

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultAPIBaseURL = "https://bench.ollama.example.com"
	configFileName    = "config.yaml"
	tokenFileName     = "token"
)

type Config struct {
	APIBaseURL string `yaml:"api_base_url"`
}

func Default() Config {
	return Config{
		APIBaseURL: DefaultAPIBaseURL,
	}
}

func Load(configDir string) (Config, error) {
	path := filepath.Join(configDir, configFileName)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = DefaultAPIBaseURL
	}

	return cfg, nil
}

func Save(configDir string, cfg Config) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(filepath.Join(configDir, configFileName), data, 0644)
}

func SaveToken(configDir, token string) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	return os.WriteFile(filepath.Join(configDir, tokenFileName), []byte(token), 0600)
}

func LoadToken(configDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(configDir, tokenFileName))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading token: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func DefaultConfigDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "ollama-bench")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ollama-bench")
}

func DefaultCacheDir() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "ollama-bench", "suites")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "ollama-bench", "suites")
}

func DefaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "ollama-bench", "pending")
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/config/ -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: add config and token management"
```

---

### Task 11: Pending Results Store

**Files:**
- Create: `internal/pending/store.go`
- Create: `internal/pending/store_test.go`

- [ ] **Step 1: Write the test**

Create `internal/pending/store_test.go`:

```go
package pending_test

import (
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/pending"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestSaveAndList(t *testing.T) {
	dir := t.TempDir()
	store := pending.NewStore(dir)

	req := models.PublishRequest{
		SuiteID:   "suite-1",
		ModelName: "llama3:8b",
		Hardware: models.HardwareInfo{
			CPUModel: "AMD Ryzen 9",
			RAMGB:    64,
		},
		Results: []models.PromptResult{
			{PromptID: "p1", RunNumber: 1, Success: true},
		},
	}

	if err := store.Save(req); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	items, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 pending item, got %d", len(items))
	}

	if items[0].Request.ModelName != "llama3:8b" {
		t.Errorf("expected model 'llama3:8b', got %q", items[0].Request.ModelName)
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	store := pending.NewStore(dir)

	store.Save(models.PublishRequest{ModelName: "model1"})
	store.Save(models.PublishRequest{ModelName: "model2"})

	items, _ := store.List()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if err := store.Remove(items[0].Filename); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	items, _ = store.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 item after removal, got %d", len(items))
	}
}

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	store := pending.NewStore(dir)

	items, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/pending/ -v
```

Expected: FAIL — `pending.NewStore` not defined

- [ ] **Step 3: Write the implementation**

Create `internal/pending/store.go`:

```go
package pending

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type Store struct {
	dir string
}

type PendingItem struct {
	Filename string
	Request  models.PublishRequest
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) Save(req models.PublishRequest) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("creating pending dir: %w", err)
	}

	filename := fmt.Sprintf("pending_%s.json", time.Now().Format("20060102_150405.000"))
	path := filepath.Join(s.dir, filename)

	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling pending result: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (s *Store) List() ([]PendingItem, error) {
	entries, err := os.ReadDir(s.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading pending dir: %w", err)
	}

	var items []PendingItem
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var req models.PublishRequest
		if err := json.Unmarshal(data, &req); err != nil {
			continue
		}

		items = append(items, PendingItem{
			Filename: e.Name(),
			Request:  req,
		})
	}

	return items, nil
}

func (s *Store) Remove(filename string) error {
	return os.Remove(filepath.Join(s.dir, filename))
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/pending/ -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/pending/store.go internal/pending/store_test.go
git commit -m "feat: add pending results store for offline support"
```

---

### Task 12: Output Formatters

**Files:**
- Create: `internal/output/formatter.go`
- Create: `internal/output/table.go`
- Create: `internal/output/json.go`
- Create: `internal/output/formatter_test.go`

- [ ] **Step 1: Write the test**

Create `internal/output/formatter_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/output/ -v
```

Expected: FAIL — types not defined

- [ ] **Step 3: Write the formatter interface**

Create `internal/output/formatter.go`:

```go
package output

import (
	"io"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type Formatter interface {
	FormatComparison(resp models.CompareResponse) error
	FormatRun(run models.RunResponse) error
	FormatSuites(suites []models.Suite) error
	FormatSuite(suite models.Suite) error
}

func New(format string, w io.Writer) Formatter {
	switch format {
	case "json":
		return NewJSONFormatter(w)
	default:
		return NewTableFormatter(w)
	}
}
```

- [ ] **Step 4: Write the table formatter**

Create `internal/output/table.go`:

```go
package output

import (
	"fmt"
	"io"

	"github.com/olekukonez/tablewriter"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type TableFormatter struct {
	w io.Writer
}

func NewTableFormatter(w io.Writer) *TableFormatter {
	return &TableFormatter{w: w}
}

func (f *TableFormatter) FormatComparison(resp models.CompareResponse) error {
	table := tablewriter.NewWriter(f.w)
	table.SetHeader([]string{"GPU", "CPU", "RAM (GB)", "Eval Rate", "Stddev", "Prompt Eval Rate", "Runs"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})

	for _, g := range resp.Groups {
		gpuName := ""
		cpuModel := ""
		ramGB := ""
		modelName := ""

		if g.Hardware != nil {
			gpuName = g.Hardware.GPUName
			cpuModel = g.Hardware.CPUModel
			ramGB = fmt.Sprintf("%.0f", g.Hardware.RAMGB)
		}
		if g.ModelName != "" {
			modelName = g.ModelName
		}

		row := []string{}
		if modelName != "" {
			row = append(row, modelName)
		}
		if gpuName != "" || cpuModel != "" {
			row = append(row, gpuName, cpuModel, ramGB)
		}
		row = append(row,
			fmt.Sprintf("%.1f tok/s", g.Stats.MeanEvalRate),
			fmt.Sprintf("%.1f", g.Stats.StddevEvalRate),
			fmt.Sprintf("%.1f tok/s", g.Stats.MeanPromptEvalRate),
			fmt.Sprintf("%d", g.RunCount),
		)
		table.Append(row)
	}

	table.Render()
	return nil
}

func (f *TableFormatter) FormatRun(run models.RunResponse) error {
	fmt.Fprintf(f.w, "Run: %s\n", run.ID)
	fmt.Fprintf(f.w, "Model: %s\n", run.ModelName)
	if run.Hardware.GPUName != "" {
		fmt.Fprintf(f.w, "GPU: %s\n", run.Hardware.GPUName)
	}
	fmt.Fprintf(f.w, "CPU: %s | RAM: %.0f GB\n\n", run.Hardware.CPUModel, run.Hardware.RAMGB)

	table := tablewriter.NewWriter(f.w)
	table.SetHeader([]string{"Prompt", "Run #", "Eval Rate", "Prompt Eval Rate", "Total Duration", "Success"})

	for _, r := range run.Results {
		success := "yes"
		if !r.Success {
			success = "FAIL"
		}
		table.Append([]string{
			r.PromptID,
			fmt.Sprintf("%d", r.RunNumber),
			fmt.Sprintf("%.1f tok/s", r.Stats.EvalRate),
			fmt.Sprintf("%.1f tok/s", r.Stats.PromptEvalRate),
			fmt.Sprintf("%.2fs", r.Stats.TotalDurationS),
			success,
		})
	}

	table.Render()
	return nil
}

func (f *TableFormatter) FormatSuites(suites []models.Suite) error {
	table := tablewriter.NewWriter(f.w)
	table.SetHeader([]string{"Slug", "Name", "Version", "Prompts"})

	for _, s := range suites {
		table.Append([]string{
			s.Slug,
			s.Name,
			s.Version,
			fmt.Sprintf("%d", len(s.Prompts)),
		})
	}

	table.Render()
	return nil
}

func (f *TableFormatter) FormatSuite(suite models.Suite) error {
	fmt.Fprintf(f.w, "Suite: %s (v%s)\n", suite.Name, suite.Version)
	fmt.Fprintf(f.w, "%s\n\n", suite.Description)
	fmt.Fprintf(f.w, "Runs per prompt: %d | Timeout: %ds | Cooldown: %ds\n\n",
		suite.Parameters.RunsPerPrompt,
		suite.Parameters.TimeoutSeconds,
		suite.Parameters.CooldownSeconds,
	)

	table := tablewriter.NewWriter(f.w)
	table.SetHeader([]string{"#", "Name", "Category", "Prompt (truncated)"})

	for i, p := range suite.Prompts {
		text := p.Text
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		table.Append([]string{
			fmt.Sprintf("%d", i+1),
			p.Name,
			p.Category,
			text,
		})
	}

	table.Render()
	return nil
}
```

- [ ] **Step 5: Write the JSON formatter**

Create `internal/output/json.go`:

```go
package output

import (
	"encoding/json"
	"io"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type JSONFormatter struct {
	w io.Writer
}

func NewJSONFormatter(w io.Writer) *JSONFormatter {
	return &JSONFormatter{w: w}
}

func (f *JSONFormatter) FormatComparison(resp models.CompareResponse) error {
	return f.encode(resp)
}

func (f *JSONFormatter) FormatRun(run models.RunResponse) error {
	return f.encode(run)
}

func (f *JSONFormatter) FormatSuites(suites []models.Suite) error {
	return f.encode(suites)
}

func (f *JSONFormatter) FormatSuite(suite models.Suite) error {
	return f.encode(suite)
}

func (f *JSONFormatter) encode(v any) error {
	enc := json.NewEncoder(f.w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
```

- [ ] **Step 6: Run tests to verify they pass**

```bash
go test ./internal/output/ -v
```

Expected: all PASS

- [ ] **Step 7: Commit**

```bash
git add internal/output/
git commit -m "feat: add table and JSON output formatters"
```

---

### Task 13: Benchmark Runner (Orchestrator)

**Files:**
- Create: `internal/benchmark/runner.go`
- Create: `internal/benchmark/runner_test.go`

- [ ] **Step 1: Write the test**

Create `internal/benchmark/runner_test.go`:

```go
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
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/benchmark/ -v
```

Expected: FAIL — `benchmark.NewRunner` not defined

- [ ] **Step 3: Write the implementation**

Create `internal/benchmark/runner.go`:

```go
package benchmark

import (
	"context"
	"fmt"
	"time"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type OllamaRunner interface {
	Run(ctx context.Context, model, prompt string) (models.OllamaStats, error)
}

type Progress struct {
	Completed   int
	Total       int
	Model       string
	PromptName  string
	RunNumber   int
	Success     bool
	EvalRate    float64
}

type ModelRun struct {
	ModelName string
	Results   []models.PromptResult
	StartedAt time.Time
	CompletedAt time.Time
}

type Runner struct {
	ollama OllamaRunner
}

func NewRunner(ollama OllamaRunner) *Runner {
	return &Runner{ollama: ollama}
}

func (r *Runner) Execute(
	ctx context.Context,
	suite models.Suite,
	modelNames []string,
	onProgress func(Progress),
) ([]ModelRun, error) {
	total := len(modelNames) * len(suite.Prompts) * suite.Parameters.RunsPerPrompt
	completed := 0

	var modelRuns []ModelRun

	for _, model := range modelNames {
		mr := ModelRun{
			ModelName: model,
			StartedAt: time.Now(),
		}

		for _, prompt := range suite.Prompts {
			for run := 1; run <= suite.Parameters.RunsPerPrompt; run++ {
				// Apply cooldown between runs (but not before the first)
				if completed > 0 && suite.Parameters.CooldownSeconds > 0 {
					select {
					case <-ctx.Done():
						return modelRuns, ctx.Err()
					case <-time.After(time.Duration(suite.Parameters.CooldownSeconds) * time.Second):
					}
				}

				timeout := time.Duration(suite.Parameters.TimeoutSeconds) * time.Second
				runCtx, cancel := context.WithTimeout(ctx, timeout)

				stats, err := r.ollama.Run(runCtx, model, prompt.Text)
				cancel()

				result := models.PromptResult{
					PromptID:  prompt.ID,
					RunNumber: run,
					Stats:     stats,
					Success:   err == nil,
				}
				if err != nil {
					result.ErrorMessage = err.Error()
				}

				mr.Results = append(mr.Results, result)
				completed++

				if onProgress != nil {
					onProgress(Progress{
						Completed:  completed,
						Total:      total,
						Model:      model,
						PromptName: prompt.Name,
						RunNumber:  run,
						Success:    result.Success,
						EvalRate:   stats.EvalRate,
					})
				}
			}
		}

		mr.CompletedAt = time.Now()
		modelRuns = append(modelRuns, mr)
	}

	return modelRuns, nil
}

func FormatBriefSummary(runs []ModelRun) string {
	var summary string
	for _, mr := range runs {
		var totalEvalRate, totalPromptEvalRate float64
		var count int
		for _, r := range mr.Results {
			if r.Success {
				totalEvalRate += r.Stats.EvalRate
				totalPromptEvalRate += r.Stats.PromptEvalRate
				count++
			}
		}
		if count > 0 {
			summary += fmt.Sprintf("  %s: avg eval %.1f tok/s, avg prompt eval %.1f tok/s (%d/%d succeeded)\n",
				mr.ModelName,
				totalEvalRate/float64(count),
				totalPromptEvalRate/float64(count),
				count, len(mr.Results),
			)
		}
	}
	return summary
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
go test ./internal/benchmark/ -v
```

Expected: all PASS

- [ ] **Step 5: Commit**

```bash
git add internal/benchmark/runner.go internal/benchmark/runner_test.go
git commit -m "feat: add benchmark runner orchestrator with progress and cooldown"
```

---

### Task 14: CLI Commands — `suites` and `suite`

**Files:**
- Modify: `internal/cli/root.go`
- Create: `internal/cli/suites.go`

- [ ] **Step 1: Write the `suites` and `suite` commands**

Create `internal/cli/suites.go`:

```go
package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/joelhelbling/ollama-bench/internal/config"
	"github.com/joelhelbling/ollama-bench/internal/output"
	"github.com/joelhelbling/ollama-bench/pkg/apiclient"
)

func newSuitesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suites",
		Short: "List available benchmark suites",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.DefaultConfigDir())
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			client := apiclient.New(cfg.APIBaseURL, Version)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			suites, err := client.ListSuites(ctx)
			if err != nil {
				return fmt.Errorf("fetching suites: %w", err)
			}

			formatter := output.New(formatFlag, os.Stdout)
			return formatter.FormatSuites(suites)
		},
	}
}

func newSuiteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "suite [slug-or-id]",
		Short: "Show details of a benchmark suite",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.DefaultConfigDir())
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			client := apiclient.New(cfg.APIBaseURL, Version)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			suite, err := client.GetSuite(ctx, args[0])
			if err != nil {
				return fmt.Errorf("fetching suite: %w", err)
			}

			formatter := output.New(formatFlag, os.Stdout)
			return formatter.FormatSuite(suite)
		},
	}
}
```

- [ ] **Step 2: Register commands in root**

Modify `internal/cli/root.go` — add to `NewRootCmd()` before the return:

```go
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ollama-bench",
		Short: "Benchmark Ollama models and share results",
		Long:  "Run benchmarks against local Ollama models, publish results, and compare performance across models and hardware.",
	}

	cmd.PersistentFlags().StringVar(&formatFlag, "format", "table", "Output format: table or json")

	cmd.AddCommand(newSuitesCmd())
	cmd.AddCommand(newSuiteCmd())

	return cmd
}
```

- [ ] **Step 3: Build and verify help output**

```bash
go build ./cmd/ollama-bench && ./ollama-bench --help
```

Expected: shows `suites` and `suite` in available commands

- [ ] **Step 4: Commit**

```bash
git add internal/cli/
git commit -m "feat: add suites and suite CLI commands"
```

---

### Task 15: CLI Commands — `run`

**Files:**
- Create: `internal/cli/run.go`
- Create: `internal/ollama/adapter.go`
- Modify: `internal/cli/root.go`

- [ ] **Step 1: Create the ollama adapter (bridges OllamaRunner interface to real ollama)**

Create `internal/ollama/adapter.go`:

```go
package ollama

import (
	"context"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

// Adapter implements benchmark.OllamaRunner using the real ollama CLI
type Adapter struct{}

func NewAdapter() *Adapter {
	return &Adapter{}
}

func (a *Adapter) Run(ctx context.Context, model, prompt string) (models.OllamaStats, error) {
	result, err := Run(ctx, model, prompt)
	if err != nil {
		return models.OllamaStats{}, err
	}
	return result.Stats, nil
}
```

- [ ] **Step 2: Write the `run` command**

Create `internal/cli/run.go`:

```go
package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/joelhelbling/ollama-bench/internal/benchmark"
	"github.com/joelhelbling/ollama-bench/internal/cache"
	"github.com/joelhelbling/ollama-bench/internal/config"
	"github.com/joelhelbling/ollama-bench/internal/hardware"
	"github.com/joelhelbling/ollama-bench/internal/ollama"
	"github.com/joelhelbling/ollama-bench/internal/pending"
	"github.com/joelhelbling/ollama-bench/pkg/apiclient"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func newRunCmd() *cobra.Command {
	var modelsFlag string
	var suiteFlag string

	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run benchmarks against local Ollama models",
		RunE: func(cmd *cobra.Command, args []string) error {
			if modelsFlag == "" {
				return fmt.Errorf("--models is required (comma-separated list of model names)")
			}
			modelNames := strings.Split(modelsFlag, ",")
			for i := range modelNames {
				modelNames[i] = strings.TrimSpace(modelNames[i])
			}

			suiteSlug := "default"
			if suiteFlag != "" {
				suiteSlug = suiteFlag
			}

			// Load config
			cfg, err := config.Load(config.DefaultConfigDir())
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			// Fetch or load cached suite
			client := apiclient.New(cfg.APIBaseURL, Version)
			suiteCache := cache.NewSuiteCache(config.DefaultCacheDir())
			suite, err := fetchOrLoadSuite(client, suiteCache, suiteSlug)
			if err != nil {
				return fmt.Errorf("getting suite: %w", err)
			}

			fmt.Printf("Suite: %s (v%s) — %d prompts, %d runs each\n",
				suite.Name, suite.Version, len(suite.Prompts), suite.Parameters.RunsPerPrompt)
			fmt.Printf("Models: %s\n\n", strings.Join(modelNames, ", "))

			// Detect hardware
			hw, err := hardware.Detect()
			if err != nil {
				return fmt.Errorf("detecting hardware: %w", err)
			}
			fmt.Printf("Hardware: %s | %s | %.0f GB RAM", hw.CPUModel, hw.GPUName, hw.RAMGB)
			if hw.VRAMGB > 0 {
				fmt.Printf(" | %.0f GB VRAM", hw.VRAMGB)
			}
			fmt.Println("\n")

			// Run benchmarks
			adapter := ollama.NewAdapter()
			runner := benchmark.NewRunner(adapter)

			total := len(modelNames) * len(suite.Prompts) * suite.Parameters.RunsPerPrompt
			modelRuns, err := runner.Execute(context.Background(), suite, modelNames, func(p benchmark.Progress) {
				status := "ok"
				if !p.Success {
					status = "FAIL"
				}
				fmt.Printf("  [%d/%d] %s — %s #%d: %s (%.1f tok/s)\n",
					p.Completed, total, p.Model, p.PromptName, p.RunNumber, status, p.EvalRate)
			})
			if err != nil {
				return fmt.Errorf("benchmark execution: %w", err)
			}

			// Print brief summary
			fmt.Printf("\nSummary:\n%s", benchmark.FormatBriefSummary(modelRuns))

			// Load token
			token, _ := config.LoadToken(config.DefaultConfigDir())

			// Publish results for each model
			fmt.Println("\nPublishing results...")
			pendingStore := pending.NewStore(config.DefaultDataDir())

			for _, mr := range modelRuns {
				req := models.PublishRequest{
					SuiteID:      suite.ID,
					SuiteVersion: suite.Version,
					ModelName:    mr.ModelName,
					Hardware:     hw,
					Results:      mr.Results,
					Token:        token,
					StartedAt:    mr.StartedAt,
					CompletedAt:  mr.CompletedAt,
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				resp, err := client.PublishResults(ctx, req)
				cancel()

				if err != nil {
					fmt.Fprintf(os.Stderr, "  Failed to publish %s: %v\n", mr.ModelName, err)
					fmt.Fprintf(os.Stderr, "  Saving locally — run 'ollama-bench publish' later\n")
					pendingStore.Save(req)
					continue
				}

				// Save token if this is the first submission
				if token == "" && resp.Token != "" {
					config.SaveToken(config.DefaultConfigDir(), resp.Token)
					token = resp.Token
				}

				fmt.Printf("  %s: %s\n", mr.ModelName, resp.URL)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&modelsFlag, "models", "m", "", "Comma-separated list of models to benchmark (required)")
	cmd.Flags().StringVarP(&suiteFlag, "suite", "s", "", "Suite slug or ID (default: 'default')")

	return cmd
}

func fetchOrLoadSuite(client *apiclient.Client, suiteCache *cache.SuiteCache, slug string) (models.Suite, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	suite, err := client.GetSuite(ctx, slug)
	if err == nil {
		// Cache for offline use
		suiteCache.Save(suite)
		return suite, nil
	}

	// Try cached version
	fmt.Fprintf(os.Stderr, "Warning: could not fetch suite from server (%v), trying cache...\n", err)
	cached, cacheErr := suiteCache.Load(slug)
	if cacheErr != nil {
		return models.Suite{}, fmt.Errorf("suite not available from server (%v) or cache (%v)", err, cacheErr)
	}

	fmt.Fprintf(os.Stderr, "Using cached suite v%s\n", cached.Version)
	return cached, nil
}
```

- [ ] **Step 3: Register the run command in root**

Add to `NewRootCmd()` in `internal/cli/root.go`:

```go
	cmd.AddCommand(newRunCmd())
```

- [ ] **Step 4: Build and verify**

```bash
go build ./cmd/ollama-bench && ./ollama-bench run --help
```

Expected: shows `run` command help with `--models` and `--suite` flags

- [ ] **Step 5: Commit**

```bash
git add internal/cli/run.go internal/ollama/adapter.go internal/cli/root.go
git commit -m "feat: add run CLI command with suite fetch, benchmark, and publish"
```

---

### Task 16: CLI Commands — `compare`

**Files:**
- Create: `internal/cli/compare.go`
- Modify: `internal/cli/root.go`

- [ ] **Step 1: Write the `compare` command**

Create `internal/cli/compare.go`:

```go
package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/joelhelbling/ollama-bench/internal/config"
	"github.com/joelhelbling/ollama-bench/internal/output"
	"github.com/joelhelbling/ollama-bench/pkg/apiclient"
)

func newCompareCmd() *cobra.Command {
	var (
		modelFlag    string
		hardwareFlag string
		gpuFlag      string
		cpuFlag      string
		osFlag       string
		archFlag     string
		ramMinFlag   float64
		ramMaxFlag   float64
		vramMinFlag  float64
		vramMaxFlag  float64
		expandFlag   bool
	)

	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare benchmark results across hardware or models",
		Long: `Compare benchmark results. Use --model to compare hardware for a model,
or --hardware to compare models on specific hardware. Additional flags filter results.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if modelFlag == "" && hardwareFlag == "" {
				return fmt.Errorf("specify --model or --hardware (or both with additional filters)")
			}

			cfg, err := config.Load(config.DefaultConfigDir())
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			client := apiclient.New(cfg.APIBaseURL, Version)
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			params := apiclient.CompareParams{
				Model:      modelFlag,
				HardwareID: hardwareFlag,
				GPU:        gpuFlag,
				CPU:        cpuFlag,
				OS:         osFlag,
				Arch:       archFlag,
				RAMMin:     ramMinFlag,
				RAMMax:     ramMaxFlag,
				VRAMMin:    vramMinFlag,
				VRAMMax:    vramMaxFlag,
			}
			if expandFlag {
				params.Expand = "runs"
			}

			resp, err := client.GetComparison(ctx, params)
			if err != nil {
				return fmt.Errorf("fetching comparison: %w", err)
			}

			if len(resp.Groups) == 0 {
				fmt.Println("No results found matching your criteria.")
				return nil
			}

			formatter := output.New(formatFlag, os.Stdout)
			return formatter.FormatComparison(resp)
		},
	}

	cmd.Flags().StringVar(&modelFlag, "model", "", "Compare hardware configs for this model")
	cmd.Flags().StringVar(&hardwareFlag, "hardware", "", "Compare models on this hardware config ID")
	cmd.Flags().StringVar(&gpuFlag, "gpu", "", "Filter by GPU name")
	cmd.Flags().StringVar(&cpuFlag, "cpu", "", "Filter by CPU model")
	cmd.Flags().StringVar(&osFlag, "os", "", "Filter by OS")
	cmd.Flags().StringVar(&archFlag, "arch", "", "Filter by architecture")
	cmd.Flags().Float64Var(&ramMinFlag, "ram-min", 0, "Filter by minimum RAM (GB)")
	cmd.Flags().Float64Var(&ramMaxFlag, "ram-max", 0, "Filter by maximum RAM (GB)")
	cmd.Flags().Float64Var(&vramMinFlag, "vram-min", 0, "Filter by minimum VRAM (GB)")
	cmd.Flags().Float64Var(&vramMaxFlag, "vram-max", 0, "Filter by maximum VRAM (GB)")
	cmd.Flags().BoolVar(&expandFlag, "expand", false, "Show individual runs alongside summary stats")

	return cmd
}
```

- [ ] **Step 2: Register in root command**

Add to `NewRootCmd()` in `internal/cli/root.go`:

```go
	cmd.AddCommand(newCompareCmd())
```

- [ ] **Step 3: Build and verify**

```bash
go build ./cmd/ollama-bench && ./ollama-bench compare --help
```

Expected: shows all filter flags

- [ ] **Step 4: Commit**

```bash
git add internal/cli/compare.go internal/cli/root.go
git commit -m "feat: add compare CLI command with hardware/model filters"
```

---

### Task 17: CLI Commands — `results` and `publish`

**Files:**
- Create: `internal/cli/results.go`
- Create: `internal/cli/publish.go`
- Modify: `internal/cli/root.go`

- [ ] **Step 1: Write the `results` command**

Create `internal/cli/results.go`:

```go
package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/joelhelbling/ollama-bench/internal/config"
	"github.com/joelhelbling/ollama-bench/internal/output"
	"github.com/joelhelbling/ollama-bench/pkg/apiclient"
)

func newResultsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "results [run-id]",
		Short: "Show results for a specific benchmark run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.DefaultConfigDir())
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			client := apiclient.New(cfg.APIBaseURL, Version)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			run, err := client.GetRun(ctx, args[0])
			if err != nil {
				return fmt.Errorf("fetching run: %w", err)
			}

			formatter := output.New(formatFlag, os.Stdout)
			return formatter.FormatRun(run)
		},
	}
}
```

- [ ] **Step 2: Write the `publish` command**

Create `internal/cli/publish.go`:

```go
package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/joelhelbling/ollama-bench/internal/config"
	"github.com/joelhelbling/ollama-bench/internal/pending"
	"github.com/joelhelbling/ollama-bench/pkg/apiclient"
)

func newPublishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "publish",
		Short: "Publish pending benchmark results that failed to upload",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.DefaultConfigDir())
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}

			store := pending.NewStore(config.DefaultDataDir())
			items, err := store.List()
			if err != nil {
				return fmt.Errorf("listing pending results: %w", err)
			}

			if len(items) == 0 {
				fmt.Println("No pending results to publish.")
				return nil
			}

			fmt.Printf("Found %d pending result(s). Publishing...\n", len(items))

			client := apiclient.New(cfg.APIBaseURL, Version)
			token, _ := config.LoadToken(config.DefaultConfigDir())

			var failures int
			for _, item := range items {
				req := item.Request
				if req.Token == "" && token != "" {
					req.Token = token
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				resp, err := client.PublishResults(ctx, req)
				cancel()

				if err != nil {
					fmt.Printf("  FAIL: %s — %v\n", req.ModelName, err)
					failures++
					continue
				}

				// Save token if first
				if token == "" && resp.Token != "" {
					config.SaveToken(config.DefaultConfigDir(), resp.Token)
					token = resp.Token
				}

				fmt.Printf("  OK: %s — %s\n", req.ModelName, resp.URL)
				store.Remove(item.Filename)
			}

			if failures > 0 {
				return fmt.Errorf("%d result(s) failed to publish — try again later", failures)
			}

			fmt.Println("All pending results published successfully.")
			return nil
		},
	}
}
```

- [ ] **Step 3: Register commands in root**

Add to `NewRootCmd()` in `internal/cli/root.go`:

```go
	cmd.AddCommand(newResultsCmd())
	cmd.AddCommand(newPublishCmd())
```

- [ ] **Step 4: Build and verify**

```bash
go build ./cmd/ollama-bench && ./ollama-bench --help
```

Expected: all commands listed — `run`, `compare`, `results`, `suites`, `suite`, `publish`

- [ ] **Step 5: Commit**

```bash
git add internal/cli/results.go internal/cli/publish.go internal/cli/root.go
git commit -m "feat: add results and publish CLI commands"
```

---

### Task 18: Integration Test — Full Run Flow

**Files:**
- Create: `internal/cli/run_test.go`

- [ ] **Step 1: Write the integration test**

Create `internal/cli/run_test.go`:

```go
package cli_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/cli"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestRunCommandPublishesResults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Set up a fake backend
	var receivedRequest models.PublishRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/suites/default":
			json.NewEncoder(w).Encode(models.Suite{
				ID:      "suite-1",
				Slug:    "default",
				Name:    "Test Suite",
				Version: "1.0.0",
				Prompts: []models.Prompt{
					{ID: "p1", Name: "simple", Category: "general", Text: "What is 2+2?"},
				},
				Parameters: models.SuiteParameters{
					RunsPerPrompt:  1,
					TimeoutSeconds: 60,
				},
			})
		case "/api/v1/results":
			json.NewDecoder(r.Body).Decode(&receivedRequest)
			json.NewEncoder(w).Encode(models.PublishResponse{
				RunID: "run-123",
				Token: "tok-abc",
				URL:   "https://example.com/results/run-123",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Configure CLI to use test server
	configDir := t.TempDir()
	configFile := filepath.Join(configDir, "config.yaml")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(configFile, []byte("api_base_url: "+server.URL), 0644)

	// Note: this test requires ollama to be running locally with the model available.
	// In CI, mock the ollama binary or skip.
	t.Log("This test requires 'ollama run llama3.2:1b' to work locally")

	// The actual test would execute the run command.
	// For unit testing without ollama, see benchmark/runner_test.go which uses FakeOllamaRunner.
}
```

- [ ] **Step 2: Run test**

```bash
go test ./internal/cli/ -v -short
```

Expected: SKIP (short mode)

- [ ] **Step 3: Commit**

```bash
git add internal/cli/run_test.go
git commit -m "test: add integration test scaffold for run command"
```

---

### Task 19: Build, Version Injection, and README

**Files:**
- Create: `Makefile`
- Create: `README.md`
- Modify: `cmd/ollama-bench/main.go`

- [ ] **Step 1: Add version flag to main**

Update `cmd/ollama-bench/main.go`:

```go
package main

import (
	"github.com/joelhelbling/ollama-bench/internal/cli"
)

// Set via ldflags at build time
var version = "dev"

func main() {
	cli.Version = version
	cli.Execute()
}
```

- [ ] **Step 2: Create Makefile**

Create `Makefile`:

```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

.PHONY: build test clean

build:
	go build $(LDFLAGS) -o ollama-bench ./cmd/ollama-bench

test:
	go test -short ./...

test-integration:
	go test ./...

clean:
	rm -f ollama-bench
```

- [ ] **Step 3: Create README**

Create `README.md`:

```markdown
# ollama-bench

Benchmark your local Ollama models and compare performance across hardware.

## Install

### From source

```bash
git clone https://github.com/joelhelbling/ollama-bench.git
cd ollama-bench
make build
```

### Homebrew (coming soon)

```bash
brew install joelhelbling/tap/ollama-bench
```

## Usage

### Run benchmarks

```bash
ollama-bench run --models llama3:8b,mistral:7b
```

Runs the default benchmark suite against the specified models, publishes results, and prints a link to view them.

### Compare hardware for a model

```bash
ollama-bench compare --model llama3:8b
```

### Compare models on hardware

```bash
ollama-bench compare --hardware <config-id>
```

### Filter comparisons

```bash
ollama-bench compare --model llama3:8b --gpu "RTX 4090" --ram-min 32
```

### View a specific run

```bash
ollama-bench results <run-id>
```

### List benchmark suites

```bash
ollama-bench suites
```

### JSON output (for agentic use)

All commands support `--format json` for structured output:

```bash
ollama-bench compare --model llama3:8b --format json
```

## Configuration

Config file: `~/.config/ollama-bench/config.yaml`

```yaml
api_base_url: https://bench.ollama.example.com
```

## Building from source

```bash
make build    # build binary
make test     # run unit tests
```
```

- [ ] **Step 4: Build and verify version**

```bash
make build && ./ollama-bench --help
```

Expected: builds successfully, shows help

- [ ] **Step 5: Run full test suite**

```bash
make test
```

Expected: all unit tests pass

- [ ] **Step 6: Commit**

```bash
git add Makefile README.md cmd/ollama-bench/main.go
git commit -m "feat: add build system, version injection, and README"
```

---

## Self-Review Notes

**Spec coverage check:**
- Suite model, versioning, fetching: Tasks 2, 9, 14 (suites command), 15 (run fetches suite)
- Hardware detection + fingerprinting: Tasks 3, 7
- Ollama execution + parsing: Tasks 5, 6
- API client for all endpoints: Task 8
- CLI commands (run, compare, results, suites, suite, publish): Tasks 14-17
- Output formats (table + JSON): Task 12
- Offline support (caching, pending store): Tasks 9, 11
- Config + token management: Task 10
- CLI version header: Task 8 (client sends it), Task 19 (version injection)
- Build from source (trust story): Task 19

**Missing from spec:** Signed binary distribution (Sigstore/cosign) — this is a release pipeline concern, not implementation code. Would be a CI/CD task (GitHub Actions with goreleaser + cosign). Noted but not included as a code task.

**Type consistency:** All types reference `pkg/models` consistently. `OllamaStats`, `HardwareInfo`, `PublishRequest`, `CompareResponse` etc. are defined once and used throughout. The `benchmark.OllamaRunner` interface matches the `ollama.Adapter` implementation.

**Placeholder scan:** No TBDs, TODOs, or "implement later" found. Integration test in Task 18 is explicitly scoped as a scaffold (requires real ollama), with the real logic tested via `FakeOllamaRunner` in Task 13.
