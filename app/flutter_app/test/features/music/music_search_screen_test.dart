import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/music/presentation/music_search_screen.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('debounces queries and sends type and offset for pagination', (
    tester,
  ) async {
    final adapter = _MusicSearchAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(
        home: MusicSearchScreen(
          mode: MusicSearchMode.pickForPost,
          apiClient: apiClient,
        ),
      ),
    );
    final appBar = tester.widget<PreferredSize>(
      find.byType(PreferredSize).first,
    );
    expect(appBar.preferredSize.height, 55);
    expect(find.byType(Image), findsOneWidget);

    await tester.enterText(find.byType(TextField), 'r');
    await tester.pump(const Duration(seconds: 1));
    expect(adapter.requests, isEmpty);

    await tester.enterText(find.byType(TextField), 'Rosalía');
    await tester.pump(const Duration(milliseconds: 399));
    expect(adapter.requests, isEmpty);
    await tester.pump(const Duration(milliseconds: 1));
    await tester.pumpAndSettle();

    expect(adapter.requests, hasLength(1));
    expect(adapter.requests.first.queryParameters, {
      'q': 'Rosalía',
      'type': 'track',
      'offset': '0',
    });
    expect(find.text('Canción de prueba'), findsOneWidget);

    await tester.tap(find.text('Álbumes'));
    await tester.pumpAndSettle();
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pumpAndSettle();
    expect(adapter.requests.last.queryParameters, {
      'q': 'Rosalía',
      'type': 'album',
      'offset': '0',
    });
    expect(find.text('Álbum de prueba'), findsOneWidget);

    await tester.tap(find.text('Cargar más'));
    await tester.pumpAndSettle();
    expect(adapter.requests.last.queryParameters['offset'], '1');
    expect(find.text('Segundo álbum'), findsOneWidget);
    apiClient.close();
  });

  testWidgets(
    'muestra el mensaje específico cuando Spotify no está disponible',
    (tester) async {
      final adapter = _MusicSearchAdapter(unavailable: true);
      final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
      await tester.pumpWidget(
        MaterialApp(
          home: MusicSearchScreen(
            mode: MusicSearchMode.browse,
            apiClient: apiClient,
          ),
        ),
      );

      await tester.enterText(find.byType(TextField), 'Rosalía');
      await tester.pump(const Duration(milliseconds: 400));
      await tester.pumpAndSettle();
      expect(
        find.text(
          'No pudimos conectar con Spotify. Intenta nuevamente en unos minutos.',
        ),
        findsOneWidget,
      );
      expect(find.text('Reintentar'), findsOneWidget);
      apiClient.close();
    },
  );

  testWidgets('cancels the previous request and ignores its late response', (
    tester,
  ) async {
    final adapter = _MusicSearchAdapter(holdFirstResponse: true);
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(
        home: MusicSearchScreen(
          mode: MusicSearchMode.browse,
          apiClient: apiClient,
        ),
      ),
    );

    await tester.enterText(find.byType(TextField), 'old');
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pump();
    expect(adapter.requests, hasLength(1));

    await tester.enterText(find.byType(TextField), 'new');
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pumpAndSettle();
    expect(adapter.cancelledQueries, contains('old'));
    expect(find.text('New result'), findsOneWidget);

    adapter.firstResponse!.complete(
      adapter.successResponse('Old result', offset: 0, total: 1),
    );
    await tester.pumpAndSettle();
    expect(find.text('Old result'), findsNothing);
    expect(find.text('New result'), findsOneWidget);
    apiClient.close();
  });

  testWidgets('returns a selected result in preference mode', (tester) async {
    final adapter = _MusicSearchAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    MusicTrack? selected;
    await tester.pumpWidget(
      MaterialApp(
        home: Builder(
          builder: (context) => Scaffold(
            body: TextButton(
              onPressed: () async {
                selected = await Navigator.of(context).push<MusicTrack>(
                  MaterialPageRoute(
                    builder: (_) => MusicSearchScreen(
                      mode: MusicSearchMode.pickForPreference,
                      apiClient: apiClient,
                    ),
                  ),
                );
              },
              child: const Text('Abrir búsqueda'),
            ),
          ),
        ),
      ),
    );

    await tester.tap(find.text('Abrir búsqueda'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), 'Rosalía');
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Canción de prueba'));
    await tester.pumpAndSettle();

    expect(selected?.title, 'Canción de prueba');
    expect(selected?.type, 'track');
    apiClient.close();
  });

  test('parses genre names and pagination metadata', () {
    final page = MusicSearchPage.fromJson({
      'genres': [
        {'name': 'Pop latino'},
      ],
      'offset': 0,
      'total': 4,
    }, type: 'genre');

    expect(page.items.single.title, 'Pop latino');
    expect(page.items.single.type, 'genre');
    expect(page.hasMore, isTrue);
  });
}

class _MusicSearchAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  final List<String> cancelledQueries = [];
  final bool unavailable;
  final bool holdFirstResponse;
  Completer<ResponseBody>? firstResponse;

  _MusicSearchAdapter({
    this.unavailable = false,
    this.holdFirstResponse = false,
  });

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    final query = options.queryParameters['q'] as String;
    cancelFuture?.then((_) => cancelledQueries.add(query));
    if (unavailable) {
      return _response({
        'success': false,
        'data': null,
        'error': {
          'code': 'MUSIC_UNAVAILABLE',
          'message': 'Spotify no está disponible.',
        },
      }, 503);
    }
    if (holdFirstResponse && requests.length == 1) {
      firstResponse = Completer<ResponseBody>();
      return firstResponse!.future;
    }

    final type = options.queryParameters['type'];
    final offset = int.parse(options.queryParameters['offset'] as String);
    final items = switch (type) {
      'album' when offset == 0 => [
        {'id': 'album-1', 'name': 'Álbum de prueba'},
      ],
      'album' => [
        {'id': 'album-2', 'name': 'Segundo álbum'},
      ],
      _ => [
        {
          'id': 'track-1',
          'name': query == 'new' ? 'New result' : 'Canción de prueba',
          'artists': [
            {'name': 'Artista de prueba'},
          ],
        },
      ],
    };
    final total = type == 'album' ? 2 : 1;
    return successResponse(
      items.first['name'] as String,
      offset: offset,
      total: total,
      items: items,
    );
  }

  ResponseBody successResponse(
    String title, {
    required int offset,
    required int total,
    List<Map<String, Object?>>? items,
  }) {
    return _response({
      'success': true,
      'data': {
        'items':
            items ??
            [
              {'id': title, 'name': title},
            ],
        'offset': offset,
        'total': total,
      },
      'error': null,
    }, 200);
  }

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
