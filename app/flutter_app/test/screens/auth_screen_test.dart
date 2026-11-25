import 'dart:async';
import 'dart:convert';
import 'dart:ui';

import 'package:flutter/material.dart';
import 'package:flutter_app/features/notifications/presentation/notification_scope.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_app/screens/auth_screen.dart';
import 'package:flutter_app/screens/kubernetes_connectivity_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/parallax_album_covers.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets(
    'pressing Enter in password submits login and switch links toggle tabs',
    (tester) async {
      tester.view.physicalSize = const Size(1920, 1080);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      var loginRequests = 0;
      final client = ApiClient(
        httpClient: MockClient((request) async {
          if (request.url.path.endsWith('/login')) {
            loginRequests++;
            return http.Response(
              jsonEncode({
                'success': false,
                'data': null,
                'error': {
                  'code': 'INVALID_CREDENTIALS',
                  'message': 'Credenciales inválidas.',
                },
              }),
              401,
              headers: {'content-type': 'application/json; charset=utf-8'},
            );
          }
          return http.Response('{}', 200);
        }),
      );
      addTearDown(client.close);

      final notificationController = NotificationController(
        repository: MockNotificationRepository(),
        playSound: false,
      );
      addTearDown(notificationController.dispose);

      await tester.pumpWidget(
        NotificationScope(
          controller: notificationController,
          child: MaterialApp(home: AuthScreen(apiClient: client)),
        ),
      );

      // Enlaces de cambio entre Iniciar Sesión y Registrarse
      expect(find.text('Crea una aquí'), findsOneWidget);
      await tester.tap(find.text('Crea una aquí'));
      await tester.pumpAndSettle();
      expect(find.text('Crear cuenta'), findsOneWidget);
      expect(find.text('Inicia sesión aquí'), findsOneWidget);

      await tester.tap(find.text('Inicia sesión aquí'));
      await tester.pumpAndSettle();
      expect(find.text('Iniciar sesión'), findsOneWidget);

      // Presionar Enter en la caja de contraseña envía el formulario de login
      final fields = find.byType(TextField);
      await tester.enterText(fields.at(0), 'user@meloop.cl');
      await tester.enterText(fields.at(1), 'secret123');
      await tester.testTextInput.receiveAction(TextInputAction.done);
      await tester.pumpAndSettle();

      expect(loginRequests, 1);
      expect(find.text('Credenciales inválidas.'), findsOneWidget);
    },
  );

  testWidgets(
    'diagnostic button stays hidden until hovering bottom-right corner for 3 seconds',
    (tester) async {
      tester.view.physicalSize = const Size(1920, 1080);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      await tester.pumpWidget(const MaterialApp(home: AuthScreen()));

      final diagnosticFinder = find.text('Diagnóstico de conexión');
      expect(diagnosticFinder, findsOneWidget);

      // Inicialmente invisible (opacidad 0) y no clickeable
      AnimatedOpacity opacityWidget() => tester.widget<AnimatedOpacity>(
        find.ancestor(
          of: diagnosticFinder,
          matching: find.byType(AnimatedOpacity),
        ),
      );
      expect(opacityWidget().opacity, 0.0);

      final gesture = await tester.createGesture(kind: PointerDeviceKind.mouse);
      await gesture.addPointer(location: Offset.zero);
      addTearDown(gesture.removePointer);

      // Hover en la esquina inferior derecha durante 2 segundos: sigue oculto
      await gesture.moveTo(const Offset(1880, 1050));
      await tester.pump(const Duration(seconds: 2));
      expect(opacityWidget().opacity, 0.0);

      // Al completar 3 segundos de hover, aparece y es clickeable
      await tester.pump(const Duration(seconds: 1));
      await tester.pumpAndSettle();
      expect(opacityWidget().opacity, 1.0);

      await tester.tap(diagnosticFinder);
      await tester.pumpAndSettle();
      expect(find.byType(KubernetesConnectivityScreen), findsOneWidget);
    },
  );

  testWidgets(
    'parallax album covers render in copy-number order and drift upward',
    (tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: SizedBox(
            width: 1200,
            height: 900,
            child: ParallaxAlbumCovers(animate: true),
          ),
        ),
      );
      await tester.pump();

      final initialImages = tester
          .widgetList<Image>(find.byType(Image))
          .toList();
      expect(initialImages.length, 22);
      for (final img in initialImages) {
        expect(img.width, img.height);
        expect(img.width, lessThanOrEqualTo(360));
        expect(img.fit, BoxFit.cover);
      }

      final firstCover = find.byType(Image).first;
      final startY = tester.getTopLeft(firstCover).dy;
      await tester.pump(const Duration(milliseconds: 50));
      await tester.pump(const Duration(milliseconds: 50));
      final movedY = tester.getTopLeft(firstCover).dy;
      expect(movedY, lessThan(startY));
    },
  );

  testWidgets(
    'splash dismissal waits until login images finish preloading and rendering',
    (tester) async {
      tester.view.physicalSize = const Size(1920, 1080);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final preloadCompleter = Completer<void>();
      var splashDismissCount = 0;
      bool? preloadedDesktopFlag;

      await tester.pumpWidget(
        MaterialApp(
          home: AuthScreen(
            preloadImages: (context, isDesktop) {
              preloadedDesktopFlag = isDesktop;
              return preloadCompleter.future;
            },
            onReadyForSplashDismiss: () => splashDismissCount++,
          ),
        ),
      );

      // Mientras las imágenes aún están cargando, el splash NO debe ocultarse.
      await tester.pump();
      expect(preloadedDesktopFlag, isTrue);
      expect(splashDismissCount, 0);

      // Una vez terminada la precarga y pintado el siguiente cuadro, se oculta.
      preloadCompleter.complete();
      await tester.pump();
      await tester.pump();
      expect(splashDismissCount, 1);
    },
  );
}
