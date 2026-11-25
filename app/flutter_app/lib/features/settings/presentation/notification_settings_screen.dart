import 'package:flutter/material.dart';

import '../../../services/api_client.dart';

class NotificationSettingsScreen extends StatefulWidget {
  const NotificationSettingsScreen({super.key, this.apiClient});

  final ApiClient? apiClient;

  @override
  State<NotificationSettingsScreen> createState() =>
      _NotificationSettingsScreenState();
}

class _NotificationSettingsScreenState
    extends State<NotificationSettingsScreen> {
  late final ApiClient _apiClient = widget.apiClient ?? ApiClient();
  late final bool _ownsApiClient = widget.apiClient == null;

  Map<String, bool> _preferences = <String, bool>{};
  String? _error;
  String? _savingType;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadPreferences();
  }

  @override
  void dispose() {
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  Future<void> _loadPreferences() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      final response = await _apiClient.getNotificationPreferences();
      final preferences = response.data;
      if (!response.success || preferences == null) {
        throw ApiException(
          message:
              response.error?.message ??
              'No se pudieron cargar las preferencias de notificaciones.',
        );
      }
      if (mounted) {
        setState(
          () => _preferences = {
            for (final item in preferences) item.type: item.enabled,
          },
        );
      }
    } on ApiException catch (error) {
      if (mounted) setState(() => _error = error.message);
    } on Exception catch (error) {
      if (mounted) setState(() => _error = error.toString());
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  Future<void> _updatePreference(String type, bool enabled) async {
    setState(() => _savingType = type);
    try {
      final response = await _apiClient.updateNotificationPreference(
        type: type,
        enabled: enabled,
      );
      final preference = response.data;
      if (!response.success || preference == null) {
        throw ApiException(
          message:
              response.error?.message ?? 'No se pudo guardar esta preferencia.',
        );
      }
      if (mounted) {
        setState(() => _preferences[type] = preference.enabled);
        _showMessage('Preferencia guardada.');
      }
    } on ApiException catch (error) {
      if (mounted) _showMessage(error.message, isError: true);
    } on Exception catch (error) {
      if (mounted) {
        _showMessage(
          'No se pudo guardar la preferencia: $error',
          isError: true,
        );
      }
    } finally {
      if (mounted) setState(() => _savingType = null);
    }
  }

  void _showMessage(String message, {bool isError = false}) {
    if (!mounted) return;
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: isError ? Colors.red.shade700 : Colors.teal.shade700,
        ),
      );
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Preferencias de notificaciones')),
    body: _isLoading
        ? const Center(child: CircularProgressIndicator())
        : _error != null
        ? _buildError()
        : _buildPreferences(),
  );

  Widget _buildError() => Center(
    child: Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.cloud_off_outlined, size: 40),
          const SizedBox(height: 12),
          Text(_error!, textAlign: TextAlign.center),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: _loadPreferences,
            icon: const Icon(Icons.refresh),
            label: const Text('Reintentar'),
          ),
        ],
      ),
    ),
  );

  Widget _buildPreferences() {
    if (_preferences.isEmpty) {
      return const Center(
        child: Text('El servidor no devolvió preferencias configurables.'),
      );
    }

    final entries = _preferences.entries.toList(growable: false);
    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 760),
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            const Text(
              'Activa o desactiva los tipos de notificación de tu cuenta. '
              'Cada cambio se guarda inmediatamente.',
            ),
            const SizedBox(height: 12),
            Card(
              child: Column(
                children: [
                  for (var index = 0; index < entries.length; index++) ...[
                    if (index > 0) const Divider(height: 1),
                    _buildPreference(entries[index]),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPreference(MapEntry<String, bool> entry) {
    final saving = _savingType == entry.key;
    return SwitchListTile(
      title: Text(_labels[entry.key] ?? entry.key.replaceAll('_', ' ')),
      value: entry.value,
      onChanged: _savingType == null
          ? (value) => _updatePreference(entry.key, value)
          : null,
      secondary: saving
          ? const SizedBox(
              width: 20,
              height: 20,
              child: CircularProgressIndicator(strokeWidth: 2),
            )
          : const Icon(Icons.notifications_active_outlined),
    );
  }

  static const _labels = {
    'SOLICITUD_AMISTAD': 'Solicitudes de amistad',
    'SOLICITUD_AMISTAD_ACEPTADA': 'Solicitudes aceptadas',
    'LIKE': 'Me gusta',
    'COMENTARIO': 'Comentarios',
    'RESPUESTA': 'Respuestas',
    'MENSAJE': 'Mensajes',
    'SUBIDA_NIVEL': 'Subidas de nivel',
    'RECOMPENSA_DESBLOQUEADA': 'Recompensas desbloqueadas',
    'ACCION_MODERACION': 'Acciones de moderación',
  };
}
