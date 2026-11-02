import 'package:flutter/material.dart';
import 'package:flutter_app/features/moderation/data/moderation_models.dart';
import 'package:flutter_app/features/moderation/data/moderation_repository.dart';
import 'package:flutter_app/features/moderation/presentation/report_sheet.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('clips the desktop report dialog to its rounded shape', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(1200, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Builder(
            builder: (context) => TextButton(
              onPressed: () => showReportSheet(
                context,
                targetType: ModerationTargetType.post,
                targetId: 'post-1',
              ),
              child: const Text('Abrir reporte'),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.text('Abrir reporte'));
    await tester.pumpAndSettle();

    expect(
      tester.widget<Dialog>(find.byType(Dialog)).clipBehavior,
      Clip.antiAlias,
    );
  });

  testWidgets('requires a reason and details for Other before submitting', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ReportSheet(
            targetType: ModerationTargetType.post,
            targetId: 'post-1',
            repository: DemoModerationRepository(),
          ),
        ),
      ),
    );

    final submitButton = tester.widget<FilledButton>(
      find.widgetWithText(FilledButton, 'Enviar reporte'),
    );
    expect(submitButton.onPressed, isNull);

    await tester.tap(find.text('Otro'));
    await tester.pump();
    expect(
      tester
          .widget<FilledButton>(
            find.widgetWithText(FilledButton, 'Enviar reporte'),
          )
          .onPressed,
      isNull,
    );

    await tester.enterText(find.byType(TextField), 'Detalles del reporte');
    await tester.pump();
    expect(
      tester
          .widget<FilledButton>(
            find.widgetWithText(FilledButton, 'Enviar reporte'),
          )
          .onPressed,
      isNotNull,
    );
  });

  testWidgets('shows the success message and the block shortcut', (
    tester,
  ) async {
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ReportSheet(
            targetType: ModerationTargetType.comment,
            targetId: 'comment-1',
            targetUsername: 'Camila',
            repository: DemoModerationRepository(),
          ),
        ),
      ),
    );

    await tester.tap(find.text('Spam'));
    await tester.pump();
    await tester.tap(find.text('Enviar reporte'));
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pump();

    expect(find.text('Gracias. Revisaremos tu reporte.'), findsOneWidget);
    expect(find.text('Bloquear también a @Camila'), findsOneWidget);
  });

  testWidgets('maps each simulated backend result to its expected message', (
    tester,
  ) async {
    for (final outcome in ModerationReportOutcome.values) {
      await tester.pumpWidget(
        MaterialApp(
          key: ValueKey(outcome),
          home: Scaffold(
            body: ReportSheet(
              targetType: ModerationTargetType.user,
              targetId: 'user-1',
              repository: DemoModerationRepository(reportOutcome: outcome),
            ),
          ),
        ),
      );
      await tester.tap(find.text('Spam'));
      await tester.pump();
      await tester.tap(find.text('Enviar reporte'));
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pump();
      expect(find.text(outcome.message), findsOneWidget);
    }
  });
}
