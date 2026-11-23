import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/services/auth_session.dart';
import 'package:flutter_app/services/user_permissions.dart';
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
        expect(session.user?.role, UserRole.user);
        client.close();
        await session.clear();
      },
    );

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
            role: UserRole.user,
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

    test(
      'refreshes the saved role from the authenticated backend response',
      () async {
        final session = AuthSession(persistent: false);
        await session.establish(
          sessionToken: 'role-token',
          expiresInSeconds: 3600,
          sessionUser: const AuthUser(
            id: 'user-1',
            username: 'Ana',
            email: 'ana@example.test',
            role: UserRole.user,
          ),
        );
        final client = ApiClient(
          session: session,
          httpClient: MockClient((request) async {
            expect(request.url.path, '/auth/validate');
            return _jsonResponse(
              statusCode: 200,
              body: {
                'success': true,
                'data': {
                  'id': 'user-1',
                  'username': 'Ana',
                  'email': 'ana@example.test',
                  'role': 'moderador',
                },
                'error': null,
              },
            );
          }),
        );

        final user = await client.validateSession();

        expect(user.role, UserRole.moderator);
        expect(session.user?.role, UserRole.moderator);
        expect(session.permissions.canModerateReports, isTrue);
        client.close();
        await session.clear();
      },
    );

    test('syncs permissions after a forbidden response', () async {
      final session = AuthSession(persistent: false);
      await session.establish(
        sessionToken: 'role-token',
        expiresInSeconds: 3600,
        sessionUser: const AuthUser(
          id: 'user-1',
          username: 'Ana',
          email: 'ana@example.test',
          role: UserRole.admin,
        ),
      );
      final client = ApiClient(
        session: session,
        httpClient: MockClient((request) async {
          if (request.url.path == '/api/demo') {
            return _jsonResponse(
              statusCode: 403,
              body: {
                'success': false,
                'data': null,
                'error': {'code': 'FORBIDDEN', 'message': 'Permiso revocado.'},
              },
            );
          }
          expect(request.url.path, '/auth/validate');
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'id': 'user-1',
                'username': 'Ana',
                'email': 'ana@example.test',
                'role': 'usuario',
              },
              'error': null,
            },
          );
        }),
      );

      await expectLater(
        client.getDemo(),
        throwsA(
          isA<ApiException>().having(
            (error) => error.statusCode,
            'statusCode',
            403,
          ),
        ),
      );

      expect(session.user?.role, UserRole.user);
      expect(session.permissions.canAccessAdministration, isFalse);
      client.close();
      await session.clear();
    });

    test('syncs permissions after a Dio request is forbidden', () async {
      final session = AuthSession(persistent: false);
      await session.establish(
        sessionToken: 'role-token',
        expiresInSeconds: 3600,
        sessionUser: const AuthUser(
          id: 'user-1',
          username: 'Ana',
          email: 'ana@example.test',
          role: UserRole.admin,
        ),
      );
      final dio = Dio();
      dio.httpClientAdapter = _DioMockAdapter((options) {
        expect(options.uri.path, '/v1/search/users');
        expect(options.headers['Authorization'], 'Bearer role-token');
        return ResponseBody.fromString(
          jsonEncode({
            'success': false,
            'data': null,
            'error': {'code': 'FORBIDDEN', 'message': 'Permiso revocado.'},
          }),
          403,
          headers: {
            Headers.contentTypeHeader: ['application/json'],
          },
        );
      });
      final client = ApiClient(
        dio: dio,
        session: session,
        httpClient: MockClient((request) async {
          expect(request.url.path, '/auth/validate');
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'id': 'user-1',
                'username': 'Ana',
                'email': 'ana@example.test',
                'role': 'usuario',
              },
              'error': null,
            },
          );
        }),
      );

      await expectLater(
        client.searchUsers('Ana'),
        throwsA(
          isA<ApiException>().having(
            (error) => error.statusCode,
            'statusCode',
            403,
          ),
        ),
      );

      expect(session.user?.role, UserRole.user);
      dio.close(force: true);
      client.close();
      await session.clear();
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

    test('requests the selected feed page from the gateway', () async {
      final client = ApiClient(
        httpClient: MockClient((request) async {
          expect(request.method, 'GET');
          expect(request.url.path, '/api/feed');
          expect(request.url.queryParameters, {
            'category': 'friends',
            'page': '2',
            'pageSize': '5',
          });
          return _jsonResponse(
            statusCode: 200,
            body: {
              'success': true,
              'data': {
                'items': [_feedPost(6)],
                'page': 2,
                'pageSize': 5,
                'hasMore': false,
              },
              'error': null,
            },
          );
        }),
      );

      final response = await client.getFeedPage(category: 'friends', page: 2);

      expect(response.data?.items.single.id, 6);
      expect(response.data?.hasMore, isFalse);
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

    test(
      'uses authenticated account settings and password endpoints',
      () async {
        final session = AuthSession(persistent: false);
        await session.establish(
          sessionToken: 'settings-token',
          expiresInSeconds: 3600,
          sessionUser: const AuthUser(
            id: 'user-1',
            username: 'Ana',
            email: 'ana@example.test',
            role: UserRole.user,
          ),
        );
        final client = ApiClient(
          session: session,
          httpClient: MockClient((request) async {
            expect(request.headers['authorization'], 'Bearer settings-token');
            switch ('${request.method} ${request.url.path}') {
              case 'GET /users/me/privacy':
                return _jsonResponse(
                  statusCode: 200,
                  body: {
                    'success': true,
                    'data': {'visibilidad_perfil': 'AMIGOS'},
                    'error': null,
                  },
                );
              case 'PATCH /users/me/privacy':
                expect(jsonDecode(request.body), {
                  'visibilidad_perfil': 'PRIVADO',
                });
                return _jsonResponse(
                  statusCode: 200,
                  body: {
                    'success': true,
                    'data': {'visibilidad_perfil': 'PRIVADO'},
                    'error': null,
                  },
                );
              case 'GET /users/me/notifications/settings':
                return _jsonResponse(
                  statusCode: 200,
                  body: {
                    'success': true,
                    'data': [
                      {'tipo_notificacion': 'LIKE', 'habilitada': false},
                    ],
                    'error': null,
                  },
                );
              case 'PATCH /users/me/notifications/settings/LIKE':
                expect(jsonDecode(request.body), {'habilitada': true});
                return _jsonResponse(
                  statusCode: 200,
                  body: {
                    'success': true,
                    'data': {'tipo_notificacion': 'LIKE', 'habilitada': true},
                    'error': null,
                  },
                );
              case 'POST /auth/change-password':
                expect(jsonDecode(request.body), {
                  'current_password': 'current-pass',
                  'new_password': 'new-password',
                });
                return _jsonResponse(
                  statusCode: 200,
                  body: {
                    'success': true,
                    'data': {'message': 'Contraseña actualizada correctamente'},
                    'error': null,
                  },
                );
              default:
                fail('Solicitud inesperada: ${request.method} ${request.url}');
            }
          }),
        );

        expect(
          (await client.getPrivacySettings())['visibilidad_perfil'],
          'AMIGOS',
        );
        expect(
          (await client.updatePrivacySetting(
            'visibilidad_perfil',
            'PRIVADO',
          ))['visibilidad_perfil'],
          'PRIVADO',
        );
        expect(
          (await client.getNotificationSettings()).single['habilitada'],
          isFalse,
        );
        await client.updateNotificationSetting(type: 'LIKE', enabled: true);
        await client.changePassword(
          currentPassword: 'current-pass',
          newPassword: 'new-password',
        );

        client.close();
        await session.clear();
      },
    );
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
  String role = 'usuario',
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
          'role': role,
        },
      },
      'error': null,
    },
  );
}
