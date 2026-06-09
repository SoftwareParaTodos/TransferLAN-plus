#!/usr/bin/env bash
# Cambiar IP y archivo según corresponda.
go run ./cmd/transferlan-discovery --mode send-chunked --target http://192.168.1.10:47231 --file "/ruta/video.mp4"
