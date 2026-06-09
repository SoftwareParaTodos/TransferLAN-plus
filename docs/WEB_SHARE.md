# TransferLAN+ v0.12.0 - Modo Invitado Web

Esta versión agrega un modo invitado pensado para casos rápidos: un dispositivo abre TransferLAN+ como receptor, genera un enlace local temporal y otro equipo de la misma red puede subir un archivo desde el navegador, sin instalar la app.

## Flujo

1. En la PC receptora ejecutar el servidor:

```bash
go run ./cmd/transferlan-discovery --mode server --port 47231
```

2. Crear un enlace invitado:

```bash
go run ./cmd/transferlan-discovery --mode guest-link --target http://localhost:47231
```

3. Abrir la URL devuelta desde otro dispositivo de la misma red local.

4. Seleccionar archivo y enviar.

Los archivos recibidos quedan en:

```text
downloads/guest_uploads/
```

## Endpoints nuevos

- `POST /guest/share/start`
- `GET /guest/{token}`
- `POST /guest/{token}/upload`

## Seguridad

- El enlace es temporal.
- El token se genera aleatoriamente.
- No requiere cuenta ni nube.
- Solo funciona si el receptor es accesible dentro de la red local.

## Próximo paso

La UI Flutter debe generar el enlace, mostrarlo como texto y luego convertirlo en QR para escanear desde Android.
