import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_app/features/music/presentation/music_track_detail_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('loads track detail and adds it to preferences', (tester) async {
    final adapter = _MusicDetailAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(
        home: MusicDetailScreen(
          type: 'track',
          spotifyId: 'track-1',
          apiClient: apiClient,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(adapter.requests.first.path, '/v1/music/tracks/track-1');
    expect(find.text('Canción demo'), findsOneWidget);
    expect(find.text('Artista demo'), findsOneWidget);
    expect(find.text('Álbum demo'), findsOneWidget);
    expect(find.text('Agregar a mis preferencias'), findsOneWidget);

    await tester.tap(find.text('Agregar a mis preferencias'));
    await tester.pumpAndSettle();
    expect(adapter.addRequests, 1);
    expect(find.text('En tus preferencias'), findsOneWidget);
    apiClient.close();
  });

  testWidgets('shows an already-saved preference and artist genres', (
    tester,
  ) async {
    final adapter = _MusicDetailAdapter(savedTrack: true, type: 'artist');
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(
        home: MusicDetailScreen(
          type: 'artist',
          spotifyId: 'artist-1',
          apiClient: apiClient,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(adapter.requests.first.path, '/v1/music/artists/artist-1');
    expect(find.text('Artista demo'), findsOneWidget);
    expect(find.text('Pop'), findsOneWidget);
    expect(find.text('Rock'), findsOneWidget);
    expect(find.text('En tus preferencias'), findsOneWidget);
    expect(find.text('Agregar a mis preferencias'), findsNothing);
    apiClient.close();
  });

  testWidgets('opens the validated Spotify URL with the external launcher', (
    tester,
  ) async {
    const channel = MethodChannel('plugins.flutter.io/url_launcher');
    MethodCall? launchCall;
    TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
        .setMockMethodCallHandler(channel, (call) async {
          launchCall = call;
          return true;
        });
    addTearDown(
      () => TestDefaultBinaryMessengerBinding.instance.defaultBinaryMessenger
          .setMockMethodCallHandler(channel, null),
    );

    final apiClient = ApiClient(
      dio: Dio()..httpClientAdapter = _MusicDetailAdapter(),
    );
    await tester.pumpWidget(
      MaterialApp(
        home: MusicDetailScreen(
          type: 'track',
          spotifyId: 'track-1',
          apiClient: apiClient,
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.tap(find.text('Abrir en Spotify'));
    await tester.pumpAndSettle();

    expect(launchCall?.method, 'launch');
    expect(
      (launchCall?.arguments as Map<Object?, Object?>)['url'],
      'https://open.spotify.com/track/track-1',
    );
    expect(
      (launchCall?.arguments as Map<Object?, Object?>)['useWebView'],
      isFalse,
    );
    apiClient.close();
  });

  testWidgets(
    'shows album tracks and maps not-found to the requested message',
    (tester) async {
      final adapter = _MusicDetailAdapter(type: 'album');
      final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
      await tester.pumpWidget(
        MaterialApp(
          home: MusicDetailScreen(
            type: 'album',
            spotifyId: 'album-1',
            apiClient: apiClient,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Canciones del álbum'), findsOneWidget);
      expect(find.text('Pista del álbum'), findsOneWidget);
      apiClient.close();

      final unavailableClient = ApiClient(
        dio: Dio()..httpClientAdapter = _MusicDetailAdapter(notFound: true),
      );
      await tester.pumpWidget(
        MaterialApp(
          home: MusicDetailScreen(
            key: const ValueKey('not-found'),
            type: 'track',
            spotifyId: 'deleted-track',
            apiClient: unavailableClient,
          ),
        ),
      );
      await tester.pumpAndSettle();
      expect(find.text('Este contenido ya no está disponible'), findsOneWidget);
      expect(find.text('Reintentar'), findsOneWidget);
      unavailableClient.close();
    },
  );

  testWidgets('shows Spotify unavailable error for a 503 response', (
    tester,
  ) async {
    final apiClient = ApiClient(
      dio: Dio()..httpClientAdapter = _MusicDetailAdapter(unavailable: true),
    );
    await tester.pumpWidget(
      MaterialApp(
        home: MusicDetailScreen(
          type: 'track',
          spotifyId: 'offline-track',
          apiClient: apiClient,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(
      find.text(
        'Spotify no está disponible. Intenta nuevamente en unos minutos.',
      ),
      findsOneWidget,
    );
    apiClient.close();
  });
}

class _MusicDetailAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  final bool savedTrack;
  final bool notFound;
  final bool unavailable;
  final String type;
  int addRequests = 0;

  _MusicDetailAdapter({
    this.savedTrack = false,
    this.notFound = false,
    this.unavailable = false,
    this.type = 'track',
  });

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    if (notFound && options.path.startsWith('/v1/music/tracks/')) {
      return _response({
        'success': false,
        'data': null,
        'error': {'code': 'NOT_FOUND', 'message': 'Gone'},
      }, 404);
    }
    if (unavailable && options.path.startsWith('/v1/music/tracks/')) {
      return _response({
        'success': false,
        'data': null,
        'error': {'code': 'MUSIC_UNAVAILABLE', 'message': 'Offline'},
      }, 503);
    }

    if (options.path == '/v1/music/preferences' && options.method == 'GET') {
      final saved = savedTrack
          ? {
              'songs': [
                {'id': 'track-1', 'name': 'Canción demo'},
              ],
              'artists': [
                {'id': 'artist-1', 'name': 'Artista demo'},
              ],
              'genres': [],
            }
          : {'songs': [], 'artists': [], 'genres': []};
      return _response(_success({'preferences': saved}), 200);
    }

    if (options.path == '/v1/music/preferences' && options.method == 'POST') {
      addRequests++;
      return _response({'success': true, 'data': {}, 'error': null}, 201);
    }

    if (options.path.startsWith('/v1/music/')) {
      final data = switch (type) {
        'artist' => {
          'id': 'artist-1',
          'name': 'Artista demo',
          'images': [
            {'url': 'https://example.test/artist.png'},
          ],
          'genres': ['Pop', 'Rock'],
          'external_url': 'https://open.spotify.com/artist/artist-1',
        },
        'album' => {
          'id': 'album-1',
          'name': 'Álbum demo',
          'images': [
            {'url': 'https://example.test/album.png'},
          ],
          'artists': [
            {'name': 'Artista demo'},
          ],
          'tracks': {
            'items': [
              {'id': 'track-2', 'name': 'Pista del álbum'},
            ],
          },
          'external_url': 'https://open.spotify.com/album/album-1',
        },
        _ => {
          'id': 'track-1',
          'name': 'Canción demo',
          'artists': [
            {'name': 'Artista demo'},
          ],
          'album': {
            'name': 'Álbum demo',
            'images': [
              {'url': 'https://example.test/album.png'},
            ],
          },
          'external_url': 'https://open.spotify.com/track/track-1',
        },
      };
      return _response(_success(data), 200);
    }
    throw StateError('Unexpected ${options.method} ${options.path}');
  }

  Map<String, Object?> _success(Object data) => {
    'success': true,
    'data': data,
    'error': null,
  };

  ResponseBody _response(Object data, int statusCode) {
    return ResponseBody.fromString(
      jsonEncode(data),
      statusCode,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
