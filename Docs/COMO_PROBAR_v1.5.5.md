# Cómo probar TransferLAN+ v1.5.5-beta

## Prueba principal

1. Enviar un archivo desde Android.
2. Al finalizar, Android debe mostrar:
   - Windows confirmó recepción;
   - nombre del archivo;
   - tamaño recibido;
   - SHA-256 si Windows lo devuelve.

## Verificación

En Windows el endpoint `/transfer/upload` debe responder JSON con:

```json
{
  "ok": true,
  "filename": "...",
  "size": 123,
  "sha256": "...",
  "path": "downloads/..."
}
```

## Resultado esperado

Android no marca completado solo porque terminó de subir: marca completado cuando Windows responde OK.
