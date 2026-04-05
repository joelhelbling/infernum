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
