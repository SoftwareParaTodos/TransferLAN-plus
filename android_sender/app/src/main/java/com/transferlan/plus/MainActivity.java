package com.transferlan.plus;

import android.app.Activity;
import android.os.Bundle;
import android.content.Intent;
import android.net.Uri;
import android.provider.OpenableColumns;
import android.database.Cursor;
import android.view.Gravity;
import android.widget.*;
import android.graphics.Color;

import java.io.DataOutputStream;
import java.io.InputStream;
import java.net.HttpURLConnection;
import java.net.URL;

public class MainActivity extends Activity {
    private static final int PICK_FILE = 1201;

    private EditText urlInput;
    private TextView statusText;
    private Uri selectedUri;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        buildUi();
        handleSharedFile(getIntent());
    }

    @Override
    protected void onNewIntent(Intent intent) {
        super.onNewIntent(intent);
        handleSharedFile(intent);
    }

    private void buildUi() {
        LinearLayout layout = new LinearLayout(this);
        layout.setOrientation(LinearLayout.VERTICAL);
        layout.setPadding(32, 48, 32, 32);
        layout.setGravity(Gravity.CENTER_HORIZONTAL);
        layout.setBackgroundColor(Color.rgb(15, 23, 42));

        TextView title = new TextView(this);
        title.setText("TransferLAN+");
        title.setTextSize(28);
        title.setTextColor(Color.rgb(229, 231, 235));

        TextView subtitle = new TextView(this);
        subtitle.setText("Sin cuentas. Sin nube. Sin cables.");
        subtitle.setTextSize(15);
        subtitle.setTextColor(Color.rgb(56, 189, 248));
        subtitle.setPadding(0, 0, 0, 28);

        urlInput = new EditText(this);
        urlInput.setHint("http://IP-DE-TU-PC:5050");
        urlInput.setText("http://");
        urlInput.setSingleLine(true);
        urlInput.setTextColor(Color.rgb(229, 231, 235));
        urlInput.setHintTextColor(Color.rgb(148, 163, 184));

        Button pickButton = new Button(this);
        pickButton.setText("Elegir archivo");
        pickButton.setOnClickListener(v -> pickFile());

        Button sendButton = new Button(this);
        sendButton.setText("Enviar a la PC");
        sendButton.setOnClickListener(v -> sendSelectedFile());

        statusText = new TextView(this);
        statusText.setText("Esperando archivo...");
        statusText.setTextColor(Color.rgb(229, 231, 235));
        statusText.setPadding(0, 24, 0, 0);

        layout.addView(title);
        layout.addView(subtitle);
        layout.addView(urlInput);
        layout.addView(pickButton);
        layout.addView(sendButton);
        layout.addView(statusText);

        setContentView(layout);
    }

    private void handleSharedFile(Intent intent) {
        if (intent != null && Intent.ACTION_SEND.equals(intent.getAction())) {
            Uri uri = intent.getParcelableExtra(Intent.EXTRA_STREAM);
            if (uri != null) {
                selectedUri = uri;
                statusText.setText("Archivo recibido desde Compartir: " + getFileName(uri));
            }
        }
    }

    private void pickFile() {
        Intent intent = new Intent(Intent.ACTION_OPEN_DOCUMENT);
        intent.setType("*/*");
        intent.addCategory(Intent.CATEGORY_OPENABLE);
        startActivityForResult(intent, PICK_FILE);
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (requestCode == PICK_FILE && resultCode == RESULT_OK && data != null) {
            selectedUri = data.getData();
            if (selectedUri != null) {
                statusText.setText("Seleccionado: " + getFileName(selectedUri));
            }
        }
    }

    private void sendSelectedFile() {
        if (selectedUri == null) {
            toast("Primero elegí un archivo");
            return;
        }

        String baseUrl = urlInput.getText().toString().trim();
        while (baseUrl.endsWith("/")) {
            baseUrl = baseUrl.substring(0, baseUrl.length() - 1);
        }

        if (!baseUrl.startsWith("http://") && !baseUrl.startsWith("https://")) {
            toast("Poné una URL válida");
            return;
        }

        final String finalBaseUrl = baseUrl;
        statusText.setText("Enviando...");

        new Thread(() -> {
            try {
                uploadFile(finalBaseUrl, selectedUri);
                runOnUiThread(() -> statusText.setText("Archivo enviado correctamente"));
            } catch (Exception e) {
                runOnUiThread(() -> statusText.setText("Error: " + e.getMessage()));
            }
        }).start();
    }

    private void uploadFile(String baseUrl, Uri uri) throws Exception {
        String boundary = "TransferLANBoundary" + System.currentTimeMillis();
        URL url = new URL(baseUrl + "/transfer/upload");
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();

        conn.setRequestMethod("POST");
        conn.setDoOutput(true);
        conn.setRequestProperty("Content-Type", "multipart/form-data; boundary=" + boundary);

        String filename = getFileName(uri);
        InputStream input = getContentResolver().openInputStream(uri);
        if (input == null) throw new Exception("No se pudo abrir archivo");

        DataOutputStream out = new DataOutputStream(conn.getOutputStream());
        out.writeBytes("--" + boundary + "\r\n");
        out.writeBytes("Content-Disposition: form-data; name=\"file\"; filename=\"" + filename + "\"\r\n");
        out.writeBytes("Content-Type: application/octet-stream\r\n\r\n");

        byte[] buffer = new byte[1024 * 256];
        int read;
        while ((read = input.read(buffer)) > 0) {
            out.write(buffer, 0, read);
        }
        input.close();

        out.writeBytes("\r\n--" + boundary + "--\r\n");
        out.flush();
        out.close();

        int code = conn.getResponseCode();
        if (code < 200 || code > 299) {
            throw new Exception("HTTP " + code);
        }
    }

    private String getFileName(Uri uri) {
        String name = "archivo_transferlan";
        Cursor cursor = getContentResolver().query(uri, null, null, null, null);
        if (cursor != null) {
            try {
                int index = cursor.getColumnIndex(OpenableColumns.DISPLAY_NAME);
                if (index >= 0 && cursor.moveToFirst()) {
                    name = cursor.getString(index);
                }
            } finally {
                cursor.close();
            }
        }
        return name;
    }

    private void toast(String msg) {
        Toast.makeText(this, msg, Toast.LENGTH_SHORT).show();
    }
}
