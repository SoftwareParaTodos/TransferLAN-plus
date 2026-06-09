# TransferLAN+ v0.15.0

Transferencia local rápida entre Android, Windows y Linux.

Sin cuentas. Sin nube. Sin cables. Sin vueltas.

## Novedades v0.15.0

- Cola local de transferencias.
- Base para múltiples envíos simultáneos.
- Estados: pendiente, en progreso, completado, fallido y cancelado.
- Endpoints `/queue`.
- Cancelación de ítems.
- Limpieza de transferencias finalizadas.
- Documentación `docs/TRANSFER_QUEUE.md`.

## Estado del proyecto

Ya incluye base para:

- discovery LAN/mDNS;
- emparejamiento PIN/QR base;
- envío de archivos;
- envío por bloques;
- reanudar transferencias;
- envío de carpetas como ZIP temporal;
- modo invitado web;
- drag & drop base;
- historial;
- notificaciones internas;
- cola de transferencias.

## Próximo hito sugerido

v0.16.0: limitar concurrencia real y pantalla de cola en Flutter.
