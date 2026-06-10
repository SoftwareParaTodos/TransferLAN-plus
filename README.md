# TransferLAN+ v1.2.5-beta — QR Windows + Android seguro

**Sin cuentas. Sin nube. Sin cables.**

Esta versión corrige el bloqueo de compilación Android de v1.2.4.

## Qué cambia

- Windows genera un QR real en `/pairing/qr.png`.
- La página local muestra el QR real.
- Android mantiene `Pegar código de PC`.
- Se elimina temporalmente el scanner interno para evitar fallos de compilación/cierres.
- Podés escanear el QR con Cámara/Google Lens, copiar el link y pegarlo en Android.
- Mantiene historial, progreso, agregar por IP y PC conocida.

## Motivo

La integración del scanner interno estaba rompiendo el build de GitHub Actions. Esta versión deja el QR real en Windows y mantiene Android estable.
