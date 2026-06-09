package transfer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

const DefaultChunkSize int64 = 4 * 1024 * 1024 // 4 MB

type ChunkedSendOptions struct {
	Target     string
	FilePath   string
	Token      string
	ChunkSize  int64
	MaxRetries int
}

type ChunkedSendResult struct {
	UploadID  string
	Status    string
	SHA256    string
	SizeBytes int64
	Response  string
	Duration  time.Duration
}

type initRequest struct {
	FileName  string `json:"file_name"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
	ChunkSize int64  `json:"chunk_size"`
}

type initResponse struct {
	Accepted     bool   `json:"accepted"`
	UploadID     string `json:"upload_id"`
	NextOffset   int64  `json:"next_offset"`
	ReceivedSize int64  `json:"received_size"`
	Message      string `json:"message"`
}

type statusResponse struct {
	UploadID     string `json:"upload_id"`
	FileName     string `json:"file_name"`
	SizeBytes    int64  `json:"size_bytes"`
	ReceivedSize int64  `json:"received_size"`
	Complete     bool   `json:"complete"`
	SHA256       string `json:"sha256"`
	Message      string `json:"message"`
}

func SendFileChunked(opts ChunkedSendOptions, onProgress func(sent int64, total int64)) (*ChunkedSendResult, error) {
	if opts.ChunkSize <= 0 {
		opts.ChunkSize = DefaultChunkSize
	}
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}
	file, err := os.Open(opts.FilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, fmt.Errorf("por ahora se envían archivos, no carpetas")
	}

	hash, err := FileSHA256(opts.FilePath)
	if err != nil {
		return nil, err
	}

	started := time.Now()
	initResp, err := initChunkedUpload(opts.Target, filepath.Base(opts.FilePath), info.Size(), hash, opts.ChunkSize, opts.Token)
	if err != nil {
		return nil, err
	}
	if !initResp.Accepted || initResp.UploadID == "" {
		return nil, fmt.Errorf("receptor no aceptó la transferencia: %s", initResp.Message)
	}

	offset := initResp.NextOffset
	if offset < 0 || offset > info.Size() {
		return nil, fmt.Errorf("offset inválido informado por receptor: %d", offset)
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}
	if onProgress != nil {
		onProgress(offset, info.Size())
	}

	buffer := make([]byte, opts.ChunkSize)
	for offset < info.Size() {
		remaining := info.Size() - offset
		readSize := opts.ChunkSize
		if remaining < readSize {
			readSize = remaining
		}
		n, err := io.ReadFull(file, buffer[:readSize])
		if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
			return nil, err
		}
		if n == 0 {
			break
		}
		if err := sendChunkWithRetry(opts.Target, initResp.UploadID, offset, buffer[:n], opts.Token, opts.MaxRetries); err != nil {
			return nil, err
		}
		offset += int64(n)
		if onProgress != nil {
			onProgress(offset, info.Size())
		}
	}

	status, raw, err := finishChunkedUpload(opts.Target, initResp.UploadID, opts.Token)
	if err != nil {
		return nil, err
	}
	return &ChunkedSendResult{UploadID: initResp.UploadID, Status: status.Message, SHA256: hash, SizeBytes: info.Size(), Response: raw, Duration: time.Since(started)}, nil
}

func initChunkedUpload(target, fileName string, size int64, sha string, chunkSize int64, token string) (*initResponse, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	u.Path = "/transfer/chunked/init"
	body, _ := json.Marshal(initRequest{FileName: fileName, SizeBytes: size, SHA256: sha, ChunkSize: chunkSize})
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("X-TransferLAN-Token", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out initResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("init falló: %s", out.Message)
	}
	return &out, nil
}

func sendChunkWithRetry(target, uploadID string, offset int64, data []byte, token string, maxRetries int) error {
	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := sendChunk(target, uploadID, offset, data, token); err != nil {
			lastErr = err
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
			continue
		}
		return nil
	}
	return fmt.Errorf("no se pudo enviar bloque en offset %d después de %d intentos: %w", offset, maxRetries, lastErr)
}

func sendChunk(target, uploadID string, offset int64, data []byte, token string) error {
	u, err := url.Parse(target)
	if err != nil {
		return err
	}
	u.Path = "/transfer/chunked/chunk"
	q := u.Query()
	q.Set("upload_id", uploadID)
	q.Set("offset", fmt.Sprintf("%d", offset))
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/octet-stream")
	if token != "" {
		req.Header.Set("X-TransferLAN-Token", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("chunk rechazado: %s", string(body))
	}
	return nil
}

func finishChunkedUpload(target, uploadID string, token string) (*statusResponse, string, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, "", err
	}
	u.Path = "/transfer/chunked/finish"
	q := u.Query()
	q.Set("upload_id", uploadID)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return nil, "", err
	}
	if token != "" {
		req.Header.Set("X-TransferLAN-Token", token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	var out statusResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, string(body), err
	}
	if resp.StatusCode >= 400 {
		return nil, string(body), fmt.Errorf("finish falló: %s", out.Message)
	}
	return &out, string(body), nil
}
