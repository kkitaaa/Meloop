import 'package:flutter/material.dart';
import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/models/demo_data.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/app_background.dart';

class ApiTestScreen extends StatefulWidget {
  final ApiClient? apiClient;

  const ApiTestScreen({super.key, this.apiClient});

  @override
  State<ApiTestScreen> createState() => _ApiTestScreenState();
}

class _ApiTestScreenState extends State<ApiTestScreen> {
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  final _messageController = TextEditingController(text: 'Hola desde Flutter');

  DemoData? _getResult;
  DemoPostData? _postResult;
  ApiException? _error;
  bool _isLoading = true;
  bool _isPosting = false;

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    _loadDemo();
  }

  @override
  void dispose() {
    _messageController.dispose();
    if (_ownsApiClient) {
      _apiClient.close();
    }
    super.dispose();
  }

  Future<void> _loadDemo() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final response = await _apiClient.getDemo();
      if (!mounted) return;
      setState(() {
        _getResult = response.data;
        _isLoading = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error;
        _isLoading = false;
      });
    }
  }

  Future<void> _sendDemo() async {
    final message = _messageController.text.trim();
    if (message.isEmpty) {
      setState(() {
        _error = const ApiException(
          message: 'Escribe un mensaje antes de enviar.',
        );
      });
      return;
    }

    setState(() {
      _isPosting = true;
      _error = null;
      _postResult = null;
    });

    try {
      final response = await _apiClient.postDemo(message);
      if (!mounted) return;
      setState(() {
        _postResult = response.data;
        _isPosting = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error;
        _isPosting = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: AppBar(
        title: const Text('Conexión con API Gateway'),
        backgroundColor: const Color(0xFF1ABC9C),
        foregroundColor: Colors.white,
      ),
      body: Stack(
        children: [
          const Positioned.fill(child: AppBackground()),
          Positioned.fill(
            child: Image.asset(
              'assets/imagenes/FondoAdmin.png',
              fit: BoxFit.cover,
              frameBuilder: (context, child, frame, wasSynchronouslyLoaded) =>
                  frame == null && !wasSynchronouslyLoaded
                  ? const SizedBox.shrink()
                  : child,
            ),
          ),
          ListView(
            padding: const EdgeInsets.all(20),
            children: [
              Text('URL base: ${ApiConfig.baseUrl}'),
              const SizedBox(height: 20),
              const Text(
                'GET /api/demo',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              if (_isLoading)
                const Center(child: CircularProgressIndicator())
              else if (_getResult != null)
                Text(_getResult!.message)
              else
                const Text('No se pudo cargar la respuesta GET.'),
              Align(
                alignment: Alignment.centerLeft,
                child: TextButton.icon(
                  onPressed: _isLoading ? null : _loadDemo,
                  icon: const Icon(Icons.refresh),
                  label: const Text('Reintentar GET'),
                ),
              ),
              const Divider(height: 32),
              const Text(
                'POST /api/demo',
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              TextField(
                controller: _messageController,
                decoration: const InputDecoration(
                  labelText: 'Mensaje para el servidor',
                  border: OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 12),
              FilledButton(
                onPressed: _isPosting ? null : _sendDemo,
                child: _isPosting
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Text('Enviar POST'),
              ),
              if (_postResult != null) ...[
                const SizedBox(height: 12),
                Text('Respuesta del servidor: ${_postResult!.receivedMessage}'),
              ],
              if (_error != null) ...[
                const SizedBox(height: 16),
                Text(
                  [
                    if (_error!.statusCode != null)
                      'HTTP ${_error!.statusCode}: ',
                    _error!.message,
                  ].join(),
                  style: const TextStyle(color: Colors.red),
                ),
              ],
            ],
          ),
        ],
      ),
    );
  }
}
