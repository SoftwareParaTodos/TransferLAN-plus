# TransferLAN+ v1.0.6-beta Android Native Java

Versión Android simplificada para compilar APK en GitHub Actions sin Kotlin, AppCompat ni Material.

## Objetivo

Evitar errores de dependencias y generar un APK debug funcional.

## Funciona así

1. Ejecutás TransferLAN+ en Windows.
2. En Android ponés la URL de la PC:

```text
http://192.168.1.45:5050
```

3. Elegís archivo.
4. Enviás a `/transfer/upload`.

## Compilar en GitHub

Subí esta versión y entrá a:

```text
Actions → Build Android APK → Run workflow
```

Al terminar descargás el artifact:

```text
TransferLANPlus-v1.0.6-beta-debug-apk
```
