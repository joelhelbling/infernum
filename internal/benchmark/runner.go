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
	Completed  int
	Total      int
	Model      string
	PromptName string
	RunNumber  int
	Success    bool
	EvalRate   float64
}

type ModelRun struct {
	ModelName   string
	Results     []models.PromptResult
	StartedAt   time.Time
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
