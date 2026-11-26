import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/posts/presentation/saved_posts_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/post_card.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('lists saved posts and removes unavailable posts', (
    tester,
  ) async {
    final adapter = _PostsAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(home: SavedPostsScreen(apiClient: apiClient)),
    );
    await tester.pumpAndSettle();

    expect(find.text('Guardados'), findsOneWidget);
    expect(find.text('Publicación guardada visible'), findsOneWidget);
    expect(find.text('Ya no está disponible'), findsOneWidget);

    await tester.tap(find.text('Quitar'));
    await tester.pumpAndSettle();

    expect(
      adapter.requests.any(
        (request) =>
            request.method == 'DELETE' &&
            request.path == '/v1/me/saved/deleted-post',
      ),
      isTrue,
    );
    expect(find.text('Ya no está disponible'), findsNothing);
    apiClient.close();
  });

  testWidgets('Post Card saves and shares a post with a friend', (
    tester,
  ) async {
    final adapter = _PostsAdapter();
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ListView(
            children: [
              PostCard(
                postData: {
                  'id': 42,
                  'user': 'Ana',
                  'time': 'Hace un minuto',
                  'content': 'Contenido de prueba',
                  'likes': 1,
                  'comments': 0,
                  'isAuthor': false,
                  'isEdited': false,
                  'within24Hours': false,
                },
                tealAccent: const Color(0xFF1ABC9C),
                apiClient: apiClient,
              ),
            ],
          ),
        ),
      ),
    );

    await tester.tap(find.byTooltip('Guardar publicación'));
    await tester.pumpAndSettle();
    expect(find.byTooltip('Quitar de Guardados'), findsOneWidget);

    await tester.tap(find.byTooltip('Compartir publicación'));
    await tester.pumpAndSettle();
    expect(find.text('Compartir'), findsOneWidget);
    expect(find.text('Ana'), findsNWidgets(2));

    await tester.tap(find.byTooltip('Enviar a Ana'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));
    expect(
      adapter.requests.any(
        (request) =>
            request.method == 'POST' && request.path == '/v1/posts/42/share',
      ),
      isTrue,
    );
    expect(find.text('Enviada a @Ana'), findsOneWidget);
    apiClient.close();
  });

  test('maps blocked share errors to the generic message', () async {
    final adapter = _PostsAdapter(blockShare: true);
    final apiClient = ApiClient(dio: Dio()..httpClientAdapter = adapter);

    await expectLater(
      apiClient.sharePost(postId: '42', recipientId: 'friend-1'),
      throwsA(
        isA<ApiException>()
            .having((error) => error.errorCode, 'errorCode', 'SHARE_RESTRICTED')
            .having(
              (error) => error.message,
              'message',
              'No puedes compartir con esta persona.',
            ),
      ),
    );
    apiClient.close();
  });
}

class _PostsAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  final bool blockShare;

  _PostsAdapter({this.blockShare = false});

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);
    if (options.path == '/v1/me/saved' && options.method == 'GET') {
      return _response(200, {
        'success': true,
        'data': {
          'items': [
            {'post_id': 'deleted-post', 'post': null, 'is_available': false},
            {
              'post': {
                'id': 55,
                'user': 'Artista',
                'time': 'Ayer',
                'content': 'Publicación guardada visible',
                'likes': 2,
                'comments': 1,
              },
            },
          ],
          'hasMore': false,
        },
        'error': null,
      });
    }
    if (blockShare && options.path.endsWith('/share')) {
      return _response(403, {
        'success': false,
        'data': null,
        'error': {
          'code': 'BLOCKED',
          'message': 'Detalle que no debe mostrarse',
        },
      });
    }
    if (options.path == '/v1/social/friends') {
      return _response(200, {
        'success': true,
        'data': {
          'items': [
            {
              'id_usuario': 'friend-1',
              'username': 'Ana',
              'nivel': 3,
              'friend_status': 'friend',
            },
          ],
        },
        'error': null,
      });
    }
    return _response(200, {'success': true, 'data': null, 'error': null});
  }

  ResponseBody _response(int statusCode, Map<String, Object?> body) =>
      ResponseBody.fromString(
        jsonEncode(body),
        statusCode,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );

  @override
  void close({bool force = false}) {}
}
