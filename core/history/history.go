package history

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type Direction string
type Status string

const (
	DirectionSent     Direction = "sent"
	DirectionReceived Direction = "received"

	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

type Entry struct {
	ID        string    `json:"id"`
	Direction Direction `json:"direction"`
	Status    Status    `json:"status"`
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path,omitempty"`
	SizeBytes int64     `json:"size_bytes"`
	PeerName  string    `json:"peer_name,omitempty"`
	PeerHost  string    `json:"peer_host,omitempty"`
	SHA256    string    `json:"sha256,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) *Store {
	return &Store{path: path}
}

func DefaultStore() *Store {
	return NewStore(filepath.Join("data", "history.json"))
}

func (s *Store) Add(entry Entry) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.loadLocked()
	if err != nil {
		return Entry{}, err
	}

	now := time.Now()
	if entry.ID == "" {
		entry.ID = now.Format("20060102-150405.000000000")
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	if entry.Status == "" {
		entry.Status = StatusPending
	}

	entries = append(entries, entry)
	return entry, s.saveLocked(entries)
}

func (s *Store) UpdateStatus(id string, status Status, sha256 string, errText string) (Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.loadLocked()
	if err != nil {
		return Entry{}, err
	}

	for i := range entries {
		if entries[i].ID == id {
			entries[i].Status = status
			entries[i].SHA256 = sha256
			entries[i].Error = errText
			entries[i].UpdatedAt = time.Now()
			return entries[i], s.saveLocked(entries)
		}
	}
	return Entry{}, errors.New("history entry not found")
}

func (s *Store) List(limit int) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := s.loadLocked()
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].UpdatedAt.After(entries[j].UpdatedAt)
	})

	if limit > 0 && len(entries) > limit {
		entries = entries[:limit]
	}
	return entries, nil
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveLocked([]Entry{})
}

func (s *Store) loadLocked() ([]Entry, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Entry{}, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return []Entry{}, nil
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Store) saveLocked(entries []Entry) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0644)
}
