package server

import (
    "encoding/json"
    "net/http"

    "github.com/softwareparatodos/transferlan-plus/core/queue"
)

type queueAddRequest struct {
    FileName  string `json:"file_name"`
    Target    string `json:"target"`
    SizeBytes int64  `json:"size_bytes"`
}

type queueProgressRequest struct {
    ID        string       `json:"id"`
    SentBytes int64        `json:"sent_bytes"`
    Status    queue.Status `json:"status"`
    Error     string       `json:"error,omitempty"`
}

type queueIDRequest struct { ID string `json:"id"` }

func (s *Server) queueStore() *queue.Manager {
    if s.Queue == nil { s.Queue = queue.NewManager() }
    return s.Queue
}

func (s *Server) handleQueueList(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    writeJSON(w, map[string]any{"items": s.queueStore().List()})
}

func (s *Server) handleQueueAdd(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    var req queueAddRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
    if req.FileName == "" { http.Error(w, "file_name is required", http.StatusBadRequest); return }
    item := s.queueStore().Add(req.FileName, req.Target, req.SizeBytes)
    writeJSON(w, map[string]any{"accepted": true, "item": item})
}

func (s *Server) handleQueueProgress(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    var req queueProgressRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
    if req.ID == "" { http.Error(w, "id is required", http.StatusBadRequest); return }
    if req.Status == "" { req.Status = queue.StatusRunning }
    if err := s.queueStore().UpdateProgress(req.ID, req.SentBytes, req.Status, req.Error); err != nil { http.Error(w, err.Error(), http.StatusNotFound); return }
    writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleQueueCancel(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    var req queueIDRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, "invalid json", http.StatusBadRequest); return }
    if req.ID == "" { http.Error(w, "id is required", http.StatusBadRequest); return }
    if err := s.queueStore().Cancel(req.ID); err != nil { http.Error(w, err.Error(), http.StatusBadRequest); return }
    writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleQueueClearFinished(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
    removed := s.queueStore().ClearFinished()
    writeJSON(w, map[string]any{"ok": true, "removed": removed})
}
