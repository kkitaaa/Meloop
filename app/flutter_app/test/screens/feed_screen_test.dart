import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/features/notifications/presentation/notification_scope.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:flutter_app/screens/feed_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/friends_sidebar.dart';
import 'package:flutter_app/widgets/paper/stacked_paper_card.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets(
    'loads feed pages via infinite scroll and requests a new category',
    (tester) async {
      tester.view.physicalSize = const Size(1920, 1080);
      tester.view.devicePixelRatio = 1;
      addTearDown(tester.view.resetPhysicalSize);
      addTearDown(tester.view.resetDevicePixelRatio);

      final notificationController = NotificationController(
        repository: MockNotificationRepository(),
        playSound: false,
      )..start();
      addTearDown(notificationController.dispose);

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
                'items': List.generate(
                  5,
                  (i) => _post((page - 1) * 5 + i + 1, category),
                ),
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
      addTearDown(client.close);

      await tester.pumpWidget(
        NotificationScope(
          controller: notificationController,
          child: MaterialApp(home: FeedScreen(apiClient: client)),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('Post recent 1'), findsOneWidget);
      expect(find.text('Cargar más blogs'), findsNothing);
      expect(requests.single.queryParameters['category'], 'recent');
      expect(find.byType(StackedPaperCard), findsWidgets);
      expect(find.byType(CustomizablePin), findsOneWidget);
      expect(find.byType(TapeCornerOverlay), findsOneWidget);
      expect(find.byType(DesktopFriendsSidebar), findsOneWidget);
      expect(find.text('Amigos conectados'), findsOneWidget);
      expect(find.text('Perfiles que te podrían interesar'), findsOneWidget);
      expect(
        find.text('En base a tus recomendaciones de Spotify'),
        findsOneWidget,
      );
      expect(
        tester.getCenter(find.byType(TapeCornerOverlay)).dx,
        closeTo(960, 2),
      );

      // Los 3 botones de filtro de blogs tienen el mismo ancho
      final recWidth = tester
          .getSize(
            find.ancestor(
              of: find.text('Recientes'),
              matching: find.byType(AnimatedContainer),
            ),
          )
          .width;
      final tendWidth = tester
          .getSize(
            find.ancestor(
              of: find.text('Tendencia'),
              matching: find.byType(AnimatedContainer),
            ),
          )
          .width;
      final amiWidth = tester
          .getSize(
            find.ancestor(
              of: find.text('Amigos').last,
              matching: find.byType(AnimatedContainer),
            ),
          )
          .width;
      expect(recWidth, tendWidth);
      expect(tendWidth, amiWidth);

      // Infinite scroll triggers page 2 automatically
      await tester.drag(find.byType(ListView), const Offset(0, -1600));
      await tester.pumpAndSettle();
      expect(requests.length, greaterThanOrEqualTo(2));
      expect(requests[1].queryParameters['page'], '2');

      // Switch category using the 3-button pill bar in Ultimos Blogs
      await tester.tap(find.text('Tendencia'));
      await tester.pumpAndSettle();
      expect(find.text('Post suggested 1'), findsOneWidget);
      expect(requests.last.queryParameters['category'], 'suggested');
      expect(requests.last.queryParameters['page'], '1');

      await tester.tap(find.text('Amigos').last);
      await tester.pumpAndSettle();
      expect(find.text('Post friends 1'), findsOneWidget);
      expect(requests.last.queryParameters['category'], 'friends');
      expect(requests.last.queryParameters['page'], '1');
    },
  );

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
    addTearDown(client.close);

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
    addTearDown(client.close);

    await tester.pumpWidget(MaterialApp(home: FeedScreen(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.byTooltip('Probar conexión con API'), findsNothing);
    await tester.tap(find.byTooltip('Ajustes'));
    await tester.pumpAndSettle();
    expect(find.text('Preferencias musicales'), findsOneWidget);
    expect(find.text('Usuarios bloqueados'), findsOneWidget);
  });
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
