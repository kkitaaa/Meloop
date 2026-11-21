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

    test('requests the selected feed page from the gateway', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'GET');
          expect(request.url.path, '/api/feed');
          expect(request.url.queryParameters, {
            'category': 'friends',
            'page': '2',
            'pageSize': '5',
          });
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'items': [_feedPost(6)],
                'page': 2,
                'pageSize': 5,
                'hasMore': false,
              },
              'error': null,
            },
          );
        }),
      );

      final response = await client.getFeedPage(category: 'friends', page: 2);

      expect(response.data?.items.single.id, 6);
      expect(response.data?.hasMore, isFalse);
    });

    test('requests mutual friends using the cursor API', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'GET');
          expect(
            request.url.path,
            '/v1/social/users/target-user/mutual-friends',
          );
          expect(request.url.queryParameters, {
            'limit': '2',
            'cursor': 'cursor-1',
          });
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'items': [
                  {
                    'id_usuario': 'friend-1',
                    'username': 'Ana',
                    'avatar_url': 'https://example.test/ana.png',
                  },
                ],
                'totalCount': 3,
                'hidden': false,
                'nextCursor': 'cursor-2',
                'hasMore': true,
              },
              'error': null,
            },
          );
        }),
      );

      final response = await client.getMutualFriends(
        'target-user',
        cursor: 'cursor-1',
        limit: 2,
      );

      expect(response.data?.items.single.id, 'friend-1');
      expect(response.data?.totalCount, 3);
      expect(response.data?.nextCursor, 'cursor-2');
      expect(response.data?.hasMore, isTrue);
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

Map<String, Object?> _feedPost(int id) => {
  'id': id,
  'user': 'Usuario_$id',
  'time': 'Hace 5 minutos',
  'content': 'Contenido de prueba',
  'song': 'Canción de prueba',
  'artist': 'Artista de prueba',
  'likes': 10,
  'comments': 1,
  'isAuthor': false,
  'isEdited': false,
  'within24Hours': true,
};

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
