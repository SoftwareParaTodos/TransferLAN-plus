# TransferLAN+ v1.0.3-beta Android Sender

Esta versión agrega una base Android para enviar archivos desde el celular a la PC por LAN.

## Qué trae

- Proyecto Android nativo Kotlin.
- Pantalla simple para cargar la IP de la PC.
- Selector de archivos de Android.
- Envío multipart a `/transfer/upload`.
- Permiso de internet/red.
- Guía para compilar APK desde Android Studio.

## Requisitos para compilar

- Android Studio
- JDK incluido con Android Studio
- SDK Android instalado

## Cómo usar

1. Ejecutar TransferLAN+ en Windows.
2. Ver la IP de la PC, por ejemplo:

```text
192.168.1.45
```

3. Abrir la app Android.
4. Poner:

```text
http://192.168.1.45:5050
```

5. Elegir archivo.
6. Enviar.

## Estado

Beta inicial. Todavía no incluye detección automática mDNS desde Android ni QR.
