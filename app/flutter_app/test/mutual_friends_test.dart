import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_app/features/social/presentation/mutual_friends_row.dart';
import 'package:flutter_app/features/social/presentation/mutual_friends_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  testWidgets('hides the profile row when the count is zero', (tester) async {
    final apiClient = ApiClient(
      httpClient: MockClient(
        (_) async => _response(200, {
          'success': true,
          'data': {
            'items': [],
            'totalCount': 0,
            'hidden': false,
            'hasMore': false,
          },
          'error': null,
        }),
      ),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: MutualFriendsRow(
            targetUserId: 'target-user',
            apiClient: apiClient,
          ),
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('0 amigos en común'), findsNothing);
    apiClient.close();
  });

  testWidgets('shows only count and privacy message for a hidden list', (
    tester,
  ) async {
    final apiClient = ApiClient(
      httpClient: MockClient(
        (_) async => _response(200, {
          'success': true,
          'data': {
            'items': [],
            'totalCount': 2,
            'hidden': true,
            'hasMore': false,
          },
          'error': null,
        }),
      ),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: MutualFriendsScreen(
          targetUserId: 'target-user',
          apiClient: apiClient,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('2 amigos en común'), findsOneWidget);
    expect(find.text('Este usuario oculta su lista de amigos'), findsOneWidget);
    expect(find.byType(ListTile), findsNothing);
    apiClient.close();
  });

  testWidgets('renders neutral unavailable-profile state on HTTP 404', (
    tester,
  ) async {
    final apiClient = ApiClient(
      httpClient: MockClient(
        (_) async => _response(404, {
          'success': false,
          'data': null,
          'error': {
            'code': 'NOT_FOUND',
            'message': 'La persona no está disponible.',
          },
        }),
      ),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: MutualFriendsScreen(
          targetUserId: 'target-user',
          apiClient: apiClient,
        ),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Perfil no disponible'), findsOneWidget);
    expect(find.textContaining('bloque'), findsNothing);
    apiClient.close();
  });

  testWidgets('does not repeat a friend returned on the next cursor page', (
    tester,
  ) async {
    tester.view.physicalSize = const Size(400, 200);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.resetPhysicalSize);
    addTearDown(tester.view.resetDevicePixelRatio);

    var requests = 0;
    final apiClient = ApiClient(
      httpClient: MockClient((request) async {
        requests++;
        if (request.url.queryParameters.containsKey('cursor')) {
          return _response(200, {
            'success': true,
            'data': {
              'items': [
                {'id_usuario': 'friend-3', 'username': 'Carla'},
                {'id_usuario': 'friend-4', 'username': 'Diego'},
              ],
              'totalCount': 4,
              'hidden': false,
              'hasMore': false,
            },
            'error': null,
          });
        }
        return _response(200, {
          'success': true,
          'data': {
            'items': [
              {'id_usuario': 'friend-1', 'username': 'Ana'},
              {'id_usuario': 'friend-2', 'username': 'Beto'},
              {'id_usuario': 'friend-3', 'username': 'Carla'},
            ],
            'totalCount': 4,
            'hidden': false,
            'nextCursor': 'cursor-1',
            'hasMore': true,
          },
          'error': null,
        });
      }),
    );

    await tester.pumpWidget(
      MaterialApp(
        home: MutualFriendsScreen(
          targetUserId: 'target-user',
          apiClient: apiClient,
        ),
      ),
    );
    await tester.pumpAndSettle();
    await tester.drag(find.byType(ListView), const Offset(0, -100));
    await tester.pumpAndSettle();
    await tester.drag(find.byType(ListView), const Offset(0, -100));
    await tester.pumpAndSettle();

    expect(requests, 2);
    expect(find.text('Carla'), findsOneWidget);
    expect(find.text('Diego'), findsOneWidget);
    apiClient.close();
  });
}

http.Response _response(int statusCode, Map<String, Object?> body) {
  return http.Response(
    jsonEncode(body),
    statusCode,
    headers: {'content-type': 'application/json; charset=utf-8'},
  );
}
