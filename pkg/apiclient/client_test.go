package apiclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/joelhelbling/ollama-bench/pkg/apiclient"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestGetSuite(t *testing.T) {
	suite := models.Suite{
		ID:      "suite-1",
		Slug:    "default",
		Name:    "Default Suite",
		Version: "1.0.0",
		Prompts: []models.Prompt{
			{ID: "p1", Name: "test", Category: "general", Text: "Hello"},
		},
		Parameters: models.SuiteParameters{RunsPerPrompt: 3, TimeoutSeconds: 300, CooldownSeconds: 5},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/suites/default" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(suite)
	}))
	defer server.Close()

	client := apiclient.New(server.URL, "test-version")
	got, err := client.GetSuite(context.Background(), "default")
	if err != nil {
		t.Fatalf("GetSuite failed: %v", err)
	}

	if got.Slug != "default" {
		t.Errorf("expected slug 'default', got %q", got.Slug)
	}
	if len(got.Prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(got.Prompts))
	}
}

func TestPublishResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/results" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-OllamaBench-Version") != "test-version" {
			t.Errorf("missing version header")
		}

		var req models.PublishRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.ModelName != "llama3:8b" {
			t.Errorf("expected model 'llama3:8b', got %q", req.ModelName)
		}

		json.NewEncoder(w).Encode(models.PublishResponse{
			RunID: "run-123",
			Token: "tok-abc",
			URL:   "https://bench.example.com/results/run-123",
		})
	}))
	defer server.Close()

	client := apiclient.New(server.URL, "test-version")
	resp, err := client.PublishResults(context.Background(), models.PublishRequest{
		ModelName: "llama3:8b",
	})
	if err != nil {
		t.Fatalf("PublishResults failed: %v", err)
	}

	if resp.RunID != "run-123" {
		t.Errorf("expected run ID 'run-123', got %q", resp.RunID)
	}
	if resp.Token != "tok-abc" {
		t.Errorf("expected token 'tok-abc', got %q", resp.Token)
	}
}

func TestGetComparison(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/compare" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("model") != "llama3:8b" {
			t.Errorf("expected model query param")
		}

		json.NewEncoder(w).Encode(models.CompareResponse{
			Groups: []models.CompareGroup{
				{
					Hardware: &models.HardwareInfo{GPUName: "RTX 4090"},
					Stats:    models.SummaryStats{MeanEvalRate: 85.2},
					RunCount: 5,
				},
			},
		})
	}))
	defer server.Close()

	client := apiclient.New(server.URL, "test-version")
	params := apiclient.CompareParams{Model: "llama3:8b"}
	resp, err := client.GetComparison(context.Background(), params)
	if err != nil {
		t.Fatalf("GetComparison failed: %v", err)
	}

	if len(resp.Groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(resp.Groups))
	}
	if resp.Groups[0].Stats.MeanEvalRate != 85.2 {
		t.Errorf("expected mean eval rate 85.2, got %f", resp.Groups[0].Stats.MeanEvalRate)
	}
}

func TestClientHandlesServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.ErrorResponse{Error: "internal error"})
	}))
	defer server.Close()

	client := apiclient.New(server.URL, "test-version")
	_, err := client.GetSuite(context.Background(), "default")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
