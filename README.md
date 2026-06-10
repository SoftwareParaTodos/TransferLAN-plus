# TransferLAN+ v1.1.2-beta — Android Compile Fix

**Sin cuentas. Sin nube. Sin cables.**

Esta versión corrige el build Android de v1.1.1-beta.

## Corrección principal

- Se corrigieron variables usadas dentro de lambdas Java que no estaban como efectivamente finales.
- Se mantiene el progreso real de transferencia.
- Se mantiene diseño AirDrop.
- Se mantiene discovery LAN.
- Se mantiene logo oficial.

## Subir y compilar

```bash
git add .
git commit -m "Fix Android compile error in progress version"
git push origin main
```

Después ejecutar:

```text
Actions → Build Android APK → Run workflow
```
