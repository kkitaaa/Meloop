import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/search/presentation/global_search_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('debounces queries and isolates a music search failure', (
    tester,
  ) async {
    final adapter = _SearchAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: GlobalSearchScreen(apiClient: apiClient)),
    );

    final appBar = tester.widget<PreferredSize>(
      find.byType(PreferredSize).first,
    );
    expect(appBar.preferredSize.height, 55);
    expect(find.byType(Image), findsOneWidget);

    await tester.enterText(find.byType(TextField), 'ana');
    await tester.pump(const Duration(milliseconds: 399));
    expect(adapter.paths, isEmpty);

    await tester.pump(const Duration(milliseconds: 1));
    await tester.pump();
    await tester.pump();
    expect(
      adapter.paths,
      containsAll(['/v1/search/users', '/v1/search/posts', '/v1/music/search']),
    );
    expect(
      adapter.requests
          .where((request) => request.path == '/v1/search/posts')
          .single
          .queryParameters,
      {'q': 'ana', 'filter': 'text'},
    );

    await tester.tap(find.text('Música'));
    await tester.pumpAndSettle();
    expect(find.text('Spotify no está disponible.'), findsOneWidget);

    await tester.tap(find.text('Usuarios'));
    await tester.pumpAndSettle();
    expect(find.text('Ana'), findsOneWidget);
    expect(find.text('Nivel 4'), findsOneWidget);
  });

  testWidgets('Android search header and content respect system insets', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(400, 800);
    tester.view.devicePixelRatio = 1;
    tester.view.viewPadding = const FakeViewPadding(top: 24, bottom: 24);
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);
    addTearDown(tester.view.resetViewPadding);
    final apiClient = ApiClient(
      dio: Dio()..httpClientAdapter = _SearchAdapter(),
    );

    await tester.pumpWidget(
      MaterialApp(
        theme: ThemeData(platform: TargetPlatform.android),
        home: GlobalSearchScreen(apiClient: apiClient),
      ),
    );

    final appBar = tester.widget<PreferredSize>(
      find.byType(PreferredSize).first,
    );
    expect(appBar.preferredSize.height, 79);
    final headerSafeArea = tester.widget<SafeArea>(
      find.descendant(
        of: find.byType(PreferredSize).first,
        matching: find.byType(SafeArea),
      ),
    );
    expect(headerSafeArea.top, isTrue);
    final screenScaffold = tester.widget<Scaffold>(find.byType(Scaffold));
    final bodySafeArea = screenScaffold.body! as SafeArea;
    expect(bodySafeArea.top, isFalse);
    expect(bodySafeArea.bottom, isTrue);
  });

  testWidgets('muestra sugerencia, recent searches y límite de cinco', (
    tester,
  ) async {
    final adapter = _SearchAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: GlobalSearchScreen(apiClient: apiClient)),
    );

    expect(
      find.text('Tus búsquedas recientes aparecerán aquí.'),
      findsOneWidget,
    );
    await tester.enterText(find.byType(TextField), 'a');
    await tester.pump(const Duration(seconds: 1));
    expect(adapter.paths, isEmpty);

    for (final query in ['ana', 'bea', 'cata', 'dora', 'eva', 'fina']) {
      await tester.enterText(find.byType(TextField), query);
      await tester.pump(const Duration(milliseconds: 400));
      await tester.pump();
      await tester.pump();
    }

    await tester.enterText(find.byType(TextField), '');
    await tester.pump();
    for (final query in ['fina', 'eva', 'dora', 'cata', 'bea']) {
      expect(find.text(query), findsOneWidget);
    }
    expect(find.text('ana'), findsNothing);
  });
}

class _SearchAdapter implements HttpClientAdapter {
  final List<String> paths = [];
  final List<RequestOptions> requests = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    paths.add(options.path);
    requests.add(options);
    if (options.path == '/v1/music/search') {
      return ResponseBody.fromString(
        jsonEncode({
          'success': false,
          'data': null,
          'error': {
            'code': 'MUSIC_UNAVAILABLE',
            'message': 'Spotify no está disponible.',
          },
        }),
        503,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    final data = options.path == '/v1/search/users'
        ? {
            'items': [
              {
                'id_usuario': 'user-1',
                'username': 'Ana',
                'nivel': 4,
                'friend_status': 'none',
              },
            ],
          }
        : {'items': []};
    return ResponseBody.fromString(
      jsonEncode({'success': true, 'data': data, 'error': null}),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
