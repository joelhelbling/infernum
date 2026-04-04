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
