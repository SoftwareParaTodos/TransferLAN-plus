# TransferLAN+ v1.5.9-beta — Smart Auto Reconnect

**Sin cuentas. Sin nube. Sin cables.**

Esta versión prepara la reconexión automática cuando cambia la IP de la PC.

## Objetivo

Que Android deje de depender de una IP fija y busque la misma PC por `device_id`.

## Cambios

- Android intenta conectar primero a la última IP conocida.
- Si falla, busca la misma PC por `device_id`.
- Si encuentra la misma PC con otra IP, actualiza la dirección guardada.
- Evita duplicados por identidad.
- Mantiene IP manual como respaldo.
- Mantiene código de PC como respaldo.
- Mantiene TransferService y verificación por SHA-256.

## Flujo

```text
Abrir TransferLAN+
↓
Probar PC conocida
↓
Si falla, buscar por device_id
↓
Actualizar IP automáticamente
↓
Enviar archivo
```
