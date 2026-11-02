import 'package:flutter/material.dart';
import 'package:flutter_app/features/admin/data/admin_models.dart';
import 'package:flutter_app/features/admin/data/admin_repository.dart';
import 'package:flutter_app/features/admin/presentation/admin_dashboard_screen.dart';
import 'package:flutter_app/features/admin/presentation/admin_moderators_screen.dart';
import 'package:flutter_app/features/admin/presentation/admin_rewards_screen.dart';
import 'package:flutter_app/features/admin/presentation/admin_users_screen.dart';
import 'package:flutter_app/features/moderation/data/moderation_models.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_app/main.dart' as app;
import 'package:flutter_app/screens/auth_screen.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('notification banners only start after a regular user login', (
    tester,
  ) async {
    await tester.pumpWidget(
      app.MeloopApp(
        key: const ValueKey('admin-login-app'),
        notificationRepository: MockNotificationRepository(
          interval: Duration.zero,
        ),
      ),
    );
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.text('A Camila le gustó tu publicación'), findsNothing);

    final inputs = find.byType(TextField);
    await tester.enterText(inputs.at(0), 'usuario@example.com');
    await tester.enterText(inputs.at(1), 'password');
    await tester.tap(find.widgetWithText(ElevatedButton, 'Iniciar Sesion'));
    await tester.pumpAndSettle();
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.text('A Camila le gustó tu publicación'), findsOneWidget);

    await tester.pumpWidget(
      app.MeloopApp(
        key: const ValueKey('new-admin-login-app'),
        notificationRepository: MockNotificationRepository(
          interval: Duration.zero,
        ),
      ),
    );
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pumpAndSettle();
    final adminInputs = find.byType(TextField);
    await tester.enterText(adminInputs.at(0), 'admin@admin.com');
    await tester.enterText(adminInputs.at(1), 'ADMIN');
    await tester.tap(find.widgetWithText(ElevatedButton, 'Iniciar Sesion'));
    await tester.pumpAndSettle();
    await tester.pump(const Duration(milliseconds: 100));
    expect(find.byType(AdminDashboardScreen), findsOneWidget);
    expect(find.text('A Camila le gustó tu publicación'), findsNothing);
  });

  testWidgets('admin demo login opens the admin dashboard', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: AuthScreen()));
    await tester.pumpAndSettle();

    final inputs = find.byType(TextField);
    await tester.enterText(inputs.at(0), 'admin@admin.com');
    await tester.enterText(inputs.at(1), 'ADMIN');
    await tester.tap(find.widgetWithText(ElevatedButton, 'Iniciar Sesion'));
    await tester.pumpAndSettle();

    expect(find.byType(AdminDashboardScreen), findsOneWidget);
  });

  testWidgets('admin tools reject non-admin roles', (tester) async {
    await tester.pumpWidget(
      const MaterialApp(home: AdminDashboardScreen(role: ModerationRole.user)),
    );

    expect(
      find.text('Esta sección está disponible solo para administradores.'),
      findsOneWidget,
    );
    expect(find.byType(AdminUsersScreen), findsNothing);
  });

  testWidgets('mobile admin logout action returns to login', (tester) async {
    tester.view.physicalSize = const Size(500, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    await tester.pumpWidget(
      const MaterialApp(home: AdminDashboardScreen(role: ModerationRole.admin)),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.byTooltip('Cerrar sesión'));
    await tester.pumpAndSettle();

    expect(find.byType(AuthScreen), findsOneWidget);
  });

  testWidgets('admin dashboard opens all three panels', (tester) async {
    tester.view.physicalSize = const Size(500, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final repository = DemoAdminRepository();
    await tester.pumpWidget(
      MaterialApp(
        home: AdminDashboardScreen(
          role: ModerationRole.admin,
          repository: repository,
        ),
      ),
    );
    await tester.pumpAndSettle();
    expect(find.byType(AdminUsersScreen), findsOneWidget);
    expect(
      find.byWidgetPredicate(
        (widget) =>
            widget is Image &&
            widget.image is AssetImage &&
            (widget.image as AssetImage).assetName ==
                'assets/imagenes/FondoAdmin.png',
      ),
      findsOneWidget,
    );

    await tester.tap(find.text('Moderadores'));
    await tester.pumpAndSettle();
    expect(find.byType(AdminModeratorsScreen), findsOneWidget);
    await tester.enterText(find.byType(TextField), 'vicente');
    await tester.pumpAndSettle();
    await tester.tap(find.text('@vicente'));
    await tester.pump();
    await tester.tap(find.widgetWithText(FilledButton, 'Asignar rol'));
    await tester.pumpAndSettle();
    expect(
      find.text('¿Quieres asignar el rol de moderador a @vicente?'),
      findsOneWidget,
    );
    await tester.tap(find.widgetWithText(FilledButton, 'Asignar'));
    await tester.pumpAndSettle();
    expect(
      find.text(
        'Rol actualizado. La persona deberá volver a iniciar sesión para que el cambio se aplique.',
      ),
      findsOneWidget,
    );
    expect(find.text('@vicente'), findsOneWidget);

    final vicenteTile = find.ancestor(
      of: find.text('@vicente'),
      matching: find.byType(ListTile),
    );
    await tester.tap(
      find.descendant(
        of: vicenteTile.first,
        matching: find.byTooltip('Retirar rol'),
      ),
    );
    await tester.pumpAndSettle();
    expect(
      find.textContaining('Sus reportes en revisión volverán a la bandeja.'),
      findsOneWidget,
    );
    await tester.tap(find.widgetWithText(FilledButton, 'Retirar rol'));
    await tester.pumpAndSettle();
    expect(find.text('@vicente'), findsNothing);

    await tester.tap(find.text('Recompensas'));
    await tester.pumpAndSettle();
    expect(find.byType(AdminRewardsScreen), findsOneWidget);
    for (final rewardName in [
      'Marco Neón',
      'Marco Dorado',
      'Pin Zafiro',
      'Modo Nocturno',
      'Marco RGB',
    ]) {
      expect(find.text(rewardName), findsOneWidget);
    }
    await tester.tap(find.widgetWithText(TextButton, 'Deshabilitar').first);
    await tester.pumpAndSettle();
    expect(
      find.textContaining(
        'Quienes ya la tienen la conservan; deja de poder desbloquearse para el resto.',
      ),
      findsOneWidget,
    );
    await tester.tap(find.widgetWithText(FilledButton, 'Deshabilitar'));
    await tester.pumpAndSettle();
    expect(find.textContaining('Deshabilitada'), findsOneWidget);

    await tester.tap(find.text('Usuarios'));
    await tester.pumpAndSettle();
    expect(find.byType(AdminUsersScreen), findsOneWidget);
    expect(find.textContaining('@'), findsWidgets);
    expect(find.textContaining('@admin.com'), findsNothing);
    expect(find.text('Suspender'), findsNothing);
  });

  testWidgets(
    'desktop user actions stay in one row and filters use white fill',
    (tester) async {
      tester.view.physicalSize = const Size(1280, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(
        MaterialApp(
          home: AdminDashboardScreen(
            role: ModerationRole.admin,
            repository: DemoAdminRepository(),
          ),
        ),
      );
      await tester.pumpAndSettle();

      final filters = find.byType(DropdownButtonFormField<AdminUserStatus?>);
      expect(filters, findsOneWidget);
      final inputTheme = InputDecorationTheme.of(tester.element(filters));
      expect(inputTheme.fillColor, Colors.white);

      final horizontalTable = find.byWidgetPredicate(
        (widget) =>
            widget is SingleChildScrollView &&
            widget.scrollDirection == Axis.horizontal,
      );
      await tester.drag(horizontalTable, const Offset(-1200, 0));
      await tester.pumpAndSettle();

      final suspendButton = find.byTooltip('Suspender').first;
      final actionsRow = find
          .ancestor(of: suspendButton, matching: find.byType(Row))
          .first;
      final historyButton = find.descendant(
        of: actionsRow,
        matching: find.byTooltip('Ver historial'),
      );
      expect(historyButton, findsOneWidget);
      expect(
        tester.getCenter(historyButton).dy,
        tester.getCenter(suspendButton).dy,
      );
      expect(
        tester.getCenter(suspendButton).dx,
        greaterThan(tester.getCenter(historyButton).dx),
      );

      await tester.tap(suspendButton);
      await tester.pumpAndSettle();
      await tester.enterText(
        find.widgetWithText(TextField, 'Motivo (5 a 300 caracteres)'),
        'Incumplimiento de las normas',
      );
      await tester.tap(find.widgetWithText(FilledButton, 'Suspender'));
      await tester.pumpAndSettle();

      final reactivationButton = find.byTooltip('Reactivar').first;
      final updatedActionsRow = find
          .ancestor(of: reactivationButton, matching: find.byType(Row))
          .first;
      final updatedHistoryButton = find.descendant(
        of: updatedActionsRow,
        matching: find.byTooltip('Ver historial'),
      );
      expect(updatedHistoryButton, findsOneWidget);
      expect(
        tester.getCenter(updatedHistoryButton).dy,
        tester.getCenter(reactivationButton).dy,
      );
      expect(
        tester.getCenter(reactivationButton).dx,
        greaterThan(tester.getCenter(updatedHistoryButton).dx),
      );

      await tester.pump(const Duration(seconds: 5));
      await tester.tap(find.byKey(const ValueKey('admin-logout')));
      await tester.pumpAndSettle();
      expect(find.byType(AuthScreen), findsOneWidget);
    },
  );
}
