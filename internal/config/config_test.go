package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joelhelbling/ollama-bench/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Default()
	if cfg.APIBaseURL == "" {
		t.Error("default API base URL should not be empty")
	}
}

func TestLoadCreatesDefault(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "ollama-bench")

	cfg, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.APIBaseURL == "" {
		t.Error("loaded config should have default API URL")
	}
}

func TestTokenPersistence(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "ollama-bench")

	if err := config.SaveToken(configDir, "tok-abc-123"); err != nil {
		t.Fatalf("SaveToken failed: %v", err)
	}

	token, err := config.LoadToken(configDir)
	if err != nil {
		t.Fatalf("LoadToken failed: %v", err)
	}

	if token != "tok-abc-123" {
		t.Errorf("expected token 'tok-abc-123', got %q", token)
	}
}

func TestLoadTokenMissing(t *testing.T) {
	dir := t.TempDir()

	token, err := config.LoadToken(dir)
	if err != nil {
		t.Fatalf("LoadToken should not error for missing token: %v", err)
	}
	if token != "" {
		t.Errorf("expected empty token, got %q", token)
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	configDir := filepath.Join(dir, "ollama-bench")

	cfg := config.Config{APIBaseURL: "https://custom.example.com"}
	if err := config.Save(configDir, cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := config.Load(configDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.APIBaseURL != "https://custom.example.com" {
		t.Errorf("expected custom URL, got %q", loaded.APIBaseURL)
	}

	// Verify file was created
	if _, err := os.Stat(filepath.Join(configDir, "config.yaml")); os.IsNotExist(err) {
		t.Fatal("config file should exist")
	}
}
