package memory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zieckey/ai-study/internal/trace"
)

type Entry struct {
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updated_at"`
}

type Store struct {
	path string
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) Path() string {
	return s.path
}

func (s *Store) Set(ctx context.Context, key string, value string) error {
	trace.Log(ctx, "memory.Store.Set", map[string]any{"path": s.path, "key": key, "value": value})
	key = strings.TrimSpace(key)
	if key == "" {
		return errors.New("memory key is required")
	}

	entries, err := s.load(ctx)
	if err != nil {
		return err
	}
	entries[key] = Entry{Key: key, Value: value, UpdatedAt: time.Now().Format(time.RFC3339)}
	return s.save(ctx, entries)
}

func (s *Store) Get(ctx context.Context, key string) (Entry, bool, error) {
	trace.Log(ctx, "memory.Store.Get", map[string]any{"path": s.path, "key": key})
	entries, err := s.load(ctx)
	if err != nil {
		return Entry{}, false, err
	}
	entry, ok := entries[strings.TrimSpace(key)]
	return entry, ok, nil
}

func (s *Store) List(ctx context.Context) ([]Entry, error) {
	trace.Log(ctx, "memory.Store.List", map[string]any{"path": s.path})
	entries, err := s.load(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		list = append(list, entry)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Key < list[j].Key })
	return list, nil
}

func (s *Store) FormatContext(ctx context.Context) (string, error) {
	entries, err := s.List(ctx)
	if err != nil {
		return "", err
	}
	if len(entries) == 0 {
		return "", nil
	}
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		lines = append(lines, fmt.Sprintf("- %s = %s", entry.Key, entry.Value))
	}
	return strings.Join(lines, "\n"), nil
}

func (s *Store) Delete(ctx context.Context, key string) (bool, error) {
	trace.Log(ctx, "memory.Store.Delete", map[string]any{"path": s.path, "key": key})
	entries, err := s.load(ctx)
	if err != nil {
		return false, err
	}
	key = strings.TrimSpace(key)
	if _, ok := entries[key]; !ok {
		return false, nil
	}
	delete(entries, key)
	return true, s.save(ctx, entries)
}

func (s *Store) load(ctx context.Context) (map[string]Entry, error) {
	trace.Log(ctx, "memory.Store.load", map[string]any{"path": s.path})
	entries := map[string]Entry{}
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return entries, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return entries, nil
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("parse memory file %s: %w", s.path, err)
	}
	return entries, nil
}

func (s *Store) save(ctx context.Context, entries map[string]Entry) error {
	trace.Log(ctx, "memory.Store.save", map[string]any{"path": s.path, "entries": len(entries)})
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0o644)
}
