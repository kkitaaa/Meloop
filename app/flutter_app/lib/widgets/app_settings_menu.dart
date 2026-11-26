import 'package:flutter/material.dart';

import '../features/music/presentation/music_preferences_screen.dart';
import '../features/profile/presentation/blocked_users_screen.dart';
import '../features/settings/presentation/account_settings_screen.dart';
import '../features/settings/presentation/settings_page_shell.dart';
import '../services/api_client.dart';
import '../theme/meloop_typography.dart';
import 'paper/stacked_paper_card.dart';
import 'session_logout.dart';

enum _AppSettingsAction {
  accountSettings,
  musicPreferences,
  blockedUsers,
  logout,
}

class AppSettingsMenu extends StatefulWidget {
  const AppSettingsMenu({super.key, this.apiClient});

  final ApiClient? apiClient;

  @override
  State<AppSettingsMenu> createState() => _AppSettingsMenuState();
}

class _AppSettingsMenuState extends State<AppSettingsMenu> {
  final MenuController _menuController = MenuController();

  Future<void> _selectAction(_AppSettingsAction action) async {
    _menuController.close();
    if (action == _AppSettingsAction.logout) {
      await SessionLogout.perform(context, apiClient: widget.apiClient);
      return;
    }

    final screen = switch (action) {
      _AppSettingsAction.accountSettings => AccountSettingsScreen(
        apiClient: widget.apiClient,
      ),
      _AppSettingsAction.musicPreferences => MusicPreferencesScreen(
        apiClient: widget.apiClient,
      ),
      _AppSettingsAction.blockedUsers => const BlockedUsersScreen(),
      _AppSettingsAction.logout => const SizedBox.shrink(),
    };
    if (!mounted) return;
    await Navigator.push(
      context,
      MaterialPageRoute<void>(builder: (_) => screen),
    );
  }

  @override
  Widget build(BuildContext context) {
    return MenuAnchor(
      controller: _menuController,
      style: const MenuStyle(
        padding: WidgetStatePropertyAll(EdgeInsets.zero),
        elevation: WidgetStatePropertyAll(0),
        backgroundColor: WidgetStatePropertyAll(Colors.transparent),
        surfaceTintColor: WidgetStatePropertyAll(Colors.transparent),
        shadowColor: WidgetStatePropertyAll(Colors.transparent),
      ),
      menuChildren: [
        SizedBox(
          width: 248,
          child: PhotoPaperSheetStack(
            sheetCount: 1,
            shadowBlur: 16,
            padding: const EdgeInsets.symmetric(vertical: 6),
            child: Material(
              color: Colors.transparent,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  _SettingsPaperMenuItem(
                    icon: Icons.manage_accounts_rounded,
                    iconColor: SettingsColors.accent,
                    label: 'Configuración de cuenta',
                    textColor: SettingsColors.text,
                    onTap: () =>
                        _selectAction(_AppSettingsAction.accountSettings),
                  ),
                  _SettingsPaperMenuItem(
                    icon: Icons.queue_music_rounded,
                    iconColor: SettingsColors.accent,
                    label: 'Preferencias musicales',
                    textColor: SettingsColors.text,
                    onTap: () =>
                        _selectAction(_AppSettingsAction.musicPreferences),
                  ),
                  _SettingsPaperMenuItem(
                    icon: Icons.block_rounded,
                    iconColor: SettingsColors.accent,
                    label: 'Usuarios bloqueados',
                    textColor: SettingsColors.text,
                    onTap: () => _selectAction(_AppSettingsAction.blockedUsers),
                  ),
                  Divider(
                    height: 10,
                    thickness: 1,
                    indent: 14,
                    endIndent: 14,
                    color: Colors.black.withValues(alpha: 0.08),
                  ),
                  _SettingsPaperMenuItem(
                    icon: Icons.logout_rounded,
                    iconColor: const Color(0xFFD94444),
                    label: 'Cerrar sesión',
                    textColor: const Color(0xFFD94444),
                    onTap: () => _selectAction(_AppSettingsAction.logout),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
      builder: (context, controller, child) => IconButton(
        tooltip: 'Ajustes',
        icon: const Icon(Icons.settings, color: Colors.white),
        onPressed: () =>
            controller.isOpen ? controller.close() : controller.open(),
      ),
    );
  }
}

class _SettingsPaperMenuItem extends StatelessWidget {
  const _SettingsPaperMenuItem({
    required this.icon,
    required this.iconColor,
    required this.label,
    required this.textColor,
    required this.onTap,
  });

  final IconData icon;
  final Color iconColor;
  final String label;
  final Color textColor;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        child: Row(
          children: [
            Icon(icon, color: iconColor, size: 20),
            const SizedBox(width: 12),
            Expanded(
              child: Text(
                label,
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: MeloopTypography.lexend(
                  fontSize: 13.5,
                  fontWeight: FontWeight.w500,
                  color: textColor,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
