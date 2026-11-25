import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/auth_session.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/core/navigation/responsive_navigation_shell.dart';
import 'package:flutter_app/features/search/presentation/global_search_screen.dart';
import 'package:flutter_app/features/notifications/presentation/notification_scope.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_app/features/settings/presentation/account_settings_screen.dart';
import 'package:flutter_app/screens/auth_screen.dart';
import 'package:flutter_app/widgets/mobile_app_header.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets(
    'Android screens share the header and profile remains in navbar',
    (tester) async {
      debugDefaultTargetPlatformOverride = TargetPlatform.android;
      addTearDown(() => debugDefaultTargetPlatformOverride = null);

      tester.view.physicalSize = const Size(599, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final controller = NotificationController(
        repository: MockNotificationRepository(),
        playSound: false,
      )..start();
      addTearDown(controller.dispose);
      final session = AuthSession(persistent: false);
      await session.establish(
        sessionToken: 'test-session-token',
        expiresInSeconds: 3600,
        sessionUser: const AuthUser(
          id: 'user-1',
          username: 'test-user',
          email: 'test@example.com',
        ),
      );
      var logoutRequested = false;
      final apiClient = ApiClient(
        httpClient: MockClient((request) async {
          if (request.url.path.endsWith('/auth/logout')) {
            logoutRequested = true;
            return http.Response(
              '{"success":true,"data":null,"error":null}',
              200,
            );
          }
          return http.Response('{}', 500);
        }),
        session: session,
      );
      addTearDown(apiClient.close);

      await tester.pumpWidget(
        NotificationScope(
          controller: controller,
          child: MaterialApp(
            home: ResponsiveNavigationShell(apiClient: apiClient),
          ),
        ),
      );
      await tester.pump(const Duration(seconds: 2));
      await tester.pump();

      for (final section in [
        'Inicio',
        'Amigos',
        'Artistas y canciones',
        'Perfil',
      ]) {
        await tester.tap(find.bySemanticsLabel(section));
        await tester.pump();

        expect(find.byType(MobileAppHeader), findsOneWidget);
        expect(find.byTooltip('Ajustes'), findsOneWidget);
        expect(
          find.descendant(
            of: find.byType(MobileAppHeader),
            matching: find.byType(CircleAvatar),
          ),
          findsNothing,
        );
      }

      expect(find.text('Editar perfil'), findsOneWidget);
      await tester.tap(find.bySemanticsLabel('Inicio'));
      await tester.pump();
      await tester.tap(find.byTooltip('Búsqueda global'));
      await tester.pumpAndSettle();
      expect(find.byType(GlobalSearchScreen), findsOneWidget);
      await tester.tap(find.byTooltip('Volver'));
      await tester.pumpAndSettle();
      expect(find.bySemanticsLabel('Inicio'), findsOneWidget);

      await tester.tap(find.byTooltip('Ajustes'));
      await tester.pumpAndSettle();
      expect(find.text('Configuración de cuenta'), findsOneWidget);
      await tester.tap(find.text('Configuración de cuenta'));
      await tester.pumpAndSettle();
      expect(find.byType(AccountSettingsScreen), findsOneWidget);
      await tester.pageBack();
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Ajustes'));
      await tester.pumpAndSettle();
      expect(find.text('Cerrar sesión'), findsOneWidget);
      await tester.tap(find.text('Cerrar sesión'));
      await tester.pumpAndSettle();
      expect(find.byType(AuthScreen), findsOneWidget);
      expect(logoutRequested, isTrue);
      expect(session.token, isNull);
      debugDefaultTargetPlatformOverride = null;
    },
  );
}
