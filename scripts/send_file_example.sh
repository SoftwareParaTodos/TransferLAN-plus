#!/usr/bin/env bash
# Uso: ./scripts/send_file_example.sh http://192.168.1.10:47231 /ruta/video.mp4
go run ./cmd/transferlan-discovery --mode send --target "$1" --file "$2"
