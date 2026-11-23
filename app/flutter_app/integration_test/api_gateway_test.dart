import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/screens/api_test_screen.dart';
import 'package:flutter_app/screens/feed_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/auth_session.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';
import 'package:http/http.dart' as http;

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('GET and POST communicate with the running API Gateway', (
    tester,
  ) async {
    await tester.pumpWidget(const MaterialApp(home: ApiTestScreen()));
    await tester.pumpAndSettle();

    expect(
      find.text('Conexión con el API Gateway funcionando.'),
      findsOneWidget,
    );

    await tester.tap(find.text('Enviar POST'));
    await tester.pumpAndSettle();

    expect(
      find.text('Respuesta del servidor: Hola desde Flutter'),
      findsOneWidget,
    );
    expect(find.textContaining('No se pudo conectar'), findsNothing);
  });

  testWidgets('feed loads posts from the running API Gateway', (tester) async {
    await tester.pumpWidget(const MaterialApp(home: FeedScreen()));
    await tester.pumpAndSettle(const Duration(seconds: 1));

    expect(
      find.textContaining('Contenido de demostración servido'),
      findsWidgets,
    );
    expect(find.text('Gateway no disponible.'), findsNothing);
  });

  testWidgets('recommendations reach the Kubernetes services', (tester) async {
    final response = await http
        .get(
          Uri.parse(
            '${ApiConfig.baseUrl}${ApiConfig.recommendationsEndpoint}/1/music',
          ),
        )
        .timeout(const Duration(seconds: 10));
    expect(response.statusCode, 200);

    final payload = jsonDecode(response.body) as Map<String, dynamic>;
    expect(payload['user_id'], 1);
    expect(payload['recommendations'], isA<List<dynamic>>());
    expect(payload['model'], isNot('fallback'));
  });

  testWidgets(
    'auth service registers, logs in, renews, validates, and logs out',
    (tester) async {
      final apiUri = Uri.parse(ApiConfig.baseUrl);
      if (ApiConfig.environment != Environment.development ||
          apiUri.host != 'localhost') {
        throw StateError(
          'El test de autenticación solo puede ejecutarse contra localhost en development.',
        );
      }

      final stamp = DateTime.now().toUtc().microsecondsSinceEpoch;
      final username = 'flutter_it_$stamp';
      final email = '$username@example.com';
      const password = 'Integration-Test-Password-42!';
      final session = AuthSession.shared;
      final apiClient = ApiClient(session: session);
      addTearDown(() async {
        await session.clear();
        apiClient.close();
      });

      await session.clear();
      await apiClient.register(
        username: username,
        email: email,
        password: password,
      );
      final user = await apiClient.login(email: email, password: password);
      expect(user.email, email);

      await session.establish(
        sessionToken: session.token!,
        expiresInSeconds: 30,
        sessionUser: user,
      );
      final validatedUser = await apiClient.validateSession();
      expect(validatedUser.id, user.id);
      expect(session.isAuthenticated, isTrue);

      await apiClient.logout();
      expect(session.isAuthenticated, isFalse);
    },
    skip: Platform.environment['MELOOP_RUN_AUTH_TESTS'] != '1',
  );
}
