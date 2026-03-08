package state

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Store struct {
	filePath string
	mu sync.Mutex
	sentIDS map[string]time.Time
}

func NewStore(filePath string) (*Store, error) {
	s := &Store{
		filePath: filePath,
		sentIDS: make(map[string]time.Time),
	}

	s.load()
	return s, nil
}

func (s *Store) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		_ = json.Unmarshal(data, &s.sentIDS)
	}
}

func (s *Store) save() error {
	data, err := json.MarshalIndent(s.sentIDS, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

func (s *Store) HasBeenSent(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.sentIDS[id]
	return exists
}

func (s *Store) MarkAsSent(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sentIDS[id] = time.Now()
	return s.save()
}
