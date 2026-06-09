import 'package:flutter/material.dart';
import '../models/history_entry.dart';
import '../services/transferlan_api.dart';

class HistoryScreen extends StatefulWidget {
  final String apiBaseUrl;

  const HistoryScreen({super.key, required this.apiBaseUrl});

  @override
  State<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends State<HistoryScreen> {
  late Future<List<HistoryEntry>> _future;

  @override
  void initState() {
    super.initState();
    _future = _load();
  }

  Future<List<HistoryEntry>> _load() => TransferLanApi(widget.apiBaseUrl).history(limit: 100);

  String _size(int bytes) {
    if (bytes <= 0) return '-';
    final mb = bytes / (1024 * 1024);
    if (mb < 1024) return '${mb.toStringAsFixed(1)} MB';
    return '${(mb / 1024).toStringAsFixed(2)} GB';
  }

  Future<void> _clear() async {
    await TransferLanApi(widget.apiBaseUrl).clearHistory();
    if (!mounted) return;
    setState(() => _future = _load());
  }

  IconData _icon(HistoryEntry e) {
    if (e.status == 'failed') return Icons.error_outline;
    if (e.direction == 'sent') return Icons.north_east;
    return Icons.south_west;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Historial'),
        actions: [
          IconButton(onPressed: () => setState(() => _future = _load()), icon: const Icon(Icons.refresh)),
          IconButton(onPressed: _clear, icon: const Icon(Icons.delete_outline)),
        ],
      ),
      body: FutureBuilder<List<HistoryEntry>>(
        future: _future,
        builder: (context, snapshot) {
          if (snapshot.connectionState != ConnectionState.done) {
            return const Center(child: CircularProgressIndicator());
          }
          if (snapshot.hasError) {
            return Center(child: Text('No se pudo cargar el historial: ${snapshot.error}'));
          }
          final entries = snapshot.data ?? const [];
          if (entries.isEmpty) {
            return const Center(child: Text('Todavía no hay transferencias registradas.'));
          }
          return ListView.builder(
            padding: const EdgeInsets.all(12),
            itemCount: entries.length,
            itemBuilder: (context, index) {
              final e = entries[index];
              final date = e.updatedAt ?? e.createdAt;
              return Card(
                child: ListTile(
                  leading: Icon(_icon(e)),
                  title: Text(e.fileName),
                  subtitle: Text('${e.direction} · ${e.status} · ${_size(e.sizeBytes)}\n${e.peerName.isNotEmpty ? e.peerName : e.peerHost}\n${date?.toLocal().toString() ?? ''}'),
                  isThreeLine: true,
                ),
              );
            },
          );
        },
      ),
    );
  }
}
