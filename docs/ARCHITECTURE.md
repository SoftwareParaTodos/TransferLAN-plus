# Arquitectura TransferLAN+ v0.11.0

## Objetivo

TransferLAN+ permite enviar archivos y carpetas grandes entre dispositivos de la misma red local sin nube, sin login y sin suscripción.

## Componentes

- `core/discovery`: detección por mDNS/Zeroconf.
- `core/pairing`: PIN temporal y tokens de dispositivos confiables.
- `core/transfer`: envío básico, envío por bloques, reanudación y envío de carpetas.
- `core/server`: endpoints HTTP locales.
- `apps/flutter_ui`: interfaz inicial Flutter.

## Envío de carpetas

En v0.11.0 se agrega `SendFolderChunked`.

Flujo:

1. Se valida que la ruta sea una carpeta.
2. Se crea un ZIP temporal en la carpeta temporal del sistema.
3. El ZIP conserva estructura interna de archivos y subcarpetas.
4. El ZIP se envía usando `SendFileChunked`.
5. El receptor lo guarda como archivo `.zip`.
6. Se valida SHA256 al finalizar.

## Decisión de seguridad

La extracción automática todavía no se realiza. Esto evita que una transferencia sobrescriba archivos existentes o genere contenido inesperado en el equipo receptor.
