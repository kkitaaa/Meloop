import 'package:flutter/material.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter_app/features/notifications/presentation/notification_scope.dart';
import 'package:flutter_app/features/notifications/state/app_notification.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_app/widgets/notifications_menu.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('Android notifications can be dismissed by swiping either way', (
    tester,
  ) async {
    final controller = _createController();
    addTearDown(controller.dispose);
    await _showMenu(tester, controller, TargetPlatform.android);

    await tester.tap(find.byTooltip('Notificaciones'));
    await tester.pumpAndSettle();
    expect(find.text('Aviso de prueba'), findsOneWidget);

    await tester.drag(find.byType(Dismissible).first, const Offset(-500, 0));
    await tester.pumpAndSettle();
    expect(find.text('Segundo aviso'), findsOneWidget);

    await tester.drag(find.byType(Dismissible).first, const Offset(500, 0));
    await tester.pumpAndSettle();
    expect(find.text('No hay notificaciones para mostrar.'), findsOneWidget);
    controller.stop();
  });

  testWidgets('desktop dismiss button appears on hover only', (tester) async {
    final controller = _createController();
    addTearDown(controller.dispose);
    await _showMenu(tester, controller, TargetPlatform.windows);

    await tester.tap(find.byTooltip('Notificaciones'));
    await tester.pumpAndSettle();
    expect(find.text('Aviso de prueba'), findsOneWidget);

    final removeButton = find.byTooltip('Eliminar notificación');
    expect(removeButton, findsNWidgets(2));
    final firstDismissButton = find.byKey(
      const ValueKey('notification-dismiss-notice-1'),
    );
    expect(tester.widget<AnimatedOpacity>(firstDismissButton).opacity, 0);

    final mouse = await tester.createGesture(kind: PointerDeviceKind.mouse);
    addTearDown(mouse.removePointer);
    final noticeCenter = tester.getCenter(find.text('Aviso de prueba'));
    await mouse.addPointer(location: noticeCenter);
    await mouse.moveTo(noticeCenter);
    await tester.pumpAndSettle();
    expect(tester.widget<AnimatedOpacity>(firstDismissButton).opacity, 1);

    await tester.tap(
      find.descendant(of: firstDismissButton, matching: removeButton),
    );
    await tester.pumpAndSettle();
    expect(find.text('Segundo aviso'), findsOneWidget);
    controller.stop();
  });
}

NotificationController _createController() => NotificationController(
  repository: _NotificationRepository(),
  playSound: false,
);

Future<void> _showMenu(
  WidgetTester tester,
  NotificationController controller,
  TargetPlatform platform,
) async {
  await tester.pumpWidget(
    NotificationScope(
      controller: controller,
      child: MaterialApp(
        theme: ThemeData(platform: platform),
        home: const Scaffold(body: NotificationsMenu()),
      ),
    ),
  );
  controller.start();
  tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
  await tester.pump();
  await tester.pumpAndSettle();
}

class _NotificationRepository implements NotificationRepository {
  @override
  Future<List<AppNotification>> fetchNotifications({DateTime? since}) async => [
    AppNotification(
      id: 'notice-1',
      kind: NotificationKind.like,
      title: 'Aviso de prueba',
      detail: 'Hace un momento',
      createdAt: DateTime(2026, 1, 3),
    ),
    AppNotification(
      id: 'notice-2',
      kind: NotificationKind.comment,
      title: 'Segundo aviso',
      detail: 'Hace dos minutos',
      createdAt: DateTime(2026, 1, 2),
    ),
  ];

  @override
  Future<void> updatePreference({
    required String type,
    required bool enabled,
  }) async {}

  @override
  void dispose() {}
}
