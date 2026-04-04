package models

type Suite struct {
	ID          string          `json:"id"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Version     string          `json:"version"`
	Prompts     []Prompt        `json:"prompts"`
	Parameters  SuiteParameters `json:"parameters"`
}

type Prompt struct {
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	Category              string `json:"category"`
	Text                  string `json:"text"`
	ExpectedTokenRangeMin *int   `json:"expected_token_range_min,omitempty"`
	ExpectedTokenRangeMax *int   `json:"expected_token_range_max,omitempty"`
	SortOrder             int    `json:"sort_order"`
}

type SuiteParameters struct {
	RunsPerPrompt   int `json:"runs_per_prompt"`
	TimeoutSeconds  int `json:"timeout_seconds"`
	CooldownSeconds int `json:"cooldown_seconds"`
}
