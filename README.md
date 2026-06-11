# TransferLAN+ v1.5.4-beta — Transfer State Recovery

**Sin cuentas. Sin nube. Sin cables.**

Esta versión mejora la recuperación visual de transferencias cuando Android vuelve a abrir la app.

## Cambios principales

- `TransferService` guarda el último estado de transferencia.
- `MainActivity` lee el último estado al abrir.
- Si hay una transferencia en curso, la pantalla lo muestra.
- Si la última transferencia terminó, se muestra como completada.
- Si falló o fue cancelada, se muestra con opción de reintentar.
- Se guarda:
  - estado;
  - progreso;
  - mensaje;
  - archivo;
  - destino;
  - enviados;
  - total;
  - hora.
- Mantiene progreso por notificación.
- Mantiene botón cancelar.
- Mantiene envío desde servicio.

## Objetivo

Que el usuario pueda volver a abrir TransferLAN+ y entender qué pasó con su transferencia.
