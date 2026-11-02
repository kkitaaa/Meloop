import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/music/presentation/music_preferences_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('shows sections, counters, empty states and refreshes', (
    tester,
  ) async {
    final adapter = _PreferencesAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: MusicPreferencesScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();

    expect(find.text('Preferencias musicales'), findsOneWidget);
    expect(
      find.text(
        'Estas preferencias se usan para sugerirte música y personas afines.',
      ),
      findsOneWidget,
    );
    expect(find.text('Canciones'), findsOneWidget);
    expect(find.text('Artistas'), findsOneWidget);
    expect(find.text('Géneros'), findsOneWidget);
    expect(find.text('0 / 30'), findsNWidgets(2));
    expect(find.text('0 / 12'), findsOneWidget);
    expect(
      find.text('Aún no agregas artistas. Añade tu primera preferencia.'),
      findsOneWidget,
    );

    await tester.drag(find.byType(ListView), const Offset(0, 500));
    await tester.pumpAndSettle();
    expect(adapter.getRequests, 2);
    apiClient.close();
  });

  testWidgets('opens the song search and communicates duplicate preferences', (
    tester,
  ) async {
    final adapter = _PreferencesAdapter(duplicateOnPost: true);
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: MusicPreferencesScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(OutlinedButton, 'Agregar').first);
    await tester.pumpAndSettle();
    expect(find.text('Canciones'), findsOneWidget);
    expect(find.text('Álbumes'), findsNothing);

    await tester.enterText(find.byType(TextField), 'prueba');
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Canción demo'));
    await tester.pumpAndSettle();

    expect(adapter.addRequests, 1);
    expect(find.text('Ya está en tus preferencias'), findsOneWidget);
    apiClient.close();
  });

  testWidgets('adds artists and autocomplete-selected genres', (tester) async {
    final adapter = _PreferencesAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: MusicPreferencesScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(OutlinedButton, 'Agregar').at(1));
    await tester.pumpAndSettle();
    expect(find.text('Artistas'), findsOneWidget);
    await tester.enterText(find.byType(TextField), 'artist');
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Artista demo').first);
    await tester.pumpAndSettle();
    expect(adapter.addedTypes.last, 'artist');

    final addGenreButton = find.widgetWithText(OutlinedButton, 'Agregar').last;
    await tester.ensureVisible(addGenreButton);
    await tester.pumpAndSettle();
    await tester.tap(addGenreButton);
    await tester.pumpAndSettle();
    expect(find.text('Agregar género'), findsOneWidget);
    await tester.enterText(find.byType(TextField), 'pop');
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Pop latino'));
    await tester.pumpAndSettle();
    expect(adapter.addedTypes.last, 'genre');
    expect(adapter.searchTypes, containsAll(['artist', 'genre']));
    apiClient.close();
  });

  testWidgets('disables adding when backend limit is reached', (tester) async {
    final adapter = _PreferencesAdapter(trackLimit: 1, withSong: true);
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: MusicPreferencesScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();

    expect(find.text('1 / 1'), findsOneWidget);
    expect(find.text('Límite alcanzado'), findsOneWidget);
    apiClient.close();
  });

  testWidgets('maps the backend preference limit error', (tester) async {
    final adapter = _PreferencesAdapter(limitOnPost: true);
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: MusicPreferencesScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.widgetWithText(OutlinedButton, 'Agregar').first);
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), 'prueba');
    await tester.pump(const Duration(milliseconds: 400));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Canción demo'));
    await tester.pumpAndSettle();

    expect(find.text('Llegaste al límite de 30'), findsOneWidget);
    apiClient.close();
  });

  testWidgets('confirms and removes a preference', (tester) async {
    final adapter = _PreferencesAdapter(withSong: true);
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: MusicPreferencesScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();

    await tester.tap(find.byTooltip('Quitar Canción favorita'));
    await tester.pumpAndSettle();
    expect(
      find.text('¿Quieres quitar «Canción favorita» de tus preferencias?'),
      findsOneWidget,
    );
    await tester.tap(find.widgetWithText(FilledButton, 'Quitar'));
    await tester.pumpAndSettle();

    expect(adapter.removeRequests, 1);
    expect(find.text('Canción favorita'), findsNothing);
    expect(find.text('Preferencia quitada.'), findsOneWidget);
    apiClient.close();
  });
}

class _PreferencesAdapter implements HttpClientAdapter {
  int getRequests = 0;
  int addRequests = 0;
  int removeRequests = 0;
  final List<String> addedTypes = [];
  final List<String> searchTypes = [];
  final bool duplicateOnPost;
  final bool limitOnPost;
  final bool withSong;
  final int trackLimit;

  _PreferencesAdapter({
    this.duplicateOnPost = false,
    this.limitOnPost = false,
    this.withSong = false,
    this.trackLimit = 30,
  });

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (options.path == '/v1/music/preferences' && options.method == 'GET') {
      getRequests++;
      return _response(_success(_data()), 200);
    }
    if (options.path == '/v1/music/preferences' && options.method == 'POST') {
      addRequests++;
      final requestBody = options.data as Map<String, dynamic>;
      addedTypes.add(requestBody['type'] as String);
      if (duplicateOnPost) {
        return _response({
          'success': false,
          'data': null,
          'error': {'code': 'ALREADY_EXISTS', 'message': 'Already exists'},
        }, 409);
      }
      if (limitOnPost) {
        return _response({
          'success': false,
          'data': null,
          'error': {
            'code': 'PREFERENCE_LIMIT',
            'message': 'Preference limit reached',
          },
        }, 422);
      }
      return _response(_success(_data()), 201);
    }
    if (options.path == '/v1/music/preferences/track/song-1' &&
        options.method == 'DELETE') {
      removeRequests++;
      return _response({'success': true, 'data': null, 'error': null}, 200);
    }
    if (options.path == '/v1/music/search') {
      final type = options.queryParameters['type'] as String;
      searchTypes.add(type);
      final name = switch (type) {
        'artist' => 'Artista demo',
        'genre' => 'Pop latino',
        _ => 'Canción demo',
      };
      return _response(
        _success({
          'items': [
            {
              'id': '$type-1',
              'name': name,
              'artists': [
                {'name': 'Artista demo'},
              ],
            },
          ],
          'offset': 0,
          'total': 1,
        }),
        200,
      );
    }
    throw StateError('Unexpected ${options.method} ${options.path}');
  }

  Map<String, Object?> _data() => {
    'preferences': {
      'songs': withSong
          ? [
              {'id': 'song-1', 'name': 'Canción favorita', 'artwork_url': null},
            ]
          : [],
      'artists': [],
      'genres': [],
    },
    'limits': {'track': trackLimit, 'artist': 12, 'genre': 30},
  };

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
