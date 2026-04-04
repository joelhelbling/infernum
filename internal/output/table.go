package output

import (
	"fmt"
	"io"

	"github.com/joelhelbling/ollama-bench/pkg/models"
	"github.com/olekukonko/tablewriter"
)

type TableFormatter struct {
	w io.Writer
}

func NewTableFormatter(w io.Writer) *TableFormatter {
	return &TableFormatter{w: w}
}

func (f *TableFormatter) FormatComparison(resp models.CompareResponse) error {
	table := tablewriter.NewWriter(f.w)
	table.Header("GPU", "CPU", "RAM (GB)", "Eval Rate", "Stddev", "Prompt Eval Rate", "Runs")

	for _, g := range resp.Groups {
		gpuName := ""
		cpuModel := ""
		ramGB := ""

		if g.Hardware != nil {
			gpuName = g.Hardware.GPUName
			cpuModel = g.Hardware.CPUModel
			ramGB = fmt.Sprintf("%.0f", g.Hardware.RAMGB)
		}

		if err := table.Append(
			gpuName,
			cpuModel,
			ramGB,
			fmt.Sprintf("%.1f tok/s", g.Stats.MeanEvalRate),
			fmt.Sprintf("%.1f", g.Stats.StddevEvalRate),
			fmt.Sprintf("%.1f tok/s", g.Stats.MeanPromptEvalRate),
			fmt.Sprintf("%d", g.RunCount),
		); err != nil {
			return fmt.Errorf("appending row: %w", err)
		}
	}

	return table.Render()
}

func (f *TableFormatter) FormatRun(run models.RunResponse) error {
	fmt.Fprintf(f.w, "Run: %s\n", run.ID)
	fmt.Fprintf(f.w, "Model: %s\n", run.ModelName)
	if run.Hardware.GPUName != "" {
		fmt.Fprintf(f.w, "GPU: %s\n", run.Hardware.GPUName)
	}
	fmt.Fprintf(f.w, "CPU: %s | RAM: %.0f GB\n\n", run.Hardware.CPUModel, run.Hardware.RAMGB)

	table := tablewriter.NewWriter(f.w)
	table.Header("Prompt", "Run #", "Eval Rate", "Prompt Eval Rate", "Total Duration", "Success")

	for _, r := range run.Results {
		success := "yes"
		if !r.Success {
			success = "FAIL"
		}
		if err := table.Append(
			r.PromptID,
			fmt.Sprintf("%d", r.RunNumber),
			fmt.Sprintf("%.1f tok/s", r.Stats.EvalRate),
			fmt.Sprintf("%.1f tok/s", r.Stats.PromptEvalRate),
			fmt.Sprintf("%.2fs", r.Stats.TotalDurationS),
			success,
		); err != nil {
			return fmt.Errorf("appending row: %w", err)
		}
	}

	return table.Render()
}

func (f *TableFormatter) FormatSuites(suites []models.Suite) error {
	table := tablewriter.NewWriter(f.w)
	table.Header("Slug", "Name", "Version", "Prompts")

	for _, s := range suites {
		if err := table.Append(
			s.Slug,
			s.Name,
			s.Version,
			fmt.Sprintf("%d", len(s.Prompts)),
		); err != nil {
			return fmt.Errorf("appending row: %w", err)
		}
	}

	return table.Render()
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
	table.Header("#", "Name", "Category", "Prompt (truncated)")

	for i, p := range suite.Prompts {
		text := p.Text
		if len(text) > 60 {
			text = text[:57] + "..."
		}
		if err := table.Append(
			fmt.Sprintf("%d", i+1),
			p.Name,
			p.Category,
			text,
		); err != nil {
			return fmt.Errorf("appending row: %w", err)
		}
	}

	return table.Render()
}
