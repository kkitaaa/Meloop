import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/screens/feed_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/post_card.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets('loads feed pages and requests a new category', (tester) async {
    tester.view.physicalSize = const Size(800, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final requests = <Uri>[];
    final client = ApiClient(
      httpClient: MockClient((request) async {
        requests.add(request.url);
        final category = request.url.queryParameters['category']!;
        final page = int.parse(request.url.queryParameters['page']!);
        final hasMore = page == 1;
        return http.Response(
          jsonEncode({
            'success': true,
            'data': {
              'items': [
                for (var index = 0; index < 5; index++)
                  _post((page - 1) * 5 + index + 1, category),
              ],
              'page': page,
              'pageSize': 5,
              'hasMore': hasMore,
            },
            'error': null,
          }),
          200,
          headers: {'content-type': 'application/json; charset=utf-8'},
        );
      }),
    );

    await tester.pumpWidget(MaterialApp(home: FeedScreen(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.text('Post recent 1'), findsOneWidget);
    expect(requests.single.queryParameters['category'], 'recent');

    await tester.dragFrom(
      tester.getCenter(find.text('Post recent 1')),
      const Offset(0, -700),
    );
    await tester.pumpAndSettle();
    await tester.pumpAndSettle();
    expect(find.text('Post recent 6'), findsOneWidget);
    expect(requests[1].queryParameters['page'], '2');

    await tester.tap(find.text('Amigos'));
    await tester.pumpAndSettle();
    expect(find.text('Post friends 1'), findsOneWidget);
    expect(requests.last.queryParameters['category'], 'friends');
    expect(requests.last.queryParameters['page'], '1');
  });

  testWidgets('shows gateway errors with a retry action', (tester) async {
    tester.view.physicalSize = const Size(800, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    var requestCount = 0;
    final client = ApiClient(
      httpClient: MockClient((_) async {
        requestCount++;
        return http.Response(
          jsonEncode({
            'success': false,
            'data': null,
            'error': {
              'code': 'SERVICE_UNAVAILABLE',
              'message': 'Gateway no disponible.',
            },
          }),
          503,
          headers: {'content-type': 'application/json; charset=utf-8'},
        );
      }),
    );

    await tester.pumpWidget(MaterialApp(home: FeedScreen(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.text('Gateway no disponible.'), findsOneWidget);
    await tester.ensureVisible(find.text('Reintentar'));
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();
    expect(requestCount, 2);
  });

  testWidgets('settings menu exposes music preferences and blocked users', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(800, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final client = ApiClient(
      httpClient: MockClient(
        (_) async => http.Response(
          jsonEncode({
            'success': true,
            'data': {'items': [], 'page': 1, 'pageSize': 5, 'hasMore': false},
            'error': null,
          }),
          200,
          headers: {'content-type': 'application/json; charset=utf-8'},
        ),
      ),
    );
    await tester.pumpWidget(MaterialApp(home: FeedScreen(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.byTooltip('Probar conexión con API'), findsNothing);
    await tester.tap(find.byTooltip('Ajustes'));
    await tester.pumpAndSettle();
    expect(find.text('Preferencias musicales'), findsOneWidget);
    expect(find.text('Usuarios bloqueados'), findsOneWidget);
  });

  testWidgets(
    'mobile header starts expanded with pin and listening card, then collapses on scroll to emphasize XP',
    (tester) async {
      tester.view.physicalSize = const Size(412, 915);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final client = ApiClient(
        httpClient: MockClient(
          (_) async => http.Response(
            jsonEncode({
              'success': true,
              'data': {
                'items': [
                  for (var index = 0; index < 8; index++)
                    _post(index + 1, 'recent'),
                ],
                'page': 1,
                'pageSize': 8,
                'hasMore': false,
              },
              'error': null,
            }),
            200,
            headers: {'content-type': 'application/json; charset=utf-8'},
          ),
        ),
      );
      addTearDown(client.close);

      await tester.pumpWidget(MaterialApp(home: FeedScreen(apiClient: client)));
      await tester.pumpAndSettle();

      final pfpFinder = find.byKey(const ValueKey<String>('mobile-header-pfp'));
      final xpFinder = find.byKey(const ValueKey<String>('mobile-header-xp'));
      final pinFinder = find.byKey(const ValueKey<String>('mobile-header-pin'));
      final listeningFinder = find.byKey(
        const ValueKey<String>('mobile-header-listening'),
      );

      expect(pfpFinder, findsOneWidget);
      expect(xpFinder, findsOneWidget);
      expect(pinFinder, findsOneWidget);
      expect(listeningFinder, findsOneWidget);

      final expandedPfpSize = tester.getSize(pfpFinder);
      final expandedXpWidth = tester.getSize(xpFinder).width;

      // Mitad del scroll: sin desbordes (overflow) en la tarjeta XP
      await tester.drag(find.text('Post recent 1'), const Offset(0, -85));
      await tester.pump();
      expect(tester.takeException(), isNull);

      // Al hacer scroll hacia abajo: escala PFP, quita el pin, oculta Escuchando y expande XP
      await tester.drag(find.text('Post recent 1'), const Offset(0, -350));
      await tester.pumpAndSettle();

      final collapsedPfpSize = tester.getSize(pfpFinder);
      final collapsedXpWidth = tester.getSize(xpFinder).width;

      expect(collapsedPfpSize.width, lessThan(expandedPfpSize.width));
      expect(collapsedPfpSize.height, lessThan(expandedPfpSize.height));
      expect(collapsedXpWidth, greaterThan(expandedXpWidth));
      expect(pinFinder, findsNothing);
      expect(listeningFinder, findsNothing);

      // En móvil, compartir y el menú de 3 puntos se mueven al long-press
      expect(find.byIcon(Icons.more_horiz), findsNothing);
      expect(find.byTooltip('Compartir publicación'), findsNothing);
      await tester.longPress(find.byType(PostCard).first);
      await tester.pumpAndSettle();
      expect(find.text('Compartir publicación'), findsOneWidget);
      expect(find.text('Reportar'), findsOneWidget);
    },
  );
}

Map<String, Object?> _post(int id, String category) => {
  'id': id,
  'user': 'Usuario_$id',
  'time': 'Hace 5 minutos',
  'content': 'Post $category $id',
  'song': null,
  'artist': null,
  'likes': id,
  'comments': 0,
  'isAuthor': false,
  'isEdited': false,
  'within24Hours': true,
};
