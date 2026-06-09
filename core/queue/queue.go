package queue

import (
    "errors"
    "sort"
    "sync"
    "time"
)

type Status string

const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusCompleted Status = "completed"
    StatusFailed    Status = "failed"
    StatusCanceled  Status = "canceled"
)

type Item struct {
    ID        string `json:"id"`
    FileName  string `json:"file_name"`
    Target    string `json:"target"`
    SizeBytes int64  `json:"size_bytes"`
    SentBytes int64  `json:"sent_bytes"`
    Status    Status `json:"status"`
    Error     string `json:"error,omitempty"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type Manager struct {
    mu    sync.Mutex
    items map[string]*Item
    order []string
}

func NewManager() *Manager {
    return &Manager{items: map[string]*Item{}, order: []string{}}
}

func (m *Manager) Add(fileName, target string, sizeBytes int64) Item {
    m.mu.Lock()
    defer m.mu.Unlock()
    now := time.Now().UTC().Format(time.RFC3339)
    id := "q_" + time.Now().UTC().Format("20060102T150405.000000000")
    it := &Item{ID: id, FileName: fileName, Target: target, SizeBytes: sizeBytes, Status: StatusPending, CreatedAt: now, UpdatedAt: now}
    m.items[id] = it
    m.order = append(m.order, id)
    return *it
}

func (m *Manager) List() []Item {
    m.mu.Lock()
    defer m.mu.Unlock()
    out := make([]Item, 0, len(m.items))
    for _, id := range m.order {
        if it, ok := m.items[id]; ok { out = append(out, *it) }
    }
    sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt > out[j].CreatedAt })
    return out
}

func (m *Manager) UpdateProgress(id string, sentBytes int64, status Status, errText string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    it, ok := m.items[id]
    if !ok { return errors.New("queue item not found") }
    it.SentBytes = sentBytes
    it.Status = status
    it.Error = errText
    it.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
    return nil
}

func (m *Manager) Cancel(id string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    it, ok := m.items[id]
    if !ok { return errors.New("queue item not found") }
    if it.Status == StatusCompleted { return errors.New("completed items cannot be canceled") }
    it.Status = StatusCanceled
    it.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
    return nil
}

func (m *Manager) ClearFinished() int {
    m.mu.Lock()
    defer m.mu.Unlock()
    removed := 0
    next := make([]string, 0, len(m.order))
    for _, id := range m.order {
        it, ok := m.items[id]
        if !ok { continue }
        if it.Status == StatusCompleted || it.Status == StatusFailed || it.Status == StatusCanceled {
            delete(m.items, id)
            removed++
            continue
        }
        next = append(next, id)
    }
    m.order = next
    return removed
}
