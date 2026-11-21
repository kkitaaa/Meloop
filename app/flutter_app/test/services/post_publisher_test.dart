import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/post_publisher.dart';

void main() {
  test('envía los campos del post como multipart y procesa el éxito', () async {
    final adapter = _ResponseAdapter(
      statusCode: 201,
      responseBody: {
        'success': true,
        'data': {
          'post': {'id': 'post-123'},
          'xp_awarded': 15,
        },
        'error': null,
      },
    );
    final dio = Dio()..httpClientAdapter = adapter;
    final apiClient = ApiClient(dio: dio);
    final progress = <(int, int)>[];

    final result = await apiClient.publishPost(
      content: 'Texto de prueba',
      images: [
        PostImageUpload(
          bytes: Uint8List.fromList([1, 2, 3]),
          name: 'cover.jpg',
        ),
      ],
      audioBytes: Uint8List.fromList([4, 5, 6]),
      audioName: 'audio.mp3',
      song: const MusicTrack(
        id: 'spotify-track',
        title: 'Canción',
        artist: 'Artista',
      ),
      cancelToken: CancelToken(),
      onSendProgress: (sent, total) => progress.add((sent, total)),
    );

    expect(adapter.requestOptions?.method, 'POST');
    expect(adapter.requestOptions?.path, '/v1/posts');
    expect(adapter.requestBody, contains('Texto de prueba'));
    expect(adapter.requestBody, contains('name="images"'));
    expect(adapter.requestBody, contains('filename="cover.jpg"'));
    expect(adapter.requestBody, contains('filename="audio.mp3"'));
    expect(adapter.requestBody, contains('spotify-track'));
    expect(progress, isNotEmpty);
    expect(result.post['id'], 'post-123');
    expect(result.xpAwarded, 15);
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
