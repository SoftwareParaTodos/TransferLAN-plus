class HistoryEntry {
  final String id;
  final String direction;
  final String status;
  final String fileName;
  final String filePath;
  final int sizeBytes;
  final String peerName;
  final String peerHost;
  final String sha256;
  final String error;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  const HistoryEntry({
    required this.id,
    required this.direction,
    required this.status,
    required this.fileName,
    required this.filePath,
    required this.sizeBytes,
    required this.peerName,
    required this.peerHost,
    required this.sha256,
    required this.error,
    this.createdAt,
    this.updatedAt,
  });

  factory HistoryEntry.fromJson(Map<String, dynamic> json) {
    return HistoryEntry(
      id: (json['id'] ?? '').toString(),
      direction: (json['direction'] ?? '').toString(),
      status: (json['status'] ?? '').toString(),
      fileName: (json['file_name'] ?? '').toString(),
      filePath: (json['file_path'] ?? '').toString(),
      sizeBytes: json['size_bytes'] is int ? json['size_bytes'] as int : int.tryParse('${json['size_bytes']}') ?? 0,
      peerName: (json['peer_name'] ?? '').toString(),
      peerHost: (json['peer_host'] ?? '').toString(),
      sha256: (json['sha256'] ?? '').toString(),
      error: (json['error'] ?? '').toString(),
      createdAt: DateTime.tryParse((json['created_at'] ?? '').toString()),
      updatedAt: DateTime.tryParse((json['updated_at'] ?? '').toString()),
    );
  }
}
