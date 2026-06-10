# TransferLAN+ v1.1.0-beta — Experiencia AirDrop

**Sin cuentas. Sin nube. Sin cables.**

Esta versión mejora la experiencia de usuario para que TransferLAN+ deje de sentirse técnico y empiece a comportarse como una app simple de transferencia local.

## Novedades principales

- Android con interfaz tipo tarjetas.
- Logo oficial integrado.
- IP manual ocultable como modo avanzado.
- Dispositivos detectados con nombre, sistema, versión e IP.
- Endpoint `/assets/logo.png` en Windows.
- Página local Windows más visual.
- Diagnóstico básico de red y firewall.
- Descubrimiento LAN por UDP broadcast conservado.
- Envío Android → Windows conservado.

## Puertos

- TCP 5050: recepción de archivos y web local.
- UDP 5050: descubrimiento automático.

## Uso

1. Ejecutar TransferLAN+ en Windows.
2. Si Android no detecta la PC, ejecutar como administrador:

```text
Windows/PERMITIR_FIREWALL_ADMIN.bat
```

3. Instalar el APK generado por GitHub Actions.
4. Abrir Android y tocar `Buscar dispositivos`.
5. Seleccionar PC, elegir archivo y enviar.
