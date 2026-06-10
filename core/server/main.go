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
	"runtime"
	"time"
)

const version = "v1.2.1-beta"
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

func main() {
	os.MkdirAll("downloads", 0755)
	go startUDPDiscovery()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/device/info", deviceInfoHandler)
	http.HandleFunc("/network/info", networkInfoHandler)
	http.HandleFunc("/transfer/upload", uploadHandler)
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("downloads"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("web/assets"))))

	ip := localIP()
	log.Println("TransferLAN+", version)
	log.Println("HTTP:", "http://localhost:5050")
	log.Println("Android:", "http://"+ip+":5050")
	log.Println("UDP discovery activo en puerto 5050")
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
	log.Println("Discovery UDP escuchando en 0.0.0.0:5050")

	for {
		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil { continue }
		msg := string(buf[:n])
		if msg != "TRANSFERLAN_DISCOVER" { continue }
		data, _ := json.Marshal(deviceInfo())
		conn.WriteToUDP(data, remote)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ip := localIP()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html>
<html lang="es">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>TransferLAN+</title>
<style>
body{font-family:system-ui,Arial,sans-serif;background:radial-gradient(circle at top,#0b5cff 0,#0f172a 42%%,#020617 100%%);color:#e5e7eb;padding:24px;min-height:100vh}
.card{max-width:900px;margin:auto;background:rgba(15,23,42,.88);border:1px solid rgba(148,163,184,.25);border-radius:28px;padding:28px;box-shadow:0 24px 80px #000b;backdrop-filter:blur(10px)}
.logo{display:block;width:150px;height:150px;margin:0 auto 14px;border-radius:32px}
h1{text-align:center;font-size:42px;margin:8px 0}
.tag{text-align:center;color:#38bdf8;font-weight:900;font-size:18px;margin-bottom:24px}
.grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(240px,1fr));gap:16px}
.box{background:#0b1220;border:1px solid #334155;border-radius:20px;padding:18px}
code,pre{background:#020617;padding:12px;border-radius:12px;display:block;overflow:auto;color:#dbeafe}
button{background:#38bdf8;color:#00111f;border:0;border-radius:14px;padding:12px 18px;font-weight:900;cursor:pointer;margin:4px}
.ok{color:#22c55e;font-weight:900}.warn{color:#fbbf24;font-weight:900}.small{color:#94a3b8;font-size:14px}
</style>
</head>
<body>
<div class="card">
<img class="logo" src="/assets/logo.png" alt="TransferLAN+">
<h1>TransferLAN+</h1>
<div class="tag">Sin cuentas. Sin nube. Sin cables.</div>

<div class="grid">
<div class="box">
<h2>Estado</h2>
<p class="ok">Servidor activo</p>
<p class="small">Versión %s</p>
<p>Desde esta PC:</p>
<code>http://localhost:5050</code>
<p>Desde Android:</p>
<code>http://%s:5050</code>
</div>

<div class="box">
<h2>Android</h2>
<p>Tocá <b>Buscar dispositivos</b> en la app.</p>
<p class="warn">Si no aparece esta PC, ejecutá el asistente de Firewall.</p>
<code>Windows/PERMITIR_FIREWALL_ADMIN.bat</code>
</div>

<div class="box">
<h2>Subir desde navegador</h2>
<input type="file" id="file">
<br>
<button onclick="upload()">Subir archivo</button>
<pre id="out">Esperando...</pre>
</div>

<div class="box">
<h2>Diagnóstico</h2>
<button onclick="check('/health')">/health</button>
<button onclick="check('/device/info')">/device/info</button>
<button onclick="check('/network/info')">/network/info</button>
<pre id="diag">Sin verificar</pre>
</div>
</div>
</div>
<script>
async function upload(){
 const f=document.getElementById('file').files[0];
 if(!f){alert('Elegí un archivo');return}
 const fd=new FormData(); fd.append('file',f);
 out.textContent='Subiendo...';
 const r=await fetch('/transfer/upload',{method:'POST',body:fd});
 out.textContent=await r.text();
}
async function check(path){
 const r=await fetch(path);
 diag.textContent=JSON.stringify(await r.json(),null,2);
}
check('/device/info');
</script>
</body>
</html>`, version, ip)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"app":"TransferLAN+","version":version,"status":"online","local_ip":localIP(),"tcp_port":httpPort,"udp_port":udpPort})
}
func deviceInfoHandler(w http.ResponseWriter, r *http.Request) { writeJSON(w, deviceInfo()) }
func networkInfoHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"local_ip":localIP(),"tcp_listen":"0.0.0.0:5050","udp_discovery":"0.0.0.0:5050","discovery_message":"TRANSFERLAN_DISCOVER","firewall_hint":"Ejecutar Windows/PERMITIR_FIREWALL_ADMIN.bat como administrador."})
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
	writeJSON(w, map[string]any{"ok":true,"filename":name,"size":written,"sha256":hex.EncodeToString(hash.Sum(nil)),"path":final})
}
func deviceInfo() DeviceInfo {
	ip := localIP()
	host, err := os.Hostname()
	if err != nil || host == "" { host = "TransferLAN-PC" }
	return DeviceInfo{App:"TransferLAN+",Version:version,Name:host,Platform:"desktop",OS:runtime.GOOS,IP:ip,Port:httpPort,BaseURL:fmt.Sprintf("http://%s:%d",ip,httpPort),Status:"available"}
}
func writeJSON(w http.ResponseWriter, v any) { w.Header().Set("Content-Type","application/json; charset=utf-8"); json.NewEncoder(w).Encode(v) }
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
