# TransferLAN+ v1.5.1-beta — Foreground Service Foundation

**Sin cuentas. Sin nube. Sin cables.**

Esta versión prepara la arquitectura real para transferencias en segundo plano.

## Cambios principales

- Agrega `TransferService.java`.
- Declara Foreground Service en AndroidManifest.
- Agrega permiso `FOREGROUND_SERVICE`.
- Agrega permiso `WAKE_LOCK`.
- Agrega notificación persistente de transferencia.
- Agrega acción de cancelación preparada.
- Mantiene el flujo actual de envío funcionando.
- Mantiene notificación de progreso desde Activity.
- Deja el proyecto listo para migrar el upload completo al servicio en la próxima fase.

## Importante

Esta es una versión de transición segura:
- no rompe el envío actual;
- agrega la base del servicio;
- evita hacer un cambio gigante de golpe.

La próxima versión moverá definitivamente el upload pesado desde `MainActivity` hacia `TransferService`.
