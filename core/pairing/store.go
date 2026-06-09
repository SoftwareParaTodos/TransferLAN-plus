package pairing

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type TrustedDevice struct {
	DeviceID   string    `json:"device_id"`
	Name       string    `json:"name"`
	Platform   string    `json:"platform"`
	Token      string    `json:"token"`
	CreatedAt  time.Time `json:"created_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type Store struct {
	path    string
	mu      sync.Mutex
	trusted map[string]TrustedDevice
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		path = filepath.Join("transferlan-data", "trusted_devices.json")
	}
	store := &Store{path: path, trusted: map[string]TrustedDevice{}}
	if err := store.Load(); err != nil {
		return nil, err
	}
	return store, nil
}

func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, &s.trusted)
}

func (s *Store) Save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.trusted, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *Store) Trust(deviceID, name, platform string) (TrustedDevice, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, err := randomToken(32)
	if err != nil {
		return TrustedDevice{}, err
	}
	dev := TrustedDevice{
		DeviceID:   deviceID,
		Name:       name,
		Platform:   platform,
		Token:      token,
		CreatedAt:  time.Now(),
		LastSeenAt: time.Now(),
	}
	s.trusted[deviceID] = dev
	return dev, s.Save()
}

func (s *Store) IsTrusted(deviceID, token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	dev, ok := s.trusted[deviceID]
	return ok && token != "" && dev.Token == token
}

func (s *Store) HasToken(token string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if token == "" {
		return false
	}
	for _, dev := range s.trusted {
		if dev.Token == token {
			dev.LastSeenAt = time.Now()
			s.trusted[dev.DeviceID] = dev
			_ = s.Save()
			return true
		}
	}
	return false
}

func (s *Store) List() []TrustedDevice {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]TrustedDevice, 0, len(s.trusted))
	for _, dev := range s.trusted {
		out = append(out, dev)
	}
	return out
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
