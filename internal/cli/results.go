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
