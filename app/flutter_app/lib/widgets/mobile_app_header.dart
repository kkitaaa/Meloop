import 'package:flutter/material.dart';

import '../services/api_client.dart';
import '../theme/meloop_theme_controller.dart';
import 'app_settings_menu.dart';
import 'notifications_menu.dart';
import 'paper/stacked_paper_card.dart';

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

  @override
  Widget build(BuildContext context) {
    final themeCtrl = MeloopThemeController.instance;
    return AnimatedBuilder(
      animation: themeCtrl,
      builder: (context, _) => Container(
        height: 55,
        color: themeCtrl.primaryColor,
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Stack(
          alignment: Alignment.center,
          children: [
            Positioned.fill(
              child: IgnorePointer(
                child: Center(
                  child: MeloopLogoSticker(height: 26, strokeWidth: 1.6),
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
      ),
    );
  }
}
