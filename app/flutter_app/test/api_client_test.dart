import 'dart:convert';

import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  group('ApiClient', () {
    test('decodes a successful GET response', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'GET');
          expect(request.url.path, '/api/demo');
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'message': 'Gateway disponible',
                'timestamp': '2026-11-01T12:00:00Z',
              },
              'error': null,
            },
          );
        }),
      );

      final response = await client.getDemo();

      expect(response.success, isTrue);
      expect(response.data?.message, 'Gateway disponible');
    });

    test('sends and decodes a successful POST response', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'POST');
          expect(request.url.path, '/api/demo');
          expect(jsonDecode(request.body), {'message': 'Hola'});
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'receivedMessage': 'Hola',
                'receivedAt': '2026-11-01T12:00:00Z',
              },
              'error': null,
            },
          );
        }),
      );

      final response = await client.postDemo('Hola');

      expect(response.data?.receivedMessage, 'Hola');
    });

    test('surfaces the backend message and HTTP status', () async {
      final client = ApiClient(
        httpClient: MockClient(
          (_) async => _jsonResponse(
            statusCode: 400,
            body: {
              'success': false,
              'data': null,
              'error': {
                'code': 'VALIDATION_ERROR',
                'message': 'El mensaje no puede estar vacío.',
              },
            },
          ),
        ),
      );

      await expectLater(
        client.postDemo(''),
        throwsA(
          isA<ApiException>()
              .having((error) => error.statusCode, 'statusCode', 400)
              .having(
                (error) => error.message,
                'message',
                'El mensaje no puede estar vacío.',
              ),
        ),
      );
    });
  });
}

http.Response _jsonResponse({
  required int statusCode,
  required Map<String, Object?> body,
}) {
  return http.Response(
    jsonEncode(body),
    statusCode,
    headers: {'content-type': 'application/json; charset=utf-8'},
  );
}
