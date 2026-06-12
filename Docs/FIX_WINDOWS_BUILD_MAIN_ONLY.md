# Fix Windows Build — compilar solo main.go

El error anterior ocurría porque en `core/server` había archivos viejos como:

- `server.go`
- `queue_handlers.go`

con imports a paquetes que todavía no existen:

- `core/config`
- `core/discovery`
- `core/history`
- `core/notify`
- `core/pairing`
- `core/queue`
- `core/transfer`
- `core/webshare`

Este workflow compila solamente:

```text
core/server/main.go
```

con:

```powershell
go build -o TransferLAN+.exe main.go
```

Así se ignoran los archivos Go viejos que todavía no forman parte del servidor estable.
