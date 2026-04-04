package pending

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/joelhelbling/ollama-bench/pkg/models"
)

type Store struct {
	dir string
}

type PendingItem struct {
	Filename string
	Request  models.PublishRequest
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

func (s *Store) Save(req models.PublishRequest) error {
	if err := os.MkdirAll(s.dir, 0755); err != nil {
		return fmt.Errorf("creating pending dir: %w", err)
	}

	filename := fmt.Sprintf("pending_%s.json", time.Now().Format("20060102_150405.000000000"))
	path := filepath.Join(s.dir, filename)

	data, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling pending result: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}

func (s *Store) List() ([]PendingItem, error) {
	entries, err := os.ReadDir(s.dir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading pending dir: %w", err)
	}

	var items []PendingItem
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}

		path := filepath.Join(s.dir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var req models.PublishRequest
		if err := json.Unmarshal(data, &req); err != nil {
			continue
		}

		items = append(items, PendingItem{
			Filename: e.Name(),
			Request:  req,
		})
	}

	return items, nil
}

func (s *Store) Remove(filename string) error {
	return os.Remove(filepath.Join(s.dir, filename))
}
