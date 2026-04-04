package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	DefaultAPIBaseURL = "https://bench.ollama.example.com"
	configFileName    = "config.yaml"
	tokenFileName     = "token"
)

type Config struct {
	APIBaseURL string `yaml:"api_base_url"`
}

func Default() Config {
	return Config{
		APIBaseURL: DefaultAPIBaseURL,
	}
}

func Load(configDir string) (Config, error) {
	path := filepath.Join(configDir, configFileName)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Default(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = DefaultAPIBaseURL
	}

	return cfg, nil
}

func Save(configDir string, cfg Config) error {
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(filepath.Join(configDir, configFileName), data, 0644)
}

func SaveToken(configDir, token string) error {
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	return os.WriteFile(filepath.Join(configDir, tokenFileName), []byte(token), 0600)
}

func LoadToken(configDir string) (string, error) {
	data, err := os.ReadFile(filepath.Join(configDir, tokenFileName))
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reading token: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func DefaultConfigDir() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "ollama-bench")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "ollama-bench")
}

func DefaultCacheDir() string {
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "ollama-bench", "suites")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "ollama-bench", "suites")
}

func DefaultDataDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "ollama-bench", "pending")
}
