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
