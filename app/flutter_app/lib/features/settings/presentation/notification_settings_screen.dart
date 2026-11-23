import 'package:flutter/material.dart';

import '../../../services/api_client.dart';
import 'settings_page_shell.dart';

class NotificationSettingsScreen extends StatefulWidget {
  const NotificationSettingsScreen({super.key, required this.apiClient});

  final ApiClient apiClient;

  @override
  State<NotificationSettingsScreen> createState() =>
      _NotificationSettingsScreenState();
}

class _NotificationSettingsScreenState
    extends State<NotificationSettingsScreen> {
  static const _types = <String, String>{
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

  final Map<String, bool> _settings = {
    for (final type in _types.keys) type: true,
  };
  final Set<String> _saving = {};
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final settings = await widget.apiClient.getNotificationSettings();
      if (!mounted) return;
      setState(() {
        for (final setting in settings) {
          final type = setting['tipo_notificacion'];
          final enabled = setting['habilitada'];
          if (type is String && enabled is bool && _types.containsKey(type)) {
            _settings[type] = enabled;
          }
        }
        _loading = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error.message;
        _loading = false;
      });
    }
  }

  Future<void> _update(String type, bool enabled) async {
    setState(() => _saving.add(type));
    try {
      await widget.apiClient.updateNotificationSetting(
        type: type,
        enabled: enabled,
      );
      if (!mounted) return;
      setState(() {
        _settings[type] = enabled;
        _saving.remove(type);
      });
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(const SnackBar(content: Text('Preferencia guardada.')));
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _saving.remove(type));
      ScaffoldMessenger.of(context)
        ..hideCurrentSnackBar()
        ..showSnackBar(
          SnackBar(
            content: Text(error.message),
            backgroundColor: Colors.red.shade700,
          ),
        );
    }
  }

  @override
  Widget build(BuildContext context) => SettingsPageShell(
    title: 'Preferencias de notificación',
    icon: Icons.notifications_outlined,
    maxWidth: 720,
    child: _loading
        ? const Center(
            child: CircularProgressIndicator(color: SettingsColors.accent),
          )
        : _error != null
        ? _errorPanel()
        : ListView(
            padding: const EdgeInsets.all(20),
            children: [
              const Text('Activa o desactiva cada tipo de notificación.'),
              const SizedBox(height: 12),
              SettingsSectionCard(
                title: 'Tipos de aviso',
                icon: Icons.tune_rounded,
                child: Column(
                  children: [
                    for (final entry in _types.entries)
                      SwitchListTile(
                        contentPadding: EdgeInsets.zero,
                        title: Text(entry.value),
                        activeThumbColor: SettingsColors.accent,
                        value: _settings[entry.key] ?? true,
                        onChanged: _saving.contains(entry.key)
                            ? null
                            : (value) => _update(entry.key, value),
                        secondary: _saving.contains(entry.key)
                            ? const SizedBox(
                                width: 20,
                                height: 20,
                                child: CircularProgressIndicator(
                                  strokeWidth: 2,
                                  color: SettingsColors.accent,
                                ),
                              )
                            : null,
                      ),
                  ],
                ),
              ),
            ],
          ),
  );

  Widget _errorPanel() => Center(
    child: Padding(
      padding: const EdgeInsets.all(24),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(_error!, textAlign: TextAlign.center),
          const SizedBox(height: 8),
          FilledButton.icon(
            onPressed: _load,
            icon: const Icon(Icons.refresh),
            label: const Text('Reintentar'),
          ),
        ],
      ),
    ),
  );
}
