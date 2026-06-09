# Si falla GitHub Actions

Abrí el run fallido y entrá a:

```text
Build TransferLAN+ Android APK
↓
Build debug APK
```

Copiá el error que aparece debajo de:

```text
FAILURE: Build failed with an exception.
```

Con eso se corrige exacto.

Esta versión ya instala Android SDK y usa `--stacktrace`, así que el próximo error debería mostrar el motivo real.
