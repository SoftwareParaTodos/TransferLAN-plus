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

const version = "v1.0.7-beta"
const port = ":5050"

func main() {
	os.MkdirAll("downloads", 0755)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/device/info", deviceInfoHandler)
	http.HandleFunc("/network/info", networkInfoHandler)
	http.HandleFunc("/transfer/upload", uploadHandler)
	http.Handle("/downloads/", http.StripPrefix("/downloads/", http.FileServer(http.Dir("downloads"))))
	ip := localIP()
	log.Println("TransferLAN+", version)
	log.Println("PC:", "http://localhost:5050")
	log.Println("Celular:", "http://"+ip+":5050")
	log.Fatal(http.ListenAndServe(port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	ip := localIP()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!doctype html><html><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>TransferLAN+</title><style>body{font-family:system-ui;background:#0f172a;color:#e5e7eb;padding:24px}.card{max-width:820px;margin:auto;background:#111827;border:1px solid #334155;border-radius:18px;padding:24px}code,pre{background:#020617;padding:10px;border-radius:10px;display:block;overflow:auto}button{background:#38bdf8;border:0;border-radius:12px;padding:12px 18px;font-weight:800}</style></head><body><div class="card"><h1>TransferLAN+</h1><p><b>Sin cuentas. Sin nube. Sin cables.</b></p><p>Beta %s</p><h2>Estado</h2><code>http://localhost:5050</code><p>Desde Android:</p><code>http://%s:5050</code><p>Prueba:</p><code>http://%s:5050/health</code><h2>Subir archivo</h2><input type="file" id="file"><button onclick="upload()">Subir</button><pre id="out">Esperando...</pre></div><script>async function upload(){let f=document.getElementById('file').files[0];if(!f){alert('Elegí archivo');return}let fd=new FormData();fd.append('file',f);out.textContent='Subiendo...';let r=await fetch('/transfer/upload',{method:'POST',body:fd});out.textContent=await r.text();}</script></body></html>`, version, ip, ip)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"app":"TransferLAN+","version":version,"status":"online","local_ip":localIP(),"port":5050})
}
func deviceInfoHandler(w http.ResponseWriter, r *http.Request) {
	ip := localIP()
	host, _ := os.Hostname()
	if host == "" { host = "TransferLAN-PC" }
	writeJSON(w, map[string]any{"app":"TransferLAN+","version":version,"name":host,"platform":"desktop","os":runtime.GOOS,"ip":ip,"port":5050,"base_url":"http://"+ip+":5050","status":"available"})
}
func networkInfoHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"local_ip":localIP(),"port":5050,"listen":"0.0.0.0:5050","firewall_hint":"Ejecutar Windows/PERMITIR_FIREWALL_ADMIN.bat como administrador"})
}
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { http.Error(w,"usar POST",405); return }
	if err := r.ParseMultipartForm(128 << 20); err != nil { http.Error(w,err.Error(),400); return }
	file, header, err := r.FormFile("file")
	if err != nil { http.Error(w,err.Error(),400); return }
	defer file.Close()
	name := filepath.Base(header.Filename)
	if name == "." || name == "" { name = fmt.Sprintf("archivo_%d.bin", time.Now().Unix()) }
	tmp := filepath.Join("downloads", name+".part")
	final := filepath.Join("downloads", name)
	dst, err := os.Create(tmp)
	if err != nil { http.Error(w,err.Error(),500); return }
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(dst,h), file)
	dst.Close()
	if err != nil { http.Error(w,err.Error(),500); return }
	if err := os.Rename(tmp, final); err != nil { http.Error(w,err.Error(),500); return }
	writeJSON(w, map[string]any{"ok":true,"filename":name,"size":n,"sha256":hex.EncodeToString(h.Sum(nil)),"path":final})
}
func writeJSON(w http.ResponseWriter, v any) { w.Header().Set("Content-Type","application/json; charset=utf-8"); json.NewEncoder(w).Encode(v) }
func localIP() string {
	conn, err := net.Dial("udp","8.8.8.8:80")
	if err != nil { return "127.0.0.1" }
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}
