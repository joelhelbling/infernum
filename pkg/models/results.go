package models

type OllamaStats struct {
	TotalDurationS      float64 `json:"total_duration_s"`
	LoadDurationS       float64 `json:"load_duration_s"`
	PromptEvalCount     int     `json:"prompt_eval_count"`
	PromptEvalDurationS float64 `json:"prompt_eval_duration_s"`
	PromptEvalRate      float64 `json:"prompt_eval_rate"`
	EvalCount           int     `json:"eval_count"`
	EvalDurationS       float64 `json:"eval_duration_s"`
	EvalRate            float64 `json:"eval_rate"`
}

type PromptResult struct {
	PromptID     string      `json:"prompt_id"`
	RunNumber    int         `json:"run_number"`
	Stats        OllamaStats `json:"stats"`
	Success      bool        `json:"success"`
	ErrorMessage string      `json:"error_message,omitempty"`
}
