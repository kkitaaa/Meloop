import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/screens/feed_screen.dart';
import 'package:flutter_app/services/api_client.dart';
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
                _post((page - 1) * 5 + 1, category),
                _post((page - 1) * 5 + 2, category),
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

    await tester.ensureVisible(find.text('Cargar más blogs'));
    await tester.tap(find.text('Cargar más blogs'));
    await tester.pumpAndSettle();
    expect(find.text('Post recent 6'), findsOneWidget);
    expect(requests[1].queryParameters['page'], '2');

    await tester.tap(find.text('amigos'));
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
    await tester.tap(find.text('Reintentar'));
    await tester.pumpAndSettle();
    expect(requestCount, 2);
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
