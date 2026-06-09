package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/softwareparatodos/transferlan-plus/core/config"
	"github.com/softwareparatodos/transferlan-plus/core/discovery"
	"github.com/softwareparatodos/transferlan-plus/core/pairing"
	"github.com/softwareparatodos/transferlan-plus/core/server"
	"github.com/softwareparatodos/transferlan-plus/core/transfer"
)

type pairReq struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	Platform string `json:"platform"`
	PIN      string `json:"pin,omitempty"`
}

func main() {
	mode := flag.String("mode", "both", "Modo: announce, browse, both, server, pair, pair-pin, guest-link, send, send-chunked, send-folder o incomplete")
	name := flag.String("name", hostname(), "Nombre visible del dispositivo")
	port := flag.Int("port", config.DefaultPort, "Puerto local anunciado")
	seconds := flag.Int("seconds", 8, "Tiempo de búsqueda en segundos")
	target := flag.String("target", "", "URL del receptor para pair/send. Ej: http://192.168.1.10:47231")
	filePath := flag.String("file", "", "Archivo/carpeta a enviar en modo send")
	downloadDir := flag.String("download-dir", "downloads", "Carpeta donde guardar archivos recibidos")
	token := flag.String("token", "", "Token de dispositivo confiable para transferencias futuras")
	pin := flag.String("pin", "", "PIN temporal mostrado por el receptor para emparejar")
	deviceID := flag.String("device-id", hostname()+"-"+runtime.GOOS, "ID local simple para emparejamiento")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("TransferLAN+ v" + config.ProtocolVer)
	fmt.Println("Servicio:", config.ServiceType)

	if *mode == "server" || *mode == "announce" || *mode == "both" {
		store, err := pairing.NewStore("")
		if err != nil {
			fmt.Println("Error iniciando store:", err)
			os.Exit(1)
		}
		srv := &server.Server{Name: *name, Platform: runtime.GOOS, Store: store, Sessions: pairing.NewSessionManager(2 * time.Minute), DownloadDir: *downloadDir}
		go func() {
			if err := srv.ListenAndServe(*port); err != nil {
				fmt.Println("Servidor finalizado:", err)
			}
		}()
	}

	if *mode == "announce" || *mode == "both" {
		announcer := discovery.NewAnnouncer()
		if err := announcer.Start(ctx, *name, *port, runtime.GOOS); err != nil {
			fmt.Println("Error anunciando dispositivo:", err)
			os.Exit(1)
		}
		defer announcer.Shutdown()
		fmt.Printf("Disponible como %q en puerto %d\n", *name, *port)
	}

	if *mode == "browse" || *mode == "both" {
		browser := discovery.NewBrowser()
		fmt.Printf("Buscando dispositivos por %d segundos...\n", *seconds)
		devices, err := browser.Browse(ctx, time.Duration(*seconds)*time.Second)
		if err != nil {
			fmt.Println("Error buscando dispositivos:", err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Println("No se detectaron dispositivos TransferLAN+ en la red local.")
		} else {
			fmt.Println("Dispositivos detectados:")
			for i, d := range devices {
				fmt.Printf("%d) %s | %s | %s | v%s\n", i+1, d.Name, d.Address(), d.Platform, d.Version)
			}
		}
	}

	if *mode == "send" {
		if *target == "" || *filePath == "" {
			fmt.Println("Falta --target y/o --file. Ej: --mode send --target http://192.168.1.10:47231 --file video.mp4")
			os.Exit(1)
		}
		if err := sendFile(*target, *filePath, *token); err != nil {
			fmt.Println("Error enviando archivo:", err)
			os.Exit(1)
		}
	}

	if *mode == "send-chunked" {
		if *target == "" || *filePath == "" {
			fmt.Println("Falta --target y/o --file. Ej: --mode send-chunked --target http://192.168.1.10:47231 --file video.mp4")
			os.Exit(1)
		}
		if err := sendFileChunked(*target, *filePath, *token); err != nil {
			fmt.Println("Error enviando archivo por bloques:", err)
			os.Exit(1)
		}
	}

	if *mode == "send-folder" {
		if *target == "" || *filePath == "" {
			fmt.Println("Falta --target y/o --file. Ej: --mode send-folder --target http://192.168.1.10:47231 --file ./MiCarpeta")
			os.Exit(1)
		}
		if err := sendFolderChunked(*target, *filePath, *token); err != nil {
			fmt.Println("Error enviando carpeta por bloques:", err)
			os.Exit(1)
		}
	}

	if *mode == "incomplete" {
		if *target == "" {
			fmt.Println("Falta --target. Ej: --mode incomplete --target http://192.168.1.10:47231")
			os.Exit(1)
		}
		if err := listIncomplete(*target); err != nil {
			fmt.Println("Error consultando transferencias incompletas:", err)
			os.Exit(1)
		}
	}

	if *mode == "guest-link" {
		if *target == "" {
			fmt.Println("Falta --target. Ej: --mode guest-link --target http://192.168.1.10:47231")
			os.Exit(1)
		}
		if err := createGuestLink(*target); err != nil {
			fmt.Println("Error creando enlace invitado:", err)
			os.Exit(1)
		}
	}

	if *mode == "pair-pin" {
		if *target == "" {
			fmt.Println("Falta --target. Ej: --mode pair-pin --target http://192.168.1.10:47231")
			os.Exit(1)
		}
		if err := startPairPIN(*target); err != nil {
			fmt.Println("Error generando PIN:", err)
			os.Exit(1)
		}
	}

	if *mode == "pair" {
		if *target == "" {
			fmt.Println("Falta --target. Ej: --target http://192.168.1.10:47231")
			os.Exit(1)
		}
		if err := requestPair(*target, *deviceID, *name, *pin); err != nil {
			fmt.Println("Error emparejando:", err)
			os.Exit(1)
		}
	}

	if *mode == "announce" || *mode == "server" {
		fmt.Println("Activo. Presioná Ctrl+C para salir.")
		<-ctx.Done()
	}
}

func sendFolderChunked(target, folderPath, token string) error {
	lastPercent := int64(-1)
	result, err := transfer.SendFolderChunked(transfer.FolderSendOptions{Target: target, FolderPath: folderPath, Token: token}, func(sent int64, total int64) {
		if total <= 0 {
			return
		}
		percent := sent * 100 / total
		if percent != lastPercent && percent%5 == 0 {
			lastPercent = percent
			fmt.Printf("Progreso carpeta comprimida: %d%% (%d/%d bytes)\n", percent, sent, total)
		}
	})
	if err != nil {
		return err
	}
	fmt.Println("Carpeta:", result.FolderName)
	fmt.Println("Archivos incluidos:", result.FileCount)
	fmt.Println("ZIP temporal:", result.ArchivePath)
	fmt.Println("Upload ID:", result.Result.UploadID)
	fmt.Println("Respuesta:", result.Result.Response)
	fmt.Println("SHA256:", result.Result.SHA256)
	fmt.Println("Tamaño:", result.Result.SizeBytes, "bytes")
	fmt.Println("Tiempo:", result.Result.Duration.Round(time.Millisecond))
	return nil
}

func sendFileChunked(target, filePath, token string) error {
	lastPercent := int64(-1)
	result, err := transfer.SendFileChunked(transfer.ChunkedSendOptions{Target: target, FilePath: filePath, Token: token}, func(sent int64, total int64) {
		if total <= 0 {
			return
		}
		percent := sent * 100 / total
		if percent != lastPercent && percent%5 == 0 {
			lastPercent = percent
			fmt.Printf("Progreso por bloques: %d%% (%d/%d bytes)\n", percent, sent, total)
		}
	})
	if err != nil {
		return err
	}
	fmt.Println("Upload ID:", result.UploadID)
	fmt.Println("Respuesta:", result.Response)
	fmt.Println("SHA256:", result.SHA256)
	fmt.Println("Tamaño:", result.SizeBytes, "bytes")
	fmt.Println("Tiempo:", result.Duration.Round(time.Millisecond))
	return nil
}

func sendFile(target, filePath, token string) error {
	lastPercent := int64(-1)
	started := time.Now()
	result, err := transfer.SendFile(target, filePath, token, func(read int64, total int64) {
		if total <= 0 {
			return
		}
		percent := read * 100 / total
		if percent != lastPercent && percent%5 == 0 {
			lastPercent = percent
			fmt.Printf("Progreso: %d%% (%d/%d bytes)\n", percent, read, total)
		}
	})
	if err != nil {
		return err
	}
	fmt.Println("Respuesta:", result.StatusCode, result.Response)
	fmt.Println("SHA256:", result.SHA256)
	fmt.Println("Tamaño:", result.SizeBytes, "bytes")
	fmt.Println("Tiempo:", time.Since(started).Round(time.Millisecond))
	return nil
}

func startPairPIN(target string) error {
	resp, err := http.Post(target+"/pair/pin/start", "application/json", bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	fmt.Println("PIN del receptor:", out)
	return nil
}

func requestPair(target, deviceID, name, pin string) error {
	body, _ := json.Marshal(pairReq{DeviceID: deviceID, Name: name, Platform: runtime.GOOS, PIN: pin})
	resp, err := http.Post(target+"/pair/request", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	fmt.Println("Respuesta del receptor:", out)
	return nil
}

func listIncomplete(target string) error {
	resp, err := http.Get(target + "/transfer/chunked/incomplete")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var out []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if len(out) == 0 {
		fmt.Println("No hay transferencias incompletas en el receptor.")
		return nil
	}
	fmt.Println("Transferencias incompletas reanudables:")
	for i, item := range out {
		fmt.Printf("%d) %v | recibido: %v/%v bytes | upload_id: %v\n", i+1, item["file_name"], item["received_size"], item["size_bytes"], item["upload_id"])
	}
	return nil
}

func hostname() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "TransferLAN-Device"
	}
	return h
}

func createGuestLink(target string) error {
	body := bytes.NewReader([]byte(`{"mode":"upload"}`))
	resp, err := http.Post(target+"/guest/share/start", "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	fmt.Println("Respuesta:", string(data))
	return nil
}
