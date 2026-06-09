package webshare

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	Token     string    `json:"token"`
	Mode      string    `json:"mode"`
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type Manager struct {
	mu       sync.Mutex
	ttl      time.Duration
	sessions map[string]Session
}

func NewManager(ttl time.Duration) *Manager {
	return &Manager{ttl: ttl, sessions: map[string]Session{}}
}

func (m *Manager) Create(mode, baseURL string) (Session, error) {
	token, err := randomToken(6)
	if err != nil {
		return Session{}, err
	}
	now := time.Now()
	s := Session{Token: token, Mode: mode, URL: baseURL + "/guest/" + token, CreatedAt: now, ExpiresAt: now.Add(m.ttl)}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = s
	return s, nil
}

func (m *Manager) Get(token string) (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[token]
	if !ok {
		return Session{}, false
	}
	if time.Now().After(s.ExpiresAt) {
		delete(m.sessions, token)
		return Session{}, false
	}
	return s, true
}

func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for k, s := range m.sessions {
		if now.After(s.ExpiresAt) {
			delete(m.sessions, k)
		}
	}
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
