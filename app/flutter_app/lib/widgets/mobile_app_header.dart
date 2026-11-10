import 'package:flutter/material.dart';

import '../services/api_client.dart';
import 'app_settings_menu.dart';
import 'notifications_menu.dart';

class MobileAppHeader extends StatelessWidget {
  const MobileAppHeader({
    super.key,
    this.onModerationTap,
    this.onSearchTap,
    this.apiClient,
    this.actions = const [],
  });

  final VoidCallback? onModerationTap;
  final VoidCallback? onSearchTap;
  final ApiClient? apiClient;
  final List<Widget> actions;

  static const _accentColor = Color(0xFF1ABC9C);

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 55,
      color: _accentColor,
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Stack(
        alignment: Alignment.center,
        children: [
          Positioned.fill(
            child: IgnorePointer(
              child: Center(
                child: Image.asset(
                  'assets/imagenes/meloop.png',
                  height: 24,
                  fit: BoxFit.contain,
                ),
              ),
            ),
          ),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              if (onSearchTap == null && onModerationTap == null)
                const SizedBox(width: 48)
              else
                Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (onSearchTap != null)
                      IconButton(
                        tooltip: 'Búsqueda global',
                        onPressed: onSearchTap,
                        icon: const Icon(Icons.search, color: Colors.white),
                      ),
                    if (onModerationTap != null)
                      IconButton(
                        tooltip: 'Cola de moderación',
                        onPressed: onModerationTap,
                        icon: const Icon(
                          Icons.shield_outlined,
                          color: Colors.white,
                        ),
                      ),
                  ],
                ),
              Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  ...actions,
                  AppSettingsMenu(apiClient: apiClient),
                  const NotificationsMenu(),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }
}
