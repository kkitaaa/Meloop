import 'package:flutter/material.dart';

import '../features/music/presentation/music_preferences_screen.dart';
import '../features/profile/presentation/blocked_users_screen.dart';
import '../services/api_client.dart';

enum _AppSettingsAction { musicPreferences, blockedUsers }

class AppSettingsMenu extends StatelessWidget {
  const AppSettingsMenu({super.key, this.apiClient});

  final ApiClient? apiClient;

  @override
  Widget build(BuildContext context) => PopupMenuButton<_AppSettingsAction>(
    tooltip: 'Ajustes',
    icon: const Icon(Icons.settings, color: Colors.white),
    onSelected: (action) {
      final screen = switch (action) {
        _AppSettingsAction.musicPreferences => MusicPreferencesScreen(
          apiClient: apiClient,
        ),
        _AppSettingsAction.blockedUsers => const BlockedUsersScreen(),
      };
      Navigator.push(context, MaterialPageRoute<void>(builder: (_) => screen));
    },
    itemBuilder: (context) => const [
      PopupMenuItem(
        value: _AppSettingsAction.musicPreferences,
        child: ListTile(
          leading: Icon(Icons.queue_music_rounded),
          title: Text('Preferencias musicales'),
          contentPadding: EdgeInsets.zero,
        ),
      ),
      PopupMenuItem(
        value: _AppSettingsAction.blockedUsers,
        child: ListTile(
          leading: Icon(Icons.block_rounded),
          title: Text('Usuarios bloqueados'),
          contentPadding: EdgeInsets.zero,
        ),
      ),
    ],
  );
}
