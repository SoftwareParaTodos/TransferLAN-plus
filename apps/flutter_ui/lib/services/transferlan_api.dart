import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import 'package:path/path.dart' as p;
import '../models/device.dart';
import '../models/history_entry.dart';

class TransferLanApi {
  final String baseUrl;

  TransferLanApi(this.baseUrl);

  Future<bool> health() async {
    final uri = Uri.parse('$baseUrl/health');
    final response = await http.get(uri).timeout(const Duration(seconds: 3));
    return response.statusCode == 200;
  }

  Future<List<LanDevice>> discoverDevices({int seconds = 3}) async {
    final uri = Uri.parse('$baseUrl/discovery/devices?seconds=$seconds');
    final response = await http.get(uri).timeout(Duration(seconds: seconds + 3));
    if (response.statusCode != 200) {
      throw Exception('No se pudo buscar dispositivos: ${response.statusCode}');
    }
    final decoded = jsonDecode(response.body);
    if (decoded is! List) return [];
    return decoded
        .whereType<Map<String, dynamic>>()
        .map(LanDevice.fromJson)
        .toList();
  }

  Future<void> uploadFile({
    required File file,
    required void Function(int sent, int total) onProgress,
    String token = '',
  }) async {
    final total = await file.length();
    final fileName = p.basename(file.path);
    final uri = Uri.parse('$baseUrl/transfer/upload?filename=${Uri.encodeComponent(fileName)}');
    final request = http.StreamedRequest('POST', uri);
    request.headers['X-TransferLAN-File-Name'] = fileName;
    if (token.isNotEmpty) {
      request.headers['X-TransferLAN-Token'] = token;
    }
    request.contentLength = total;

    int sent = 0;
    final source = file.openRead();
    final completer = Completer<void>();
    late StreamSubscription<List<int>> sub;
    sub = source.listen(
      (chunk) {
        request.sink.add(chunk);
        sent += chunk.length;
        onProgress(sent, total);
      },
      onDone: () {
        request.sink.close();
        completer.complete();
      },
      onError: (Object e, StackTrace st) {
        request.sink.close();
        completer.completeError(e, st);
      },
      cancelOnError: true,
    );

    final streamedFuture = request.send();
    await completer.future;
    await sub.cancel();
    final streamed = await streamedFuture;

    if (streamed.statusCode < 200 || streamed.statusCode >= 300) {
      final body = await streamed.stream.bytesToString();
      throw Exception('Error al enviar archivo: ${streamed.statusCode} $body');
    }
  }

  Future<Map<String, dynamic>> startPairPIN() async {
    final uri = Uri.parse('$baseUrl/pair/pin/start');
    final response = await http.post(uri).timeout(const Duration(seconds: 5));
    if (response.statusCode != 200) {
      throw Exception('No se pudo generar PIN: ${response.statusCode}');
    }
    return jsonDecode(response.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> pairWithReceiver({
    required String deviceId,
    required String name,
    required String platform,
    required String pin,
  }) async {
    final uri = Uri.parse('$baseUrl/pair/request');
    final response = await http.post(
      uri,
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'device_id': deviceId,
        'name': name,
        'platform': platform,
        'pin': pin,
      }),
    ).timeout(const Duration(seconds: 5));
    if (response.statusCode != 200) {
      throw Exception('Emparejamiento rechazado: ${response.statusCode} ${response.body}');
    }
    return jsonDecode(response.body) as Map<String, dynamic>;
  }

  Future<List<HistoryEntry>> history({int limit = 50}) async {
    final uri = Uri.parse('$baseUrl/history?limit=$limit');
    final response = await http.get(uri).timeout(const Duration(seconds: 5));
    if (response.statusCode != 200) {
      throw Exception('No se pudo leer historial: ${response.statusCode}');
    }
    final decoded = jsonDecode(response.body);
    if (decoded is! List) return [];
    return decoded
        .whereType<Map<String, dynamic>>()
        .map(HistoryEntry.fromJson)
        .toList();
  }

  Future<void> clearHistory() async {
    final uri = Uri.parse('$baseUrl/history/clear');
    final response = await http.post(uri).timeout(const Duration(seconds: 5));
    if (response.statusCode != 200) {
      throw Exception('No se pudo limpiar historial: ${response.statusCode}');
    }
  }

  Future<List<dynamic>> incompleteTransfers() async {
    final uri = Uri.parse('$baseUrl/transfer/chunked/incomplete');
    final response = await http.get(uri).timeout(const Duration(seconds: 5));
    if (response.statusCode != 200) return [];
    return jsonDecode(response.body) as List<dynamic>;
  }
}
