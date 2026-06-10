# TransferLAN+ v1.2.4-beta — QR Real Fix

**Sin cuentas. Sin nube. Sin cables.**

Esta versión arregla el emparejamiento QR.

## Correcciones

- Windows genera un QR real offline en `/pairing/qr.png`.
- La página local muestra el QR real.
- Android vuelve a tener botón `Escanear QR`.
- Android pide permiso de cámara antes de abrir el scanner.
- Si el scanner falla, sigue disponible `Pegar código de PC`.
- Se mantiene historial, progreso, IP manual y dispositivos conocidos.

## Nota técnica

Windows usa la librería Go:

```text
github.com/skip2/go-qrcode
```

La primera compilación puede descargar esa dependencia.
