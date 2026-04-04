package pending_test

import (
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/pending"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestSaveAndList(t *testing.T) {
	dir := t.TempDir()
	store := pending.NewStore(dir)

	req := models.PublishRequest{
		SuiteID:   "suite-1",
		ModelName: "llama3:8b",
		Hardware: models.HardwareInfo{
			CPUModel: "AMD Ryzen 9",
			RAMGB:    64,
		},
		Results: []models.PromptResult{
			{PromptID: "p1", RunNumber: 1, Success: true},
		},
	}

	if err := store.Save(req); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	items, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 pending item, got %d", len(items))
	}

	if items[0].Request.ModelName != "llama3:8b" {
		t.Errorf("expected model 'llama3:8b', got %q", items[0].Request.ModelName)
	}
}

func TestRemove(t *testing.T) {
	dir := t.TempDir()
	store := pending.NewStore(dir)

	store.Save(models.PublishRequest{ModelName: "model1"})
	store.Save(models.PublishRequest{ModelName: "model2"})

	items, _ := store.List()
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	if err := store.Remove(items[0].Filename); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	items, _ = store.List()
	if len(items) != 1 {
		t.Fatalf("expected 1 item after removal, got %d", len(items))
	}
}

func TestListEmpty(t *testing.T) {
	dir := t.TempDir()
	store := pending.NewStore(dir)

	items, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items, got %d", len(items))
	}
}
