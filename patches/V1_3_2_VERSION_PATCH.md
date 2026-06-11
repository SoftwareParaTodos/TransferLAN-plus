# Patch v1.3.2-beta

Aplicar este paquete sobre el repo actual.

Además revisar manualmente:

## core/server/main.go

Debe tener:

```go
const version = "v1.3.2-beta"
```

## android_sender/app/build.gradle

Debe tener:

```gradle
versionCode 22
versionName '1.3.2-beta'
```

## .github/workflows/build-android-apk.yml

Debe tener:

```yaml
name: TransferLANPlus-v1.3.2-beta-debug-apk
```
