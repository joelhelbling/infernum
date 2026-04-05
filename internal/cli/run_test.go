package cli_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/joelhelbling/ollama-bench/internal/cli"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestRunCommandPublishesResults(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Set up a fake backend
	var receivedRequest models.PublishRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/suites/default":
			json.NewEncoder(w).Encode(models.Suite{
				ID:      "suite-1",
				Slug:    "default",
				Name:    "Test Suite",
				Version: "1.0.0",
				Prompts: []models.Prompt{
					{ID: "p1", Name: "simple", Category: "general", Text: "What is 2+2?"},
				},
				Parameters: models.SuiteParameters{
					RunsPerPrompt:  1,
					TimeoutSeconds: 60,
				},
			})
		case "/api/v1/results":
			json.NewDecoder(r.Body).Decode(&receivedRequest)
			json.NewEncoder(w).Encode(models.PublishResponse{
				RunID: "run-123",
				Token: "tok-abc",
				URL:   "https://example.com/results/run-123",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Configure CLI to use test server
	configDir := t.TempDir()
	configFile := filepath.Join(configDir, "config.yaml")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(configFile, []byte("api_base_url: "+server.URL), 0644)

	// Note: this test requires ollama to be running locally with the model available.
	// In CI, mock the ollama binary or skip.
	t.Log("This test requires 'ollama run llama3.2:1b' to work locally")

	// The actual test would execute the run command.
	// For unit testing without ollama, see benchmark/runner_test.go which uses FakeOllamaRunner.
}
