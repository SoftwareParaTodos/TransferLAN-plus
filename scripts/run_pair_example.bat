@echo off
REM Cambiar la IP por la detectada con run_browse.bat
go run ./cmd/transferlan-discovery --mode pair --name TransferLAN-Android --target http://192.168.1.10:47231
pause
