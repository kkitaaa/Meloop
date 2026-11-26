import 'package:flutter/material.dart';

import '../features/settings/presentation/account_settings_screen.dart';
import '../features/music/presentation/music_preferences_screen.dart';
import '../features/profile/presentation/blocked_users_screen.dart';
import '../features/settings/presentation/settings_page_shell.dart';
import '../services/api_client.dart';
import 'session_logout.dart';

enum _AppSettingsAction {
  accountSettings,
  musicPreferences,
  blockedUsers,
  logout,
}

class AppSettingsMenu extends StatelessWidget {
  const AppSettingsMenu({super.key, this.apiClient});

  final ApiClient? apiClient;

  @override
  Widget build(BuildContext context) => PopupMenuButton<_AppSettingsAction>(
    tooltip: 'Ajustes',
    icon: const Icon(Icons.settings, color: Colors.white),
    onSelected: (action) async {
      if (action == _AppSettingsAction.logout) {
        await SessionLogout.perform(context, apiClient: apiClient);
        return;
      }

      final screen = switch (action) {
        _AppSettingsAction.accountSettings => AccountSettingsScreen(
          apiClient: apiClient,
        ),
        _AppSettingsAction.musicPreferences => MusicPreferencesScreen(
          apiClient: apiClient,
        ),
        _AppSettingsAction.blockedUsers => const BlockedUsersScreen(),
        _AppSettingsAction.logout => const SizedBox.shrink(),
      };
      Navigator.push(context, MaterialPageRoute<void>(builder: (_) => screen));
    },
    itemBuilder: (context) => const [
      PopupMenuItem(
        value: _AppSettingsAction.accountSettings,
        child: ListTile(
          leading: Icon(
            Icons.manage_accounts_rounded,
            color: SettingsColors.accent,
          ),
          title: Text(
            'Configuración de cuenta',
            style: TextStyle(color: SettingsColors.text),
          ),
          contentPadding: EdgeInsets.zero,
        ),
      ),
      PopupMenuItem(
        value: _AppSettingsAction.musicPreferences,
        child: ListTile(
          leading: Icon(
            Icons.queue_music_rounded,
            color: SettingsColors.accent,
          ),
          title: Text(
            'Preferencias musicales',
            style: TextStyle(color: SettingsColors.text),
          ),
          contentPadding: EdgeInsets.zero,
        ),
      ),
      PopupMenuItem(
        value: _AppSettingsAction.blockedUsers,
        child: ListTile(
          leading: Icon(Icons.block_rounded, color: SettingsColors.accent),
          title: Text(
            'Usuarios bloqueados',
            style: TextStyle(color: SettingsColors.text),
          ),
          contentPadding: EdgeInsets.zero,
        ),
      ),
      PopupMenuDivider(),
      PopupMenuItem(
        value: _AppSettingsAction.logout,
        child: ListTile(
          leading: Icon(Icons.logout_rounded, color: Colors.red),
          title: Text('Cerrar sesión', style: TextStyle(color: Colors.red)),
          contentPadding: EdgeInsets.zero,
        ),
      ),
    ],
  );
}
