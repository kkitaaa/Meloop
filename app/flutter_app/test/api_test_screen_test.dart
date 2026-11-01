import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/screens/api_test_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets('loads GET data and sends a POST message', (tester) async {
    final client = ApiClient(
      httpClient: MockClient((request) async {
        if (request.method == 'GET') {
          return _jsonResponse(200, {
            'success': true,
            'data': {
              'message': 'Conexión con el API Gateway funcionando.',
              'timestamp': '2026-11-01T12:00:00Z',
            },
            'error': null,
          });
        }

        expect(request.method, 'POST');
        expect(jsonDecode(request.body), {'message': 'Hola desde Flutter'});
        return _jsonResponse(200, {
          'success': true,
          'data': {
            'receivedMessage': 'Hola desde Flutter',
            'receivedAt': '2026-11-01T12:00:00Z',
          },
          'error': null,
        });
      }),
    );

    await tester.pumpWidget(
      MaterialApp(home: ApiTestScreen(apiClient: client)),
    );
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
  });
}

http.Response _jsonResponse(int statusCode, Map<String, Object?> body) {
  return http.Response(
    jsonEncode(body),
    statusCode,
    headers: {'content-type': 'application/json; charset=utf-8'},
  );
}
