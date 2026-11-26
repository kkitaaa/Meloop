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
    final closeButton = find.byKey(const ValueKey('notification-banner-close'));
    expect(closeButton, findsOneWidget);
    await tester.tap(closeButton);
    await tester.pump();
    expect(find.byType(Card), findsNothing);
    controller.dispose();
  });

  testWidgets(
    'notification banner stacks 1, 2, or up to 3 photo paper sheets based on unread count',
    (tester) async {
      for (final unreadCount in [1, 2, 3, 5]) {
        final controller = NotificationController(
          repository: _MultiNotificationRepository(unreadCount),
          playSound: false,
        );
        await tester.pumpWidget(
          MaterialApp(
            home: const Scaffold(body: SizedBox.expand()),
            builder: (context, child) => NotificationFeedbackOverlay(
              controller: controller,
              child: child!,
            ),
          ),
        );
        controller.start();
        tester.binding.handleAppLifecycleStateChanged(
          AppLifecycleState.resumed,
        );
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 400));

        final expectedSheets = unreadCount.clamp(1, 3);
        expect(
          find.byKey(
            ValueKey<String>('notification-paper-sheets-$expectedSheets'),
          ),
          findsOneWidget,
        );
        controller.dispose();
      }
    },
  );

  testWidgets(
    'notification banner stays visible for 6s and plays reverse slide-up on close',
    (tester) async {
      final controller = NotificationController(
        repository: _NotificationRepository(),
        playSound: false,
      );
      expect(controller.bannerDuration, const Duration(seconds: 6));

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
      await tester.pump(const Duration(milliseconds: 400));

      // A los 5.5s sigue visible (duración extendida a 6s)
      await tester.pump(const Duration(milliseconds: 5100));
      expect(find.byType(Card), findsOneWidget);

      // Al pulsar X inicia la animación de salida en reversa y desaparece al terminar
      await tester.tap(find.byKey(const ValueKey('notification-banner-close')));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));
      expect(
        find.byKey(const ValueKey<String>('notification-paper-sheets-1')),
        findsOneWidget,
      );
      await tester.pumpAndSettle();
      expect(
        find.byKey(const ValueKey<String>('notification-paper-sheets-1')),
        findsNothing,
      );
      controller.dispose();
    },
  );
}

class _MultiNotificationRepository implements NotificationRepository {
  _MultiNotificationRepository(this.count);

  final int count;

  @override
  Future<List<AppNotification>> fetchNotifications({DateTime? since}) async =>
      List<AppNotification>.generate(
        count,
        (index) => AppNotification(
          id: 'multi-notice-$index',
          kind: NotificationKind.like,
          title: 'Notificación $index',
          detail: 'Hace un momento',
          createdAt: DateTime(2026, 1, 1, 0, index),
        ),
      );

  @override
  Future<void> updatePreference({
    required String type,
    required bool enabled,
  }) async {}

  @override
  void dispose() {}
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
