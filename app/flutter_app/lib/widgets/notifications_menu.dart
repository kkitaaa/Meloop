import 'dart:async';

import 'package:flutter/material.dart';

import '../features/notifications/presentation/notification_scope.dart';
import '../features/notifications/presentation/grouped_notification.dart';
import '../features/notifications/state/app_notification.dart';
import '../features/notifications/state/notification_controller.dart';
import '../features/profile/presentation/profile_screen.dart';
import '../features/posts/post_detail_screen.dart';
import '../screens/friends_screen.dart';

class NotificationsMenu extends StatefulWidget {
  const NotificationsMenu({super.key});

  @override
  State<NotificationsMenu> createState() => _NotificationsMenuState();
}

class _NotificationsMenuState extends State<NotificationsMenu> {
  final MenuController _menuController = MenuController();
  bool _showSettings = false;

  @override
  Widget build(BuildContext context) {
    final notifications = NotificationScope.of(context);
    final panelWidth = (MediaQuery.sizeOf(context).width - 24).clamp(
      0.0,
      365.0,
    );

    return AnimatedBuilder(
      animation: notifications,
      builder: (context, _) => MenuAnchor(
        controller: _menuController,
        style: const MenuStyle(
          padding: WidgetStatePropertyAll(EdgeInsets.zero),
          elevation: WidgetStatePropertyAll(12),
        ),
        menuChildren: [
          SizedBox(
            width: panelWidth,
            child: _buildPanel(context, notifications),
          ),
        ],
        builder: (context, controller, child) => Stack(
          clipBehavior: Clip.none,
          alignment: Alignment.center,
          children: [
            IconButton(
              tooltip: 'Notificaciones',
              icon: _NotificationPulseIcon(
                pulse: notifications.notificationPulse,
              ),
              onPressed: () =>
                  controller.isOpen ? controller.close() : controller.open(),
            ),
            if (notifications.unreadCount > 0)
              Positioned(
                top: 6,
                right: 6,
                child: Container(
                  width: 14,
                  height: 14,
                  alignment: Alignment.center,
                  decoration: const BoxDecoration(
                    color: Color(0xFFFF6262),
                    shape: BoxShape.circle,
                  ),
                  child: Text(
                    '${notifications.unreadCount}',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 8,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildPanel(
    BuildContext context,
    NotificationController notifications,
  ) {
    final items = notifications.notifications
        .where(
          (item) => notifications.preference(
            AppNotification.preferenceKey(item.kind),
          ),
        )
        .toList();

    return SizedBox(
      height: _showSettings ? 390 : 360,
      child: Material(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 12, 8, 8),
              child: Row(
                children: [
                  Text(
                    _showSettings ? 'Preferencias' : 'Notificaciones',
                    style: const TextStyle(
                      fontSize: 17,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const Spacer(),
                  IconButton(
                    tooltip: _showSettings
                        ? 'Ver notificaciones'
                        : 'Configurar notificaciones',
                    icon: Icon(
                      _showSettings ? Icons.arrow_back : Icons.tune,
                      color: const Color(0xFF1E7E70),
                    ),
                    onPressed: () =>
                        setState(() => _showSettings = !_showSettings),
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            if (_showSettings)
              Expanded(child: _settingsPanel(notifications))
            else
              Expanded(
                child: _notificationsList(context, items, notifications),
              ),
          ],
        ),
      ),
    );
  }

  Widget _settingsPanel(NotificationController notifications) => Padding(
    padding: const EdgeInsets.all(10),
    child: SingleChildScrollView(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Padding(
            padding: EdgeInsets.all(6),
            child: Text(
              'Elige qué alertas quieres recibir.',
              style: TextStyle(color: Colors.black54, fontSize: 12),
            ),
          ),
          _preferenceSwitch(
            notifications,
            'like',
            'Likes en publicaciones',
            Icons.favorite,
          ),
          _preferenceSwitch(
            notifications,
            'comment',
            'Comentarios',
            Icons.mode_comment_outlined,
          ),
          _preferenceSwitch(
            notifications,
            'friend',
            'Solicitudes de amistad',
            Icons.person_add_alt_1,
          ),
          _preferenceSwitch(
            notifications,
            'level',
            'Subidas de nivel',
            Icons.trending_up,
          ),
          _preferenceSwitch(
            notifications,
            'reward',
            'Recompensas desbloqueadas',
            Icons.card_giftcard,
          ),
        ],
      ),
    ),
  );

  Widget _preferenceSwitch(
    NotificationController notifications,
    String key,
    String label,
    IconData icon,
  ) => SwitchListTile(
    value: notifications.preference(key),
    activeThumbColor: const Color(0xFF00AF92),
    onChanged: (value) => unawaited(notifications.setPreference(key, value)),
    secondary: Icon(icon, color: const Color(0xFF1E7E70), size: 20),
    title: Text(label, style: const TextStyle(fontSize: 13)),
  );

  Widget _notificationsList(
    BuildContext context,
    List<AppNotification> items,
    NotificationController notifications,
  ) {
    final groups = GroupedNotification.group(items);
    if (items.isEmpty) {
      return const Padding(
        padding: EdgeInsets.all(32),
        child: Text('No hay notificaciones para mostrar.'),
      );
    }

    return ListView.separated(
      shrinkWrap: true,
      itemCount: groups.length,
      separatorBuilder: (_, _) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final group = groups[index];
        final item = group.latest;
        return InkWell(
          onTap: () => _openNotification(context, group, notifications),
          child: Container(
            color: group.isRead ? Colors.white : const Color(0xFFE9F8F5),
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
            child: Row(
              children: [
                CircleAvatar(
                  radius: 18,
                  backgroundColor: _iconColor(item.kind).withValues(alpha: .15),
                  child: Icon(
                    _icon(item.kind),
                    size: 18,
                    color: _iconColor(item.kind),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        group.title,
                        style: TextStyle(
                          fontSize: 13,
                          fontWeight: group.isRead
                              ? FontWeight.normal
                              : FontWeight.bold,
                        ),
                      ),
                      if (item.detail.isNotEmpty) ...[
                        const SizedBox(height: 3),
                        Text(
                          item.detail,
                          style: const TextStyle(
                            fontSize: 11,
                            color: Colors.black54,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
                if (group.count > 1)
                  Padding(
                    padding: const EdgeInsets.only(left: 6),
                    child: Text(
                      '${group.count}',
                      style: const TextStyle(
                        color: Color(0xFF1E7E70),
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                if (!item.isRead)
                  const Icon(Icons.circle, size: 8, color: Color(0xFF00AF92)),
              ],
            ),
          ),
        );
      },
    );
  }

  void _openNotification(
    BuildContext context,
    GroupedNotification group,
    NotificationController notifications,
  ) {
    final item = group.latest;
    for (final notification in group.notifications) {
      notifications.markRead(notification.id);
    }
    _menuController.close();
    if (item.postId != null) {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (_) => PostDetailScreen(postId: item.postId!),
        ),
      );
    } else if (item.kind == NotificationKind.friendRequest) {
      Navigator.push(
        context,
        MaterialPageRoute(builder: (_) => const FriendsScreen()),
      );
    } else if (item.kind == NotificationKind.levelUp) {
      Navigator.push(
        context,
        MaterialPageRoute(builder: (_) => const ProfileScreen()),
      );
    } else if (item.kind == NotificationKind.reward) {
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (_) => const ProfileScreen(openInventory: true),
        ),
      );
    }
  }

  IconData _icon(NotificationKind kind) => switch (kind) {
    NotificationKind.like => Icons.favorite,
    NotificationKind.comment => Icons.mode_comment,
    NotificationKind.friendRequest => Icons.person_add_alt_1,
    NotificationKind.levelUp => Icons.trending_up,
    NotificationKind.reward => Icons.card_giftcard,
  };

  Color _iconColor(NotificationKind kind) => switch (kind) {
    NotificationKind.like => const Color(0xFFFF6262),
    NotificationKind.comment => const Color(0xFF00AF92),
    NotificationKind.friendRequest => const Color(0xFF5B7FFF),
    NotificationKind.levelUp => const Color(0xFFFFB020),
    NotificationKind.reward => const Color(0xFF9C5BFF),
  };
}

class _NotificationPulseIcon extends StatefulWidget {
  const _NotificationPulseIcon({required this.pulse});

  final int pulse;

  @override
  State<_NotificationPulseIcon> createState() => _NotificationPulseIconState();
}

class _NotificationPulseIconState extends State<_NotificationPulseIcon>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller = AnimationController(
    vsync: this,
    duration: const Duration(seconds: 3),
  );

  @override
  void didUpdateWidget(covariant _NotificationPulseIcon oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.pulse != oldWidget.pulse) _controller.forward(from: 0);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => SizedBox(
    width: 30,
    height: 30,
    child: AnimatedBuilder(
      animation: _controller,
      builder: (context, _) => Stack(
        alignment: Alignment.center,
        children: [
          if (_controller.isAnimating)
            CustomPaint(
              size: const Size.square(30),
              painter: _RadarPainter(_controller.value),
            ),
          const Icon(Icons.notifications, color: Colors.white, size: 20),
        ],
      ),
    ),
  );
}

class _RadarPainter extends CustomPainter {
  const _RadarPainter(this.progress);

  final double progress;

  @override
  void paint(Canvas canvas, Size size) {
    for (var wave = 0; wave < 3; wave++) {
      final phase = progress * 3 - wave;
      if (phase < 0 || phase > 1) continue;
      final paint = Paint()
        ..color = const Color(0xFF8FFFE9).withValues(alpha: 1 - phase)
        ..style = PaintingStyle.stroke
        ..strokeWidth = 1.5;
      canvas.drawCircle(size.center(Offset.zero), 8 + phase * 10, paint);
    }
  }

  @override
  bool shouldRepaint(covariant _RadarPainter oldDelegate) =>
      progress != oldDelegate.progress;
}
