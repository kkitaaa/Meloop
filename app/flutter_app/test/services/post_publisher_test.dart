import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/post_publisher.dart';

void main() {
  test(
    'envía una publicación de texto como JSON y procesa la respuesta 201',
    () async {
      final adapter = _ResponseAdapter(
        statusCode: 201,
        responseBody: {
          'id': 'post-123',
          'author_id': 'author-1',
          'content': 'Texto de prueba',
          'media_url': '',
          'music_id': '',
          'created_at': '2026-11-24T20:00:00Z',
        },
      );
      final dio = Dio()..httpClientAdapter = adapter;
      final apiClient = ApiClient(dio: dio);

      final result = await apiClient.publishPost(
        content: 'Texto de prueba',
        images: const [],
        cancelToken: CancelToken(),
        onSendProgress: (_, _) {},
      );

      expect(adapter.requestOptions?.method, 'POST');
      expect(adapter.requestOptions?.path, '/api/v1/posts');
      expect(adapter.requestOptions?.contentType, Headers.jsonContentType);
      expect(jsonDecode(adapter.requestBody), {
        'content': 'Texto de prueba',
        'media_urls': <String>[],
        'music_id': '',
      });
      expect(result.post['id'], 'post-123');
      expect(result.post['author_id'], 'author-1');
      expect(result.xpAwarded, isNull);
    },
  );

  test(
    'rechaza adjuntos mientras solo está habilitada la publicación de texto',
    () async {
      final adapter = _ResponseAdapter(statusCode: 201, responseBody: {});
      final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);

      await expectLater(
        apiClient.publishPost(
          content: 'Texto',
          images: [
            PostImageUpload(bytes: Uint8List.fromList([1]), name: 'imagen.jpg'),
          ],
          cancelToken: CancelToken(),
          onSendProgress: (_, _) {},
        ),
        throwsA(
          isA<ApiException>().having(
            (error) => error.message,
            'message',
            contains('solo se pueden publicar textos'),
          ),
        ),
      );
      expect(adapter.requestOptions, isNull);
    },
  );

  test('expone status y error del servidor en HTTP 400', () async {
    final dio = Dio()
      ..httpClientAdapter = _ResponseAdapter(
        statusCode: 400,
        responseBody: {'error': 'datos inválidos o incompletos'},
      );
    final apiClient = ApiClient(dio: dio);

    await expectLater(
      apiClient.publishPost(
        content: 'Texto',
        images: const [],
        cancelToken: CancelToken(),
        onSendProgress: (_, _) {},
      ),
      throwsA(
        isA<ApiException>()
            .having((error) => error.statusCode, 'statusCode', 400)
            .having(
              (error) => error.message,
              'message',
              'datos inválidos o incompletos',
            ),
      ),
    );
  });

  test('traduce los códigos de error de creación de posts', () async {
    final cases = {
      'MUSIC_NOT_FOUND': 'No encontramos esa canción en Spotify.',
      'MUSIC_UNAVAILABLE':
          'No pudimos validar la canción ahora. Intenta más tarde.',
      'POSTING_RESTRICTED': 'Tienes una restricción temporal para publicar.',
    };

    for (final entry in cases.entries) {
      final dio = Dio()
        ..httpClientAdapter = _ResponseAdapter(
          statusCode: 422,
          responseBody: {
            'success': false,
            'data': null,
            'error': {'code': entry.key, 'message': 'backend message'},
          },
        );
      final apiClient = ApiClient(dio: dio);

      await expectLater(
        apiClient.publishPost(
          content: 'Texto',
          images: const [],
          cancelToken: CancelToken(),
          onSendProgress: (_, _) {},
        ),
        throwsA(
          isA<ApiException>()
              .having((error) => error.errorCode, 'errorCode', entry.key)
              .having((error) => error.message, 'message', entry.value),
        ),
      );
    }
  });
}

class _ResponseAdapter implements HttpClientAdapter {
  final int statusCode;
  final Map<String, dynamic> responseBody;
  RequestOptions? requestOptions;
  String requestBody = '';

  _ResponseAdapter({required this.statusCode, required this.responseBody});

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requestOptions = options;
    if (requestStream != null) {
      requestBody = await utf8.decoder.bind(requestStream).join();
    }
    return ResponseBody.fromString(
      jsonEncode(responseBody),
      statusCode,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
