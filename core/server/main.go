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
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

const version = "v1.2.5-beta"
const httpPort = 5050
const udpPort = 5050

type DeviceInfo struct {
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
	os.MkdirAll("downloads", 0755)
	go startUDPDiscovery()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/device/info", deviceInfoHandler)
	http.HandleFunc("/network/info", networkInfoHandler)
	http.HandleFunc("/pairing/info", pairingInfoHandler)
	http.HandleFunc("/pairing/code", pairingCodeHandler)
	http.HandleFunc("/pairing/qr.png", pairingQRPNGHandler)
	http.HandleFunc("/history", historyHandler)
	http.HandleFunc("/transfer/upload", uploadHandler)
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("downloads"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))

	ip := localIP()
	log.Println("TransferLAN+", version)
	log.Println("PC:", "http://localhost:5050")
	log.Println("Android:", "http://"+ip+":5050")
	log.Println("QR:", "http://"+ip+":5050/pairing/qr.png")
	log.Fatal(http.ListenAndServe("0.0.0.0:5050", nil))
}

func startUDPDiscovery() {
	addr := net.UDPAddr{Port: udpPort, IP: net.ParseIP("0.0.0.0")}
	conn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		log.Println("No se pudo iniciar UDP discovery:", err)
		return
	}
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil { continue }
		if string(buf[:n]) != "TRANSFERLAN_DISCOVER" { continue }
		data, _ := json.Marshal(deviceInfo())
		conn.WriteToUDP(data, remote)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ip := localIP()
	pair := pairingURL()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html lang="es"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>TransferLAN+</title>
<style>
body{font-family:system-ui,Arial,sans-serif;background:radial-gradient(circle at top,#0b5cff 0,#0f172a 44%%,#020617 100%%);color:#e5e7eb;padding:24px;min-height:100vh}
.card{max-width:1080px;margin:auto;background:rgba(15,23,42,.92);border:1px solid rgba(148,163,184,.25);border-radius:28px;padding:28px;box-shadow:0 24px 80px #000b}
.logo{display:block;width:140px;height:140px;margin:0 auto 14px;border-radius:32px}
h1{text-align:center;font-size:42px;margin:8px 0}.tag{text-align:center;color:#38bdf8;font-weight:900;font-size:18px;margin-bottom:24px}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(280px,1fr));gap:16px}.box{background:#0b1220;border:1px solid #334155;border-radius:20px;padding:18px}
code,pre,textarea{background:#020617;padding:12px;border-radius:12px;display:block;overflow:auto;color:#dbeafe;word-break:break-all;width:100%%;box-sizing:border-box;border:1px solid #334155}
button{background:#38bdf8;color:#00111f;border:0;border-radius:14px;padding:12px 18px;font-weight:900;cursor:pointer;margin:4px}
.ok{color:#22c55e;font-weight:900}.small{color:#94a3b8;font-size:14px}.pair{font-size:13px;min-height:95px}
.qr{display:block;margin:10px auto;background:#fff;border-radius:18px;padding:12px;max-width:260px;width:100%%}.item{border-bottom:1px solid #334155;padding:10px 0}
</style></head>
<body><div class="card">
<img class="logo" src="/assets/logo.png" alt="TransferLAN+" onerror="this.style.display='none'">
<h1>TransferLAN+</h1><div class="tag">Sin cuentas. Sin nube. Sin cables.</div>
<div class="grid">
<div class="box"><h2>Estado</h2><p class="ok">Servidor activo</p><p class="small">Versión %s</p><p>Desde esta PC:</p><code>http://localhost:5050</code><p>Desde Android:</p><code>http://%s:5050</code></div>
<div class="box"><h2>QR de emparejamiento</h2><p>Escanealo con Cámara/Google Lens, copiá el link y pegalo en Android.</p><img class="qr" src="/pairing/qr.png?t=%d" alt="QR TransferLAN+"></div>
<div class="box"><h2>Código alternativo</h2><textarea class="pair" id="pair">%s</textarea><button onclick="copyPair()">Copiar código</button></div>
<div class="box"><h2>Últimos recibidos</h2><button onclick="loadHistory()">Actualizar historial</button><div id="hist">Cargando...</div></div>
<div class="box"><h2>Descargas</h2><a style="color:#38bdf8" href="/downloads/" target="_blank">Ver carpeta desde navegador</a></div>
<div class="box"><h2>Diagnóstico</h2><button onclick="check('/health')">/health</button><button onclick="check('/pairing/info')">/pairing/info</button><pre id="diag">Sin verificar</pre></div>
</div></div>
<script>
async function check(path){const r=await fetch(path);const ct=r.headers.get('content-type')||'';diag.textContent=ct.includes('application/json')?JSON.stringify(await r.json(),null,2):await r.text();}
function copyPair(){const el=document.getElementById('pair');el.select();document.execCommand('copy');alert('Código copiado');}
function fmt(b){if(!b)return'0 B';let u=['B','KB','MB','GB','TB'];let i=0,v=b;while(v>=1024&&i<u.length-1){v/=1024;i++}return v.toFixed(1)+' '+u[i]}
async function loadHistory(){const r=await fetch('/history');const data=await r.json();hist.innerHTML='';if(!data.items||!data.items.length){hist.innerHTML='<p class="small">Sin archivos recibidos todavía.</p>';return}data.items.slice().reverse().slice(0,8).forEach(x=>{let d=document.createElement('div');d.className='item';d.innerHTML='<b>'+x.filename+'</b><br><span class="small">'+fmt(x.size)+' · '+x.time+'</span><br><a style="color:#38bdf8" href="/downloads/'+encodeURIComponent(x.filename)+'" target="_blank">Abrir</a>';hist.appendChild(d);});}
check('/health');loadHistory();
</script></body></html>`, version, ip, time.Now().Unix(), pair)
}

func pairingQRPNGHandler(w http.ResponseWriter, r *http.Request) {
	png, err := qrcode.Encode(pairingURL(), qrcode.Medium, 512)
	if err != nil {
		http.Error(w, "no se pudo generar QR: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"app":"TransferLAN+","version":version,"status":"online","local_ip":localIP(),"tcp_port":httpPort,"udp_port":udpPort})
}
func deviceInfoHandler(w http.ResponseWriter, r *http.Request) { writeJSON(w, deviceInfo()) }
func networkInfoHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"local_ip":localIP(),"tcp_listen":"0.0.0.0:5050","udp_discovery":"0.0.0.0:5050","discovery_message":"TRANSFERLAN_DISCOVER"})
}
func pairingInfoHandler(w http.ResponseWriter, r *http.Request) {
	info := deviceInfo()
	writeJSON(w, map[string]any{"type":"transferlan_pairing","scheme":"transferlan://connect","device":info,"pairing_url":pairingURL()})
}
func pairingCodeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, pairingURL())
}
func historyHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"items": readHistory()})
}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w,"usar POST",405); return }
	if err := r.ParseMultipartForm(128 << 20); err != nil { http.Error(w,err.Error(),400); return }
	file, header, err := r.FormFile("file")
	if err != nil { http.Error(w,err.Error(),400); return }
	defer file.Close()
	name := safeName(header.Filename)
	tmp := filepath.Join("downloads", name+".part")
	final := filepath.Join("downloads", name)
	dst, err := os.Create(tmp)
	if err != nil { http.Error(w,err.Error(),500); return }
	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(dst, hash), file)
	dst.Close()
	if err != nil { http.Error(w,err.Error(),500); return }
	if err := os.Rename(tmp, final); err != nil { http.Error(w,err.Error(),500); return }
	sum := hex.EncodeToString(hash.Sum(nil))
	appendHistory(HistoryItem{Filename:name, Size:written, SHA256:sum, Path:final, Time:time.Now().Format("2006-01-02 15:04:05")})
	writeJSON(w, map[string]any{"ok":true,"filename":name,"size":written,"sha256":sum,"path":final,"message":"Archivo recibido correctamente"})
}
func deviceInfo() DeviceInfo {
	ip := localIP()
	host, err := os.Hostname()
	if err != nil || host == "" { host = "TransferLAN-PC" }
	return DeviceInfo{App:"TransferLAN+",Version:version,Name:host,Platform:"desktop",OS:runtime.GOOS,IP:ip,Port:httpPort,BaseURL:fmt.Sprintf("http://%s:%d",ip,httpPort),Status:"available"}
}
func pairingURL() string {
	info := deviceInfo()
	q := url.Values{}
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
	if err != nil { return items }
	json.Unmarshal(data, &items)
	return items
}
func appendHistory(item HistoryItem) {
	items := readHistory()
	items = append(items, item)
	if len(items) > 100 { items = items[len(items)-100:] }
	data, _ := json.MarshalIndent(items, "", "  ")
	os.WriteFile("history.json", data, 0644)
}
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type","application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}
func localIP() string {
	conn, err := net.Dial("udp","8.8.8.8:80")
	if err != nil { return "127.0.0.1" }
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
func safeName(name string) string {
	if name == "" { return fmt.Sprintf("archivo_%d.bin", time.Now().Unix()) }
	return filepath.Base(name)
}
