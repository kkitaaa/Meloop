import 'package:flutter/material.dart';
import 'package:flutter_app/features/notifications/presentation/notification_feedback_overlay.dart';
import 'package:flutter_app/features/notifications/state/app_notification.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('desktop notification banner appears below the taller header', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1280, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final controller = NotificationController(
      repository: _NotificationRepository(),
      playSound: false,
    );

    await tester.pumpWidget(
      MaterialApp(
        home: const Scaffold(body: SizedBox.expand()),
        builder: (context, child) =>
            NotificationFeedbackOverlay(controller: controller, child: child!),
      ),
    );
    controller.start();
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump();

    final banner = find.byType(Card);
    expect(banner, findsOneWidget);
    expect(tester.getTopLeft(banner).dy, 98);
    controller.dispose();
  });
}

class _NotificationRepository implements NotificationRepository {
  @override
  Future<List<AppNotification>> fetchNotifications({DateTime? since}) async {
    return [
      AppNotification(
        id: 'notice-1',
        kind: NotificationKind.like,
        title: 'Alguien reaccionó a tu publicación',
        detail: 'Hace un momento',
        createdAt: DateTime(2026),
      ),
    ];
  }

  @override
  Future<void> updatePreference({
    required String type,
    required bool enabled,
  }) async {}

  @override
  void dispose() {}
}
