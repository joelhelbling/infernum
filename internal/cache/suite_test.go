package cache_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/cache"
	"github.com/joelhelbling/ollama-bench/pkg/models"
)

func TestSaveThenLoad(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewSuiteCache(dir)

	suite := models.Suite{
		ID:      "suite-1",
		Slug:    "default",
		Version: "1.0.0",
		Prompts: []models.Prompt{
			{ID: "p1", Name: "test", Category: "general", Text: "Hello"},
		},
		Parameters: models.SuiteParameters{RunsPerPrompt: 3},
	}

	if err := c.Save(suite); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	expectedPath := filepath.Join(dir, "default-1.0.0.json")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("expected cache file at %s", expectedPath)
	}

	loaded, err := c.Load("default")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", loaded.Version)
	}
	if len(loaded.Prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(loaded.Prompts))
	}
}

func TestLoadNotCached(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewSuiteCache(dir)

	_, err := c.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing cache")
	}
}

func TestSaveOverwritesOlderVersion(t *testing.T) {
	dir := t.TempDir()
	c := cache.NewSuiteCache(dir)

	suite1 := models.Suite{Slug: "default", Version: "1.0.0"}
	suite2 := models.Suite{Slug: "default", Version: "1.1.0"}

	c.Save(suite1)
	c.Save(suite2)

	loaded, err := c.Load("default")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Version != "1.1.0" {
		t.Errorf("expected version '1.1.0', got %q", loaded.Version)
	}
}
