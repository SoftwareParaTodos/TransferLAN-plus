class LanDevice {
  final String id;
  final String name;
  final String host;
  final int port;
  final String platform;
  final String version;
  final bool trusted;

  const LanDevice({
    required this.id,
    required this.name,
    required this.host,
    required this.port,
    this.platform = 'unknown',
    this.version = '',
    this.trusted = false,
  });

  String get baseUrl => 'http://$host:$port';

  factory LanDevice.fromJson(Map<String, dynamic> json) {
    final ips = json['ips'];
    final hostFromIps = ips is List && ips.isNotEmpty ? ips.first.toString() : null;
    return LanDevice(
      id: (json['id'] ?? '${json['name']}-${json['host']}-${json['port']}').toString(),
      name: (json['name'] ?? 'Dispositivo TransferLAN+').toString(),
      host: (json['host'] ?? hostFromIps ?? '127.0.0.1').toString(),
      port: json['port'] is int ? json['port'] as int : int.tryParse('${json['port']}') ?? 47231,
      platform: (json['platform'] ?? 'unknown').toString(),
      version: (json['version'] ?? '').toString(),
      trusted: json['trusted'] == true,
    );
  }
}
