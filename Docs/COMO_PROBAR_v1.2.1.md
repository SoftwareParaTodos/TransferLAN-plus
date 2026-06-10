# Cómo probar TransferLAN+ v1.2.1-beta

## Windows

Debe estar disponible un link de emparejamiento tipo:

```text
transferlan://connect?name=PC-MAX&ip=10.92.222.190&port=5050&base_url=http://10.92.222.190:5050
```

## Android

1. Instalar APK generado por GitHub Actions.
2. Abrir TransferLAN+.
3. Tocar `Escanear QR`.
4. Escanear un QR que contenga el link `transferlan://connect?...`.
5. La PC debe quedar guardada automáticamente.

## Nota

Si todavía no está generado el QR real en Windows, podés probar con cualquier generador de QR usando el link anterior.
