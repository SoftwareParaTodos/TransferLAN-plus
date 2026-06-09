@echo off
go run ./cmd/transferlan-discovery --mode send-folder --target http://127.0.0.1:47231 --file ./docs --token PEGAR_TOKEN
pause
