import 'package:flutter/material.dart';

import '../state/app_notification.dart';
import '../state/notification_controller.dart';

class NotificationFeedbackOverlay extends StatelessWidget {
  const NotificationFeedbackOverlay({
    required this.controller,
    required this.child,
    super.key,
  });

  final NotificationController controller;
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: controller,
      child: child,
      builder: (context, child) {
        final notification = controller.activeBanner;
        final top = MediaQuery.viewPaddingOf(context).top + 56;
        return Stack(
          children: [
            ?child,
            if (notification != null)
              Positioned(
                top: top,
                left: 16,
                right: 16,
                child: Center(
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 520),
                    child: _NotificationBanner(
                      key: ValueKey(notification.id),
                      notification: notification,
                      onDismiss: controller.dismissBanner,
                    ),
                  ),
                ),
              ),
          ],
        );
      },
    );
  }
}

class _NotificationBanner extends StatelessWidget {
  const _NotificationBanner({
    required this.notification,
    required this.onDismiss,
    super.key,
  });

  final AppNotification notification;
  final VoidCallback onDismiss;

  @override
  Widget build(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: Dismissible(
        key: ValueKey(notification.id),
        direction: DismissDirection.up,
        onDismissed: (_) => onDismiss(),
        child: Card(
          elevation: 12,
          margin: EdgeInsets.zero,
          color: const Color(0xFFFAFAFA),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 13),
            child: Row(
              children: [
                CircleAvatar(
                  backgroundColor: const Color(0xFF00AF92)
                      .withValues(alpha: .14),
                  child: Icon(
                    switch (notification.kind) {
                      NotificationKind.like => Icons.favorite,
                      NotificationKind.comment => Icons.mode_comment,
                      NotificationKind.friendRequest => Icons.person_add_alt_1,
                    },
                    color: const Color(0xFF00AF92),
                    size: 20,
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        notification.title,
                        style: const TextStyle(fontWeight: FontWeight.w600),
                      ),
                      if (notification.detail.isNotEmpty) ...[
                        const SizedBox(height: 3),
                        Text(
                          notification.detail,
                          style: const TextStyle(
                            color: Colors.black54,
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
