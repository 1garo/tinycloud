package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/1garo/tinycloud/internal/app"
)

type Store struct {
	path string
	mu   sync.RWMutex
	apps []app.App
}

func New(path string) *Store {
	return &Store{path: path, apps: []app.App{}}
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read state: %w", err)
	}
	if err := json.Unmarshal(data, &s.apps); err != nil {
		return fmt.Errorf("decode state: %w", err)
	}
	if s.apps == nil {
		s.apps = []app.App{}
	}
	return nil
}

func (s *Store) List() []app.App {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps := make([]app.App, len(s.apps))
	copy(apps, s.apps)
	sort.Slice(apps, func(i, j int) bool {
		return apps[i].UpdatedAt.After(apps[j].UpdatedAt)
	})
	return apps
}

func (s *Store) Add(item app.App) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.apps = append(s.apps, item)
	return s.saveLocked()
}

func (s *Store) Update(item app.App) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for index := range s.apps {
		if s.apps[index].ID == item.ID {
			s.apps[index] = item
			return s.saveLocked()
		}
	}
	return fmt.Errorf("app %q not found", item.ID)
}

func (s *Store) saveLocked() error {
	data, err := json.MarshalIndent(s.apps, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	if err := os.WriteFile(s.path, data, 0o600); err != nil {
		return fmt.Errorf("write state: %w", err)
	}
	return nil
}
