import 'package:flutter/material.dart';

import '../features/music/presentation/music_preferences_screen.dart';
import '../features/profile/presentation/blocked_users_screen.dart';
import '../features/settings/presentation/account_settings_screen.dart';
import '../features/settings/presentation/settings_page_shell.dart';
import '../services/api_client.dart';

enum _AppSettingsAction { accountSettings, musicPreferences, blockedUsers }

class AppSettingsMenu extends StatelessWidget {
  const AppSettingsMenu({super.key, this.apiClient});

  final ApiClient? apiClient;

  @override
  Widget build(BuildContext context) => PopupMenuButton<_AppSettingsAction>(
    tooltip: 'Ajustes',
    icon: const Icon(Icons.settings, color: Colors.white),
    color: Colors.white,
    surfaceTintColor: Colors.transparent,
    elevation: 10,
    offset: const Offset(0, 8),
    shape: RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(14),
      side: const BorderSide(color: SettingsColors.border),
    ),
    onSelected: (action) {
      final screen = switch (action) {
        _AppSettingsAction.accountSettings => AccountSettingsScreen(
          apiClient: apiClient,
        ),
        _AppSettingsAction.musicPreferences => MusicPreferencesScreen(
          apiClient: apiClient,
        ),
        _AppSettingsAction.blockedUsers => const BlockedUsersScreen(),
      };
      Navigator.push(context, MaterialPageRoute<void>(builder: (_) => screen));
    },
    itemBuilder: (context) => [
      PopupMenuItem(
        value: _AppSettingsAction.accountSettings,
        child: ListTile(
          leading: const Icon(
            Icons.manage_accounts_outlined,
            color: SettingsColors.accent,
          ),
          title: const Text(
            'Configuración de cuenta',
            style: TextStyle(color: SettingsColors.text),
          ),
          contentPadding: EdgeInsets.zero,
        ),
      ),
      PopupMenuItem(
        value: _AppSettingsAction.musicPreferences,
        child: ListTile(
          leading: const Icon(
            Icons.queue_music_rounded,
            color: SettingsColors.accent,
          ),
          title: const Text(
            'Preferencias musicales',
            style: TextStyle(color: SettingsColors.text),
          ),
          contentPadding: EdgeInsets.zero,
        ),
      ),
      PopupMenuItem(
        value: _AppSettingsAction.blockedUsers,
        child: ListTile(
          leading: const Icon(
            Icons.block_rounded,
            color: SettingsColors.accent,
          ),
          title: const Text(
            'Usuarios bloqueados',
            style: TextStyle(color: SettingsColors.text),
          ),
          contentPadding: EdgeInsets.zero,
        ),
      ),
    ],
  );
}
