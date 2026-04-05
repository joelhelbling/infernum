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
