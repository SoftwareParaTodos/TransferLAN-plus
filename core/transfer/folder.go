package transfer

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FolderSendOptions struct {
	Target      string
	FolderPath  string
	Token       string
	ChunkSize   int64
	MaxRetries  int
	KeepArchive bool
}

type FolderSendResult struct {
	ArchivePath string
	FileCount   int
	FolderName  string
	Result      *ChunkedSendResult
}

func SendFolderChunked(opts FolderSendOptions, onProgress func(sent int64, total int64)) (*FolderSendResult, error) {
	info, err := os.Stat(opts.FolderPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("la ruta no es una carpeta: %s", opts.FolderPath)
	}

	archivePath, fileCount, err := CreateFolderArchive(opts.FolderPath)
	if err != nil {
		return nil, err
	}
	if !opts.KeepArchive {
		defer os.Remove(archivePath)
	}

	result, err := SendFileChunked(ChunkedSendOptions{
		Target: opts.Target, FilePath: archivePath, Token: opts.Token, ChunkSize: opts.ChunkSize, MaxRetries: opts.MaxRetries,
	}, onProgress)
	if err != nil {
		return nil, err
	}
	return &FolderSendResult{ArchivePath: archivePath, FileCount: fileCount, FolderName: filepath.Base(opts.FolderPath), Result: result}, nil
}

func CreateFolderArchive(folderPath string) (string, int, error) {
	absFolder, err := filepath.Abs(folderPath)
	if err != nil {
		return "", 0, err
	}
	base := filepath.Base(absFolder)
	archiveName := sanitizeArchiveName(base) + "_" + time.Now().Format("20060102_150405") + ".zip"
	archivePath := filepath.Join(os.TempDir(), archiveName)

	out, err := os.Create(archivePath)
	if err != nil {
		return "", 0, err
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	fileCount := 0
	err = filepath.WalkDir(absFolder, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == absFolder {
			return nil
		}
		rel, err := filepath.Rel(absFolder, path)
		if err != nil {
			return err
		}
		zipName := filepath.ToSlash(filepath.Join(base, rel))
		if d.IsDir() {
			_, err := zw.Create(zipName + "/")
			return err
		}

		info, err := d.Info()
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = zipName
		header.Method = zip.Deflate

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		if _, err := io.Copy(writer, in); err != nil {
			return err
		}
		fileCount++
		return nil
	})
	if err != nil {
		os.Remove(archivePath)
		return "", 0, err
	}
	return archivePath, fileCount, nil
}

func sanitizeArchiveName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "carpeta"
	}
	var b strings.Builder
	for _, r := range name {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
