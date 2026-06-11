package com.transferlan.plus;

import android.app.Activity;
import android.app.AlertDialog;
import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.os.Bundle;
import android.os.Build;
import android.content.Intent;
import android.content.BroadcastReceiver;
import android.content.IntentFilter;
import android.content.SharedPreferences;
import android.net.Uri;
import android.net.wifi.WifiManager;
import android.provider.OpenableColumns;
import android.database.Cursor;
import android.view.View;
import android.view.Gravity;
import android.view.WindowManager;
import android.widget.*;
import android.graphics.Color;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.content.Context;
import android.Manifest;
import android.content.pm.PackageManager;

import java.io.*;
import java.net.*;
import java.util.*;

public class MainActivity extends Activity {
    static final int PICK = 1201;
    static final int DISCOVERY_PORT = 5050;
    static final String DISCOVERY_MESSAGE = "TRANSFERLAN_DISCOVER";
    static final String PREFS = "transferlan_prefs";
    static final String KEY_LAST_BASE = "last_base_url";
    static final String KEY_LAST_NAME = "last_device_name";
    static final String KEY_LAST_HISTORY = "last_history";
    static final String TRANSFER_STATE_PREFS = "transferlan_transfer_state";
    static final String TRANSFER_CHANNEL_ID = "transferlan_transfer";
    static final int NOTIFICATION_ID_TRANSFER = 5050;
    static final int REQ_NOTIFICATIONS = 5051;

    TextView status;
    TextView deviceCard;
    TextView selectedFileText;
    TextView progressText;
    TextView transferDiagnosticText;
    ProgressBar progressBar;
    Button retryButton;
    Button cancelTransferButton;
    LinearLayout foundDevicesBox;
    LinearLayout secondaryActionsBox;

    Uri selected;
    long selectedSize = 0;
    String selectedBaseUrl = "";
    String selectedDeviceName = "";

    SharedPreferences prefs;
    NotificationManager notificationManager;
    WifiManager.MulticastLock multicastLock;
    Set<String> seenDeviceCards = new HashSet<>();

    BroadcastReceiver transferStatusReceiver = new BroadcastReceiver() {
        @Override
        public void onReceive(Context context, Intent intent) {
            if (intent == null) return;
            if (!TransferService.ACTION_STATUS.equals(intent.getAction())) return;
            handleTransferStatus(intent);
        }
    };
    boolean isSending = false;
    long lastUiProgressUpdate = 0;

    @Override
    public void onCreate(Bundle b) {
        super.onCreate(b);
        prefs = getSharedPreferences(PREFS, MODE_PRIVATE);
        setupNotifications();
        requestNotificationPermissionIfNeeded();
        buildUi();
        handleIntent(getIntent());
        autoConnectLastDevice();
    }


    @Override
    protected void onResume() {
        super.onResume();
        try {
            IntentFilter f = new IntentFilter(TransferService.ACTION_STATUS);
            if (android.os.Build.VERSION.SDK_INT >= 33) {
                registerReceiver(transferStatusReceiver, f, Context.RECEIVER_NOT_EXPORTED);
            } else {
                registerReceiver(transferStatusReceiver, f);
            }
        } catch(Exception ignored) {}
    }

    @Override
    protected void onPause() {
        try { unregisterReceiver(transferStatusReceiver); } catch(Exception ignored) {}
        super.onPause();
    }

    @Override
    public void onNewIntent(Intent i) {
        super.onNewIntent(i);
        handleIntent(i);
    }

    void handleIntent(Intent i) {
        if (i == null) return;

        if (Intent.ACTION_VIEW.equals(i.getAction()) && i.getData() != null) {
            handlePairingUri(i.getData());
            return;
        }

        if (Intent.ACTION_SEND.equals(i.getAction())) {
            Uri u = i.getParcelableExtra(Intent.EXTRA_STREAM);
            if (u != null) setSelectedFile(u);
        }
    }

    void buildUi() {
        ScrollView scroll = new ScrollView(this);
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setPadding(28, 36, 28, 28);
        root.setBackgroundColor(Color.rgb(15,23,42));

        ImageView logo = new ImageView(this);
        logo.setImageResource(getResources().getIdentifier("transferlan_logo", "drawable", getPackageName()));
        LinearLayout.LayoutParams logoParams = new LinearLayout.LayoutParams(150, 150);
        logoParams.gravity = Gravity.CENTER_HORIZONTAL;
        logo.setLayoutParams(logoParams);
        root.addView(logo);

        TextView title = text("TransferLAN+", 31, 229,231,235, true);
        title.setGravity(Gravity.CENTER_HORIZONTAL);
        root.addView(title);

        TextView sub = text("Sin cuentas. Sin nube. Sin cables.", 16, 56,189,248, true);
        sub.setGravity(Gravity.CENTER_HORIZONTAL);
        root.addView(sub);

        TextView section = text("Mis dispositivos", 19, 229,231,235, true);
        section.setPadding(0, 26, 0, 8);
        root.addView(section);

        deviceCard = cardText("Buscando PC conocida...");
        root.addView(deviceCard);

        Button sendBig = primaryButton("Enviar archivo");
        sendBig.setTextSize(18);
        sendBig.setOnClickListener(v -> pickAndSendFlow());
        root.addView(sendBig);

        selectedFileText = cardText("Archivo: ninguno seleccionado");
        root.addView(selectedFileText);

        progressBar = new ProgressBar(this, null, android.R.attr.progressBarStyleHorizontal);
        progressBar.setMax(100);
        progressBar.setProgress(0);
        root.addView(progressBar);

        progressText = text("Listo para enviar.", 14, 148,163,184, false);
        root.addView(progressText);

        transferDiagnosticText = cardText("Transferencia: sin archivo en curso");
        root.addView(transferDiagnosticText);

        retryButton = secondaryButton("Reintentar envío");
        retryButton.setVisibility(View.GONE);
        retryButton.setOnClickListener(v -> sendFile());
        root.addView(retryButton);

        cancelTransferButton = secondaryButton("Cancelar transferencia");
        cancelTransferButton.setVisibility(View.GONE);
        cancelTransferButton.setOnClickListener(v -> cancelActiveTransfer());
        root.addView(cancelTransferButton);

        status = text("Iniciando...", 15, 229,231,235, false);
        status.setPadding(0, 14, 0, 16);
        root.addView(status);

        TextView secondaryTitle = text("¿No aparece tu dispositivo?", 17, 229,231,235, true);
        secondaryTitle.setPadding(0, 20, 0, 6);
        root.addView(secondaryTitle);

        secondaryActionsBox = new LinearLayout(this);
        secondaryActionsBox.setOrientation(LinearLayout.VERTICAL);

        Button searchAgain = secondaryButton("Buscar nuevamente");
        searchAgain.setOnClickListener(v -> discoverDevices());
        secondaryActionsBox.addView(searchAgain);

        Button pasteCode = secondaryButton("Pegar código de PC");
        pasteCode.setOnClickListener(v -> showPasteCodeDialog());
        secondaryActionsBox.addView(pasteCode);

        Button addIp = secondaryButton("Agregar por IP");
        addIp.setOnClickListener(v -> showManualIpDialog());
        secondaryActionsBox.addView(addIp);

        Button history = secondaryButton("Ver historial local");
        history.setOnClickListener(v -> showLocalHistory());
        secondaryActionsBox.addView(history);

        root.addView(secondaryActionsBox);

        foundDevicesBox = new LinearLayout(this);
        foundDevicesBox.setOrientation(LinearLayout.VERTICAL);
        foundDevicesBox.setPadding(0, 16, 0, 0);
        root.addView(foundDevicesBox);

        scroll.addView(root);
        setContentView(scroll);
        refreshDeviceCard();
        restoreLastTransferState();
    }

    void pickAndSendFlow() {
        if (selectedBaseUrl.length() == 0) {
            String base = prefs.getString(KEY_LAST_BASE, "");
            if (base.length() > 0) {
                selectedBaseUrl = base;
                selectedDeviceName = prefs.getString(KEY_LAST_NAME, "PC conocida");
            } else {
                toast("Primero conectá una PC.");
                showManualIpDialog();
                return;
            }
        }

        if (selected == null) {
            pickFile();
        } else {
            sendFile();
        }
    }

    TextView text(String s, int size, int r, int g, int b, boolean bold) {
        TextView t = new TextView(this);
        t.setText(s);
        t.setTextSize(size);
        t.setTextColor(Color.rgb(r,g,b));
        t.setPadding(0, 8, 0, 8);
        if (bold) t.setTypeface(Typeface.DEFAULT, Typeface.BOLD);
        return t;
    }

    TextView cardText(String s) {
        TextView t = text(s, 15, 229,231,235, false);
        GradientDrawable bg = new GradientDrawable();
        bg.setColor(Color.rgb(17,24,39));
        bg.setCornerRadius(24);
        bg.setStroke(2, Color.rgb(51,65,85));
        t.setBackground(bg);
        t.setPadding(22, 20, 22, 20);
        LinearLayout.LayoutParams lp = new LinearLayout.LayoutParams(-1, -2);
        lp.setMargins(0, 8, 0, 12);
        t.setLayoutParams(lp);
        return t;
    }

    Button primaryButton(String label) {
        Button b = new Button(this);
        b.setText(label);
        b.setTextColor(Color.rgb(0,17,31));
        b.setTypeface(Typeface.DEFAULT, Typeface.BOLD);
        b.setAllCaps(false);
        GradientDrawable bg = new GradientDrawable(GradientDrawable.Orientation.LEFT_RIGHT, new int[]{Color.rgb(56,189,248), Color.rgb(34,197,94)});
        bg.setCornerRadius(20);
        b.setBackground(bg);
        LinearLayout.LayoutParams lp = new LinearLayout.LayoutParams(-1, -2);
        lp.setMargins(0, 12, 0, 8);
        b.setLayoutParams(lp);
        return b;
    }

    Button secondaryButton(String label) {
        Button b = new Button(this);
        b.setText(label);
        b.setTextColor(Color.rgb(229,231,235));
        b.setAllCaps(false);
        GradientDrawable bg = new GradientDrawable();
        bg.setColor(Color.rgb(31,41,55));
        bg.setCornerRadius(18);
        bg.setStroke(2, Color.rgb(51,65,85));
        b.setBackground(bg);
        LinearLayout.LayoutParams lp = new LinearLayout.LayoutParams(-1, -2);
        lp.setMargins(0, 8, 0, 6);
        b.setLayoutParams(lp);
        return b;
    }

    void refreshDeviceCard() {
        String base = prefs.getString(KEY_LAST_BASE, "");
        String name = prefs.getString(KEY_LAST_NAME, "");
        if (selectedBaseUrl.length() > 0) {
            if (selectedDeviceName.length() == 0) selectedDeviceName = "PC conectada";
            deviceCard.setText("✓ " + selectedDeviceName + "\nDisponible\nLista para recibir archivos");
        } else if (base.length() > 0) {
            if (name.length() == 0) name = "PC conocida";
            deviceCard.setText("⭐ " + name + "\nProbando conexión...\nGuardada anteriormente");
        } else {
            deviceCard.setText("No hay PC conectada todavía.\nUsá Buscar nuevamente, Pegar código de PC o Agregar por IP.");
        }
    }

    void autoConnectLastDevice() {
        String base = prefs.getString(KEY_LAST_BASE, "");
        if (base.length() == 0) {
            status.setText("No hay PC conocida. Podés agregar una por IP o pegar código.");
            refreshDeviceCard();
            return;
        }
        status.setText("Probando PC conocida...");
        testAndSelectDevice(base, prefs.getString(KEY_LAST_NAME, "PC conocida"), false);
    }



    void restoreLastTransferState() {
        try {
            SharedPreferences state = getSharedPreferences(TRANSFER_STATE_PREFS, MODE_PRIVATE);
            String st = state.getString("status", "");
            if (st.length() == 0) return;

            int progress = state.getInt("progress", 0);
            String msg = state.getString("message", "");
            String filename = state.getString("filename", "");
            String target = state.getString("target", "");
            long sent = state.getLong("sent", 0);
            long total = state.getLong("total", 0);

            if (TransferService.STATUS_SENDING.equals(st) || TransferService.STATUS_PREPARING.equals(st) || TransferService.STATUS_FINALIZING.equals(st)) {
                progressBar.setProgress(progress);
                progressText.setText(msg.length() > 0 ? msg : "Transferencia en curso");
                transferDiagnosticText.setText("Transferencia en curso\nArchivo: " + filename + "\nDestino: " + target + "\n" + formatBytes(sent) + " / " + formatBytes(total));
                cancelTransferButton.setVisibility(View.VISIBLE);
                retryButton.setVisibility(View.GONE);
                status.setText("Transferencia en segundo plano detectada.");
            } else if (TransferService.STATUS_COMPLETED.equals(st)) {
                progressBar.setProgress(100);
                progressText.setText("Última transferencia completada.");
                transferDiagnosticText.setText("✓ Última transferencia completada\nArchivo: " + filename + "\nDestino: " + target);
                cancelTransferButton.setVisibility(View.GONE);
                retryButton.setVisibility(View.GONE);
                status.setText("Última transferencia completada.");
            } else if (TransferService.STATUS_ERROR.equals(st) || TransferService.STATUS_CANCELLED.equals(st)) {
                progressBar.setProgress(progress);
                progressText.setText("Última transferencia no completada.");
                transferDiagnosticText.setText("⚠ Última transferencia no completada\nArchivo: " + filename + "\nDestino: " + target + "\nEstado: " + st);
                cancelTransferButton.setVisibility(View.GONE);
                retryButton.setVisibility(View.VISIBLE);
                status.setText("Podés reintentar el envío.");
            }
        } catch(Exception ignored) {}
    }

    void handleTransferStatus(Intent intent) {
        String st = intent.getStringExtra(TransferService.EXTRA_STATUS);
        String msg = intent.getStringExtra(TransferService.EXTRA_MESSAGE);
        int progress = intent.getIntExtra(TransferService.EXTRA_PROGRESS, 0);
        long sent = intent.getLongExtra(TransferService.EXTRA_SENT, 0);
        long total = intent.getLongExtra(TransferService.EXTRA_TOTAL, 0);

        if (msg == null) msg = "";

        if (TransferService.STATUS_PREPARING.equals(st)) {
            progressBar.setProgress(0);
            progressText.setText("Preparando transferencia...");
            transferDiagnosticText.setText(msg);
            cancelTransferButton.setVisibility(View.VISIBLE);
            status.setText("Transferencia en segundo plano iniciada.");
        } else if (TransferService.STATUS_SENDING.equals(st)) {
            progressBar.setProgress(progress);
            progressText.setText(msg);
            transferDiagnosticText.setText("Enviando en segundo plano\n" + formatBytes(sent) + " / " + formatBytes(total));
            cancelTransferButton.setVisibility(View.VISIBLE);
            status.setText("Enviando archivo...");
        } else if (TransferService.STATUS_FINALIZING.equals(st)) {
            progressBar.setProgress(99);
            progressText.setText("Finalizando...");
            transferDiagnosticText.setText(msg);
            cancelTransferButton.setVisibility(View.VISIBLE);
            status.setText("Esperando confirmación de la PC.");
        } else if (TransferService.STATUS_COMPLETED.equals(st)) {
            progressBar.setProgress(100);
            progressText.setText("Transferencia completada.");
            transferDiagnosticText.setText("✓ Archivo recibido por la PC");
            cancelTransferButton.setVisibility(View.GONE);
            retryButton.setVisibility(View.GONE);
            status.setText("Archivo enviado correctamente.");
            addLocalHistory("✓ Transferencia completada en segundo plano");
        } else if (TransferService.STATUS_CANCELLED.equals(st)) {
            progressText.setText("Transferencia cancelada.");
            transferDiagnosticText.setText("La transferencia fue cancelada.");
            cancelTransferButton.setVisibility(View.GONE);
            retryButton.setVisibility(View.VISIBLE);
            status.setText("Transferencia cancelada.");
        } else if (TransferService.STATUS_ERROR.equals(st)) {
            progressText.setText("Transferencia interrumpida.");
            transferDiagnosticText.setText("No se pudo completar. Podés reintentar.");
            cancelTransferButton.setVisibility(View.GONE);
            retryButton.setVisibility(View.VISIBLE);
            status.setText("Transferencia interrumpida.");
        }
    }

    void cancelActiveTransfer() {
        try {
            Intent i = new Intent(this, TransferService.class);
            i.setAction(TransferService.ACTION_CANCEL);
            startService(i);
            status.setText("Cancelando transferencia...");
        } catch(Exception e) {
            status.setText("No se pudo cancelar: " + e.getMessage());
        }
    }

    void showPasteCodeDialog() {
        final EditText input = new EditText(this);
        input.setHint("transferlan://connect?...");
        input.setMinLines(3);
        input.setTextColor(Color.BLACK);

        new AlertDialog.Builder(this)
            .setTitle("Pegar código de PC")
            .setMessage("Copiá el código desde la PC y pegalo acá.")
            .setView(input)
            .setPositiveButton("Guardar", (dialog, which) -> handlePairingText(input.getText().toString().trim()))
            .setNegativeButton("Cancelar", null)
            .show();
    }

    void handlePairingText(String text) {
        if (text.length() == 0) {
            status.setText("Código vacío.");
            return;
        }
        try {
            Uri uri = Uri.parse(text);
            handlePairingUri(uri);
        } catch(Exception e) {
            status.setText("Código inválido.");
        }
    }

    void handlePairingUri(Uri uri) {
        if (uri == null) return;
        if (!"transferlan".equals(uri.getScheme()) || !"connect".equals(uri.getHost())) {
            status.setText("Código no corresponde a TransferLAN+.");
            return;
        }

        String name = uri.getQueryParameter("name");
        String baseUrl = uri.getQueryParameter("base_url");
        String ip = uri.getQueryParameter("ip");
        String port = uri.getQueryParameter("port");

        if (baseUrl == null || baseUrl.length() == 0) {
            if (ip == null || ip.length() == 0) {
                status.setText("Código sin IP/base_url.");
                return;
            }
            if (port == null || port.length() == 0) port = "5050";
            baseUrl = "http://" + ip + ":" + port;
        }

        if (name == null || name.length() == 0) name = "PC por código";
        status.setText("PC guardada. Probando conexión...");
        testAndSelectDevice(baseUrl, name, true);
    }

    void showManualIpDialog() {
        LinearLayout box = new LinearLayout(this);
        box.setOrientation(LinearLayout.VERTICAL);

        final EditText ipInput = new EditText(this);
        ipInput.setHint("IP de la PC, ej: 192.168.1.107");
        ipInput.setSingleLine(true);
        ipInput.setTextColor(Color.BLACK);
        box.addView(ipInput);

        final EditText portInput = new EditText(this);
        portInput.setHint("Puerto");
        portInput.setText("5050");
        portInput.setSingleLine(true);
        portInput.setTextColor(Color.BLACK);
        box.addView(portInput);

        new AlertDialog.Builder(this)
            .setTitle("Agregar PC por IP")
            .setMessage("Usalo solo si no aparece automáticamente.")
            .setView(box)
            .setPositiveButton("Conectar", (dialog, which) -> {
                String ip = ipInput.getText().toString().trim();
                String port = portInput.getText().toString().trim();
                if (ip.length() == 0) {
                    toast("Poné la IP de la PC");
                    return;
                }
                if (port.length() == 0) port = "5050";
                String base = (ip.startsWith("http://") || ip.startsWith("https://")) ? ip : "http://" + ip + ":" + port;
                testAndSelectDevice(base, "PC manual", true);
            })
            .setNegativeButton("Cancelar", null)
            .show();
    }

    void testAndSelectDevice(String base, String fallbackName, boolean showResult) {
        final String finalBase = trimSlash(base);
        final String finalFallbackName = fallbackName;
        new Thread(() -> {
            try {
                HttpURLConnection c = (HttpURLConnection)new URL(finalBase + "/device/info").openConnection();
                c.setConnectTimeout(1200);
                c.setReadTimeout(1500);
                if (c.getResponseCode() == 200) {
                    String body = readAll(c.getInputStream());
                    String name = extract(body, "name");
                    if (name.length() == 0) name = finalFallbackName;
                    final String finalName = name;
                    runOnUiThread(() -> {
                        selectedBaseUrl = finalBase;
                        selectedDeviceName = finalName;
                        saveKnownDevice(finalName, finalBase);
                        refreshDeviceCard();
                        status.setText("PC conectada correctamente.");
                        if (showResult) toast("PC conectada: " + finalName);
                    });
                } else {
                    runOnUiThread(() -> status.setText("La PC respondió con error."));
                }
                c.disconnect();
            } catch(Exception e) {
                runOnUiThread(() -> {
                    status.setText("No se pudo conectar. Revisá Wi-Fi, Firewall o IP.");
                    refreshDeviceCard();
                });
            }
        }).start();
    }

    void saveKnownDevice(String name, String baseUrl) {
        prefs.edit().putString(KEY_LAST_NAME, name).putString(KEY_LAST_BASE, baseUrl).apply();
    }

    String trimSlash(String s) {
        while (s.endsWith("/")) s = s.substring(0, s.length()-1);
        return s;
    }

    void setSelectedFile(Uri u) {
        selected = u;
        selectedSize = fileSize(u);
        selectedFileText.setText("Archivo: " + fileName(u) + " (" + formatBytes(selectedSize) + ")");
        progressBar.setProgress(0);
        progressText.setText("Archivo listo. Tocá Enviar archivo para comenzar.");
        retryButton.setVisibility(View.GONE);
        transferDiagnosticText.setText("Archivo listo:\n" + fileName(u) + "\nTamaño: " + formatBytes(selectedSize));
        if (selectedBaseUrl.length() > 0 || prefs.getString(KEY_LAST_BASE, "").length() > 0) {
            sendFile();
        }
    }

    void discoverDevices() {
        foundDevicesBox.removeAllViews();
        seenDeviceCards.clear();
        status.setText("Buscando dispositivos...");
        new Thread(() -> {
            final Set<String> found = new HashSet<>();
            try {
                acquireMulticastLock();
                DatagramSocket socket = new DatagramSocket();
                socket.setBroadcast(true);
                socket.setSoTimeout(1200);
                byte[] data = DISCOVERY_MESSAGE.getBytes("UTF-8");

                for (InetAddress target : broadcastTargets()) {
                    socket.send(new DatagramPacket(data, data.length, target, DISCOVERY_PORT));
                }

                long end = System.currentTimeMillis() + 2500;
                byte[] buffer = new byte[4096];

                while (System.currentTimeMillis() < end) {
                    try {
                        DatagramPacket response = new DatagramPacket(buffer, buffer.length);
                        socket.receive(response);
                        String body = new String(response.getData(), 0, response.getLength(), "UTF-8");
                        String base = extract(body, "base_url");
                        if (base.length() == 0) base = "http://" + response.getAddress().getHostAddress() + ":5050";
                        final String finalBase = trimSlash(base);
                        final String finalBody = body;
                        if (!found.contains(finalBase)) {
                            found.add(finalBase);
                            runOnUiThread(() -> addDeviceCard(finalBase, finalBody));
                        }
                    } catch (SocketTimeoutException ignored) {}
                }
                socket.close();
            } catch(Exception e) {
                runOnUiThread(() -> status.setText("Broadcast no disponible. Probando respaldo..."));
            } finally {
                releaseMulticastLock();
            }

            if (found.size() == 0) {
                runOnUiThread(() -> {
                    status.setText("No apareció por broadcast. Probando PC conocida y respaldo...");
                    autoConnectLastDevice();
                    fallbackScan();
                });
            } else {
                runOnUiThread(() -> status.setText("Dispositivos encontrados: " + found.size()));
            }
        }).start();
    }

    void fallbackScan() {
        new Thread(() -> {
            String ip = localIp();
            if (ip.length() == 0 || !ip.contains(".")) {
                runOnUiThread(() -> status.setText("No se pudo detectar red."));
                return;
            }
            String prefix = ip.substring(0, ip.lastIndexOf(".") + 1);
            final int[] count = {0};

            for (int i=1; i<=254; i++) {
                String base = "http://" + prefix + i + ":5050";
                try {
                    HttpURLConnection c = (HttpURLConnection)new URL(base + "/device/info").openConnection();
                    c.setConnectTimeout(160);
                    c.setReadTimeout(220);
                    if (c.getResponseCode() == 200) {
                        String body = readAll(c.getInputStream());
                        count[0]++;
                        runOnUiThread(() -> addDeviceCard(base, body));
                    }
                    c.disconnect();
                } catch(Exception ignored) {}
            }

            runOnUiThread(() -> {
                if (count[0] == 0) status.setText("No se encontraron dispositivos. Usá Pegar código de PC o Agregar por IP.");
                else status.setText("Dispositivos encontrados: " + count[0]);
            });
        }).start();
    }

    void addDeviceCard(String base, String body) {
        final String deviceBase = trimSlash(base);
        if (seenDeviceCards.contains(deviceBase)) return;
        seenDeviceCards.add(deviceBase);

        String parsedName = extract(body, "name");
        String parsedOs = extract(body, "os");
        String parsedVersion = extract(body, "version");

        if (parsedName.length() == 0) parsedName = "Computadora encontrada";
        if (parsedOs.length() == 0) parsedOs = "desktop";

        final String deviceName = parsedName;
        final String deviceOs = parsedOs;
        final String deviceVersion = parsedVersion;

        Button b = secondaryButton("🖥  " + deviceName + "\n" + deviceOs + " · " + deviceVersion + "\nUsar este dispositivo");
        b.setOnClickListener(v -> {
            selectedBaseUrl = deviceBase;
            selectedDeviceName = deviceName;
            saveKnownDevice(deviceName, deviceBase);
            refreshDeviceCard();
            status.setText("Dispositivo seleccionado y guardado.");
        });
        foundDevicesBox.addView(b);
    }

    List<InetAddress> broadcastTargets() {
        List<InetAddress> list = new ArrayList<>();
        try {
            list.add(InetAddress.getByName("255.255.255.255"));
            String ip = localIp();
            if (ip.contains(".")) list.add(InetAddress.getByName(ip.substring(0, ip.lastIndexOf(".") + 1) + "255"));
        } catch(Exception ignored) {}
        return list;
    }

    void acquireMulticastLock() {
        try {
            WifiManager wifi = (WifiManager)getApplicationContext().getSystemService(Context.WIFI_SERVICE);
            multicastLock = wifi.createMulticastLock("transferlan_discovery");
            multicastLock.setReferenceCounted(true);
            multicastLock.acquire();
        } catch(Exception ignored) {}
    }

    void releaseMulticastLock() {
        try {
            if (multicastLock != null && multicastLock.isHeld()) multicastLock.release();
        } catch(Exception ignored) {}
    }

    String extract(String json, String key) {
        try {
            String p = "\"" + key + "\":";
            int i = json.indexOf(p);
            if (i < 0) return "";
            i += p.length();
            while (i < json.length() && (json.charAt(i) == ' ' || json.charAt(i) == '"')) i++;
            int j = i;
            while (j < json.length() && json.charAt(j) != '"' && json.charAt(j) != ',' && json.charAt(j) != '}') j++;
            return json.substring(i,j).replace("\\/","/").trim();
        } catch(Exception e) { return ""; }
    }

    String localIp() {
        try {
            Enumeration<NetworkInterface> es = NetworkInterface.getNetworkInterfaces();
            while (es.hasMoreElements()) {
                NetworkInterface n = es.nextElement();
                Enumeration<InetAddress> as = n.getInetAddresses();
                while (as.hasMoreElements()) {
                    InetAddress a = as.nextElement();
                    if (!a.isLoopbackAddress() && a instanceof Inet4Address) return a.getHostAddress();
                }
            }
        } catch(Exception ignored) {}
        return "";
    }

    void pickFile() {
        Intent i = new Intent(Intent.ACTION_OPEN_DOCUMENT);
        i.setType("*/*");
        i.addCategory(Intent.CATEGORY_OPENABLE);
        startActivityForResult(i, PICK);
    }

    @Override
    protected void onActivityResult(int req, int res, Intent data) {
        super.onActivityResult(req, res, data);
        if (req == PICK && res == RESULT_OK && data != null) setSelectedFile(data.getData());
    }



    void sendFile() {
        if (isSending) {
            toast("Ya hay una transferencia en curso.");
            return;
        }

        if (selected == null) {
            pickFile();
            return;
        }

        String base = selectedBaseUrl;
        if (base.length() == 0) base = prefs.getString(KEY_LAST_BASE, "");
        if (base.length() == 0) {
            toast("Primero conectá una PC.");
            showManualIpDialog();
            return;
        }

        String targetName = selectedDeviceName.length() > 0 ? selectedDeviceName : prefs.getString(KEY_LAST_NAME, "PC");
        String filename = fileName(selected);
        String finalBase = trimSlash(base);

        progressBar.setProgress(1);
        retryButton.setVisibility(View.GONE);
        progressText.setText("Transferencia iniciada en segundo plano.");
        cancelTransferButton.setVisibility(View.VISIBLE);
        transferDiagnosticText.setText("Enviando en segundo plano:\nDestino: " + targetName + "\nArchivo: " + filename + "\nTamaño: " + formatBytes(selectedSize));
        status.setText("TransferService está enviando el archivo.");

        try {
            Intent i = new Intent(this, TransferService.class);
            i.setAction(TransferService.ACTION_START);
            i.putExtra(TransferService.EXTRA_FILE_URI, selected.toString());
            i.putExtra(TransferService.EXTRA_BASE_URL, finalBase);
            i.putExtra(TransferService.EXTRA_FILENAME, filename);
            i.putExtra(TransferService.EXTRA_TARGET, targetName);
            i.putExtra(TransferService.EXTRA_SIZE, selectedSize);

            if (Build.VERSION.SDK_INT >= 26) startForegroundService(i);
            else startService(i);

            addLocalHistory("↗ Envío iniciado en segundo plano\n" + filename + "\n" + formatBytes(selectedSize));
            showMessage("Envío iniciado", "La transferencia sigue desde la notificación de TransferLAN+.");
            selected = null;
            selectedSize = 0;
            selectedFileText.setText("Archivo: ninguno seleccionado");
            progressBar.setProgress(0);
            progressText.setText("El envío continúa en segundo plano.");
        } catch(Exception e) {
            retryButton.setVisibility(View.VISIBLE);
            status.setText("No se pudo iniciar servicio: " + e.getMessage());
            progressText.setText("No se pudo iniciar transferencia. Podés reintentar.");
        }
    }

    void upload(String base, Uri uri, long startTime) throws Exception {
        String boundary = "TransferLANBoundary" + System.currentTimeMillis();
        HttpURLConnection c = (HttpURLConnection)new URL(base + "/transfer/upload").openConnection();
        c.setRequestMethod("POST");
        c.setDoOutput(true);
        c.setConnectTimeout(15000);
        c.setReadTimeout(0);
        c.setChunkedStreamingMode(1024 * 128);
        c.setRequestProperty("Connection", "close");
        c.setRequestProperty("Content-Type", "multipart/form-data; boundary=" + boundary);

        InputStream in = getContentResolver().openInputStream(uri);
        if (in == null) throw new Exception("No se pudo abrir archivo");

        ProgressOutputStream out = null;
        try {
            out = new ProgressOutputStream(new BufferedOutputStream(c.getOutputStream(), 1024 * 64));
            String fn = fileName(uri);

            out.writeRaw("--" + boundary + "\r\n");
            out.writeRaw("Content-Disposition: form-data; name=\"file\"; filename=\"" + fn + "\"\r\n");
            out.writeRaw("Content-Type: application/octet-stream\r\n\r\n");

            byte[] buf = new byte[1024 * 64];
            int n;

            while ((n = in.read(buf)) > 0) {
                out.writeFile(buf, 0, n);

                long now = System.currentTimeMillis();
                if (now - lastUiProgressUpdate > 350) {
                    lastUiProgressUpdate = now;
                    updateProgress(out.fileBytesWritten, selectedSize, Math.max(1, now - startTime));
                }
            }

            out.writeRaw("\r\n--" + boundary + "--\r\n");
            out.flush();
        } finally {
            try { in.close(); } catch(Exception ignored) {}
            try { if (out != null) out.close(); } catch(Exception ignored) {}
        }

        runOnUiThread(() -> {
            progressBar.setProgress(99);
            progressText.setText("Finalizando... esperando confirmación de la PC");
            transferDiagnosticText.setText("Finalizando transferencia...\nEsperando confirmación de Windows.");
            showTransferNotification("Finalizando", "Esperando confirmación de la PC", 99, true);
        });

        int code = c.getResponseCode();
        if (code < 200 || code > 299) throw new Exception("HTTP " + code);

        try {
            InputStream response = c.getInputStream();
            while (response.read(new byte[1024]) != -1) {}
            response.close();
        } catch(Exception ignored) {}

        c.disconnect();
    }

    void updateProgress(long sent, long total, long elapsedMs) {
        int percent = total > 0 ? (int)Math.min(99, (sent * 100) / total) : 0;
        double mbps = sent / 1024.0 / 1024.0 / (elapsedMs / 1000.0);
        runOnUiThread(() -> {
            progressBar.setProgress(percent);
            progressText.setText(percent + "% · " + formatBytes(sent) + " / " + formatBytes(total) + " · " + String.format(Locale.US, "%.1f MB/s", mbps));
            showTransferNotification("Enviando archivo", percent + "% · " + formatBytes(sent) + " / " + formatBytes(total), percent, true);
        });
    }

    class ProgressOutputStream extends FilterOutputStream {
        long fileBytesWritten = 0;
        ProgressOutputStream(OutputStream out) { super(out); }
        void writeRaw(String s) throws IOException { out.write(s.getBytes("UTF-8")); }
        void writeFile(byte[] b, int off, int len) throws IOException {
            out.write(b, off, len);
            fileBytesWritten += len;
        }
    }

    String readAll(InputStream in) throws Exception {
        BufferedReader br = new BufferedReader(new InputStreamReader(in));
        StringBuilder sb = new StringBuilder();
        String line;
        while ((line = br.readLine()) != null) sb.append(line).append("\n");
        return sb.toString();
    }

    String fileName(Uri u) {
        String n = "archivo_transferlan";
        Cursor c = getContentResolver().query(u, null, null, null, null);
        if (c != null) {
            try {
                int idx = c.getColumnIndex(OpenableColumns.DISPLAY_NAME);
                if (idx >= 0 && c.moveToFirst()) n = c.getString(idx);
            } finally { c.close(); }
        }
        return n;
    }

    long fileSize(Uri u) {
        long size = 0;
        Cursor c = getContentResolver().query(u, null, null, null, null);
        if (c != null) {
            try {
                int idx = c.getColumnIndex(OpenableColumns.SIZE);
                if (idx >= 0 && c.moveToFirst()) size = c.getLong(idx);
            } finally { c.close(); }
        }
        return size;
    }

    String formatBytes(long b) {
        if (b <= 0) return "0 B";
        double v = b;
        String[] units = {"B", "KB", "MB", "GB", "TB"};
        int i = 0;
        while (v >= 1024 && i < units.length - 1) { v /= 1024; i++; }
        return String.format(Locale.US, "%.1f %s", v, units[i]);
    }

    void addLocalHistory(String line) {
        String old = prefs.getString(KEY_LAST_HISTORY, "");
        String entry = new java.text.SimpleDateFormat("yyyy-MM-dd HH:mm:ss", java.util.Locale.US).format(new java.util.Date()) + "\n" + line;
        String updated = entry + "\n\n" + old;
        if (updated.length() > 4000) updated = updated.substring(0, 4000);
        prefs.edit().putString(KEY_LAST_HISTORY, updated).apply();
    }

    void showLocalHistory() {
        String hist = prefs.getString(KEY_LAST_HISTORY, "");
        if (hist.length() == 0) hist = "Todavía no hay envíos registrados.";
        new AlertDialog.Builder(this)
            .setTitle("Historial local")
            .setMessage(hist)
            .setPositiveButton("OK", null)
            .show();
    }

    void showMessage(String title, String msg) {
        new AlertDialog.Builder(this)
            .setTitle(title)
            .setMessage(msg)
            .setPositiveButton("OK", null)
            .show();
    }



    void startTransferServiceFoundation(String filename, String target) {
        try {
            Intent i = new Intent(this, TransferService.class);
            i.setAction(TransferService.ACTION_START);
            i.putExtra(TransferService.EXTRA_FILENAME, filename);
            i.putExtra(TransferService.EXTRA_TARGET, target);
            if (Build.VERSION.SDK_INT >= 26) {
                startForegroundService(i);
            } else {
                startService(i);
            }
        } catch(Exception ignored) {}
    }

    void stopTransferServiceFoundation() {
        try {
            Intent i = new Intent(this, TransferService.class);
            i.setAction(TransferService.ACTION_CANCEL);
            startService(i);
        } catch(Exception ignored) {}
    }

    void setupNotifications() {
        notificationManager = (NotificationManager) getSystemService(Context.NOTIFICATION_SERVICE);
        if (Build.VERSION.SDK_INT >= 26 && notificationManager != null) {
            NotificationChannel channel = new NotificationChannel(
                TRANSFER_CHANNEL_ID,
                "Transferencias",
                NotificationManager.IMPORTANCE_LOW
            );
            channel.setDescription("Progreso de transferencias TransferLAN+");
            notificationManager.createNotificationChannel(channel);
        }
    }

    void requestNotificationPermissionIfNeeded() {
        if (Build.VERSION.SDK_INT >= 33) {
            try {
                if (checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
                    requestPermissions(new String[]{Manifest.permission.POST_NOTIFICATIONS}, REQ_NOTIFICATIONS);
                }
            } catch(Exception ignored) {}
        }
    }

    void showTransferNotification(String title, String message, int progress, boolean ongoing) {
        try {
            if (notificationManager == null) return;
            if (Build.VERSION.SDK_INT >= 33 && checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
                return;
            }

            Notification.Builder builder = Build.VERSION.SDK_INT >= 26
                ? new Notification.Builder(this, TRANSFER_CHANNEL_ID)
                : new Notification.Builder(this);

            builder.setSmallIcon(android.R.drawable.stat_sys_upload)
                .setContentTitle("TransferLAN+ · " + title)
                .setContentText(message)
                .setOngoing(ongoing)
                .setOnlyAlertOnce(true);

            if (ongoing) {
                builder.setProgress(100, Math.max(0, Math.min(100, progress)), false);
            } else {
                builder.setProgress(0, 0, false);
            }

            notificationManager.notify(NOTIFICATION_ID_TRANSFER, builder.build());
        } catch(Exception ignored) {}
    }

    void toast(String m) {
        Toast.makeText(this, m, Toast.LENGTH_SHORT).show();
    }
}
