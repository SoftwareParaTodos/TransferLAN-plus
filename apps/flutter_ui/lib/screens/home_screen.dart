import 'dart:io';
import 'package:desktop_drop/desktop_drop.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/material.dart';
import '../models/device.dart';
import '../services/transferlan_api.dart';
import '../widgets/action_button.dart';
import 'transfer_screen.dart';
import 'history_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final TextEditingController _hostController = TextEditingController(text: 'http://127.0.0.1:47231');
  bool _checking = false;
  bool _discovering = false;
  String _status = 'Receptor no verificado';
  String _pairPIN = '';
  String _pairURL = '';
  bool _pairing = false;
  List<LanDevice> _devices = const [];
  LanDevice? _selectedDevice;
  bool _dragging = false;

  Future<void> _checkReceiver() async {
    setState(() {
      _checking = true;
      _status = 'Verificando receptor...';
    });
    try {
      final ok = await TransferLanApi(_hostController.text.trim()).health();
      setState(() => _status = ok ? 'Receptor disponible' : 'No respondió correctamente');
    } catch (_) {
      setState(() => _status = 'No se pudo conectar al receptor');
    } finally {
      if (mounted) setState(() => _checking = false);
    }
  }


  Future<void> _generatePairPIN() async {
    setState(() {
      _pairing = true;
      _status = 'Generando PIN temporal...';
    });
    try {
      final data = await TransferLanApi(_hostController.text.trim()).startPairPIN();
      setState(() {
        _pairPIN = (data['pin'] ?? '').toString();
        _pairURL = (data['pairing_url'] ?? '').toString();
        _status = 'PIN generado. Usalo desde el otro dispositivo antes de que venza.';
      });
    } catch (e) {
      setState(() => _status = 'No se pudo generar PIN: $e');
    } finally {
      if (mounted) setState(() => _pairing = false);
    }
  }

  Future<void> _pairSelectedDevice() async {
    final target = _selectedDevice;
    if (target == null) {
      setState(() => _status = 'Primero seleccioná un dispositivo detectado');
      return;
    }
    final controller = TextEditingController();
    final pin = await showDialog<String>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Emparejar dispositivo'),
        content: TextField(
          controller: controller,
          keyboardType: TextInputType.number,
          maxLength: 6,
          decoration: const InputDecoration(labelText: 'PIN de 6 dígitos del receptor'),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context), child: const Text('Cancelar')),
          FilledButton(onPressed: () => Navigator.pop(context, controller.text.trim()), child: const Text('Emparejar')),
        ],
      ),
    );
    if (pin == null || pin.isEmpty) return;
    setState(() => _status = 'Emparejando con ${target.name}...');
    try {
      await TransferLanApi(target.baseUrl).pairWithReceiver(
        deviceId: 'flutter-ui-local',
        name: 'TransferLAN+ Flutter',
        platform: Platform.operatingSystem,
        pin: pin,
      );
      setState(() => _status = 'Dispositivo emparejado correctamente');
    } catch (e) {
      setState(() => _status = 'No se pudo emparejar: $e');
    }
  }

  Future<void> _discoverDevices() async {
    setState(() {
      _discovering = true;
      _status = 'Buscando dispositivos TransferLAN+ en la red...';
    });
    try {
      final found = await TransferLanApi(_hostController.text.trim()).discoverDevices(seconds: 4);
      setState(() {
        _devices = found;
        _selectedDevice = found.isNotEmpty ? found.first : null;
        _status = found.isEmpty ? 'No se detectaron dispositivos' : 'Dispositivos detectados: ${found.length}';
      });
    } catch (e) {
      setState(() => _status = 'No se pudo buscar en LAN: $e');
    } finally {
      if (mounted) setState(() => _discovering = false);
    }
  }

  Future<void> _pickAndSend({LanDevice? device}) async {
    final target = device ?? _selectedDevice;
    final apiBaseUrl = target?.baseUrl ?? _hostController.text.trim();
    final result = await FilePicker.platform.pickFiles(allowMultiple: false);
    if (result == null || result.files.single.path == null) return;
    final file = File(result.files.single.path!);
    if (!mounted) return;
    Navigator.of(context).push(MaterialPageRoute(
      builder: (_) => TransferScreen(
        file: file,
        apiBaseUrl: apiBaseUrl,
        deviceName: target?.name ?? 'Receptor manual',
      ),
    ));
  }


  void _sendDroppedPath(String path, {LanDevice? device}) {
    final target = device ?? _selectedDevice;
    final apiBaseUrl = target?.baseUrl ?? _hostController.text.trim();
    final file = File(path);
    if (!file.existsSync()) {
      setState(() => _status = 'El elemento arrastrado no es un archivo válido');
      return;
    }
    Navigator.of(context).push(MaterialPageRoute(
      builder: (_) => TransferScreen(
        file: file,
        apiBaseUrl: apiBaseUrl,
        deviceName: target?.name ?? 'Receptor manual',
      ),
    ));
  }

  IconData _iconForPlatform(String platform) {
    final lower = platform.toLowerCase();
    if (lower.contains('android')) return Icons.phone_android;
    if (lower.contains('windows') || lower.contains('linux') || lower.contains('darwin')) return Icons.computer;
    return Icons.devices;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 760),
            child: Padding(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const SizedBox(height: 12),
                  const Text('TransferLAN+', textAlign: TextAlign.center, style: TextStyle(fontSize: 34, fontWeight: FontWeight.w900)),
                  const SizedBox(height: 4),
                  const Text('Sin cuentas. Sin nube. Sin cables.', textAlign: TextAlign.center, style: TextStyle(color: Colors.white70)),
                  const SizedBox(height: 28),
                  Row(
                    children: [
                      Expanded(child: ActionButton(icon: Icons.upload_file, label: 'ENVIAR', onPressed: () => _pickAndSend())),
                      const SizedBox(width: 12),
                      Expanded(child: ActionButton(icon: Icons.search, label: 'BUSCAR LAN', onPressed: _discoverDevices)),
                    ],
                  ),
                  const SizedBox(height: 24),
                  TextField(
                    controller: _hostController,
                    decoration: InputDecoration(
                      labelText: 'Backend local de esta app',
                      helperText: 'Debe estar abierto con --mode both/server. Desde Android usá la IP de la PC.',
                      suffixIcon: _checking
                          ? const Padding(padding: EdgeInsets.all(12), child: CircularProgressIndicator(strokeWidth: 2))
                          : IconButton(icon: const Icon(Icons.health_and_safety), onPressed: _checkReceiver),
                    ),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: _pairing ? null : _generatePairPIN,
                          icon: const Icon(Icons.qr_code_2),
                          label: const Text('Mostrar PIN/QR'),
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: OutlinedButton.icon(
                          onPressed: _pairSelectedDevice,
                          icon: const Icon(Icons.verified_user),
                          label: const Text('Emparejar'),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  OutlinedButton.icon(
                    onPressed: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => HistoryScreen(apiBaseUrl: _hostController.text.trim()))),
                    icon: const Icon(Icons.history),
                    label: const Text('Ver historial'),
                  ),
                  if (_pairPIN.isNotEmpty) ...[
                    const SizedBox(height: 12),
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('PIN temporal del receptor', style: TextStyle(color: Colors.white70)),
                            Text(_pairPIN, style: const TextStyle(fontSize: 34, fontWeight: FontWeight.w900, letterSpacing: 4)),
                            if (_pairURL.isNotEmpty) Text(_pairURL, style: const TextStyle(color: Colors.white54)),
                          ],
                        ),
                      ),
                    ),
                  ],
                  const SizedBox(height: 12),
                  Card(
                    child: ListTile(
                      leading: Icon(_status.contains('disponible') || _status.contains('detectados') || _status.contains('emparejado') ? Icons.check_circle : Icons.info_outline),
                      title: Text(_status),
                      subtitle: const Text('v0.10.0: la UI suma PIN/QR lógico y consulta /discovery/devices y muestra equipos reales detectados por mDNS.'),
                      trailing: _discovering ? const SizedBox(width: 24, height: 24, child: CircularProgressIndicator(strokeWidth: 2)) : null,
                    ),
                  ),
                  const SizedBox(height: 20),
                  Row(
                    children: [
                      const Expanded(child: Text('Dispositivos detectados', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold))),
                      TextButton.icon(onPressed: _discovering ? null : _discoverDevices, icon: const Icon(Icons.refresh), label: const Text('Actualizar')),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Expanded(
                    child: DropTarget(
                      onDragEntered: (_) => setState(() => _dragging = true),
                      onDragExited: (_) => setState(() => _dragging = false),
                      onDragDone: (detail) {
                        setState(() => _dragging = false);
                        if (detail.files.isNotEmpty) {
                          _sendDroppedPath(detail.files.first.path);
                        }
                      },
                      child: AnimatedContainer(
                        duration: const Duration(milliseconds: 180),
                        decoration: BoxDecoration(
                          border: Border.all(color: _dragging ? Theme.of(context).colorScheme.primary : Colors.transparent, width: 2),
                          borderRadius: BorderRadius.circular(18),
                        ),
                        child: _devices.isEmpty
                            ? Center(child: Text(_dragging ? 'Soltá el archivo para enviarlo' : 'Abrí TransferLAN+ en otra PC/celular, tocá “BUSCAR LAN” o arrastrá un archivo acá.'))
                            : ListView.builder(
                                itemCount: _devices.length,
                                itemBuilder: (context, index) {
                                  final device = _devices[index];
                                  final selected = _selectedDevice?.id == device.id;
                                  return Card(
                                    child: ListTile(
                                      selected: selected,
                                      leading: Icon(_iconForPlatform(device.platform)),
                                      title: Text(device.name),
                                      subtitle: Text('${device.baseUrl}  ·  ${device.platform}  ·  v${device.version}'),
                                      trailing: FilledButton.icon(
                                        onPressed: () => _pickAndSend(device: device),
                                        icon: const Icon(Icons.send),
                                        label: const Text('Enviar'),
                                      ),
                                      onTap: () => setState(() => _selectedDevice = device),
                                    ),
                                  );
                                },
                              ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
