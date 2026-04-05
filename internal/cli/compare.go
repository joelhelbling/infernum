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
