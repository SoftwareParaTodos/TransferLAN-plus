# Cómo compilar el APK Android

1. Instalá Android Studio.
2. Abrí la carpeta:

```text
android_sender
```

3. Esperá que Gradle sincronice.
4. Conectá el celular con depuración USB o usá emulador.
5. Tocá Run.

## Generar APK

En Android Studio:

```text
Build > Build Bundle(s) / APK(s) > Build APK(s)
```

El APK queda normalmente en:

```text
android_sender/app/build/outputs/apk/debug/app-debug.apk
```

## Uso

1. Ejecutá TransferLAN+ en Windows.
2. Anotá la IP de la PC.
3. En Android cargá:

```text
http://IP-DE-LA-PC:5050
```

4. Elegí archivo y enviar.
