package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type Client struct {
	baseURL    string
	version    string
	httpClient *http.Client
}

func New(baseURL, version string) *Client {
	return &Client{
		baseURL: baseURL,
		version: version,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetSuite(ctx context.Context, slugOrID string) (models.Suite, error) {
	var suite models.Suite
	err := c.get(ctx, fmt.Sprintf("/api/v1/suites/%s", slugOrID), &suite)
	return suite, err
}

func (c *Client) ListSuites(ctx context.Context) ([]models.Suite, error) {
	var suites []models.Suite
	err := c.get(ctx, "/api/v1/suites", &suites)
	return suites, err
}

func (c *Client) PublishResults(ctx context.Context, req models.PublishRequest) (models.PublishResponse, error) {
	var resp models.PublishResponse
	err := c.post(ctx, "/api/v1/results", req, &resp)
	return resp, err
}

func (c *Client) GetRun(ctx context.Context, runID string) (models.RunResponse, error) {
	var resp models.RunResponse
	err := c.get(ctx, fmt.Sprintf("/api/v1/results/%s", runID), &resp)
	return resp, err
}

type CompareParams struct {
	Model      string
	HardwareID string
	GPU        string
	CPU        string
	RAMMin     float64
	RAMMax     float64
	VRAMMin    float64
	VRAMMax    float64
	OS         string
	Arch       string
	SuiteID    string
	Expand     string
}

func (c *Client) GetComparison(ctx context.Context, params CompareParams) (models.CompareResponse, error) {
	query := url.Values{}
	if params.Model != "" {
		query.Set("model", params.Model)
	}
	if params.HardwareID != "" {
		query.Set("hardware_id", params.HardwareID)
	}
	if params.GPU != "" {
		query.Set("gpu", params.GPU)
	}
	if params.CPU != "" {
		query.Set("cpu", params.CPU)
	}
	if params.RAMMin > 0 {
		query.Set("ram_min", fmt.Sprintf("%.0f", params.RAMMin))
	}
	if params.RAMMax > 0 {
		query.Set("ram_max", fmt.Sprintf("%.0f", params.RAMMax))
	}
	if params.VRAMMin > 0 {
		query.Set("vram_min", fmt.Sprintf("%.0f", params.VRAMMin))
	}
	if params.VRAMMax > 0 {
		query.Set("vram_max", fmt.Sprintf("%.0f", params.VRAMMax))
	}
	if params.OS != "" {
		query.Set("os", params.OS)
	}
	if params.Arch != "" {
		query.Set("arch", params.Arch)
	}
	if params.SuiteID != "" {
		query.Set("suite_id", params.SuiteID)
	}
	if params.Expand != "" {
		query.Set("expand", params.Expand)
	}

	path := "/api/v1/compare"
	if encoded := query.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp models.CompareResponse
	err := c.get(ctx, path, &resp)
	return resp, err
}

func (c *Client) GetHardwareConfigs(ctx context.Context) ([]models.HardwareConfigSummary, error) {
	var configs []models.HardwareConfigSummary
	err := c.get(ctx, "/api/v1/hardware-configs", &configs)
	return configs, err
}

func (c *Client) GetModels(ctx context.Context) ([]models.ModelSummary, error) {
	var modelList []models.ModelSummary
	err := c.get(ctx, "/api/v1/models", &modelList)
	return modelList, err
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-OllamaBench-Version", c.version)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *Client) post(ctx context.Context, path string, body any, result any) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-OllamaBench-Version", c.version)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
