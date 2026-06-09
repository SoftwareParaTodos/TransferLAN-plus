package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/softwareparatodos/transferlan-plus/core/config"
	"github.com/softwareparatodos/transferlan-plus/core/discovery"
	"github.com/softwareparatodos/transferlan-plus/core/history"
	"github.com/softwareparatodos/transferlan-plus/core/notify"
	"github.com/softwareparatodos/transferlan-plus/core/pairing"
	"github.com/softwareparatodos/transferlan-plus/core/queue"
	"github.com/softwareparatodos/transferlan-plus/core/transfer"
	"github.com/softwareparatodos/transferlan-plus/core/webshare"
)

type Server struct {
	Name        string
	Platform    string
	Store       *pairing.Store
	Sessions    *pairing.SessionManager
	DownloadDir string
	WebShare    *webshare.Manager
	History     *history.Store
	Notify      *notify.Store
	Queue       *queue.Manager
}

type healthResponse struct {
	App      string `json:"app"`
	Version  string `json:"version"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
	Status   string `json:"status"`
}

type pairRequest struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
	PIN      string `json:"pin,omitempty"`
}

type pairResponse struct {
	Accepted bool   `json:"accepted"`
	Message  string `json:"message"`
	Token    string `json:"token,omitempty"`
}

type uploadResponse struct {
	Accepted  bool   `json:"accepted"`
	FileName  string `json:"file_name"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
	SavedTo   string `json:"saved_to"`
	Message   string `json:"message"`
}

type chunkedInitRequest struct {
	FileName  string `json:"file_name"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
	ChunkSize int64  `json:"chunk_size"`
}

type chunkedInitResponse struct {
	Accepted     bool   `json:"accepted"`
	UploadID     string `json:"upload_id"`
	NextOffset   int64  `json:"next_offset"`
	ReceivedSize int64  `json:"received_size"`
	Message      string `json:"message"`
}

type chunkedStatusResponse struct {
	UploadID     string `json:"upload_id"`
	FileName     string `json:"file_name"`
	SizeBytes    int64  `json:"size_bytes"`
	ReceivedSize int64  `json:"received_size"`
	Complete     bool   `json:"complete"`
	SHA256       string `json:"sha256"`
	SavedTo      string `json:"saved_to,omitempty"`
	Message      string `json:"message"`
}

type chunkedMeta struct {
	UploadID  string `json:"upload_id"`
	FileName  string `json:"file_name"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
	ChunkSize int64  `json:"chunk_size"`
	TmpPath   string `json:"tmp_path"`
	FinalPath string `json:"final_path"`
	CreatedAt string `json:"created_at"`
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/pair/pin/start", s.handlePairPINStart)
	mux.HandleFunc("/pair/pin/current", s.handlePairPINCurrent)
	mux.HandleFunc("/pair/request", s.handlePairRequest)
	mux.HandleFunc("/pair/trusted", s.handleTrusted)
	mux.HandleFunc("/discovery/devices", s.handleDiscoveryDevices)
	mux.HandleFunc("/transfer/upload", s.handleUpload)
	mux.HandleFunc("/transfer/chunked/init", s.handleChunkedInit)
	mux.HandleFunc("/transfer/chunked/chunk", s.handleChunkedChunk)
	mux.HandleFunc("/transfer/chunked/finish", s.handleChunkedFinish)
	mux.HandleFunc("/transfer/chunked/status", s.handleChunkedStatus)
	mux.HandleFunc("/transfer/chunked/incomplete", s.handleChunkedIncomplete)
	mux.HandleFunc("/guest/share/start", s.handleGuestShareStart)
	mux.HandleFunc("/history", s.handleHistoryList)
	mux.HandleFunc("/history/clear", s.handleHistoryClear)
	mux.HandleFunc("/notifications", s.handleNotificationsList)
	mux.HandleFunc("/notifications/read-all", s.handleNotificationsReadAll)
	mux.HandleFunc("/notifications/clear", s.handleNotificationsClear)
	mux.HandleFunc("/queue", s.handleQueueList)
	mux.HandleFunc("/queue/add", s.handleQueueAdd)
	mux.HandleFunc("/queue/progress", s.handleQueueProgress)
	mux.HandleFunc("/queue/cancel", s.handleQueueCancel)
	mux.HandleFunc("/queue/clear-finished", s.handleQueueClearFinished)
	mux.HandleFunc("/guest/", s.handleGuest)
	return mux
}

func (s *Server) ListenAndServe(port int) error {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Servidor local TransferLAN+ escuchando en %s", addr)
	return http.ListenAndServe(addr, s.Handler())
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, healthResponse{
		App:      config.AppName,
		Version:  config.ProtocolVer,
		Name:     s.Name,
		Platform: s.Platform,
		Status:   "ok",
	})
}

type pairPINResponse struct {
	PIN        string `json:"pin"`
	ExpiresAt  string `json:"expires_at"`
	PairingURL string `json:"pairing_url"`
	Message    string `json:"message"`
}

func (s *Server) sessions() *pairing.SessionManager {
	if s.Sessions == nil {
		s.Sessions = pairing.NewSessionManager(2 * time.Minute)
	}
	return s.Sessions
}

func (s *Server) handlePairPINStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "usar GET o POST", http.StatusMethodNotAllowed)
		return
	}
	session, err := s.sessions().Start()
	if err != nil {
		http.Error(w, "no se pudo generar PIN", http.StatusInternalServerError)
		return
	}
	url := fmt.Sprintf("transferlan://pair?name=%s&pin=%s", s.Name, session.PIN)
	writeJSON(w, pairPINResponse{
		PIN: session.PIN, ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
		PairingURL: url, Message: "PIN temporal generado. Compartilo solo con el dispositivo que querés vincular.",
	})
}

func (s *Server) handlePairPINCurrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "usar GET", http.StatusMethodNotAllowed)
		return
	}
	session, ok := s.sessions().Current()
	if !ok {
		writeJSON(w, map[string]any{"active": false, "message": "no hay PIN activo"})
		return
	}
	url := fmt.Sprintf("transferlan://pair?name=%s&pin=%s", s.Name, session.PIN)
	writeJSON(w, map[string]any{"active": true, "pin": session.PIN, "expires_at": session.ExpiresAt.Format(time.RFC3339), "pairing_url": url})
}

func (s *Server) handlePairRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "usar POST", http.StatusMethodNotAllowed)
		return
	}
	var req pairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	if req.DeviceID == "" || req.Name == "" {
		http.Error(w, "device_id y name son obligatorios", http.StatusBadRequest)
		return
	}

	if req.PIN == "" || !s.sessions().Verify(req.PIN) {
		http.Error(w, "PIN inválido o vencido", http.StatusForbidden)
		return
	}

	trusted, err := s.Store.Trust(req.DeviceID, req.Name, req.Platform)
	if err != nil {
		http.Error(w, "no se pudo guardar dispositivo", http.StatusInternalServerError)
		return
	}

	log.Printf("Dispositivo emparejado: %s (%s) a las %s", req.Name, req.DeviceID, time.Now().Format(time.RFC3339))
	writeJSON(w, pairResponse{Accepted: true, Message: "dispositivo aceptado por PIN", Token: trusted.Token})
}

func (s *Server) handleTrusted(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, s.Store.List())
}

type discoveryDeviceResponse struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	IPs      []string `json:"ips"`
	Platform string   `json:"platform"`
	Version  string   `json:"version"`
	BaseURL  string   `json:"base_url"`
}

func (s *Server) handleDiscoveryDevices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "usar GET", http.StatusMethodNotAllowed)
		return
	}
	seconds := 3
	if raw := r.URL.Query().Get("seconds"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			seconds = parsed
		}
	}
	if seconds > 15 {
		seconds = 15
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(seconds)*time.Second)
	defer cancel()

	browser := discovery.NewBrowser()
	devices, err := browser.Browse(ctx, time.Duration(seconds)*time.Second)
	if err != nil {
		http.Error(w, "no se pudieron buscar dispositivos", http.StatusInternalServerError)
		return
	}

	out := make([]discoveryDeviceResponse, 0, len(devices))
	seen := map[string]bool{}
	for _, d := range devices {
		host := d.Host
		if len(d.IPs) > 0 {
			host = d.IPs[0]
		}
		id := fmt.Sprintf("%s-%s-%d", d.Name, host, d.Port)
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, discoveryDeviceResponse{
			ID: id, Name: d.Name, Host: host, Port: d.Port, IPs: d.IPs,
			Platform: d.Platform, Version: d.Version, BaseURL: "http://" + d.Address(),
		})
	}
	writeJSON(w, out)
}

func (s *Server) requireTrusted(w http.ResponseWriter, r *http.Request) bool {
	if s.Store == nil {
		http.Error(w, "store no inicializado", http.StatusInternalServerError)
		return false
	}
	token := r.Header.Get("X-TransferLAN-Token")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	if !s.Store.HasToken(token) {
		http.Error(w, "dispositivo no emparejado o token inválido", http.StatusUnauthorized)
		return false
	}
	return true
}

func (s *Server) notifyStore() *notify.Store {
	if s.Notify == nil {
		s.Notify = notify.DefaultStore()
	}
	return s.Notify
}

func (s *Server) addNotification(t notify.Type, title, body, fileName, filePath string, sizeBytes int64) {
	_, err := s.notifyStore().Add(t, title, body, fileName, filePath, sizeBytes)
	if err != nil {
		log.Printf("no se pudo registrar notificación: %v", err)
	}
}

func (s *Server) handleNotificationsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	events, err := s.notifyStore().List(limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, events)
}

func (s *Server) handleNotificationsReadAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := s.notifyStore().MarkAllRead(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleNotificationsClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := s.notifyStore().Clear(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) historyStore() *history.Store {
	if s.History == nil {
		s.History = history.DefaultStore()
	}
	return s.History
}

func (s *Server) handleHistoryList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "usar GET", http.StatusMethodNotAllowed)
		return
	}
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	entries, err := s.historyStore().List(limit)
	if err != nil {
		http.Error(w, "no se pudo leer historial", http.StatusInternalServerError)
		return
	}
	writeJSON(w, entries)
}

func (s *Server) handleHistoryClear(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodDelete {
		http.Error(w, "usar POST o DELETE", http.StatusMethodNotAllowed)
		return
	}
	if err := s.historyStore().Clear(); err != nil {
		http.Error(w, "no se pudo limpiar historial", http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]any{"accepted": true, "message": "historial limpiado"})
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	if !s.requireTrusted(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "usar POST", http.StatusMethodNotAllowed)
		return
	}

	filename := transfer.SafeFileName(r.URL.Query().Get("filename"))
	if headerName := r.Header.Get("X-TransferLAN-File-Name"); headerName != "" {
		filename = transfer.SafeFileName(headerName)
	}

	downloadDir := s.DownloadDir
	if downloadDir == "" {
		downloadDir = "downloads"
	}
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		http.Error(w, "no se pudo crear carpeta de descarga", http.StatusInternalServerError)
		return
	}

	finalPath := filepath.Join(downloadDir, filename)
	finalPath = uniquePath(finalPath)
	tmpPath := finalPath + ".part"

	out, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, "no se pudo crear archivo destino", http.StatusInternalServerError)
		return
	}

	h := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(out, h), r.Body)
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(tmpPath)
		http.Error(w, "transferencia interrumpida", http.StatusInternalServerError)
		return
	}

	gotHash := hex.EncodeToString(h.Sum(nil))
	expectedHash := r.URL.Query().Get("sha256")
	if expectedHeader := r.Header.Get("X-TransferLAN-SHA256"); expectedHeader != "" {
		expectedHash = expectedHeader
	}
	if expectedHash != "" && expectedHash != gotHash {
		_ = os.Remove(tmpPath)
		http.Error(w, "hash SHA256 no coincide: archivo descartado", http.StatusBadRequest)
		return
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		http.Error(w, "no se pudo finalizar archivo", http.StatusInternalServerError)
		return
	}

	log.Printf("Archivo recibido: %s (%d bytes)", finalPath, written)
	_, _ = s.historyStore().Add(history.Entry{Direction: history.DirectionReceived, Status: history.StatusCompleted, FileName: filepath.Base(finalPath), FilePath: finalPath, SizeBytes: written, SHA256: gotHash, PeerHost: r.RemoteAddr})
	s.addNotification(notify.TypeTransferDone, "Archivo recibido", filepath.Base(finalPath)+" se guardó correctamente.", filepath.Base(finalPath), finalPath, written)
	writeJSON(w, uploadResponse{Accepted: true, FileName: filepath.Base(finalPath), SizeBytes: written, SHA256: gotHash, SavedTo: finalPath, Message: "archivo recibido correctamente"})
}

func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	base := path[:len(path)-len(ext)]
	for i := 1; i < 10000; i++ {
		candidate := fmt.Sprintf("%s_%d%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
	return fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), ext)
}

func (s *Server) handleChunkedInit(w http.ResponseWriter, r *http.Request) {
	if !s.requireTrusted(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "usar POST", http.StatusMethodNotAllowed)
		return
	}
	var req chunkedInitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "json inválido", http.StatusBadRequest)
		return
	}
	req.FileName = transfer.SafeFileName(req.FileName)
	if req.FileName == "" || req.SizeBytes <= 0 || req.SHA256 == "" {
		http.Error(w, "file_name, size_bytes y sha256 son obligatorios", http.StatusBadRequest)
		return
	}
	if req.ChunkSize <= 0 {
		req.ChunkSize = transfer.DefaultChunkSize
	}

	downloadDir := s.downloadDir()
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		http.Error(w, "no se pudo crear carpeta de descarga", http.StatusInternalServerError)
		return
	}
	if err := os.MkdirAll(filepath.Join(downloadDir, ".transferlan_parts"), 0755); err != nil {
		http.Error(w, "no se pudo crear carpeta temporal", http.StatusInternalServerError)
		return
	}

	if existing, ok := s.findResumableUpload(req.FileName, req.SizeBytes, req.SHA256); ok {
		received := currentSize(existing.TmpPath)
		if received > existing.SizeBytes {
			received = existing.SizeBytes
		}
		writeJSON(w, chunkedInitResponse{Accepted: true, UploadID: existing.UploadID, NextOffset: received, ReceivedSize: received, Message: "transferencia existente encontrada; reanudando desde next_offset"})
		return
	}

	finalPath := uniquePath(filepath.Join(downloadDir, req.FileName))
	uploadID := fmt.Sprintf("%d", time.Now().UnixNano())
	tmpPath := filepath.Join(downloadDir, ".transferlan_parts", uploadID+".part")
	metaPath := tmpPath + ".json"
	meta := chunkedMeta{UploadID: uploadID, FileName: filepath.Base(finalPath), SizeBytes: req.SizeBytes, SHA256: req.SHA256, ChunkSize: req.ChunkSize, TmpPath: tmpPath, FinalPath: finalPath, CreatedAt: time.Now().Format(time.RFC3339)}
	if err := saveMeta(metaPath, meta); err != nil {
		http.Error(w, "no se pudo iniciar transferencia", http.StatusInternalServerError)
		return
	}
	file, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		http.Error(w, "no se pudo crear archivo temporal", http.StatusInternalServerError)
		return
	}
	_ = file.Close()
	writeJSON(w, chunkedInitResponse{Accepted: true, UploadID: uploadID, NextOffset: currentSize(tmpPath), ReceivedSize: currentSize(tmpPath), Message: "transferencia por bloques iniciada"})
}

func (s *Server) handleChunkedChunk(w http.ResponseWriter, r *http.Request) {
	if !s.requireTrusted(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "usar POST", http.StatusMethodNotAllowed)
		return
	}
	uploadID := r.URL.Query().Get("upload_id")
	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if uploadID == "" || err != nil || offset < 0 {
		http.Error(w, "upload_id y offset válidos son obligatorios", http.StatusBadRequest)
		return
	}
	meta, err := s.loadChunkedMeta(uploadID)
	if err != nil {
		http.Error(w, "transferencia no encontrada", http.StatusNotFound)
		return
	}
	expected := currentSize(meta.TmpPath)
	if offset != expected {
		writeJSON(w, chunkedInitResponse{Accepted: false, UploadID: uploadID, NextOffset: expected, ReceivedSize: expected, Message: "offset no coincide; reintentar desde next_offset"})
		return
	}
	out, err := os.OpenFile(meta.TmpPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		http.Error(w, "no se pudo abrir temporal", http.StatusInternalServerError)
		return
	}
	written, copyErr := io.Copy(out, r.Body)
	closeErr := out.Close()
	if copyErr != nil || closeErr != nil {
		http.Error(w, "no se pudo guardar bloque", http.StatusInternalServerError)
		return
	}
	newSize := expected + written
	writeJSON(w, chunkedStatusResponse{UploadID: uploadID, FileName: meta.FileName, SizeBytes: meta.SizeBytes, ReceivedSize: newSize, Complete: newSize >= meta.SizeBytes, SHA256: meta.SHA256, Message: "bloque recibido"})
}

func (s *Server) handleChunkedFinish(w http.ResponseWriter, r *http.Request) {
	if !s.requireTrusted(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "usar POST", http.StatusMethodNotAllowed)
		return
	}
	uploadID := r.URL.Query().Get("upload_id")
	meta, err := s.loadChunkedMeta(uploadID)
	if err != nil {
		http.Error(w, "transferencia no encontrada", http.StatusNotFound)
		return
	}
	received := currentSize(meta.TmpPath)
	if received != meta.SizeBytes {
		w.WriteHeader(http.StatusConflict)
		writeJSON(w, chunkedStatusResponse{UploadID: uploadID, FileName: meta.FileName, SizeBytes: meta.SizeBytes, ReceivedSize: received, Complete: false, SHA256: meta.SHA256, Message: "transferencia incompleta"})
		return
	}
	gotHash, err := transfer.FileSHA256(meta.TmpPath)
	if err != nil || gotHash != meta.SHA256 {
		_ = os.Remove(meta.TmpPath)
		_ = os.Remove(meta.TmpPath + ".json")
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(w, chunkedStatusResponse{UploadID: uploadID, FileName: meta.FileName, SizeBytes: meta.SizeBytes, ReceivedSize: received, Complete: false, SHA256: gotHash, Message: "SHA256 no coincide; archivo descartado"})
		return
	}
	if err := os.Rename(meta.TmpPath, meta.FinalPath); err != nil {
		http.Error(w, "no se pudo finalizar archivo", http.StatusInternalServerError)
		return
	}
	_ = os.Remove(meta.TmpPath + ".json")
	_, _ = s.historyStore().Add(history.Entry{Direction: history.DirectionReceived, Status: history.StatusCompleted, FileName: meta.FileName, FilePath: meta.FinalPath, SizeBytes: meta.SizeBytes, SHA256: gotHash, PeerHost: r.RemoteAddr})
	s.addNotification(notify.TypeTransferDone, "Transferencia completada", meta.FileName+" llegó completo y verificado.", meta.FileName, meta.FinalPath, meta.SizeBytes)
	writeJSON(w, chunkedStatusResponse{UploadID: uploadID, FileName: meta.FileName, SizeBytes: meta.SizeBytes, ReceivedSize: received, Complete: true, SHA256: gotHash, SavedTo: meta.FinalPath, Message: "archivo recibido correctamente por bloques"})
}

func (s *Server) handleChunkedStatus(w http.ResponseWriter, r *http.Request) {
	uploadID := r.URL.Query().Get("upload_id")
	meta, err := s.loadChunkedMeta(uploadID)
	if err != nil {
		http.Error(w, "transferencia no encontrada", http.StatusNotFound)
		return
	}
	received := currentSize(meta.TmpPath)
	writeJSON(w, chunkedStatusResponse{UploadID: uploadID, FileName: meta.FileName, SizeBytes: meta.SizeBytes, ReceivedSize: received, Complete: received >= meta.SizeBytes, SHA256: meta.SHA256, Message: "estado de transferencia"})
}

func (s *Server) handleChunkedIncomplete(w http.ResponseWriter, r *http.Request) {
	partsDir := filepath.Join(s.downloadDir(), ".transferlan_parts")
	entries, err := os.ReadDir(partsDir)
	if err != nil {
		writeJSON(w, []chunkedStatusResponse{})
		return
	}
	items := make([]chunkedStatusResponse, 0)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(partsDir, entry.Name()))
		if err != nil {
			continue
		}
		var meta chunkedMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		received := currentSize(meta.TmpPath)
		if received >= meta.SizeBytes {
			continue
		}
		items = append(items, chunkedStatusResponse{UploadID: meta.UploadID, FileName: meta.FileName, SizeBytes: meta.SizeBytes, ReceivedSize: received, Complete: false, SHA256: meta.SHA256, Message: "transferencia incompleta reanudable"})
	}
	writeJSON(w, items)
}

type guestShareRequest struct {
	Mode    string `json:"mode"`
	BaseURL string `json:"base_url"`
}

type guestShareResponse struct {
	Accepted  bool   `json:"accepted"`
	Token     string `json:"token"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expires_at"`
	Message   string `json:"message"`
}

type guestUploadResponse struct {
	Accepted  bool   `json:"accepted"`
	FileName  string `json:"file_name"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
	SavedTo   string `json:"saved_to"`
	Message   string `json:"message"`
}

func (s *Server) webShares() *webshare.Manager {
	if s.WebShare == nil {
		s.WebShare = webshare.NewManager(15 * time.Minute)
	}
	return s.WebShare
}

func (s *Server) handleGuestShareStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	var req guestShareRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Mode == "" {
		req.Mode = "upload"
	}
	if req.BaseURL == "" {
		req.BaseURL = "http://" + r.Host
	}
	session, err := s.webShares().Create(req.Mode, strings.TrimRight(req.BaseURL, "/"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, guestShareResponse{
		Accepted:  true,
		Token:     session.Token,
		URL:       session.URL,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
		Message:   "Enlace invitado creado. Generá el QR con esta URL desde la UI.",
	})
}

func (s *Server) handleGuest(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/guest/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}
	token := parts[0]
	session, ok := s.webShares().Get(token)
	if !ok {
		http.Error(w, "enlace invitado vencido o inválido", http.StatusUnauthorized)
		return
	}
	if len(parts) == 2 && parts[1] == "upload" {
		s.handleGuestUpload(w, r, session.Token)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, guestPageHTML, session.Token, session.ExpiresAt.Format("15:04:05"), session.Token)
}

func (s *Server) handleGuestUpload(w http.ResponseWriter, r *http.Request, token string) {
	if r.Method != http.MethodPost {
		http.Error(w, "método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseMultipartForm(64 << 20); err != nil {
		http.Error(w, "no se pudo leer el formulario", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "faltó el archivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	dir := filepath.Join(s.downloadDir(), "guest_uploads")
	if err := os.MkdirAll(dir, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	name := transfer.SafeFileName(header.Filename)
	finalPath := uniquePath(filepath.Join(dir, name))
	tmpPath := finalPath + ".part"
	out, err := os.Create(tmpPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(out, h), file)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(tmpPath)
		http.Error(w, copyErr.Error(), http.StatusInternalServerError)
		return
	}
	if closeErr != nil {
		_ = os.Remove(tmpPath)
		http.Error(w, closeErr.Error(), http.StatusInternalServerError)
		return
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = s.historyStore().Add(history.Entry{Direction: history.DirectionReceived, Status: history.StatusCompleted, FileName: filepath.Base(finalPath), FilePath: finalPath, SizeBytes: written, SHA256: hex.EncodeToString(h.Sum(nil)), PeerName: "Modo Invitado Web", PeerHost: r.RemoteAddr})
	writeJSON(w, guestUploadResponse{
		Accepted:  true,
		FileName:  filepath.Base(finalPath),
		SizeBytes: written,
		SHA256:    hex.EncodeToString(h.Sum(nil)),
		SavedTo:   finalPath,
		Message:   "Archivo recibido por Modo Invitado Web",
	})
}

const guestPageHTML = `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>TransferLAN+ Invitado</title>
<style>
body{font-family:system-ui,Arial,sans-serif;background:#0f172a;color:#e5e7eb;margin:0;padding:24px} .card{max-width:620px;margin:auto;background:#111827;border:1px solid #334155;border-radius:18px;padding:24px;box-shadow:0 18px 60px #0008} h1{margin-top:0;color:#38bdf8}.muted{color:#94a3b8}.btn{background:#38bdf8;color:#06111f;border:0;border-radius:12px;padding:14px 18px;font-weight:700} input{display:block;width:100%;padding:14px;margin:16px 0;border-radius:12px;background:#020617;color:#e5e7eb;border:1px solid #334155}.ok{color:#22c55e}.err{color:#fb7185} progress{width:100%;height:20px}
</style>
</head>
<body>
<div class="card">
<h1>TransferLAN+</h1>
<p class="muted">Modo Invitado Web. Subí un archivo a este equipo sin instalar la app.</p>
<p class="muted">Código: <b>%s</b> · vence aprox. a las %s</p>
<form id="f">
<input id="file" name="file" type="file" required>
<progress id="p" value="0" max="100"></progress><br><br>
<button class="btn" type="submit">Enviar archivo</button>
</form>
<p id="msg" class="muted"></p>
</div>
<script>
const form=document.getElementById('f'), msg=document.getElementById('msg'), p=document.getElementById('p');
form.addEventListener('submit', e=>{e.preventDefault(); const file=document.getElementById('file').files[0]; if(!file)return; const data=new FormData(); data.append('file',file); const xhr=new XMLHttpRequest(); xhr.open('POST','/guest/%s/upload'); xhr.upload.onprogress=(ev)=>{if(ev.lengthComputable)p.value=Math.round(ev.loaded*100/ev.total)}; xhr.onload=()=>{msg.className=xhr.status<300?'ok':'err'; msg.textContent=xhr.status<300?'Archivo enviado correctamente.':'Error: '+xhr.responseText}; xhr.onerror=()=>{msg.className='err'; msg.textContent='No se pudo enviar.'}; xhr.send(data);});
</script>
</body>
</html>`

func (s *Server) downloadDir() string {
	if s.DownloadDir != "" {
		return s.DownloadDir
	}
	return "downloads"
}

func (s *Server) loadChunkedMeta(uploadID string) (chunkedMeta, error) {
	var meta chunkedMeta
	if uploadID == "" {
		return meta, fmt.Errorf("upload_id vacío")
	}
	path := filepath.Join(s.downloadDir(), ".transferlan_parts", uploadID+".part.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return meta, err
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return meta, err
	}
	return meta, nil
}

func (s *Server) findResumableUpload(fileName string, sizeBytes int64, sha string) (chunkedMeta, bool) {
	partsDir := filepath.Join(s.downloadDir(), ".transferlan_parts")
	entries, err := os.ReadDir(partsDir)
	if err != nil {
		return chunkedMeta{}, false
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(partsDir, entry.Name()))
		if err != nil {
			continue
		}
		var meta chunkedMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}
		received := currentSize(meta.TmpPath)
		if meta.SHA256 == sha && meta.SizeBytes == sizeBytes && meta.FileName == fileName && received > 0 && received < sizeBytes {
			return meta, true
		}
	}
	return chunkedMeta{}, false
}

func saveMeta(path string, meta chunkedMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func currentSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
