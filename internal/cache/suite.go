package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type SuiteCache struct {
	dir string
}

func NewSuiteCache(dir string) *SuiteCache {
	return &SuiteCache{dir: dir}
}

func (c *SuiteCache) Save(suite models.Suite) error {
	if err := os.MkdirAll(c.dir, 0755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}

	// Remove older versions of the same suite
	entries, _ := os.ReadDir(c.dir)
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), suite.Slug+"-") && strings.HasSuffix(e.Name(), ".json") {
			os.Remove(filepath.Join(c.dir, e.Name()))
		}
	}

	filename := fmt.Sprintf("%s-%s.json", suite.Slug, suite.Version)
	path := filepath.Join(c.dir, filename)

	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling suite: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (c *SuiteCache) Load(slug string) (models.Suite, error) {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return models.Suite{}, fmt.Errorf("reading cache dir: %w", err)
	}

	for _, e := range entries {
		if strings.HasPrefix(e.Name(), slug+"-") && strings.HasSuffix(e.Name(), ".json") {
			path := filepath.Join(c.dir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				return models.Suite{}, fmt.Errorf("reading cached suite: %w", err)
			}
			var suite models.Suite
			if err := json.Unmarshal(data, &suite); err != nil {
				return models.Suite{}, fmt.Errorf("parsing cached suite: %w", err)
			}
			return suite, nil
		}
	}

	return models.Suite{}, fmt.Errorf("suite %q not found in cache", slug)
}
