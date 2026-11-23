import 'package:flutter/material.dart';
import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/services/health_check_service.dart';

/// Screen to test and debug Kubernetes connectivity and backend health
class KubernetesConnectivityScreen extends StatefulWidget {
  const KubernetesConnectivityScreen({super.key});

  @override
  State<KubernetesConnectivityScreen> createState() =>
      _KubernetesConnectivityScreenState();
}

class _KubernetesConnectivityScreenState
    extends State<KubernetesConnectivityScreen> {
  final _healthCheckService = HealthCheckService();

  ConnectivityReport? _report;
  bool _isLoading = false;
  String? _error;

  @override
  void dispose() {
    super.dispose();
  }

  Future<void> _runDiagnostics() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _report = null;
    });

    try {
      final report = await _healthCheckService.runDiagnostics();
      if (!mounted) return;
      setState(() {
        _report = report;
        _isLoading = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  Future<void> _testHealthEndpoint() async {
    setState(() => _isLoading = true);
    try {
      final health = await _healthCheckService.checkHealth();
      if (!mounted) return;
      _showDialog('Health Status', health.toString());
    } catch (e) {
      if (!mounted) return;
      _showDialog('Error', 'Health check failed: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _testWebSocket() async {
    setState(() => _isLoading = true);
    try {
      final result = await _healthCheckService.testWebSocket();
      if (!mounted) return;
      _showDialog(
        'WebSocket Test',
        result
            ? 'WebSocket connection successful'
            : 'WebSocket connection failed',
      );
    } catch (e) {
      if (!mounted) return;
      _showDialog('Error', 'WebSocket test failed: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _verifyEndpoints() async {
    setState(() => _isLoading = true);
    try {
      final results = await _healthCheckService.verifyEndpoints();
      if (!mounted) return;
      final message = results.entries
          .map((e) => '${e.key}: ${e.value ? '✓' : '✗'}')
          .join('\n');
      _showDialog('Endpoints Verification', message);
    } catch (e) {
      if (!mounted) return;
      _showDialog('Error', 'Endpoint verification failed: $e');
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  void _showDialog(String title, String message) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(title),
        content: SingleChildScrollView(child: SelectableText(message)),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('OK'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Kubernetes Connectivity Test'),
        elevation: 0,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Environment Info Card
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Environment Configuration',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 12),
                    _InfoRow('Environment', ApiConfig.environment.toString()),
                    _InfoRow('Base URL', ApiConfig.baseUrl),
                    _InfoRow('WebSocket URL', ApiConfig.webSocketUrl),
                    _InfoRow('Gateway Routes', ApiConfig.gatewayRouteSummary),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 20),

            // Test Buttons
            const Text(
              'Diagnostic Tests',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                ElevatedButton.icon(
                  onPressed: _isLoading ? null : _runDiagnostics,
                  icon: const Icon(Icons.health_and_safety),
                  label: const Text('Full Diagnostics'),
                ),
                ElevatedButton.icon(
                  onPressed: _isLoading ? null : _testHealthEndpoint,
                  icon: const Icon(Icons.favorite),
                  label: const Text('Health Check'),
                ),
                ElevatedButton.icon(
                  onPressed: _isLoading ? null : _verifyEndpoints,
                  icon: const Icon(Icons.check_circle),
                  label: const Text('Verify Endpoints'),
                ),
                ElevatedButton.icon(
                  onPressed: _isLoading ? null : _testWebSocket,
                  icon: const Icon(Icons.cloud),
                  label: const Text('WebSocket Test'),
                ),
              ],
            ),
            const SizedBox(height: 20),

            // Loading Indicator
            if (_isLoading) const Center(child: CircularProgressIndicator()),

            // Error Display
            if (_error != null)
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.red.shade100,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.red),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Error',
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: Colors.red,
                      ),
                    ),
                    const SizedBox(height: 8),
                    SelectableText(
                      _error ?? '',
                      style: const TextStyle(color: Colors.red),
                    ),
                  ],
                ),
              ),

            // Report Display
            if (_report != null) ...[
              const SizedBox(height: 20),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: _report!.isFullyHealthy
                      ? Colors.green.shade50
                      : Colors.orange.shade50,
                  border: Border.all(
                    color: _report!.isFullyHealthy
                        ? Colors.green
                        : Colors.orange,
                  ),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Connectivity Status',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: _report!.isFullyHealthy
                                ? Colors.green
                                : Colors.orange,
                            borderRadius: BorderRadius.circular(20),
                          ),
                          child: Text(
                            _report!.isFullyHealthy ? 'HEALTHY' : 'UNHEALTHY',
                            style: const TextStyle(
                              color: Colors.white,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 16),
                    _InfoRow('Base URL', _report!.baseUrl),
                    _InfoRow('Environment', _report!.environment),
                    _InfoRow('Backend Version', _report!.backendVersion),
                    _InfoRow(
                      'Backend Healthy',
                      _report!.backendHealthy ? '✓ Yes' : '✗ No',
                      color: _report!.backendHealthy
                          ? Colors.green
                          : Colors.red,
                    ),
                    _InfoRow(
                      'WebSocket Healthy',
                      _report!.webSocketHealthy == null
                          ? 'Not configured'
                          : _report!.webSocketHealthy == true
                          ? '✓ Yes'
                          : '✗ No',
                      color: _report!.webSocketHealthy == null
                          ? null
                          : _report!.webSocketHealthy == true
                          ? Colors.green
                          : Colors.red,
                    ),
                    const SizedBox(height: 12),
                    const Text(
                      'Endpoints',
                      style: TextStyle(fontWeight: FontWeight.bold),
                    ),
                    ..._report!.endpoints.entries.map(
                      (e) => _InfoRow(
                        e.key,
                        e.value ? '✓' : '✗',
                        color: e.value ? Colors.green : Colors.red,
                      ),
                    ),
                    if (_report!.services.isNotEmpty) ...[
                      const SizedBox(height: 12),
                      const Text(
                        'Services',
                        style: TextStyle(fontWeight: FontWeight.bold),
                      ),
                      ..._report!.services.entries.map(
                        (e) => _InfoRow(e.key, e.value.toString()),
                      ),
                    ],
                  ],
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

class _InfoRow extends StatelessWidget {
  final String label;
  final String value;
  final Color? color;

  const _InfoRow(this.label, this.value, {this.color});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(fontWeight: FontWeight.w500)),
          Flexible(
            child: SelectableText(
              value,
              textAlign: TextAlign.end,
              style: TextStyle(
                color: color,
                fontFamily: 'monospace',
                fontSize: 12,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
