import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/post_card.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('a song attached to a post opens its Spotify detail', (
    tester,
  ) async {
    final adapter = _PostMusicAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: SingleChildScrollView(
            child: PostCard(
              postData: {
                'id': 42,
                'user': 'Ana',
                'content': 'Escuchando ahora',
                'song': 'Canción demo',
                'artist': 'Artista demo',
                'song_id': 'track-42',
                'isAuthor': false,
              },
              tealAccent: const Color(0xFF1ABC9C),
              apiClient: apiClient,
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.text('Canción demo').first);
    await tester.pumpAndSettle();

    expect(adapter.detailPath, '/v1/music/tracks/track-42');
    expect(find.text('Ficha musical'), findsOneWidget);
    expect(find.text('Artista demo'), findsOneWidget);
    apiClient.close();
  });
}

class _PostMusicAdapter implements HttpClientAdapter {
  String? detailPath;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (options.path == '/v1/music/tracks/track-42') {
      detailPath = options.path;
      return ResponseBody.fromString(
        jsonEncode({
          'success': true,
          'data': {
            'id': 'track-42',
            'name': 'Canción demo',
            'artists': [
              {'name': 'Artista demo'},
            ],
            'external_url': 'https://open.spotify.com/track/track-42',
          },
          'error': null,
        }),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    if (options.path == '/v1/music/preferences') {
      return ResponseBody.fromString(
        jsonEncode({
          'success': true,
          'data': {
            'preferences': {'songs': [], 'artists': [], 'genres': []},
          },
          'error': null,
        }),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    throw StateError('Unexpected ${options.method} ${options.path}');
  }

  @override
  void close({bool force = false}) {}
}
