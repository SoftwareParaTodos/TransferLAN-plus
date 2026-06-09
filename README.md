# TransferLAN+ v1.0.1-beta

Primera beta ejecutable base.

## Qué incluye

- Backend local en Go.
- Servidor HTTP en `http://localhost:5050`.
- Página web simple para probar.
- Endpoint `/health`.
- Endpoint `/devices`.
- Endpoint `/transfer/upload`.
- Carpeta `downloads/`.
- Scripts para Windows y Linux.

## Requisitos

Instalar Go:

https://go.dev/dl/

## Ejecutar en Windows

Doble click en:

```bat
run_windows.bat
```

O desde consola:

```bat
cd core\server
go run .
```

Después abrir:

```text
http://localhost:5050
```

## Ejecutar en Linux

```bash
chmod +x run_linux.sh
./run_linux.sh
```

## Estado

Esta versión todavía no tiene app Android real ni Flutter compilado. Sirve para probar el primer receptor local desde navegador y preparar la base ejecutable.
