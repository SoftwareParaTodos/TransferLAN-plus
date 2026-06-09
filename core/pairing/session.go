package pairing

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type Session struct {
	PIN       string    `json:"pin"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SessionManager struct {
	mu      sync.Mutex
	current *Session
	ttl     time.Duration
}

func NewSessionManager(ttl time.Duration) *SessionManager {
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	return &SessionManager{ttl: ttl}
}

func (m *SessionManager) Start() (Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	pin, err := GeneratePIN()
	if err != nil {
		return Session{}, err
	}
	s := Session{PIN: pin, ExpiresAt: time.Now().Add(m.ttl)}
	m.current = &s
	return s, nil
}

func (m *SessionManager) Current() (Session, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil || time.Now().After(m.current.ExpiresAt) {
		m.current = nil
		return Session{}, false
	}
	return *m.current, true
}

func (m *SessionManager) Verify(pin string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.current == nil || time.Now().After(m.current.ExpiresAt) {
		m.current = nil
		return false
	}
	if pin != m.current.PIN {
		return false
	}
	m.current = nil
	return true
}

func GeneratePIN() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
