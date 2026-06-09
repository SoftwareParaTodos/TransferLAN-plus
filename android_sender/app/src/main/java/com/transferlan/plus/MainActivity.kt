package com.transferlan.plus

import android.app.Activity
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.provider.OpenableColumns
import android.view.Gravity
import android.widget.*
import androidx.appcompat.app.AppCompatActivity
import java.io.DataOutputStream
import java.net.HttpURLConnection
import java.net.URL
import kotlin.concurrent.thread

class MainActivity : AppCompatActivity() {
    private lateinit var urlInput: EditText
    private lateinit var statusText: TextView
    private var selectedUri: Uri? = null

    companion object {
        const val PICK_FILE = 1201
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        val layout = LinearLayout(this).apply {
            orientation = LinearLayout.VERTICAL
            setPadding(32, 48, 32, 32)
            gravity = Gravity.CENTER_HORIZONTAL
            setBackgroundColor(0xFF0F172A.toInt())
        }

        val title = TextView(this).apply {
            text = "TransferLAN+"
            textSize = 28f
            setTextColor(0xFFE5E7EB.toInt())
        }

        val subtitle = TextView(this).apply {
            text = "Sin cuentas. Sin nube. Sin cables."
            textSize = 15f
            setTextColor(0xFF38BDF8.toInt())
            setPadding(0, 0, 0, 28)
        }

        urlInput = EditText(this).apply {
            hint = "http://IP-DE-TU-PC:5050"
            setText("http://")
            setSingleLine(true)
            setTextColor(0xFFE5E7EB.toInt())
            setHintTextColor(0xFF94A3B8.toInt())
        }

        val pickButton = Button(this).apply {
            text = "Elegir archivo"
            setOnClickListener { pickFile() }
        }

        val sendButton = Button(this).apply {
            text = "Enviar a la PC"
            setOnClickListener { sendSelectedFile() }
        }

        statusText = TextView(this).apply {
            text = "Esperando archivo..."
            setTextColor(0xFFE5E7EB.toInt())
            setPadding(0, 24, 0, 0)
        }

        layout.addView(title)
        layout.addView(subtitle)
        layout.addView(urlInput)
        layout.addView(pickButton)
        layout.addView(sendButton)
        layout.addView(statusText)

        setContentView(layout)

        handleSharedFile(intent)
    }

    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        handleSharedFile(intent)
    }

    private fun handleSharedFile(intent: Intent) {
        if (intent.action == Intent.ACTION_SEND) {
            val uri = intent.getParcelableExtra<Uri>(Intent.EXTRA_STREAM)
            if (uri != null) {
                selectedUri = uri
                statusText.text = "Archivo recibido desde Compartir: ${getFileName(uri)}"
            }
        }
    }

    private fun pickFile() {
        val intent = Intent(Intent.ACTION_OPEN_DOCUMENT).apply {
            type = "*/*"
            addCategory(Intent.CATEGORY_OPENABLE)
        }
        startActivityForResult(intent, PICK_FILE)
    }

    override fun onActivityResult(requestCode: Int, resultCode: Int, data: Intent?) {
        super.onActivityResult(requestCode, resultCode, data)
        if (requestCode == PICK_FILE && resultCode == Activity.RESULT_OK) {
            selectedUri = data?.data
            selectedUri?.let {
                statusText.text = "Seleccionado: ${getFileName(it)}"
            }
        }
    }

    private fun sendSelectedFile() {
        val uri = selectedUri
        if (uri == null) {
            toast("Primero elegí un archivo")
            return
        }

        val baseUrl = urlInput.text.toString().trim().removeSuffix("/")
        if (!baseUrl.startsWith("http://") && !baseUrl.startsWith("https://")) {
            toast("Poné una URL válida")
            return
        }

        statusText.text = "Enviando..."
        thread {
            try {
                uploadFile(baseUrl, uri)
                runOnUiThread {
                    statusText.text = "Archivo enviado correctamente"
                }
            } catch (e: Exception) {
                runOnUiThread {
                    statusText.text = "Error: ${e.message}"
                }
            }
        }
    }

    private fun uploadFile(baseUrl: String, uri: Uri) {
        val boundary = "TransferLANBoundary${System.currentTimeMillis()}"
        val url = URL("$baseUrl/transfer/upload")
        val conn = url.openConnection() as HttpURLConnection

        conn.requestMethod = "POST"
        conn.doOutput = true
        conn.setRequestProperty("Content-Type", "multipart/form-data; boundary=$boundary")

        val filename = getFileName(uri)
        val input = contentResolver.openInputStream(uri) ?: error("No se pudo abrir archivo")

        DataOutputStream(conn.outputStream).use { out ->
            out.writeBytes("--$boundary\r\n")
            out.writeBytes("Content-Disposition: form-data; name=\"file\"; filename=\"$filename\"\r\n")
            out.writeBytes("Content-Type: application/octet-stream\r\n\r\n")

            input.use { stream ->
                val buffer = ByteArray(1024 * 256)
                while (true) {
                    val read = stream.read(buffer)
                    if (read <= 0) break
                    out.write(buffer, 0, read)
                }
            }

            out.writeBytes("\r\n--$boundary--\r\n")
            out.flush()
        }

        val code = conn.responseCode
        if (code !in 200..299) {
            val msg = conn.errorStream?.bufferedReader()?.readText() ?: "HTTP $code"
            error(msg)
        }
    }

    private fun getFileName(uri: Uri): String {
        var name = "archivo_transferlan"
        val cursor = contentResolver.query(uri, null, null, null, null)
        cursor?.use {
            val index = it.getColumnIndex(OpenableColumns.DISPLAY_NAME)
            if (index >= 0 && it.moveToFirst()) {
                name = it.getString(index)
            }
        }
        return name
    }

    private fun toast(msg: String) {
        Toast.makeText(this, msg, Toast.LENGTH_SHORT).show()
    }
}
