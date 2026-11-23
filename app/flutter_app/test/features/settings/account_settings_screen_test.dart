import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/features/settings/presentation/account_settings_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/auth_session.dart';
import 'package:flutter_app/services/user_permissions.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets('saves a privacy change and confirms it visibly', (tester) async {
    final session = AuthSession(persistent: false);
    await session.establish(
      sessionToken: 'settings-token',
      expiresInSeconds: 3600,
      sessionUser: const AuthUser(
        id: 'user-1',
        username: 'Ana',
        email: 'ana@example.test',
        role: UserRole.user,
      ),
    );
    var privacyUpdates = 0;
    final client = ApiClient(
      session: session,
      httpClient: MockClient((request) async {
        expect(request.headers['authorization'], 'Bearer settings-token');
        if (request.method == 'GET' &&
            request.url.path == '/users/me/privacy') {
          return _response(_privacySettings('PUBLICO'));
        }
        if (request.method == 'PATCH' &&
            request.url.path == '/users/me/privacy') {
          privacyUpdates++;
          expect(jsonDecode(request.body), {'visibilidad_perfil': 'PRIVADO'});
          return _response(_privacySettings('PRIVADO'));
        }
        fail('Solicitud inesperada: ${request.method} ${request.url}');
      }),
    );
    addTearDown(() async {
      client.close();
      await session.clear();
    });

    await tester.pumpWidget(
      MaterialApp(home: AccountSettingsScreen(apiClient: client)),
    );
    await tester.pumpAndSettle();

    final profileVisibility = find.byType(DropdownButton<String>).first;
    await tester.ensureVisible(profileVisibility);
    await tester.pumpAndSettle();
    await tester.tap(profileVisibility);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Nadie').last);
    await tester.pumpAndSettle();

    expect(privacyUpdates, 1);
    expect(find.text('Privacidad actualizada.'), findsOneWidget);
  });
}

Map<String, Object?> _privacySettings(String profileVisibility) => {
  'success': true,
  'data': {
    'visibilidad_perfil': profileVisibility,
    'visibilidad_publicaciones': 'PUBLICO',
    'visibilidad_interacciones': 'AMIGOS',
    'recepcion_mensajes': 'TODOS',
    'recepcion_solicitudes_amistad': 'AMIGOS',
  },
  'error': null,
};

http.Response _response(Map<String, Object?> body) => http.Response(
  jsonEncode(body),
  200,
  headers: {'content-type': 'application/json; charset=utf-8'},
);
