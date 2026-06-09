package notify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Type define eventos visibles para la UI. En v0.15.0 todavía no invoca APIs
// nativas del sistema operativo: centraliza eventos para que Flutter muestre
// notificaciones nativas con flutter_local_notifications o equivalente.
type Type string

const (
	TypeIncomingTransfer Type = "incoming_transfer"
	TypeTransferDone     Type = "transfer_done"
	TypeTransferError    Type = "transfer_error"
	TypeGuestUpload      Type = "guest_upload"
	TypePairing          Type = "pairing"
)

type Event struct {
	ID        string `json:"id"`
	Type      Type   `json:"type"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	FileName  string `json:"file_name,omitempty"`
	FilePath  string `json:"file_path,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
	CreatedAt string `json:"created_at"`
	Read      bool   `json:"read"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func DefaultStore() *Store {
	return NewStore(filepath.Join("data", "notifications.json"))
}

func NewStore(path string) *Store { return &Store{path: path} }

func (s *Store) Add(t Type, title, body, fileName, filePath string, sizeBytes int64) (Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	events, _ := s.readUnlocked()
	e := Event{ID: time.Now().UTC().Format("20060102150405.000000000"), Type: t, Title: title, Body: body, FileName: fileName, FilePath: filePath, SizeBytes: sizeBytes, CreatedAt: time.Now().Format(time.RFC3339), Read: false}
	events = append(events, e)
	if len(events) > 100 {
		events = events[len(events)-100:]
	}
	return e, s.writeUnlocked(events)
}

func (s *Store) List(limit int) ([]Event, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	events, err := s.readUnlocked()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].CreatedAt > events[j].CreatedAt })
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if len(events) > limit {
		events = events[:limit]
	}
	return events, nil
}

func (s *Store) MarkAllRead() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	events, err := s.readUnlocked()
	if err != nil {
		return err
	}
	for i := range events {
		events[i].Read = true
	}
	return s.writeUnlocked(events)
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.writeUnlocked([]Event{})
}

func (s *Store) readUnlocked() ([]Event, error) {
	b, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return []Event{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return []Event{}, nil
	}
	var events []Event
	if err := json.Unmarshal(b, &events); err != nil {
		return []Event{}, nil
	}
	return events, nil
}

func (s *Store) writeUnlocked(events []Event) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, b, 0o644)
}
