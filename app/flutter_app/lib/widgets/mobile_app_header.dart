import 'package:flutter/material.dart';

import 'notifications_menu.dart';

class MobileAppHeader extends StatelessWidget {
  const MobileAppHeader({super.key, this.onModerationTap});

  final VoidCallback? onModerationTap;

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
              if (onModerationTap == null)
                const SizedBox(width: 48)
              else
                IconButton(
                  tooltip: 'Cola de moderación',
                  onPressed: onModerationTap,
                  icon: const Icon(Icons.shield_outlined, color: Colors.white),
                ),
              const NotificationsMenu(),
            ],
          ),
        ],
      ),
    );
  }
}
