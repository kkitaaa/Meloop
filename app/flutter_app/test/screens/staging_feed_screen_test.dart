import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/screens/feed_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/auth_session.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets('loads the staging feed contract without category filters', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(800, 900);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    final session = AuthSession(persistent: false);
    await session.establish(
      sessionToken: 'staging-token',
      expiresInSeconds: 3600,
      sessionUser: const AuthUser(
        id: 'staging-user',
        username: 'Vicente',
        email: 'vicente@example.test',
      ),
    );
    addTearDown(session.clear);

    final client = ApiClient(
      session: session,
      httpClient: MockClient((request) async {
        expect(request.url.path, '/v1/feed');
        expect(request.url.queryParameters, {'page': '1', 'limit': '5'});
        expect(request.headers['authorization'], 'Bearer staging-token');
        expect(request.headers['x-user-id'], 'staging-user');
        return http.Response(
          jsonEncode({
            'page': 1,
            'limit': 5,
            'data': [
              {
                'id': 'post-uuid-1',
                'content': 'Publicación de staging',
                'media_url': '',
                'created_at': '2026-11-24T18:00:00Z',
                'author_id': 'author-1',
                'author_username': 'Ana',
                'author_avatar': '',
                'likes_count': 4,
                'comments_count': 2,
              },
            ],
          }),
          200,
          headers: {'content-type': 'application/json; charset=utf-8'},
        );
      }),
    );
    addTearDown(client.close);

    await tester.pumpWidget(MaterialApp(home: FeedScreen(apiClient: client)));
    await tester.pumpAndSettle();

    expect(find.text('Publicación de staging'), findsOneWidget);
    expect(find.text('Recientes'), findsNothing);
    expect(find.text('Tendencia'), findsNothing);
  });
}
