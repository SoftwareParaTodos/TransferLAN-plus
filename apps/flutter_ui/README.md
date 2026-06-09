# TransferLAN+ Flutter UI v0.13.0

Primera interfaz gráfica para TransferLAN+.

## Incluye

- Pantalla principal oscura.
- Botones Enviar / Recibir.
- Lista de dispositivos detectados.
- Selector de archivo.
- Pantalla de progreso visual.
- Cliente HTTP preparado para hablar con el backend Go local.
- Estado del receptor usando `/health`.

## Estado

Esta versión es una UI base. Todavía no reemplaza al backend Go: lo consume por HTTP.

## Ejecutar

```bash
cd apps/flutter_ui
flutter pub get
flutter run
```

Para probar contra el backend:

```bash
cd ../../
go run ./cmd/transferlan-discovery -mode receiver
```

Luego en la UI usar host:

```text
http://127.0.0.1:8787
```

En Android real, reemplazar `127.0.0.1` por la IP de la PC receptora.


## v0.13.0

- Zona Drag & Drop en escritorio.
- Pantalla Historial.
- Cliente API para `/history` y `/history/clear`.
