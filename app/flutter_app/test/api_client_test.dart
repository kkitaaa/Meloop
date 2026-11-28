import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/models/account_settings.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/auth_session.dart';
import 'package:flutter_app/services/post_publisher.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  group('ApiClient', () {
    test(
      'logs in through the gateway and attaches the session token',
      () async {
        final session = AuthSession(persistent: false);
        final client = ApiClient(
          session: session,
          httpClient: MockClient((request) async {
            expect(request.method, 'POST');
            expect(request.url.path, '/auth/login');
            expect(jsonDecode(request.body), {
              'email': 'ana@example.test',
              'password': 'password123',
            });
            return _sessionResponse(
              statusCode: 200,
              token: 'session-token',
              expiresIn: 3600,
            );
          }),
        );

        final user = await client.login(
          email: ' ana@example.test ',
          password: 'password123',
        );

        expect(user.id, 'user-1');
        expect(session.token, 'session-token');
        expect(session.user?.email, 'ana@example.test');
        client.close();
        await session.clear();
      },
    );

    test('attaches the authenticated session to text post requests', () async {
      final session = AuthSession(persistent: false);
      await session.establish(
        sessionToken: 'post-token',
        expiresInSeconds: 3600,
        sessionUser: const AuthUser(
          id: 'post-user',
          username: 'Ana',
          email: 'ana@example.test',
        ),
      );
      final dio = Dio();
      dio.httpClientAdapter = _DioMockAdapter((options) {
        expect(options.method, 'POST');
        expect(options.uri.path, '/api/v1/posts');
        expect(options.headers['Authorization'], 'Bearer post-token');
        expect(options.headers['X-User-ID'], 'post-user');
        expect(options.contentType, Headers.jsonContentType);
        expect(options.data, {
          'content': 'Publicación de prueba',
          'media_urls': <String>[],
          'music_id': '',
        });
        return ResponseBody.fromString(
          jsonEncode({
            'success': true,
            'data': {
              'post': {'id': 'post-1'},
            },
            'error': null,
          }),
          201,
          headers: {
            Headers.contentTypeHeader: ['application/json'],
          },
        );
      });
      final client = ApiClient(dio: dio, session: session);

      final result = await client.publishPost(
        content: 'Publicación de prueba',
        images: const <PostImageUpload>[],
        cancelToken: CancelToken(),
        onSendProgress: (sent, total) {},
      );

      expect(result.post['id'], 'post-1');
      dio.close(force: true);
      await session.clear();
    });

    test('loads raw author posts from the post service', () async {
      final session = AuthSession(persistent: false);
      await session.establish(
        sessionToken: 'profile-token',
        expiresInSeconds: 3600,
        sessionUser: const AuthUser(
          id: 'viewer-1',
          username: 'Ana',
          email: 'ana@example.test',
        ),
      );
      final dio = Dio();
      dio.httpClientAdapter = _DioMockAdapter((options) {
        expect(options.method, 'GET');
        expect(options.uri.path, '/api/v1/posts/author/author-1');
        expect(options.uri.queryParameters, {'limit': '10', 'offset': '0'});
        expect(options.headers['Authorization'], 'Bearer profile-token');
        expect(options.headers['X-User-ID'], 'viewer-1');
        return ResponseBody.fromString(
          jsonEncode([
            {
              'id': 'post-1',
              'author_id': 'author-1',
              'content': 'Publicación real',
              'created_at': '2026-11-24T20:00:00Z',
            },
          ]),
          200,
          headers: {
            Headers.contentTypeHeader: ['application/json'],
          },
        );
      });
      final client = ApiClient(dio: dio, session: session);

      final posts = await client.getAuthorPosts('author-1');

      expect(posts.single['content'], 'Publicación real');
      client.close();
      await session.clear();
    });

    test(
      'persists like and unlike requests with the authenticated session',
      () async {
        final session = AuthSession(persistent: false);
        await session.establish(
          sessionToken: 'likes-token',
          expiresInSeconds: 3600,
          sessionUser: const AuthUser(
            id: 'viewer-1',
            username: 'Ana',
            email: 'ana@example.test',
          ),
        );
        final dio = Dio();
        var requestCount = 0;
        dio.httpClientAdapter = _DioMockAdapter((options) {
          expect(options.uri.path, '/v1/posts/post-1/like');
          expect(options.headers['Authorization'], 'Bearer likes-token');
          expect(options.headers['X-User-ID'], 'viewer-1');
          expect(options.method, requestCount++ == 0 ? 'POST' : 'DELETE');
          return ResponseBody.fromString(
            jsonEncode({'message': 'ok'}),
            200,
            headers: {
              Headers.contentTypeHeader: ['application/json'],
            },
          );
        });
        final client = ApiClient(dio: dio, session: session);

        await client.setPostLike(postId: 'post-1', liked: true);
        await client.setPostLike(postId: 'post-1', liked: false);

        client.close();
        dio.close(force: true);
        await session.clear();
      },
    );

    test(
      'parses Spotify catalog search and detail responses without a user login',
      () async {
        final dio = Dio();
        dio.httpClientAdapter = _DioMockAdapter((options) {
          expect(options.headers['Authorization'], isNull);
          if (options.uri.path == '/v1/music/search') {
            expect(options.method, 'GET');
            expect(options.uri.queryParameters, {
              'q': 'Daft Punk',
              'type': 'track',
              'offset': '0',
              'limit': '10',
            });
            return ResponseBody.fromString(
              jsonEncode({
                'tracks': {
                  'offset': 0,
                  'limit': 10,
                  'total': 1,
                  'next': null,
                  'items': [
                    {
                      'id': 'track-1',
                      'name': 'One More Time',
                      'artists': [
                        {'id': 'artist-1', 'name': 'Daft Punk'},
                      ],
                      'album': {
                        'images': [
                          {'url': 'https://images.example.test/cover.jpg'},
                        ],
                      },
                      'external_urls': {
                        'spotify': 'https://open.spotify.com/track/track-1',
                      },
                    },
                  ],
                },
              }),
              200,
              headers: {
                Headers.contentTypeHeader: ['application/json'],
              },
            );
          }

          expect(options.uri.path, '/v1/music/tracks/track-1');
          return ResponseBody.fromString(
            jsonEncode({
              'id': 'track-1',
              'name': 'One More Time',
              'artists': [
                {'id': 'artist-1', 'name': 'Daft Punk'},
              ],
              'album': {
                'name': 'Discovery',
                'images': [
                  {'url': 'https://images.example.test/cover.jpg'},
                ],
              },
              'external_urls': {
                'spotify': 'https://open.spotify.com/track/track-1',
              },
            }),
            200,
            headers: {
              Headers.contentTypeHeader: ['application/json'],
            },
          );
        });
        final client = ApiClient(dio: dio);

        final search = await client.searchMusicTracks('Daft Punk');
        final result = search.data!.items.single;
        expect(result.id, 'track-1');
        expect(result.title, 'One More Time');
        expect(result.artist, 'Daft Punk');
        expect(result.artworkUrl, 'https://images.example.test/cover.jpg');
        expect(search.data!.hasMore, isFalse);

        final detail = await client.getMusicDetail(
          type: 'track',
          spotifyId: 'track-1',
        );
        expect(detail.data!.name, 'One More Time');
        expect(detail.data!.album, 'Discovery');
        expect(
          detail.data!.externalUrl,
          'https://open.spotify.com/track/track-1',
        );
        client.close();
        dio.close(force: true);
      },
    );

    test('parses normalized music results and empty artwork', () async {
      final dio = Dio();
      dio.httpClientAdapter = _DioMockAdapter((options) {
        expect(options.uri.queryParameters['limit'], '20');
        return ResponseBody.fromString(
          jsonEncode({
            'success': true,
            'data': {
              'results': [
                {
                  'type': 'track',
                  'spotify_id': 'track-1',
                  'name': 'Canción',
                  'artists': ['Rosalía'],
                  'album': '',
                  'image_url': '',
                  'external_url': '',
                },
              ],
              'limit': 20,
              'offset': 0,
              'total': 2,
            },
            'error': null,
          }),
          200,
          headers: {
            Headers.contentTypeHeader: ['application/json'],
          },
        );
      });
      final client = ApiClient(dio: dio);
      final page = (await client.searchMusicTracks('Rosalía', limit: 20)).data!;
      expect(page.items.single.id, 'track-1');
      expect(page.items.single.artist, 'Rosalía');
      expect(page.items.single.artworkUrl, isNull);
      expect(page.total, 2);
      expect(page.hasMore, isTrue);
      client.close();
      dio.close(force: true);
    });

    test('accepts a login response without an optional role', () async {
      final client = ApiClient(
        session: AuthSession(persistent: false),
        httpClient: MockClient((_) async {
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'session_token': 'session-token',
                'expires_in': 3600,
                'user': {
                  'id': 'user-1',
                  'username': 'Ana',
                  'email': 'ana@example.test',
                },
              },
              'error': null,
            },
          );
        }),
      );

      final user = await client.login(
        email: 'ana@example.test',
        password: 'password123',
      );
      expect(user.id, 'user-1');
      client.close();
    });

    test(
      'renews an expiring session before an authenticated request',
      () async {
        final session = AuthSession(persistent: false);
        await session.establish(
          sessionToken: 'old-token',
          expiresInSeconds: 1,
          sessionUser: const AuthUser(
            id: 'user-1',
            username: 'Ana',
            email: 'ana@example.test',
          ),
        );
        final client = ApiClient(
          session: session,
          httpClient: MockClient((request) async {
            if (request.url.path == '/auth/refresh') {
              expect(request.method, 'POST');
              expect(request.headers['authorization'], 'Bearer old-token');
              return _sessionResponse(
                statusCode: 200,
                token: 'renewed-token',
                expiresIn: 3600,
              );
            }

            expect(request.url.path, '/api/demo');
            expect(request.headers['authorization'], 'Bearer renewed-token');
            return _jsonResponse(
              statusCode: 200,
              body: {
                'success': true,
                'data': {
                  'message': 'Gateway disponible',
                  'timestamp': '2026-11-01T12:00:00Z',
                },
                'error': null,
              },
            );
          }),
        );

        final result = await client.getDemo();
        expect(result.success, isTrue);
        expect(session.token, 'renewed-token');
        client.close();
        await session.clear();
      },
    );

    test('decodes a successful GET response', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'GET');
          expect(request.url.path, '/api/demo');
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'message': 'Gateway disponible',
                'timestamp': '2026-11-01T12:00:00Z',
              },
              'error': null,
            },
          );
        }),
      );

      final response = await client.getDemo();

      expect(response.success, isTrue);
      expect(response.data?.message, 'Gateway disponible');
    });

    test('sends and decodes a successful POST response', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'POST');
          expect(request.url.path, '/api/demo');
          expect(jsonDecode(request.body), {'message': 'Hola'});
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'receivedMessage': 'Hola',
                'receivedAt': '2026-11-01T12:00:00Z',
              },
              'error': null,
            },
          );
        }),
      );

      final response = await client.postDemo('Hola');

      expect(response.data?.receivedMessage, 'Hola');
    });

    test('uses the account and privacy API contracts', () async {
      final session = AuthSession(persistent: false);
      final client = ApiClient(
        session: session,
        httpClient: MockClient((request) async {
          if (request.method == 'GET' && request.url.path == '/users/me') {
            return _jsonResponse(
              statusCode: 200,
              body: {
                'success': true,
                'data': {
                  'id_usuario': 'user-1',
                  'username': 'Ana',
                  'correo': 'ana@example.test',
                },
                'error': null,
              },
            );
          }
          if (request.method == 'GET' &&
              request.url.path == '/users/me/privacy') {
            return _jsonResponse(
              statusCode: 200,
              body: {
                'success': true,
                'data': _privacySettings(profileVisibility: 'PUBLICO'),
                'error': null,
              },
            );
          }
          if (request.method == 'PATCH' &&
              request.url.path == '/users/me/privacy') {
            expect(jsonDecode(request.body), {'visibilidad_perfil': 'PRIVADO'});
            return _jsonResponse(
              statusCode: 200,
              body: {
                'success': true,
                'data': _privacySettings(profileVisibility: 'PRIVADO'),
                'error': null,
              },
            );
          }
          if (request.method == 'POST' &&
              request.url.path == '/auth/change-password') {
            expect(jsonDecode(request.body), {
              'current_password': 'old-password',
              'new_password': 'new-password',
            });
            return _jsonResponse(
              statusCode: 200,
              body: {'success': true, 'data': null, 'error': null},
            );
          }
          fail('Unexpected request: ${request.method} ${request.url}');
        }),
      );

      final account = await client.getAccountDetails();
      expect(account.data?.email, 'ana@example.test');

      final privacy = await client.getPrivacySettings();
      expect(privacy.data?.profileVisibility, 'PUBLICO');

      final updatedPrivacy = await client.updatePrivacySettings(
        const PrivacySettingsPatch(profileVisibility: 'PRIVADO'),
      );
      expect(updatedPrivacy.data?.profileVisibility, 'PRIVADO');

      final password = await client.changePassword(
        currentPassword: 'old-password',
        newPassword: 'new-password',
      );
      expect(password.success, isTrue);

      client.close();
      await session.clear();
    });

    test('reads and updates notification preferences', () async {
      final session = AuthSession(persistent: false);
      final client = ApiClient(
        session: session,
        httpClient: MockClient((request) async {
          if (request.method == 'GET' &&
              request.url.path == '/users/me/notifications/settings') {
            return _jsonResponse(
              statusCode: 200,
              body: {
                'success': true,
                'data': [
                  {'tipo_notificacion': 'LIKE', 'habilitada': true},
                ],
                'error': null,
              },
            );
          }
          if (request.method == 'PATCH' &&
              request.url.path == '/users/me/notifications/settings/LIKE') {
            expect(jsonDecode(request.body), {'habilitada': false});
            return _jsonResponse(
              statusCode: 200,
              body: {
                'success': true,
                'data': {'tipo_notificacion': 'LIKE', 'habilitada': false},
                'error': null,
              },
            );
          }
          fail('Unexpected request: ${request.method} ${request.url}');
        }),
      );

      final preferences = await client.getNotificationPreferences();
      expect(preferences.data?.single.type, 'LIKE');
      expect(preferences.data?.single.enabled, isTrue);

      final updated = await client.updateNotificationPreference(
        type: 'LIKE',
        enabled: false,
      );
      expect(updated.data?.enabled, isFalse);

      client.close();
      await session.clear();
    });

    test('requests a feed page using the environment API contract', () async {
      final isDevelopment = ApiConfig.environment == Environment.development;
      final session = AuthSession(persistent: false);
      await session.establish(
        sessionToken: 'feed-token',
        expiresInSeconds: 3600,
        sessionUser: const AuthUser(
          id: 'feed-user',
          username: 'Ana',
          email: 'ana@example.test',
        ),
      );
      final client = ApiClient(
        session: session,
        httpClient: MockClient((request) async {
          expect(request.method, 'GET');
          expect(
            request.url.path,
            ApiConfig.feedEndpointFor(ApiConfig.environment),
          );
          expect(request.url.queryParameters, {
            'page': '2',
            if (isDevelopment) 'category': 'friends',
            if (isDevelopment) 'pageSize': '5',
            if (!isDevelopment) 'limit': '5',
          });
          if (!isDevelopment) {
            expect(request.headers['authorization'], 'Bearer feed-token');
            expect(request.headers['x-user-id'], 'feed-user');
          }
          return _jsonResponse(
            statusCode: 200,
            body: {
              if (isDevelopment) ...{
                'success': true,
                'data': {
                  'items': [_feedPost(6)],
                  'page': 2,
                  'pageSize': 5,
                  'hasMore': false,
                },
                'error': null,
              } else ...{
                'page': 2,
                'limit': 5,
                'data': [_backendFeedPost('feed-post-6')],
              },
            },
          );
        }),
      );

      final response = await client.getFeedPage(category: 'friends', page: 2);

      expect(response.data?.items.single.id, isDevelopment ? 6 : 'feed-post-6');
      expect(response.data?.hasMore, isFalse);
      if (!isDevelopment) {
        expect(response.data?.items.single.user, 'Ana Backend');
        expect(response.data?.items.single.userId, 'author-6');
        expect(response.data?.items.single.likes, 12);
      }
      client.close();
      await session.clear();
    });

    test('reports HTTP status when the feed route returns non-JSON', () async {
      final client = ApiClient(
        httpClient: MockClient((_) async {
          return http.Response(
            '404 page not found',
            404,
            headers: {'content-type': 'text/plain'},
          );
        }),
      );

      await expectLater(
        client.getFeedPage(category: 'recent', page: 1),
        throwsA(
          isA<ApiException>()
              .having((error) => error.statusCode, 'statusCode', 404)
              .having(
                (error) => error.message,
                'message',
                contains('HTTP 404'),
              ),
        ),
      );
      client.close();
    });

    test('requests mutual friends using the cursor API', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'GET');
          expect(
            request.url.path,
            '/v1/social/users/target-user/mutual-friends',
          );
          expect(request.url.queryParameters, {
            'limit': '2',
            'cursor': 'cursor-1',
          });
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'items': [
                  {
                    'id_usuario': 'friend-1',
                    'username': 'Ana',
                    'avatar_url': 'https://example.test/ana.png',
                  },
                ],
                'totalCount': 3,
                'hidden': false,
                'nextCursor': 'cursor-2',
                'hasMore': true,
              },
              'error': null,
            },
          );
        }),
      );

      final response = await client.getMutualFriends(
        'target-user',
        cursor: 'cursor-1',
        limit: 2,
      );

      expect(response.data?.items.single.id, 'friend-1');
      expect(response.data?.totalCount, 3);
      expect(response.data?.nextCursor, 'cursor-2');
      expect(response.data?.hasMore, isTrue);
    });

    test('sends friend requests through the Gateway friends route', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'POST');
          expect(request.url.path, '/v1/social/friends/requests');
          expect(jsonDecode(request.body), {'id_usuario': 'friend-1'});
          return _jsonResponse(
            statusCode: 201,
            body: {
              'success': true,
              'data': {'id': 'request-1'},
              'error': null,
            },
          );
        }),
      );

      final response = await client.sendFriendRequest('friend-1');

      expect(response.data?['id'], 'request-1');
      client.close();
    });

    test('surfaces the backend message and HTTP status', () async {
      final client = ApiClient(
        httpClient: MockClient(
          (_) async => _jsonResponse(
            statusCode: 400,
            body: {
              'success': false,
              'data': null,
              'error': {
                'code': 'VALIDATION_ERROR',
                'message': 'El mensaje no puede estar vacío.',
              },
            },
          ),
        ),
      );

      await expectLater(
        client.postDemo(''),
        throwsA(
          isA<ApiException>()
              .having((error) => error.statusCode, 'statusCode', 400)
              .having(
                (error) => error.message,
                'message',
                'El mensaje no puede estar vacío.',
              ),
        ),
      );
    });
  });
}

class _DioMockAdapter implements HttpClientAdapter {
  _DioMockAdapter(this.respond);

  final ResponseBody Function(RequestOptions options) respond;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return respond(options);
  }

  @override
  void close({bool force = false}) {}
}

Map<String, Object?> _privacySettings({required String profileVisibility}) => {
  'id_usuario': 'user-1',
  'visibilidad_perfil': profileVisibility,
  'visibilidad_publicaciones': 'AMIGOS',
  'visibilidad_interacciones': 'PRIVADO',
  'recepcion_mensajes': 'TODOS',
  'recepcion_solicitudes_amistad': 'AMIGOS',
};

Map<String, Object?> _feedPost(int id) => {
  'id': id,
  'user': 'Usuario_$id',
  'time': 'Hace 5 minutos',
  'content': 'Contenido de prueba',
  'song': 'Canción de prueba',
  'artist': 'Artista de prueba',
  'likes': 10,
  'comments': 1,
  'isAuthor': false,
  'isEdited': false,
  'within24Hours': true,
};

Map<String, Object?> _backendFeedPost(String id) => {
  'id': id,
  'content': 'Contenido del backend',
  'media_url': 'https://cdn.example.test/post.jpg',
  'created_at': '2026-11-24T18:00:00Z',
  'author_id': 'author-6',
  'author_username': 'Ana Backend',
  'author_avatar': '',
  'likes_count': 12,
  'comments_count': 3,
};

http.Response _jsonResponse({
  required int statusCode,
  required Map<String, Object?> body,
}) {
  return http.Response(
    jsonEncode(body),
    statusCode,
    headers: {'content-type': 'application/json; charset=utf-8'},
  );
}

http.Response _sessionResponse({
  required int statusCode,
  required String token,
  required int expiresIn,
}) {
  return _jsonResponse(
    statusCode: statusCode,
    body: {
      'success': true,
      'data': {
        'session_token': token,
        'expires_in': expiresIn,
        'user': {
          'id': 'user-1',
          'username': 'Ana',
          'email': 'ana@example.test',
        },
      },
      'error': null,
    },
  );
}
