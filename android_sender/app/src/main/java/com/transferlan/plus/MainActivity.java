package com.transferlan.plus;

import android.app.Activity;
import android.os.Bundle;
import android.content.Intent;
import android.net.Uri;
import android.net.wifi.WifiManager;
import android.provider.OpenableColumns;
import android.database.Cursor;
import android.view.Gravity;
import android.view.View;
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

    EditText urlInput;
    TextView status;
    TextView selectedDeviceText;
    TextView selectedFileText;
    LinearLayout devices;
    LinearLayout advancedBox;
    Uri selected;
    String selectedBaseUrl = "";
    WifiManager.MulticastLock multicastLock;

    @Override
    public void onCreate(Bundle b) {
        super.onCreate(b);
        buildUi();
        handleShared(getIntent());
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

        Button scan = primaryButton("BUSCAR DISPOSITIVOS");
        scan.setOnClickListener(v -> discoverDevices());
        root.addView(scan);

        devices = new LinearLayout(this);
        devices.setOrientation(LinearLayout.VERTICAL);
        devices.setPadding(0, 14, 0, 14);
        root.addView(devices);

        selectedDeviceText = cardText("Destino: ningún dispositivo seleccionado");
        root.addView(selectedDeviceText);

        Button pick = primaryButton("ELEGIR ARCHIVO");
        pick.setOnClickListener(v -> pickFile());
        root.addView(pick);

        selectedFileText = cardText("Archivo: ninguno seleccionado");
        root.addView(selectedFileText);

        Button send = primaryButton("ENVIAR");
        send.setOnClickListener(v -> sendFile());
        root.addView(send);

        Button advanced = secondaryButton("CONFIGURACIÓN MANUAL");
        advanced.setOnClickListener(v -> toggleAdvanced());
        root.addView(advanced);

        advancedBox = new LinearLayout(this);
        advancedBox.setOrientation(LinearLayout.VERTICAL);
        advancedBox.setVisibility(View.GONE);

        urlInput = new EditText(this);
        urlInput.setHint("http://IP-DE-TU-PC:5050");
        urlInput.setText("http://");
        urlInput.setSingleLine(true);
        urlInput.setTextColor(Color.WHITE);
        urlInput.setHintTextColor(Color.GRAY);
        advancedBox.addView(urlInput);
        root.addView(advancedBox);

        status = text("Listo para buscar dispositivos.", 15, 229,231,235, false);
        status.setPadding(0, 22, 0, 0);
        root.addView(status);

        scroll.addView(root);
        setContentView(scroll);
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
        b.setTextSize(16);
        GradientDrawable bg = new GradientDrawable(GradientDrawable.Orientation.LEFT_RIGHT, new int[]{Color.rgb(56,189,248), Color.rgb(34,197,94)});
        bg.setCornerRadius(18);
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
        return b;
    }

    void toggleAdvanced() {
        advancedBox.setVisibility(advancedBox.getVisibility() == View.VISIBLE ? View.GONE : View.VISIBLE);
    }

    void handleShared(Intent i) {
        if (i != null && Intent.ACTION_SEND.equals(i.getAction())) {
            Uri u = i.getParcelableExtra(Intent.EXTRA_STREAM);
            if (u != null) {
                selected = u;
                if (selectedFileText != null) selectedFileText.setText("Archivo: " + fileName(u));
            }
        }
    }

    void discoverDevices() {
        devices.removeAllViews();
        status.setText("Buscando dispositivos en la red...");

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
            } catch (Exception e) {
                runOnUiThread(() -> status.setText("Error buscando: " + e.getMessage()));
            } finally {
                releaseMulticastLock();
            }

            if (found.size() == 0) {
                runOnUiThread(() -> {
                    status.setText("No apareció por broadcast. Probando respaldo...");
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
                runOnUiThread(() -> status.setText("No se pudo detectar la red."));
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
            runOnUiThread(() -> status.setText(count[0] == 0 ? "No se encontraron dispositivos." : "Dispositivos encontrados: " + count[0]));
        }).start();
    }

    void addDeviceCard(String base, String body) {
        String name = extract(body, "name");
        String os = extract(body, "os");
        String version = extract(body, "version");
        if (name.length() == 0) name = "Computadora encontrada";
        if (os.length() == 0) os = "desktop";

        Button b = secondaryButton("🖥  " + name + "\n" + os + " · " + version + "\nSeleccionar");
        b.setOnClickListener(v -> {
            selectedBaseUrl = base;
            urlInput.setText(base);
            selectedDeviceText.setText("Destino: " + name + " (" + base + ")");
            status.setText("Dispositivo seleccionado.");
        });
        devices.addView(b);
    }

    List<InetAddress> broadcastTargets() {
        List<InetAddress> list = new ArrayList<>();
        try {
            list.add(InetAddress.getByName("255.255.255.255"));
            String ip = localIp();
            if (ip.contains(".")) {
                String prefix = ip.substring(0, ip.lastIndexOf(".") + 1);
                list.add(InetAddress.getByName(prefix + "255"));
            }
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
        try { if (multicastLock != null && multicastLock.isHeld()) multicastLock.release(); } catch(Exception ignored) {}
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
        if (req == PICK && res == RESULT_OK && data != null) {
            selected = data.getData();
            selectedFileText.setText("Archivo: " + fileName(selected));
        }
    }

    void sendFile() {
        if (selected == null) { toast("Primero elegí un archivo"); return; }
        String base = selectedBaseUrl.length() > 0 ? selectedBaseUrl : urlInput.getText().toString().trim();
        while (base.endsWith("/")) base = base.substring(0, base.length()-1);
        if (!base.startsWith("http://") && !base.startsWith("https://")) { toast("Seleccioná un dispositivo o cargá una URL válida"); return; }

        final String finalBase = base;
        status.setText("Enviando...");
        new Thread(() -> {
            try {
                upload(finalBase, selected);
                runOnUiThread(() -> status.setText("Archivo enviado correctamente."));
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
        DataOutputStream out = new DataOutputStream(c.getOutputStream());
        String fn = fileName(uri);
        out.writeBytes("--" + boundary + "\r\n");
        out.writeBytes("Content-Disposition: form-data; name=\"file\"; filename=\"" + fn + "\"\r\n");
        out.writeBytes("Content-Type: application/octet-stream\r\n\r\n");
        byte[] buf = new byte[1024 * 256];
        int n;
        while ((n = in.read(buf)) > 0) out.write(buf, 0, n);
        in.close();
        out.writeBytes("\r\n--" + boundary + "--\r\n");
        out.close();
        int code = c.getResponseCode();
        if (code < 200 || code > 299) throw new Exception("HTTP " + code);
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

    void toast(String m) { Toast.makeText(this, m, Toast.LENGTH_SHORT).show(); }
}
