# TransferLAN+ v1.2.2-beta — QR Safe Fix

**Sin cuentas. Sin nube. Sin cables.**

Esta versión corrige los problemas detectados:

- El QR/pareo de la PC no se mostraba correctamente.
- El botón QR de Android podía colgar la app y cerrarla.

## Solución segura

Para no romper la transferencia que ya funciona, esta versión usa un sistema de emparejamiento seguro por **código/link local**:

1. La PC muestra un link `transferlan://connect?...`.
2. Android permite pegar ese link desde la opción `Pegar código de PC`.
3. Android guarda la PC automáticamente.
4. Después se envía normalmente.

## Importante

Se desactiva temporalmente el scanner QR nativo para evitar cierres de la app.
El QR real con cámara vuelve en una versión posterior, con permisos y librería probados.
