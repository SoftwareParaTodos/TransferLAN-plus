# TransferLAN+ v0.15.0 - Notificaciones y bandeja

Esta versión agrega la base de notificaciones internas para que la interfaz Flutter pueda mostrar avisos nativos en Android, Windows y Linux.

## Objetivo

Que TransferLAN+ pueda quedar abierto en segundo plano y avisar cuando:

- llega un archivo;
- termina una transferencia por bloques;
- se recibe un archivo desde Modo Invitado Web;
- se complete un emparejamiento;
- ocurra un error de transferencia.

## Endpoints nuevos

```http
GET /notifications?limit=50
POST /notifications/read-all
POST /notifications/clear
```

Los eventos se guardan localmente en:

```text
data/notifications.json
```

## Diseño previsto para Flutter

### Android

Usar `flutter_local_notifications` para mostrar avisos del sistema.

Ejemplo:

```text
TransferLAN+
Archivo recibido: video.mp4
```

### Windows / Linux

La UI debe leer `/notifications` cada pocos segundos o mediante polling liviano.
Cuando aparezca un evento nuevo, mostrar notificación nativa.

## Bandeja del sistema

Esta versión deja preparada la arquitectura para agregar bandeja del sistema en la próxima iteración visual:

- minimizar a bandeja;
- mantener receptor activo;
- abrir carpeta de descargas;
- pausar recepción;
- cerrar TransferLAN+.

## Filosofía

Las notificaciones no deben molestar. Solo deben aparecer en eventos importantes.
