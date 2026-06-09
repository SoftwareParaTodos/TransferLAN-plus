@echo off
REM Uso: scripts\send_file_example.bat http://192.168.1.10:47231 C:\ruta\video.mp4
go run ./cmd/transferlan-discovery --mode send --target %1 --file %2
