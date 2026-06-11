# Patch v1.5.9-beta — Smart Auto Reconnect

Aplicar sobre `android_sender/app/src/main/java/com/transferlan/plus/MainActivity.java`.

Esta versión depende de que v1.5.8 ya haya agregado:

```java
static final String KEY_LAST_DEVICE_ID = "last_device_id";
```

y que Windows devuelva `device_id` en `/device/info`.

---

## 1. Modificar `autoConnectLastDevice()`

Reemplazar el método por:

```java
void autoConnectLastDevice() {
    String base = prefs.getString(KEY_LAST_BASE, "");
    String name = prefs.getString(KEY_LAST_NAME, "PC conocida");
    String deviceId = prefs.getString(KEY_LAST_DEVICE_ID, "");

    if (base.length() == 0) {
        status.setText("No hay PC conocida. Podés agregar una por IP o pegar código.");
        refreshDeviceCard();
        return;
    }

    status.setText("Probando PC conocida...");
    transferDiagnosticText.setText("Buscando " + name + "...");

    testKnownDeviceOrRecover(base, name, deviceId);
}
```

---

## 2. Agregar método `testKnownDeviceOrRecover`

Agregar antes de `showPasteCodeDialog()` o cerca de `testAndSelectDevice()`:

```java
void testKnownDeviceOrRecover(String base, String fallbackName, String expectedDeviceId) {
    final String finalBase = trimSlash(base);
    final String finalFallbackName = fallbackName;
    final String finalExpectedDeviceId = expectedDeviceId;

    new Thread(() -> {
        try {
            HttpURLConnection c = (HttpURLConnection)new URL(finalBase + "/device/info").openConnection();
            c.setConnectTimeout(1200);
            c.setReadTimeout(1500);

            if (c.getResponseCode() == 200) {
                String body = readAll(c.getInputStream());
                String name = extract(body, "name");
                String deviceId = extract(body, "device_id");
                if (deviceId.length() == 0) deviceId = extract(body, "id");
                if (name.length() == 0) name = finalFallbackName;

                final String finalName = name;
                final String finalDeviceId = deviceId;

                runOnUiThread(() -> {
                    selectedBaseUrl = finalBase;
                    selectedDeviceName = finalName;
                    saveKnownDeviceWithId(finalName, finalBase, finalDeviceId);
                    refreshDeviceCard();
                    status.setText("PC conectada correctamente.");
                    transferDiagnosticText.setText("✓ " + finalName + " disponible");
                });
                c.disconnect();
                return;
            }
            c.disconnect();
        } catch(Exception ignored) {}

        if (finalExpectedDeviceId.length() > 0) {
            runOnUiThread(() -> {
                status.setText("La IP anterior no responde. Buscando la misma PC...");
                transferDiagnosticText.setText("Buscando " + finalFallbackName + " por identidad...");
            });
            scanNetworkForKnownDevice(finalExpectedDeviceId, finalFallbackName);
        } else {
            runOnUiThread(() -> {
                status.setText("No se pudo conectar con la PC conocida.");
                transferDiagnosticText.setText("La IP anterior no responde. Usá Buscar nuevamente, Pegar código de PC o Agregar por IP.");
                refreshDeviceCard();
            });
        }
    }).start();
}
```

---

## 3. Agregar escaneo por `device_id`

```java
void scanNetworkForKnownDevice(String expectedDeviceId, String fallbackName) {
    new Thread(() -> {
        String ip = localIp();

        if (ip.length() == 0 || !ip.contains(".")) {
            runOnUiThread(() -> {
                status.setText("No se pudo detectar la red.");
                transferDiagnosticText.setText("Revisá que el Wi‑Fi esté activo.");
            });
            return;
        }

        String prefix = ip.substring(0, ip.lastIndexOf(".") + 1);
        final boolean[] found = {false};

        for (int i = 1; i <= 254 && !found[0]; i++) {
            String base = "http://" + prefix + i + ":5050";

            try {
                HttpURLConnection c = (HttpURLConnection)new URL(base + "/device/info").openConnection();
                c.setConnectTimeout(180);
                c.setReadTimeout(250);

                if (c.getResponseCode() == 200) {
                    String body = readAll(c.getInputStream());
                    String deviceId = extract(body, "device_id");
                    if (deviceId.length() == 0) deviceId = extract(body, "id");

                    if (expectedDeviceId.equals(deviceId)) {
                        String name = extract(body, "name");
                        if (name.length() == 0) name = fallbackName;

                        final String finalBase = base;
                        final String finalName = name;
                        final String finalDeviceId = deviceId;
                        found[0] = true;

                        runOnUiThread(() -> {
                            selectedBaseUrl = trimSlash(finalBase);
                            selectedDeviceName = finalName;
                            saveKnownDeviceWithId(finalName, selectedBaseUrl, finalDeviceId);
                            refreshDeviceCard();
                            status.setText("PC encontrada en nueva dirección.");
                            transferDiagnosticText.setText("✓ " + finalName + " encontrada\\nDirección actualizada automáticamente.");
                        });
                    }
                }

                c.disconnect();
            } catch(Exception ignored) {}
        }

        if (!found[0]) {
            runOnUiThread(() -> {
                status.setText("No se encontró la PC conocida.");
                transferDiagnosticText.setText("No se encontró " + fallbackName + ".\\n\\nProbá abrir TransferLAN+ en Windows o usar Pegar código de PC.");
                refreshDeviceCard();
            });
        }
    }).start();
}
```

---

## 4. Mejorar `refreshDeviceCard()`

Donde dice “Probando conexión...”, se puede cambiar por:

```java
deviceCard.setText("⭐ " + name + "\\nBuscando disponibilidad...\\nGuardada anteriormente" + deviceIdSuffix());
```

Y cuando conecte:

```java
deviceCard.setText("🟢 " + selectedDeviceName + "\\nDisponible\\nLista para recibir archivos" + deviceIdSuffix());
```

---

## Resultado

Con esto, si la PC cambia de IP pero conserva `device_id`, Android la encuentra y actualiza la dirección guardada automáticamente.
