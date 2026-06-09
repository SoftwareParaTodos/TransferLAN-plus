package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const version = "v1.0.1-beta"
const port = ":5050"

type HealthResponse struct {
	App       string `json:"app"`
	Version   string `json:"version"`
	Status    string `json:"status"`
	LocalIP   string `json:"local_ip"`
	Timestamp string `json:"timestamp"`
}

type Device struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Platform string `json:"platform"`
	Status   string `json:"status"`
}

func main() {
	must(os.MkdirAll("downloads", 0755))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/devices", devicesHandler)
	http.HandleFunc("/transfer/upload", uploadHandler)
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("downloads"))))

	log.Println("TransferLAN+ " + version)
	log.Println("Servidor iniciado en http://localhost" + port)
	log.Println("IP local probable:", localIP())
	log.Fatal(http.ListenAndServe(port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join("web", "index.html"))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, HealthResponse{
		App:       "TransferLAN+",
		Version:   version,
		Status:    "online",
		LocalIP:   localIP(),
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

func devicesHandler(w http.ResponseWriter, r *http.Request) {
	// Base simulada: después se reemplaza por mDNS real.
	writeJSON(w, []Device{
		{Name: "Este equipo", Address: "http://" + localIP() + "5050", Platform: runtimePlatform(), Status: "online"},
	})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "usar POST", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(64 << 20)
	if err != nil {
		http.Error(w, "no se pudo leer multipart: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "campo file requerido: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	dstPath := filepath.Join("downloads", safeName(header.Filename))
	dst, err := os.Create(dstPath + ".part")
	if err != nil {
		http.Error(w, "no se pudo crear archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(dst, hasher), file)
	if err != nil {
		http.Error(w, "error guardando archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}
	dst.Close()

	finalPath := filepath.Join("downloads", safeName(header.Filename))
	if err := os.Rename(dstPath+".part", finalPath); err != nil {
		http.Error(w, "error finalizando archivo: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"ok":       true,
		"filename": header.Filename,
		"size":     written,
		"sha256":   hex.EncodeToString(hasher.Sum(nil)),
		"path":     finalPath,
		"url":      "/downloads/" + safeName(header.Filename),
	})
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

func localIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	addr := conn.LocalAddr().(*net.UDPAddr)
	return addr.IP.String()
}

func safeName(name string) string {
	if name == "" {
		return fmt.Sprintf("archivo_%d.bin", time.Now().Unix())
	}
	return filepath.Base(name)
}

func runtimePlatform() string {
	return "desktop"
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
