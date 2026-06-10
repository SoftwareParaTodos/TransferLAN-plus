package com.transferlan.plus;

import android.app.Activity;
import android.os.Bundle;
import android.content.Intent;
import android.content.SharedPreferences;
import android.net.Uri;
import android.net.wifi.WifiManager;
import android.provider.OpenableColumns;
import android.database.Cursor;
import android.view.View;
import android.view.Gravity;
import android.widget.*;
import android.graphics.Color;
import android.graphics.Typeface;
import android.graphics.drawable.GradientDrawable;
import android.content.Context;

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

    EditText ipInput;
    EditText portInput;
    TextView status;
    TextView selectedDeviceText;
    TextView selectedFileText;
    TextView progressText;
    TextView knownDeviceText;
    ProgressBar progressBar;
    LinearLayout devices;
    LinearLayout manualBox;
    Uri selected;
    long selectedSize = 0;
    String selectedBaseUrl = "";
    SharedPreferences prefs;
    WifiManager.MulticastLock multicastLock;

    @Override
    public void onCreate(Bundle b) {
        super.onCreate(b);
        prefs = getSharedPreferences(PREFS, MODE_PRIVATE);
        buildUi();
        handleShared(getIntent());
        autoConnectLastDevice();
    }

    @Override
    public void onNewIntent(Intent i) {
        super.onNewIntent(i);
        handleShared(i);
    }

    void buildUi() {
        ScrollView scroll = new ScrollView(this);
        LinearLayout root = new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setPadding(28, 36, 28, 28);
        root.setBackgroundColor(Color.rgb(15,23,42));

        ImageView logo = new ImageView(this);
        logo.setImageResource(getResources().getIdentifier("transferlan_logo", "drawable", getPackageName()));
        LinearLayout.LayoutParams logoParams = new LinearLayout.LayoutParams(180, 180);
        logoParams.gravity = Gravity.CENTER_HORIZONTAL;
        logo.setLayoutParams(logoParams);
        root.addView(logo);

        TextView title = text("TransferLAN+", 30, 229,231,235, true);
        title.setGravity(Gravity.CENTER_HORIZONTAL);
        root.addView(title);

        TextView sub = text("Sin cuentas. Sin nube. Sin cables.", 16, 56,189,248, true);
        sub.setGravity(Gravity.CENTER_HORIZONTAL);
        root.addView(sub);

        knownDeviceText = cardText("Dispositivo conocido: ninguno");
        root.addView(knownDeviceText);

        Button reconnect = secondaryButton("Reconectar última PC");
        reconnect.setOnClickListener(v -> autoConnectLastDevice());
        root.addView(reconnect);

        Button scan = primaryButton("Buscar dispositivos");
        scan.setOnClickListener(v -> discoverDevices());
        root.addView(scan);

        devices = new LinearLayout(this);
        devices.setOrientation(LinearLayout.VERTICAL);
        devices.setPadding(0, 14, 0, 14);
        root.addView(devices);

        selectedDeviceText = cardText("Destino: ningún dispositivo seleccionado");
        root.addView(selectedDeviceText);

        Button pick = primaryButton("Elegir archivo");
        pick.setOnClickListener(v -> pickFile());
        root.addView(pick);

        selectedFileText = cardText("Archivo: ninguno seleccionado");
        root.addView(selectedFileText);

        progressBar = new ProgressBar(this, null, android.R.attr.progressBarStyleHorizontal);
        progressBar.setMax(100);
        progressBar.setProgress(0);
        root.addView(progressBar);

        progressText = text("Progreso: 0%", 14, 148,163,184, false);
        root.addView(progressText);

        Button send = primaryButton("Enviar");
        send.setOnClickListener(v -> sendFile());
        root.addView(send);

        Button manual = secondaryButton("Agregar PC por IP");
        manual.setOnClickListener(v -> toggleManual());
        root.addView(manual);

        manualBox = new LinearLayout(this);
        manualBox.setOrientation(LinearLayout.VERTICAL);
        manualBox.setVisibility(View.GONE);

        ipInput = new EditText(this);
        ipInput.setHint("IP de la PC, ej: 10.92.222.190");
        ipInput.setSingleLine(true);
        ipInput.setTextColor(Color.WHITE);
        ipInput.setHintTextColor(Color.GRAY);
        manualBox.addView(ipInput);

        portInput = new EditText(this);
        portInput.setHint("Puerto");
        portInput.setText("5050");
        portInput.setSingleLine(true);
        portInput.setTextColor(Color.WHITE);
        portInput.setHintTextColor(Color.GRAY);
        manualBox.addView(portInput);

        Button connect = primaryButton("Guardar y conectar");
        connect.setOnClickListener(v -> connectManualIp());
        manualBox.addView(connect);

        root.addView(manualBox);

        status = text("Listo.", 15, 229,231,235, false);
        status.setPadding(0, 22, 0, 0);
        root.addView(status);

        scroll.addView(root);
        setContentView(scroll);
        refreshKnownDeviceLabel();
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
        bg.setCornerRadius(22);
        bg.setStroke(2, Color.rgb(51,65,85));
        t.setBackground(bg);
        t.setPadding(22, 18, 22, 18);
        return t;
    }

    Button primaryButton(String label) {
        Button b = new Button(this);
        b.setText(label);
        b.setTextColor(Color.rgb(0,17,31));
        b.setTypeface(Typeface.DEFAULT, Typeface.BOLD);
        b.setAllCaps(false);
        GradientDrawable bg = new GradientDrawable(GradientDrawable.Orientation.LEFT_RIGHT, new int[]{Color.rgb(56,189,248), Color.rgb(34,197,94)});
        bg.setCornerRadius(18);
        b.setBackground(bg);
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
        return b;
    }

    void toggleManual() {
        manualBox.setVisibility(manualBox.getVisibility() == View.VISIBLE ? View.GONE : View.VISIBLE);
    }

    void refreshKnownDeviceLabel() {
        String base = prefs.getString(KEY_LAST_BASE, "");
        String name = prefs.getString(KEY_LAST_NAME, "");
        if (base.length() > 0) {
            if (name.length() == 0) name = "PC conocida";
            knownDeviceText.setText("Dispositivo conocido: " + name + " (" + base + ")");
        } else {
            knownDeviceText.setText("Dispositivo conocido: ninguno");
        }
    }

    void saveKnownDevice(String name, String baseUrl) {
        prefs.edit().putString(KEY_LAST_NAME, name).putString(KEY_LAST_BASE, baseUrl).apply();
        refreshKnownDeviceLabel();
    }

    void autoConnectLastDevice() {
        String base = prefs.getString(KEY_LAST_BASE, "");
        if (base.length() == 0) {
            status.setText("No hay PC conocida guardada.");
            return;
        }
        status.setText("Probando PC conocida...");
        testAndSelectDevice(base, prefs.getString(KEY_LAST_NAME, "PC conocida"));
    }

    void connectManualIp() {
        String ip = ipInput.getText().toString().trim();
        String port = portInput.getText().toString().trim();
        if (ip.length() == 0) {
            toast("Poné la IP de la PC");
            return;
        }
        if (port.length() == 0) port = "5050";
        String base = (ip.startsWith("http://") || ip.startsWith("https://")) ? ip : "http://" + ip + ":" + port;
        status.setText("Probando conexión...");
        testAndSelectDevice(base, "PC manual");
    }

    void testAndSelectDevice(String base, String fallbackName) {
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
                        selectedDeviceText.setText("Destino: " + finalName + " (" + finalBase + ")");
                        saveKnownDevice(finalName, finalBase);
                        status.setText("PC conectada correctamente.");
                    });
                } else {
                    runOnUiThread(() -> status.setText("La PC respondió con error."));
                }
                c.disconnect();
            } catch(Exception e) {
                runOnUiThread(() -> status.setText("No se pudo conectar: " + e.getMessage()));
            }
        }).start();
    }

    String trimSlash(String s) {
        while (s.endsWith("/")) s = s.substring(0, s.length()-1);
        return s;
    }

    void handleShared(Intent i) {
        if (i != null && Intent.ACTION_SEND.equals(i.getAction())) {
            Uri u = i.getParcelableExtra(Intent.EXTRA_STREAM);
            if (u != null) setSelectedFile(u);
        }
    }

    void setSelectedFile(Uri u) {
        selected = u;
        selectedSize = fileSize(u);
        selectedFileText.setText("Archivo: " + fileName(u) + " (" + formatBytes(selectedSize) + ")");
        progressBar.setProgress(0);
        progressText.setText("Progreso: 0%");
    }

    void discoverDevices() {
        devices.removeAllViews();
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
                        final String finalBase = base;
                        final String finalBody = body;
                        if (!found.contains(finalBase)) {
                            found.add(finalBase);
                            runOnUiThread(() -> addDeviceCard(finalBase, finalBody));
                        }
                    } catch (SocketTimeoutException ignored) {}
                }
                socket.close();
            } catch(Exception e) {
                runOnUiThread(() -> status.setText("Error buscando: " + e.getMessage()));
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
                    c.setConnectTimeout(180);
                    c.setReadTimeout(250);
                    if (c.getResponseCode() == 200) {
                        String body = readAll(c.getInputStream());
                        count[0]++;
                        runOnUiThread(() -> addDeviceCard(base, body));
                    }
                    c.disconnect();
                } catch(Exception ignored) {}
            }
            runOnUiThread(() -> {
                if (count[0] == 0) status.setText("No se encontraron dispositivos. Usá Agregar PC por IP.");
                else status.setText("Dispositivos encontrados: " + count[0]);
            });
        }).start();
    }

    void addDeviceCard(String base, String body) {
        String parsedName = extract(body, "name");
        String parsedOs = extract(body, "os");
        String parsedVersion = extract(body, "version");
        if (parsedName.length() == 0) parsedName = "Computadora encontrada";
        if (parsedOs.length() == 0) parsedOs = "desktop";
        final String deviceName = parsedName;
        final String deviceOs = parsedOs;
        final String deviceVersion = parsedVersion;
        final String deviceBase = trimSlash(base);

        Button b = secondaryButton("🖥  " + deviceName + "\n" + deviceOs + " · " + deviceVersion + "\nSeleccionar");
        b.setOnClickListener(v -> {
            selectedBaseUrl = deviceBase;
            selectedDeviceText.setText("Destino: " + deviceName + " (" + deviceBase + ")");
            saveKnownDevice(deviceName, deviceBase);
            status.setText("Dispositivo seleccionado y guardado.");
        });
        devices.addView(b);
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
        } catch(Exception e) {
            return "";
        }
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
        if (req == PICK && res == RESULT_OK && data != null) {
            setSelectedFile(data.getData());
        }
    }

    void sendFile() {
        if (selected == null) {
            toast("Primero elegí un archivo");
            return;
        }
        String base = selectedBaseUrl;
        if (base.length() == 0) base = prefs.getString(KEY_LAST_BASE, "");
        if (base.length() == 0) {
            toast("Seleccioná una PC o agregala por IP");
            return;
        }
        final String finalBase = trimSlash(base);
        progressBar.setProgress(0);
        progressText.setText("Progreso: 0%");
        status.setText("Enviando...");
        new Thread(() -> {
            try {
                upload(finalBase, selected);
                runOnUiThread(() -> {
                    progressBar.setProgress(100);
                    progressText.setText("Progreso: 100% · Completado");
                    status.setText("Archivo enviado correctamente.");
                });
            } catch(Exception e) {
                runOnUiThread(() -> status.setText("Error: " + e.getMessage()));
            }
        }).start();
    }

    void upload(String base, Uri uri) throws Exception {
        String boundary = "TransferLANBoundary" + System.currentTimeMillis();
        HttpURLConnection c = (HttpURLConnection)new URL(base + "/transfer/upload").openConnection();
        c.setRequestMethod("POST");
        c.setDoOutput(true);
        c.setRequestProperty("Content-Type", "multipart/form-data; boundary=" + boundary);
        InputStream in = getContentResolver().openInputStream(uri);
        if (in == null) throw new Exception("No se pudo abrir archivo");
        ProgressOutputStream out = new ProgressOutputStream(c.getOutputStream());
        String fn = fileName(uri);
        out.writeRaw("--" + boundary + "\r\n");
        out.writeRaw("Content-Disposition: form-data; name=\"file\"; filename=\"" + fn + "\"\r\n");
        out.writeRaw("Content-Type: application/octet-stream\r\n\r\n");
        byte[] buf = new byte[1024 * 256];
        int n;
        long start = System.currentTimeMillis();
        while ((n = in.read(buf)) > 0) {
            out.writeFile(buf, 0, n);
            updateProgress(out.fileBytesWritten, selectedSize, Math.max(1, System.currentTimeMillis() - start));
        }
        in.close();
        out.writeRaw("\r\n--" + boundary + "--\r\n");
        out.close();
        int code = c.getResponseCode();
        if (code < 200 || code > 299) throw new Exception("HTTP " + code);
    }

    void updateProgress(long sent, long total, long elapsedMs) {
        int percent = total > 0 ? (int)Math.min(100, (sent * 100) / total) : 0;
        double mbps = sent / 1024.0 / 1024.0 / (elapsedMs / 1000.0);
        runOnUiThread(() -> {
            progressBar.setProgress(percent);
            progressText.setText("Progreso: " + percent + "% · " + formatBytes(sent) + " / " + formatBytes(total) + " · " + String.format(Locale.US, "%.1f MB/s", mbps));
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
        while (v >= 1024 && i < units.length - 1) {
            v /= 1024;
            i++;
        }
        return String.format(Locale.US, "%.1f %s", v, units[i]);
    }

    void toast(String m) {
        Toast.makeText(this, m, Toast.LENGTH_SHORT).show();
    }
}
