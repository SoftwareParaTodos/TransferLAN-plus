package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type SendResult struct {
	StatusCode int
	Response   string
	SHA256     string
	SizeBytes  int64
}

type progressReader struct {
	reader io.Reader
	total  int64
	read   int64
	onRead func(read int64, total int64)
}

func (p *progressReader) Read(buf []byte) (int, error) {
	n, err := p.reader.Read(buf)
	if n > 0 {
		p.read += int64(n)
		if p.onRead != nil {
			p.onRead(p.read, p.total)
		}
	}
	return n, err
}

func SendFile(target, filePath, token string, onProgress func(read int64, total int64)) (*SendResult, error) {
	file, err := os.Open(filePath)
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

	hash, err := FileSHA256(filePath)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	u.Path = "/transfer/upload"
	q := u.Query()
	q.Set("filename", filepath.Base(filePath))
	q.Set("sha256", hash)
	u.RawQuery = q.Encode()

	reader := &progressReader{reader: file, total: info.Size(), onRead: onProgress}
	req, err := http.NewRequest(http.MethodPost, u.String(), reader)
	if err != nil {
		return nil, err
	}
	req.ContentLength = info.Size()
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("X-TransferLAN-File-Name", filepath.Base(filePath))
	req.Header.Set("X-TransferLAN-SHA256", hash)
	if token != "" {
		req.Header.Set("X-TransferLAN-Token", token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return &SendResult{StatusCode: resp.StatusCode, Response: string(body), SHA256: hash, SizeBytes: info.Size()}, nil
}

func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
