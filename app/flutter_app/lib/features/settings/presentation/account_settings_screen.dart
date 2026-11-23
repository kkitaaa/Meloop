import 'package:flutter/material.dart';

import '../../../services/api_client.dart';
import '../../../services/auth_session.dart';
import '../../profile/presentation/blocked_users_screen.dart';
import 'notification_settings_screen.dart';
import 'settings_page_shell.dart';

class AccountSettingsScreen extends StatefulWidget {
  const AccountSettingsScreen({super.key, this.apiClient});

  final ApiClient? apiClient;

  @override
  State<AccountSettingsScreen> createState() => _AccountSettingsScreenState();
}

class _AccountSettingsScreenState extends State<AccountSettingsScreen> {
  late final ApiClient _apiClient = widget.apiClient ?? ApiClient();
  final Map<String, String> _privacy = {};
  final Set<String> _savingPrivacy = {};
  bool _ownsApiClient = false;
  bool _loadingPrivacy = true;
  String? _privacyError;

  static const _visibilityValues = ['PUBLICO', 'AMIGOS', 'PRIVADO'];
  static const _receptionValues = ['TODOS', 'AMIGOS', 'NADIE'];

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _loadPrivacy();
  }

  @override
  void dispose() {
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  Future<void> _loadPrivacy() async {
    setState(() {
      _loadingPrivacy = true;
      _privacyError = null;
    });
    try {
      final settings = await _apiClient.getPrivacySettings();
      if (!mounted) return;
      setState(() {
        _privacy
          ..clear()
          ..addEntries(
            settings.entries
                .where((entry) => entry.value is String)
                .map((entry) => MapEntry(entry.key, entry.value as String)),
          );
        _loadingPrivacy = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _privacyError = error.message;
        _loadingPrivacy = false;
      });
    }
  }

  Future<void> _savePrivacy(String field, String value) async {
    setState(() => _savingPrivacy.add(field));
    try {
      final settings = await _apiClient.updatePrivacySetting(field, value);
      if (!mounted) return;
      setState(() {
        _privacy
          ..clear()
          ..addEntries(
            settings.entries
                .where((entry) => entry.value is String)
                .map((entry) => MapEntry(entry.key, entry.value as String)),
          );
        _savingPrivacy.remove(field);
      });
      _showMessage('Privacidad actualizada.');
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _savingPrivacy.remove(field));
      _showMessage(error.message, isError: true);
    }
  }

  Future<void> _changePassword() async {
    final formKey = GlobalKey<FormState>();
    final currentPassword = TextEditingController();
    final newPassword = TextEditingController();
    final confirmation = TextEditingController();

    final password = await showDialog<String>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Cambiar contraseña'),
        content: Form(
          key: formKey,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextFormField(
                controller: currentPassword,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Contraseña actual',
                ),
                validator: (value) => value == null || value.isEmpty
                    ? 'Ingresa tu contraseña actual.'
                    : null,
              ),
              TextFormField(
                controller: newPassword,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Nueva contraseña',
                ),
                validator: (value) => value == null || value.length < 8
                    ? 'Usa al menos 8 caracteres.'
                    : null,
              ),
              TextFormField(
                controller: confirmation,
                obscureText: true,
                decoration: const InputDecoration(
                  labelText: 'Confirmar nueva contraseña',
                ),
                validator: (value) => value != newPassword.text
                    ? 'Las contraseñas no coinciden.'
                    : null,
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () {
              if (formKey.currentState?.validate() ?? false) {
                Navigator.pop(dialogContext, newPassword.text);
              }
            },
            child: const Text('Guardar'),
          ),
        ],
      ),
    );

    try {
      if (password == null || !mounted) return;
      await _apiClient.changePassword(
        currentPassword: currentPassword.text,
        newPassword: password,
      );
      if (mounted) _showMessage('Contraseña actualizada.');
    } on ApiException catch (error) {
      if (mounted) _showMessage(error.message, isError: true);
    } finally {
      currentPassword.dispose();
      newPassword.dispose();
      confirmation.dispose();
    }
  }

  void _showMessage(String message, {bool isError = false}) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(
        SnackBar(
          content: Text(message),
          backgroundColor: isError ? Colors.red.shade700 : null,
        ),
      );
  }

  @override
  Widget build(BuildContext context) => SettingsPageShell(
    title: 'Configuración de cuenta',
    icon: Icons.manage_accounts_outlined,
    maxWidth: 820,
    child: ListView(
      padding: const EdgeInsets.all(20),
      children: [
        _section(
          title: 'Cuenta',
          icon: Icons.person_outline,
          children: [
            _emailChangePanel(),
            const Divider(height: 28),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.password),
              title: const Text('Cambiar contraseña'),
              subtitle: const Text(
                'Confirma tu contraseña actual para guardar una nueva.',
              ),
              trailing: const Icon(Icons.chevron_right),
              onTap: _changePassword,
            ),
          ],
        ),
        _section(
          title: 'Privacidad',
          icon: Icons.visibility_outlined,
          children: [
            const Text('Controla quién puede ver tu contenido y contactarte.'),
            const SizedBox(height: 12),
            if (_loadingPrivacy)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(20),
                  child: CircularProgressIndicator(),
                ),
              )
            else if (_privacyError != null)
              _loadError(_privacyError!, _loadPrivacy)
            else ...[
              _privacySelector(
                field: 'visibilidad_perfil',
                title: 'Visibilidad del perfil',
                values: _visibilityValues,
              ),
              _privacySelector(
                field: 'visibilidad_publicaciones',
                title: 'Visibilidad de publicaciones',
                values: _visibilityValues,
              ),
              _privacySelector(
                field: 'visibilidad_interacciones',
                title: 'Visibilidad de interacciones',
                values: _visibilityValues,
              ),
              _privacySelector(
                field: 'recepcion_mensajes',
                title: 'Quién puede escribirme',
                values: _receptionValues,
              ),
              _privacySelector(
                field: 'recepcion_solicitudes_amistad',
                title: 'Quién puede enviarme solicitudes',
                values: _receptionValues,
              ),
            ],
            const ListTile(
              contentPadding: EdgeInsets.zero,
              title: Text('Mostrar estado en línea'),
              subtitle: Text(
                'No disponible: el backend aún no guarda esta preferencia.',
              ),
              trailing: Switch(value: false, onChanged: null),
            ),
          ],
        ),
        _section(
          title: 'Seguridad',
          icon: Icons.shield_outlined,
          children: [
            const ListTile(
              contentPadding: EdgeInsets.zero,
              leading: Icon(Icons.devices),
              title: Text('Cerrar sesión en todos los dispositivos'),
              subtitle: Text(
                'No disponible: el backend no tiene un endpoint para revocar todas las sesiones.',
              ),
              trailing: Icon(Icons.lock_outline),
            ),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: const Icon(Icons.block_outlined),
              title: const Text('Usuarios bloqueados'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => Navigator.push(
                context,
                MaterialPageRoute<void>(
                  builder: (_) => const BlockedUsersScreen(),
                ),
              ),
            ),
          ],
        ),
        _section(
          title: 'Notificaciones',
          icon: Icons.notifications_outlined,
          children: [
            ListTile(
              contentPadding: EdgeInsets.zero,
              title: const Text('Preferencias de notificación'),
              subtitle: const Text('Configura cada tipo de aviso.'),
              trailing: const Icon(Icons.chevron_right),
              onTap: () => Navigator.push(
                context,
                MaterialPageRoute<void>(
                  builder: (_) =>
                      NotificationSettingsScreen(apiClient: _apiClient),
                ),
              ),
            ),
          ],
        ),
        _section(
          title: 'Mis datos',
          icon: Icons.download_outlined,
          children: const [
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: Icon(Icons.file_download_outlined),
              title: Text('Descargar mis datos'),
              subtitle: Text(
                'No disponible: el backend aún no ofrece exportación de datos.',
              ),
              trailing: Icon(Icons.lock_outline),
            ),
          ],
        ),
        _section(
          title: 'Zona peligrosa',
          icon: Icons.warning_amber_outlined,
          children: const [
            Text(
              'La eliminación debe explicar el período de gracia de 30 días y requerir tu contraseña.',
            ),
            SizedBox(height: 8),
            ListTile(
              contentPadding: EdgeInsets.zero,
              leading: Icon(Icons.delete_forever_outlined, color: Colors.red),
              title: Text(
                'Eliminar mi cuenta',
                style: TextStyle(color: Colors.red),
              ),
              subtitle: Text(
                'No disponible: el backend aún no implementa eliminación ni período de gracia.',
              ),
              trailing: Icon(Icons.lock_outline),
            ),
          ],
        ),
      ],
    ),
  );

  Widget _section({
    required String title,
    required IconData icon,
    required List<Widget> children,
  }) => SettingsSectionCard(
    title: title,
    icon: icon,
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: children,
    ),
  );

  Widget _emailChangePanel() {
    final email = AuthSession.shared.user?.email;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('Correo actual: ${email ?? 'No disponible'}'),
        const SizedBox(height: 8),
        const Text(
          'El cambio requiere tu contraseña, verificación por enlace y un límite de 30 días (RN-22). El endpoint actual no implementa esos controles, por lo que la solicitud está deshabilitada.',
        ),
        const SizedBox(height: 8),
        const TextField(
          enabled: false,
          decoration: InputDecoration(labelText: 'Nuevo correo'),
        ),
        const TextField(
          enabled: false,
          obscureText: true,
          decoration: InputDecoration(labelText: 'Contraseña actual'),
        ),
        const SizedBox(height: 8),
        const OutlinedButton(
          onPressed: null,
          child: Text('Solicitar cambio de correo'),
        ),
      ],
    );
  }

  Widget _privacySelector({
    required String field,
    required String title,
    required List<String> values,
  }) {
    final selected = _privacy[field];
    final validValue = values.contains(selected) ? selected : null;
    return ListTile(
      contentPadding: EdgeInsets.zero,
      title: Text(title),
      trailing: _savingPrivacy.contains(field)
          ? const SizedBox(
              width: 22,
              height: 22,
              child: CircularProgressIndicator(strokeWidth: 2),
            )
          : DropdownButton<String>(
              value: validValue,
              hint: const Text('Seleccionar'),
              items: values
                  .map(
                    (value) => DropdownMenuItem(
                      value: value,
                      child: Text(_privacyLabel(value)),
                    ),
                  )
                  .toList(),
              onChanged: validValue == null || _savingPrivacy.contains(field)
                  ? null
                  : (value) {
                      if (value != null && value != validValue) {
                        _savePrivacy(field, value);
                      }
                    },
            ),
    );
  }

  String _privacyLabel(String value) => switch (value) {
    'PUBLICO' || 'TODOS' => 'Todos',
    'AMIGOS' => 'Amigos',
    'PRIVADO' || 'NADIE' => 'Nadie',
    _ => value,
  };

  Widget _loadError(String message, VoidCallback retry) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(message, style: TextStyle(color: Colors.red.shade700)),
      TextButton.icon(
        onPressed: retry,
        icon: const Icon(Icons.refresh),
        label: const Text('Reintentar'),
      ),
    ],
  );
}
