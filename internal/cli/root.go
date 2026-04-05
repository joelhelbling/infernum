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

	cmd.AddCommand(newSuitesCmd())
	cmd.AddCommand(newSuiteCmd())

	return cmd
}

func Execute() {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
