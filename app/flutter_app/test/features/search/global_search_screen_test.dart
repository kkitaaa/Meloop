import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/core/navigation/artists_and_songs_screen.dart';
import 'package:flutter_app/core/navigation/desktop_navigation_shell.dart';
import 'package:flutter_app/features/notifications/presentation/notification_scope.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_app/features/search/presentation/global_search_screen.dart';
import 'package:flutter_app/screens/friends_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/post_card.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('submits queries explicitly and isolates a music search failure', (
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

    // Typing and pausing does NOT auto-trigger search or pollute recent history.
    await tester.enterText(find.byType(TextField), 'ana');
    await tester.pump(const Duration(seconds: 1));
    expect(adapter.paths, isEmpty);

    // Pressing the Buscar button executes the search.
    await tester.tap(find.text('Buscar'));
    await tester.pumpAndSettle();
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
    await tester.tap(find.text('Buscar'));
    await tester.pump();
    expect(adapter.paths, isEmpty);

    for (final query in ['ana', 'bea', 'cata', 'dora', 'eva', 'fina']) {
      await tester.enterText(find.byType(TextField), query);
      await tester.testTextInput.receiveAction(TextInputAction.search);
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

  testWidgets(
    'desktop two-section search shows interactive posts and jumps songs/friends to tabs',
    (tester) async {
      tester.view.physicalSize = const Size(1440, 900);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final notifController = NotificationController(
        repository: MockNotificationRepository(),
        playSound: false,
      )..start();
      addTearDown(notifController.dispose);

      final adapter = _RichSearchAdapter();
      final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
      addTearDown(apiClient.close);

      await tester.pumpWidget(
        NotificationScope(
          controller: notifController,
          child: MaterialApp(home: DesktopNavigationShell(apiClient: apiClient)),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byTooltip('Búsqueda global'));
      await tester.pumpAndSettle();

      expect(find.text('Publicaciones y música'), findsOneWidget);
      expect(find.text('Amigos y personas'), findsOneWidget);

      await tester.enterText(find.byType(TextField), 'daft');
      await tester.testTextInput.receiveAction(TextInputAction.search);
      await tester.pumpAndSettle();

      // Posts are rendered directly as interactive PostCard widgets.
      expect(find.byType(PostCard), findsOneWidget);
      expect(find.text('Escuchando Daft Punk hoy'), findsOneWidget);
      // Songs appear in the left sheet and friends in the right sheet.
      expect(find.text('Get Lucky'), findsOneWidget);
      expect(find.text('Vicente'), findsOneWidget);

      // Tapping a song jumps directly to the Artists & Songs tab with query applied.
      await tester.tap(find.text('Get Lucky'));
      await tester.pumpAndSettle();
      expect(find.byType(ArtistsAndSongsScreen), findsOneWidget);

      // Open search again and tap a friend -> jumps directly to the Friends tab.
      await tester.tap(find.byTooltip('Búsqueda global'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Vicente'));
      await tester.pumpAndSettle();
      expect(find.byType(FriendsScreen), findsOneWidget);
    },
  );
}

class _RichSearchAdapter implements HttpClientAdapter {
  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    final Object data = switch (options.path) {
      '/v1/search/users' => {
        'items': [
          {
            'id_usuario': 'user-1',
            'username': 'Vicente',
            'nivel': 7,
            'friend_status': 'friends',
          },
        ],
      },
      '/v1/search/posts' => {
        'items': [
          {
            'id': 'post-1',
            'user': 'Andrea',
            'user_id': 'user-2',
            'content': 'Escuchando Daft Punk hoy',
            'song': 'Instant Crush',
            'artist': 'Daft Punk',
            'likes': 5,
            'comments': 2,
          },
        ],
      },
      '/v1/music/search' => {
        'items': [
          {
            'id': 'track-1',
            'title': 'Get Lucky',
            'artist': 'Daft Punk',
            'type': 'track',
          },
        ],
      },
      _ => {'items': [], 'page': 1, 'pageSize': 5, 'hasMore': false},
    };
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
