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
			fmt.Print("\n\n")

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
