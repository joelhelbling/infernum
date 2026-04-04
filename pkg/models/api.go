package models

import "time"

// --- Requests ---

type PublishRequest struct {
	SuiteID      string         `json:"suite_id"`
	SuiteVersion string         `json:"suite_version"`
	ModelName    string         `json:"model_name"`
	Hardware     HardwareInfo   `json:"hardware"`
	Results      []PromptResult `json:"results"`
	Token        string         `json:"token,omitempty"`
	StartedAt    time.Time      `json:"started_at"`
	CompletedAt  time.Time      `json:"completed_at"`
}

// --- Responses ---

type PublishResponse struct {
	RunID string `json:"run_id"`
	Token string `json:"token"`
	URL   string `json:"url"`
}

type RunResponse struct {
	ID           string         `json:"id"`
	SuiteID      string         `json:"suite_id"`
	SuiteVersion string         `json:"suite_version"`
	ModelName    string         `json:"model_name"`
	Hardware     HardwareInfo   `json:"hardware"`
	Results      []PromptResult `json:"results"`
	StartedAt    time.Time      `json:"started_at"`
	CompletedAt  time.Time      `json:"completed_at"`
}

type CompareResponse struct {
	SuiteID      string         `json:"suite_id"`
	SuiteVersion string         `json:"suite_version"`
	Groups       []CompareGroup `json:"groups"`
}

type CompareGroup struct {
	ModelName string        `json:"model_name,omitempty"`
	Hardware  *HardwareInfo `json:"hardware,omitempty"`
	Stats     SummaryStats  `json:"stats"`
	RunCount  int           `json:"run_count"`
	Runs      []RunResponse `json:"runs,omitempty"`
}

type SummaryStats struct {
	MeanEvalRate         float64 `json:"mean_eval_rate"`
	StddevEvalRate       float64 `json:"stddev_eval_rate"`
	MeanPromptEvalRate   float64 `json:"mean_prompt_eval_rate"`
	StddevPromptEvalRate float64 `json:"stddev_prompt_eval_rate"`
}

type HardwareConfigSummary struct {
	ID       string       `json:"id"`
	Hardware HardwareInfo `json:"hardware"`
	RunCount int          `json:"run_count"`
}

type ModelSummary struct {
	Name     string `json:"name"`
	RunCount int    `json:"run_count"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
