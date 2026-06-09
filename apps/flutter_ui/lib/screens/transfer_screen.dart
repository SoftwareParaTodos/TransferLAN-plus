import 'dart:io';
import 'package:flutter/material.dart';
import 'package:path/path.dart' as p;
import '../services/transferlan_api.dart';

class TransferScreen extends StatefulWidget {
  final File file;
  final String apiBaseUrl;
  final String deviceName;

  const TransferScreen({super.key, required this.file, required this.apiBaseUrl, this.deviceName = 'Receptor'});

  @override
  State<TransferScreen> createState() => _TransferScreenState();
}

class _TransferScreenState extends State<TransferScreen> {
  double _progress = 0;
  String _status = 'Listo para enviar';
  final TextEditingController _tokenController = TextEditingController();
  bool _sending = false;

  Future<void> _send() async {
    setState(() {
      _sending = true;
      _status = 'Enviando archivo...';
      _progress = 0.05;
    });
    try {
      final api = TransferLanApi(widget.apiBaseUrl);
      await api.uploadFile(
        file: widget.file,
        token: _tokenController.text.trim(),
        onProgress: (sent, total) {
          if (!mounted || total == 0) return;
          setState(() => _progress = sent / total);
        },
      );
      setState(() {
        _progress = 1;
        _status = 'Transferencia completada';
      });
    } catch (e) {
      setState(() => _status = 'Error: $e');
    } finally {
      setState(() => _sending = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final fileName = p.basename(widget.file.path);
    return Scaffold(
      appBar: AppBar(title: const Text('Enviar archivo')),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Card(
              child: ListTile(
                leading: const Icon(Icons.insert_drive_file),
                title: Text(fileName),
                subtitle: Text('${widget.file.path}\nDestino: ${widget.deviceName} · ${widget.apiBaseUrl}'),
              ),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _tokenController,
              decoration: const InputDecoration(
                labelText: 'Token de dispositivo confiable',
                helperText: 'v0.10.0: pegá el token devuelto al emparejar. En v0.11.0 se guardará automático.',
              ),
            ),
            const SizedBox(height: 24),
            LinearProgressIndicator(value: _progress),
            const SizedBox(height: 12),
            Text('${(_progress * 100).toStringAsFixed(0)}%', textAlign: TextAlign.center, style: const TextStyle(fontSize: 28, fontWeight: FontWeight.bold)),
            const SizedBox(height: 8),
            Text(_status, textAlign: TextAlign.center),
            const Spacer(),
            FilledButton.icon(
              onPressed: _sending ? null : _send,
              icon: const Icon(Icons.send),
              label: const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Text('Enviar ahora'),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
