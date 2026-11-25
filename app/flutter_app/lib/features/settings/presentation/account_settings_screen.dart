import 'package:flutter/material.dart';

import '../../../models/account_settings.dart';
import '../../../services/api_client.dart';
import '../../../widgets/session_logout.dart';
import '../../profile/presentation/blocked_users_screen.dart';
import 'notification_settings_screen.dart';

class AccountSettingsScreen extends StatefulWidget {
  const AccountSettingsScreen({super.key, this.apiClient});

  final ApiClient? apiClient;

  @override
  State<AccountSettingsScreen> createState() => _AccountSettingsScreenState();
}

class _AccountSettingsScreenState extends State<AccountSettingsScreen> {
  late final ApiClient _apiClient = widget.apiClient ?? ApiClient();
  late final bool _ownsApiClient = widget.apiClient == null;

  final _currentPasswordController = TextEditingController();
  final _newPasswordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();

  AccountDetails? _account;
  PrivacySettings? _privacy;
  String? _accountError;
  String? _privacyError;
  String? _savingPrivacyField;
  bool _loadingAccount = true;
  bool _loadingPrivacy = true;
  bool _savingPassword = false;

  @override
  void initState() {
    super.initState();
    _loadAccount();
    _loadPrivacy();
  }

  @override
  void dispose() {
    _currentPasswordController.dispose();
    _newPasswordController.dispose();
    _confirmPasswordController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  Future<void> _loadAccount() async {
    setState(() {
      _loadingAccount = true;
      _accountError = null;
    });
    try {
      final response = await _apiClient.getAccountDetails();
      final account = response.data;
      if (!response.success || account == null) {
        throw ApiException(
          message:
              response.error?.message ??
              'No se pudieron cargar los datos de la cuenta.',
        );
      }
      if (mounted) setState(() => _account = account);
    } on ApiException catch (error) {
      if (mounted) setState(() => _accountError = error.message);
    } on Exception catch (error) {
      if (mounted) setState(() => _accountError = error.toString());
    } finally {
      if (mounted) setState(() => _loadingAccount = false);
    }
  }

  Future<void> _loadPrivacy() async {
    setState(() {
      _loadingPrivacy = true;
      _privacyError = null;
    });
    try {
      final response = await _apiClient.getPrivacySettings();
      final privacy = response.data;
      if (!response.success || privacy == null) {
        throw ApiException(
          message:
              response.error?.message ??
              'No se pudo cargar la configuración de privacidad.',
        );
      }
      if (mounted) setState(() => _privacy = privacy);
    } on ApiException catch (error) {
      if (mounted) setState(() => _privacyError = error.message);
    } on Exception catch (error) {
      if (mounted) setState(() => _privacyError = error.toString());
    } finally {
      if (mounted) setState(() => _loadingPrivacy = false);
    }
  }

  Future<void> _changePassword() async {
    final currentPassword = _currentPasswordController.text;
    final newPassword = _newPasswordController.text;
    if (currentPassword.isEmpty || newPassword.isEmpty) {
      _showMessage('Ingresa tu contraseña actual y la nueva contraseña.');
      return;
    }
    if (newPassword != _confirmPasswordController.text) {
      _showMessage('La confirmación no coincide con la nueva contraseña.');
      return;
    }

    setState(() => _savingPassword = true);
    try {
      final response = await _apiClient.changePassword(
        currentPassword: currentPassword,
        newPassword: newPassword,
      );
      if (!response.success) {
        throw ApiException(
          message:
              response.error?.message ?? 'No se pudo actualizar la contraseña.',
        );
      }
      _currentPasswordController.clear();
      _newPasswordController.clear();
      _confirmPasswordController.clear();
      _showMessage('Contraseña actualizada correctamente.');
    } on ApiException catch (error) {
      _showMessage(error.message, isError: true);
    } on Exception catch (error) {
      _showMessage(
        'No se pudo actualizar la contraseña: $error',
        isError: true,
      );
    } finally {
      if (mounted) setState(() => _savingPassword = false);
    }
  }

  Future<void> _savePrivacy(String field, String value) async {
    final changes = switch (field) {
      'visibilidad_perfil' => PrivacySettingsPatch(profileVisibility: value),
      'visibilidad_publicaciones' => PrivacySettingsPatch(
        postsVisibility: value,
      ),
      'visibilidad_interacciones' => PrivacySettingsPatch(
        interactionsVisibility: value,
      ),
      'recepcion_mensajes' => PrivacySettingsPatch(messageReception: value),
      'recepcion_solicitudes_amistad' => PrivacySettingsPatch(
        friendRequestReception: value,
      ),
      _ => throw ArgumentError.value(field, 'field'),
    };

    setState(() => _savingPrivacyField = field);
    try {
      final response = await _apiClient.updatePrivacySettings(changes);
      final privacy = response.data;
      if (!response.success || privacy == null) {
        throw ApiException(
          message:
              response.error?.message ??
              'No se pudo guardar la configuración de privacidad.',
        );
      }
      if (mounted) {
        setState(() => _privacy = privacy);
        _showMessage('Configuración de privacidad guardada.');
      }
    } on ApiException catch (error) {
      if (mounted) _showMessage(error.message, isError: true);
    } on Exception catch (error) {
      if (mounted) {
        _showMessage('No se pudo guardar la privacidad: $error', isError: true);
      }
    } finally {
      if (mounted) setState(() => _savingPrivacyField = null);
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
    appBar: AppBar(title: const Text('Configuración de cuenta')),
    body: Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 920),
        child: ListView(
          padding: const EdgeInsets.all(20),
          children: [
            _buildAccountSection(),
            const SizedBox(height: 12),
            _buildPrivacySection(),
            const SizedBox(height: 12),
            _buildSecuritySection(),
            const SizedBox(height: 12),
            _buildNotificationsSection(),
            const SizedBox(height: 12),
            _buildDataSection(),
            const SizedBox(height: 12),
            _buildDangerSection(),
            const SizedBox(height: 24),
          ],
        ),
      ),
    ),
  );

  Widget _buildAccountSection() => _SettingsSection(
    title: 'Cuenta',
    icon: Icons.manage_accounts_outlined,
    children: [
      if (_loadingAccount)
        const LinearProgressIndicator()
      else if (_accountError != null)
        _loadError(_accountError!, _loadAccount)
      else if (_account != null)
        ListTile(
          contentPadding: EdgeInsets.zero,
          title: const Text('Correo actual'),
          subtitle: Text(_account!.email),
          leading: const Icon(Icons.email_outlined),
        ),
      const SizedBox(height: 4),
      const Text(
        'Cambiar correo está pendiente: el backend aún no valida la contraseña '
        'actual ni envía el enlace de confirmación. Al activarlo, también debe '
        'informar el plazo RN-22 para EMAIL_CHANGE_TOO_SOON.',
      ),
      const SizedBox(height: 8),
      OutlinedButton.icon(
        onPressed: null,
        icon: const Icon(Icons.email_outlined),
        label: const Text('Cambiar correo (pendiente del backend)'),
      ),
      const Divider(height: 28),
      const Text(
        'Para proteger tu cuenta, confirma la contraseña actual antes de '
        'elegir una nueva.',
      ),
      const SizedBox(height: 12),
      TextField(
        controller: _currentPasswordController,
        obscureText: true,
        autofillHints: const [AutofillHints.password],
        decoration: const InputDecoration(
          labelText: 'Contraseña actual',
          border: OutlineInputBorder(),
        ),
      ),
      const SizedBox(height: 10),
      TextField(
        controller: _newPasswordController,
        obscureText: true,
        decoration: const InputDecoration(
          labelText: 'Nueva contraseña',
          border: OutlineInputBorder(),
        ),
      ),
      const SizedBox(height: 10),
      TextField(
        controller: _confirmPasswordController,
        obscureText: true,
        decoration: const InputDecoration(
          labelText: 'Confirmar nueva contraseña',
          border: OutlineInputBorder(),
        ),
      ),
      const SizedBox(height: 12),
      Align(
        alignment: Alignment.centerRight,
        child: FilledButton(
          onPressed: _savingPassword ? null : _changePassword,
          child: _savingPassword
              ? const SizedBox(
                  width: 18,
                  height: 18,
                  child: CircularProgressIndicator(strokeWidth: 2),
                )
              : const Text('Actualizar contraseña'),
        ),
      ),
    ],
  );

  Widget _buildPrivacySection() => _SettingsSection(
    title: 'Privacidad',
    icon: Icons.lock_outline,
    children: [
      if (_loadingPrivacy)
        const LinearProgressIndicator()
      else if (_privacyError != null)
        _loadError(_privacyError!, _loadPrivacy)
      else if (_privacy != null) ...[
        _privacyDropdown(
          title: 'Visibilidad del perfil',
          value: _privacy!.profileVisibility,
          field: 'visibilidad_perfil',
          options: _visibilityOptions,
        ),
        _privacyDropdown(
          title: 'Visibilidad de publicaciones',
          value: _privacy!.postsVisibility,
          field: 'visibilidad_publicaciones',
          options: _visibilityOptions,
        ),
        _privacyDropdown(
          title: 'Visibilidad de interacciones',
          value: _privacy!.interactionsVisibility,
          field: 'visibilidad_interacciones',
          options: _visibilityOptions,
        ),
        _privacyDropdown(
          title: 'Quién puede escribirme',
          value: _privacy!.messageReception,
          field: 'recepcion_mensajes',
          options: _receptionOptions,
        ),
        _privacyDropdown(
          title: 'Quién puede enviarme solicitudes',
          value: _privacy!.friendRequestReception,
          field: 'recepcion_solicitudes_amistad',
          options: _receptionOptions,
        ),
      ],
      const Divider(),
      const ListTile(
        contentPadding: EdgeInsets.zero,
        leading: Icon(Icons.visibility_outlined),
        title: Text('Mostrar estado en línea'),
        subtitle: Text(
          'Pendiente del backend: todavía no existe una preferencia para '
          'guardar ni aplicar esta opción.',
        ),
        trailing: Icon(Icons.lock_outline),
      ),
    ],
  );

  Widget _privacyDropdown({
    required String title,
    required String value,
    required String field,
    required List<String> options,
  }) {
    final saving = _savingPrivacyField == field;
    return ListTile(
      contentPadding: EdgeInsets.zero,
      title: Text(title),
      trailing: saving
          ? const SizedBox(
              width: 22,
              height: 22,
              child: CircularProgressIndicator(strokeWidth: 2),
            )
          : DropdownButton<String>(
              value: value,
              items: [
                for (final option in options)
                  DropdownMenuItem(
                    value: option,
                    child: Text(_privacyLabels[option] ?? option),
                  ),
              ],
              onChanged: _savingPrivacyField == null
                  ? (selected) {
                      if (selected != null && selected != value) {
                        _savePrivacy(field, selected);
                      }
                    }
                  : null,
            ),
    );
  }

  Widget _buildSecuritySection() => _SettingsSection(
    title: 'Seguridad',
    icon: Icons.shield_outlined,
    children: [
      const Text(
        'Cerrar sesión en todos los dispositivos requiere un endpoint para '
        'revocar todas las sesiones. Por ahora solo está disponible cerrar '
        'sesión en este dispositivo.',
      ),
      const SizedBox(height: 8),
      OutlinedButton.icon(
        onPressed: null,
        icon: const Icon(Icons.phonelink_erase_outlined),
        label: const Text('Cerrar sesión en todos los dispositivos'),
      ),
      const Divider(height: 24),
      ListTile(
        contentPadding: EdgeInsets.zero,
        leading: const Icon(Icons.block_outlined),
        title: const Text('Usuarios bloqueados'),
        subtitle: const Text(
          'La pantalla actual usa datos de demostración; conectar el backend '
          'de bloqueos está pendiente.',
        ),
        trailing: const Icon(Icons.chevron_right),
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute<void>(builder: (_) => const BlockedUsersScreen()),
        ),
      ),
      TextButton.icon(
        onPressed: () => SessionLogout.perform(context, apiClient: _apiClient),
        icon: const Icon(Icons.logout),
        label: const Text('Cerrar sesión en este dispositivo'),
      ),
    ],
  );

  Widget _buildNotificationsSection() => _SettingsSection(
    title: 'Notificaciones',
    icon: Icons.notifications_outlined,
    children: [
      const Text('Configura qué tipos de notificaciones quieres recibir.'),
      const SizedBox(height: 8),
      ListTile(
        contentPadding: EdgeInsets.zero,
        title: const Text('Preferencias de notificaciones'),
        subtitle: const Text('Los cambios se guardan en tu cuenta.'),
        trailing: const Icon(Icons.chevron_right),
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute<void>(
            builder: (_) => NotificationSettingsScreen(apiClient: _apiClient),
          ),
        ),
      ),
    ],
  );

  Widget _buildDataSection() => _SettingsSection(
    title: 'Mis datos',
    icon: Icons.download_outlined,
    children: [
      const Text(
        'Descargar tus datos está pendiente de un endpoint de exportación. '
        'Al habilitarlo, en escritorio se abrirá el diálogo para guardar y en '
        'móvil se podrá compartir el archivo.',
      ),
      const SizedBox(height: 8),
      OutlinedButton.icon(
        onPressed: null,
        icon: const Icon(Icons.download),
        label: const Text('Descargar mis datos (pendiente del backend)'),
      ),
    ],
  );

  Widget _buildDangerSection() => _SettingsSection(
    title: 'Zona peligrosa',
    icon: Icons.warning_amber_outlined,
    children: [
      const Text(
        'La eliminación de cuenta requiere que el backend valide tu contraseña '
        'y programe la baja con 30 días de gracia antes de eliminar los datos. '
        'La acción permanecerá deshabilitada hasta que ese flujo esté disponible.',
      ),
      const SizedBox(height: 8),
      FilledButton.icon(
        onPressed: null,
        style: FilledButton.styleFrom(
          backgroundColor: Colors.red.shade700,
          disabledBackgroundColor: Colors.red.shade100,
          disabledForegroundColor: Colors.red.shade900,
        ),
        icon: const Icon(Icons.delete_forever_outlined),
        label: const Text('Eliminar mi cuenta (pendiente del backend)'),
      ),
    ],
  );

  Widget _loadError(String message, VoidCallback retry) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(message, style: TextStyle(color: Colors.red.shade700)),
      Align(
        alignment: Alignment.centerRight,
        child: TextButton.icon(
          onPressed: retry,
          icon: const Icon(Icons.refresh),
          label: const Text('Reintentar'),
        ),
      ),
    ],
  );

  static const _visibilityOptions = ['PUBLICO', 'AMIGOS', 'PRIVADO'];
  static const _receptionOptions = ['TODOS', 'AMIGOS', 'NADIE'];
  static const _privacyLabels = {
    'PUBLICO': 'Público',
    'AMIGOS': 'Amigos',
    'PRIVADO': 'Privado',
    'TODOS': 'Todos',
    'NADIE': 'Nadie',
  };
}

class _SettingsSection extends StatelessWidget {
  const _SettingsSection({
    required this.title,
    required this.icon,
    required this.children,
  });

  final String title;
  final IconData icon;
  final List<Widget> children;

  @override
  Widget build(BuildContext context) => Card(
    clipBehavior: Clip.antiAlias,
    child: Padding(
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, color: Theme.of(context).colorScheme.primary),
              const SizedBox(width: 10),
              Expanded(
                child: Text(
                  title,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: Theme.of(context).textTheme.titleLarge,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          ...children,
        ],
      ),
    ),
  );
}
