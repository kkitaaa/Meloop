import 'package:flutter/material.dart';
import 'package:flutter_app/features/moderation/data/moderation_models.dart';
import 'package:flutter_app/features/moderation/data/moderation_repository.dart';
import 'package:flutter_app/features/moderation/presentation/moderation_queue_screen.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('denies access to regular users', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: ModerationQueueScreen(role: ModerationRole.user)),
    );
    expect(
      find.text('Esta sección está disponible solo para moderadores.'),
      findsOneWidget,
    );
  });

  testWidgets('moderators see urgent reports and own content is disabled', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(800, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      MaterialApp(
        home: ModerationQueueScreen(
          role: ModerationRole.moderator,
          repository: DemoModerationRepository(),
        ),
      ),
    );
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('Vencido'), findsOneWidget);
    expect(find.text('Vence en 5 h'), findsOneWidget);
    expect(find.text('Pendientes (3)'), findsOneWidget);
    await tester.scrollUntilVisible(
      find.text('Contenido propio'),
      300,
      scrollable: find.byType(Scrollable).last,
    );
    expect(find.text('Contenido propio'), findsOneWidget);
  });

  testWidgets('admins can filter escalated reports', (tester) async {
    tester.view.physicalSize = const Size(800, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      MaterialApp(
        home: ModerationQueueScreen(
          role: ModerationRole.admin,
          repository: DemoModerationRepository(),
        ),
      ),
    );
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    await tester.ensureVisible(find.text('Escalados (1)'));
    await tester.tap(find.text('Escalados (1)'));
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(
      find.text('Comentario con información que requiere verificación.'),
      findsOneWidget,
    );
  });
}
