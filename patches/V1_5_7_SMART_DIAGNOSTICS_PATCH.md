# Patch v1.5.7-beta — Smart Connection Diagnostics

Aplicar sobre `android_sender/app/src/main/java/com/transferlan/plus/MainActivity.java`.

## 1. Agregar botón en opciones secundarias

Buscar:

```java
Button history = secondaryButton("Ver historial local");
history.setOnClickListener(v -> showLocalHistory());
secondaryActionsBox.addView(history);
```

Debajo agregar:

```java
Button diag = secondaryButton("Diagnosticar conexión");
diag.setOnClickListener(v -> runConnectionDiagnostic());
secondaryActionsBox.addView(diag);
```

## 2. Agregar métodos antes de `showPasteCodeDialog()`

```java
String explainConnectionError(String raw) {
    if (raw == null) raw = "";
    String lower = raw.toLowerCase(Locale.US);

    StringBuilder sb = new StringBuilder();
    sb.append("No se pudo conectar con la PC.\n\n");

    if (lower.contains("connect") || lower.contains("timed out") || lower.contains("timeout")) {
        sb.append("Posibles causas:\n");
        sb.append("• TransferLAN+ no está abierto en Windows.\n");
        sb.append("• La IP de la PC cambió.\n");
        sb.append("• El Firewall de Windows bloquea el puerto 5050.\n");
        sb.append("• El celular y la PC no están en la misma Wi‑Fi/hotspot.\n\n");
    } else if (lower.contains("http 404")) {
        sb.append("La PC respondió, pero parece una versión vieja de TransferLAN+.\n\n");
    } else if (lower.contains("http 500")) {
        sb.append("La PC recibió la conexión, pero tuvo un error interno.\n\n");
    } else if (lower.contains("cancel")) {
        sb.append("La transferencia fue cancelada.\n\n");
    } else {
        sb.append("Puede ser un corte de red o una versión antigua abierta en Windows.\n\n");
    }

    sb.append("Qué probar:\n");
    sb.append("1. Abrí TransferLAN+ en la PC.\n");
    sb.append("2. Desde Chrome del celular probá: http://IP-DE-LA-PC:5050/device/info\n");
    sb.append("3. Si no abre, ejecutá PERMITIR_FIREWALL_ADMIN.bat como administrador.\n");
    sb.append("4. Si cambió la IP, usá Agregar por IP o pegá el código de PC.\n");

    if (raw.length() > 0) sb.append("\nDetalle técnico:\n").append(raw);
    return sb.toString();
}

void runConnectionDiagnostic() {
    String base = selectedBaseUrl;
    if (base.length() == 0) base = prefs.getString(KEY_LAST_BASE, "");

    if (base.length() == 0) {
        showMessage("Diagnóstico", "No hay una PC guardada.\n\nUsá Pegar código de PC o Agregar por IP.");
        return;
    }

    final String finalBase = trimSlash(base);
    status.setText("Diagnosticando conexión...");
    transferDiagnosticText.setText("Probando conexión con la PC...\n" + finalBase);

    new Thread(() -> {
        String result;
        try {
            HttpURLConnection c = (HttpURLConnection)new URL(finalBase + "/device/info").openConnection();
            c.setConnectTimeout(2500);
            c.setReadTimeout(3000);
            int code = c.getResponseCode();
            if (code == 200) {
                String body = readAll(c.getInputStream());
                String name = extract(body, "name");
                if (name.length() == 0) name = "PC encontrada";
                result = "✓ Conexión correcta\n\nPC: " + name + "\nDirección: " + finalBase + "\n\nYa podés enviar archivos.";
            } else {
                result = "La PC respondió con HTTP " + code + ".\n\nPuede ser una versión vieja o un endpoint no disponible.";
            }
            c.disconnect();
        } catch(Exception e) {
            result = explainConnectionError(e.getMessage());
        }

        final String finalResult = result;
        runOnUiThread(() -> {
            status.setText("Diagnóstico terminado.");
            transferDiagnosticText.setText(finalResult);
            showMessage("Diagnóstico de conexión", finalResult);
        });
    }).start();
}
```

## 3. Usar diagnóstico en errores

Donde haya errores tipo:

```java
status.setText("No se pudo conectar...");
```

reemplazar por:

```java
transferDiagnosticText.setText(explainConnectionError(e.getMessage()));
status.setText("No se pudo conectar.");
```
