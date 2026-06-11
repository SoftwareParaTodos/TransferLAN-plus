package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

const version = "v1.5.9-beta"
const httpPort = 5050
const udpPort = 5050

type DeviceInfo struct {
	ID       string `json:"id"`
	DeviceID string `json:"device_id"`
	App      string `json:"app"`
	Version  string `json:"version"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
	OS       string `json:"os"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	BaseURL  string `json:"base_url"`
	Status   string `json:"status"`
}

type HistoryItem struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	SHA256   string `json:"sha256"`
	Path     string `json:"path"`
	Time     string `json:"time"`
}

func main() {
	_ = os.MkdirAll("downloads", 0755)
	go startUDPDiscovery()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/device/info", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, deviceInfo()) })
	http.HandleFunc("/network/info", networkInfoHandler)
	http.HandleFunc("/pairing/info", pairingInfoHandler)
	http.HandleFunc("/pairing/code", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, pairingURL()) })
	http.HandleFunc("/pairing/qr.png", pairingQRPNGHandler)
	http.HandleFunc("/history", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, map[string]any{"items": readHistory()}) })
	http.HandleFunc("/transfer/upload", uploadHandler)
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("downloads"))))

	ip := localIP()
	log.Println("TransferLAN+", version)
	log.Println("PC: http://localhost:5050")
	log.Println("Android: http://" + ip + ":5050")
	log.Println("Codigo de pareo:", pairingURL())
	log.Fatal(http.ListenAndServe("0.0.0.0:5050", nil))
}

func startUDPDiscovery() {
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{Port: udpPort, IP: net.ParseIP("0.0.0.0")})
	if err != nil {
		log.Println("Aviso: no se pudo iniciar UDP discovery:", err)
		return
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil || string(buf[:n]) != "TRANSFERLAN_DISCOVER" {
			continue
		}
		data, _ := json.Marshal(deviceInfo())
		_, _ = conn.WriteToUDP(data, remote)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	info := deviceInfo()
	pair := pairingURL()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>TransferLAN+</title>
<style>body{font-family:system-ui;background:#020617;color:#e5e7eb;padding:24px}.card{max-width:980px;margin:auto;background:#0f172a;border:1px solid #334155;border-radius:28px;padding:28px}h1{text-align:center;font-size:42px}.tag{text-align:center;color:#38bdf8;font-weight:900}.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px}.box{background:#111827;border:1px solid #334155;border-radius:20px;padding:18px}code,pre,textarea{background:#020617;color:#dbeafe;padding:12px;border-radius:12px;display:block;word-break:break-all;width:100%%;box-sizing:border-box}button{background:#38bdf8;color:#00111f;border:0;border-radius:14px;padding:12px 18px;font-weight:900}.qr{display:block;margin:auto;background:white;border-radius:18px;padding:12px;max-width:260px;width:100%%}.ok{color:#22c55e;font-weight:900}a{color:#38bdf8}</style></head>
<body><div class="card"><h1>TransferLAN+</h1><div class="tag">Sin cuentas. Sin nube. Sin cables.</div><div class="grid">
<div class="box"><h2>Estado</h2><p class="ok">Servidor activo</p><p>Versión</p><code>%s</code><p>Equipo</p><code>%s</code><p>Device ID</p><code>%s</code><p>Android</p><code>%s</code></div>
<div class="box"><h2>QR</h2><img class="qr" src="/pairing/qr.png"></div>
<div class="box"><h2>Código</h2><textarea id="pair" rows="5">%s</textarea><button onclick="copyPair()">Copiar</button></div>
<div class="box"><h2>Descargas</h2><a href="/downloads/" target="_blank">Ver carpeta</a></div>
<div class="box"><h2>Historial</h2><button onclick="loadHistory()">Actualizar</button><pre id="hist">Cargando...</pre></div>
<div class="box"><h2>Diagnóstico</h2><button onclick="check('/device/info')">/device/info</button><button onclick="check('/health')">/health</button><pre id="diag"></pre></div>
</div></div><script>
function copyPair(){pair.select();document.execCommand('copy');alert('Código copiado')}
async function check(p){let r=await fetch(p);diag.textContent=JSON.stringify(await r.json(),null,2)}
async function loadHistory(){let r=await fetch('/history');hist.textContent=JSON.stringify(await r.json(),null,2)}
loadHistory()
</script></body></html>`, version, info.Name, info.DeviceID, info.BaseURL, pair)
}

func pairingQRPNGHandler(w http.ResponseWriter, r *http.Request) {
	png, err := qrcode.Encode(pairingURL(), qrcode.Medium, 512)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	_, _ = w.Write(png)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"app": "TransferLAN+", "version": version, "status": "online", "local_ip": localIP(), "device_id": deviceInfo().DeviceID})
}

func networkInfoHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"local_ip": localIP(), "tcp_listen": "0.0.0.0:5050", "udp_discovery": "0.0.0.0:5050", "discovery_message": "TRANSFERLAN_DISCOVER"})
}

func pairingInfoHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"type": "transferlan_pairing", "scheme": "transferlan://connect", "device": deviceInfo(), "pairing_url": pairingURL()})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "usar POST", 405)
		return
	}
	if err := r.ParseMultipartForm(128 << 20); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	defer file.Close()

	name := safeName(header.Filename)
	tmp := filepath.Join("downloads", name+".part")
	final := filepath.Join("downloads", name)
	dst, err := os.Create(tmp)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(dst, hash), file)
	_ = dst.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := os.Rename(tmp, final); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	sum := hex.EncodeToString(hash.Sum(nil))
	appendHistory(HistoryItem{Filename: name, Size: written, SHA256: sum, Path: final, Time: time.Now().Format("2006-01-02 15:04:05")})
	writeJSON(w, map[string]any{"ok": true, "filename": name, "size": written, "sha256": sum, "path": final, "message": "Archivo recibido correctamente"})
}

func deviceInfo() DeviceInfo {
	ip := localIP()
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "TransferLAN-PC"
	}
	id := deviceID(host)
	return DeviceInfo{ID: id, DeviceID: id, App: "TransferLAN+", Version: version, Name: host, Platform: "desktop", OS: runtime.GOOS, IP: ip, Port: httpPort, BaseURL: fmt.Sprintf("http://%s:%d", ip, httpPort), Status: "available"}
}

func deviceID(host string) string {
	path := "device_id.txt"
	if data, err := os.ReadFile(path); err == nil && len(string(data)) >= 8 {
		return string(data)
	}
	sum := sha1.Sum([]byte(fmt.Sprintf("%s|%s|%d", host, runtime.GOOS, time.Now().UnixNano())))
	id := hex.EncodeToString(sum[:])[:12]
	_ = os.WriteFile(path, []byte(id), 0644)
	return id
}

func pairingURL() string {
	info := deviceInfo()
	q := url.Values{}
	q.Set("device_id", info.DeviceID)
	q.Set("name", info.Name)
	q.Set("ip", info.IP)
	q.Set("port", fmt.Sprintf("%d", info.Port))
	q.Set("version", info.Version)
	q.Set("base_url", info.BaseURL)
	return "transferlan://connect?" + q.Encode()
}

func readHistory() []HistoryItem {
	var items []HistoryItem
	data, err := os.ReadFile("history.json")
	if err != nil {
		return items
	}
	_ = json.Unmarshal(data, &items)
	return items
}

func appendHistory(item HistoryItem) {
	items := append(readHistory(), item)
	if len(items) > 100 {
		items = items[len(items)-100:]
	}
	data, _ := json.MarshalIndent(items, "", "  ")
	_ = os.WriteFile("history.json", data, 0644)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(v)
}

func localIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func safeName(name string) string {
	if name == "" {
		return fmt.Sprintf("archivo_%d.bin", time.Now().Unix())
	}
	return filepath.Base(name)
}
