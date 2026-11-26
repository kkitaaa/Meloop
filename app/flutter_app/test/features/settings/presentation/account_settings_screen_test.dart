import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/features/settings/presentation/account_settings_screen.dart';
import 'package:flutter_app/features/settings/presentation/notification_settings_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/auth_session.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets('account settings save supported fields and flag pending APIs', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(390, 1800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    var profileVisibility = 'PUBLICO';
    var passwordChanged = false;
    final apiClient = ApiClient(
      session: AuthSession(persistent: false),
      httpClient: MockClient((request) async {
        if (request.method == 'GET' && request.url.path == '/users/me') {
          return _jsonResponse({
            'id_usuario': 'user-1',
            'username': 'test-user',
            'correo': 'test@example.com',
          });
        }
        if (request.method == 'GET' &&
            request.url.path == '/users/me/privacy') {
          return _jsonResponse(_privacy(profileVisibility));
        }
        if (request.method == 'PATCH' &&
            request.url.path == '/users/me/privacy') {
          final body = jsonDecode(request.body) as Map<String, dynamic>;
          profileVisibility = body['visibilidad_perfil'] as String;
          return _jsonResponse(_privacy(profileVisibility));
        }
        if (request.method == 'POST' &&
            request.url.path == '/auth/change-password') {
          expect(jsonDecode(request.body), {
            'current_password': 'old-password',
            'new_password': 'new-password',
          });
          passwordChanged = true;
          return _jsonResponse(null);
        }
        fail('Unexpected request: ${request.method} ${request.url}');
      }),
    );
    addTearDown(apiClient.close);

    await tester.pumpWidget(
      MaterialApp(home: AccountSettingsScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();

    expect(find.text('Configuración de cuenta'), findsOneWidget);
    expect(find.text('test@example.com'), findsOneWidget);
    expect(find.text('Mostrar estado en línea'), findsOneWidget);
    expect(
      tester
          .widget<OutlinedButton>(
            find.widgetWithText(
              OutlinedButton,
              'Cambiar correo (pendiente del backend)',
            ),
          )
          .onPressed,
      isNull,
    );

    final passwordFields = find.byType(TextField);
    await tester.enterText(passwordFields.at(0), 'old-password');
    await tester.enterText(passwordFields.at(1), 'new-password');
    await tester.enterText(passwordFields.at(2), 'new-password');
    await tester.tap(find.text('Actualizar contraseña'));
    await tester.pumpAndSettle();
    expect(passwordChanged, isTrue);
    expect(find.text('Contraseña actualizada correctamente.'), findsOneWidget);

    await tester.ensureVisible(find.text('Visibilidad del perfil'));
    final profileTile = find.ancestor(
      of: find.text('Visibilidad del perfil'),
      matching: find.byType(ListTile),
    );
    await tester.tap(
      find.descendant(
        of: profileTile,
        matching: find.byType(DropdownButton<String>),
      ),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Privado').last);
    await tester.pumpAndSettle();

    expect(profileVisibility, 'PRIVADO');
    expect(find.text('Configuración de privacidad guardada.'), findsOneWidget);
  });

  testWidgets('notification preferences load and persist a switch change', (
    tester,
  ) async {
    var likesEnabled = true;
    final apiClient = ApiClient(
      session: AuthSession(persistent: false),
      httpClient: MockClient((request) async {
        if (request.method == 'GET' &&
            request.url.path == '/users/me/notifications/settings') {
          return _jsonResponse([
            {'tipo_notificacion': 'LIKE', 'habilitada': likesEnabled},
          ]);
        }
        if (request.method == 'PATCH' &&
            request.url.path == '/users/me/notifications/settings/LIKE') {
          likesEnabled =
              (jsonDecode(request.body) as Map<String, dynamic>)['habilitada']
                  as bool;
          return _jsonResponse({
            'tipo_notificacion': 'LIKE',
            'habilitada': likesEnabled,
          });
        }
        fail('Unexpected request: ${request.method} ${request.url}');
      }),
    );
    addTearDown(apiClient.close);

    await tester.pumpWidget(
      MaterialApp(home: NotificationSettingsScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();
    expect(find.text('Me gusta'), findsOneWidget);

    await tester.tap(find.widgetWithText(SwitchListTile, 'Me gusta'));
    await tester.pumpAndSettle();

    expect(likesEnabled, isFalse);
    expect(
      tester
          .widget<SwitchListTile>(
            find.widgetWithText(SwitchListTile, 'Me gusta'),
          )
          .value,
      isFalse,
    );
    expect(find.text('Preferencia guardada.'), findsOneWidget);
  });
}

http.Response _jsonResponse(Object? data) => http.Response(
  jsonEncode({'success': true, 'data': data, 'error': null}),
  200,
  headers: {'content-type': 'application/json; charset=utf-8'},
);

Map<String, Object?> _privacy(String profileVisibility) => {
  'id_usuario': 'user-1',
  'visibilidad_perfil': profileVisibility,
  'visibilidad_publicaciones': 'AMIGOS',
  'visibilidad_interacciones': 'PRIVADO',
  'recepcion_mensajes': 'TODOS',
  'recepcion_solicitudes_amistad': 'AMIGOS',
};
