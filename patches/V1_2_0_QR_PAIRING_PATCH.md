# Patch v1.2.0-beta QR Pairing Parte 1

En `core/server/main.go`, agregar imports si faltan:

```go
"net/url"
```

Agregar handlers dentro de `main()`:

```go
http.HandleFunc("/pairing/info", pairingInfoHandler)
http.HandleFunc("/pairing/qr", pairingQRHandler)
```

Agregar estas funciones al final:

```go
func pairingInfoHandler(w http.ResponseWriter, r *http.Request) {
	info := deviceInfo()
	writeJSON(w, map[string]any{
		"type": "transferlan_pairing",
		"scheme": "transferlan://connect",
		"device": info,
		"pairing_url": pairingURL(),
	})
}

func pairingQRHandler(w http.ResponseWriter, r *http.Request) {
	ip := localIP()
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="512" height="512" viewBox="0 0 512 512">
<rect width="512" height="512" rx="60" fill="#ffffff"/>
<rect x="40" y="40" width="120" height="120" fill="#020617"/>
<rect x="70" y="70" width="60" height="60" fill="#ffffff"/>
<rect x="352" y="40" width="120" height="120" fill="#020617"/>
<rect x="382" y="70" width="60" height="60" fill="#ffffff"/>
<rect x="40" y="352" width="120" height="120" fill="#020617"/>
<rect x="70" y="382" width="60" height="60" fill="#ffffff"/>
<text x="256" y="235" text-anchor="middle" font-family="Arial" font-size="34" font-weight="700" fill="#020617">TransferLAN+</text>
<text x="256" y="280" text-anchor="middle" font-family="Arial" font-size="22" fill="#020617">%s:5050</text>
<text x="256" y="318" text-anchor="middle" font-family="Arial" font-size="18" fill="#334155">QR pairing base</text>
</svg>`, ip)
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
```

Cambiar versión:

```go
const version = "v1.2.0-beta"
```
