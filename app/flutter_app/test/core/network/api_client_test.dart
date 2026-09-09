import 'dart:convert';

import 'package:flutter_app/core/errors/api_exception.dart';
import 'package:flutter_app/core/network/api_client.dart';
import 'package:flutter_app/features/users/services/user_service.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  group('ApiClient', () {
    test('convierte la respuesta de GET /users a modelos de usuario', () async {
      final client = ApiClient(
        baseUrl: 'http://gateway.test:8080',
        httpClient: MockClient((request) async {
          expect(request.method, 'GET');
          expect(request.url.path, '/users');
          return http.Response(
            jsonEncode({
              'success': true,
              'data': [
                {'id': 1, 'username': 'alanp', 'email': 'alanp@example.com'},
              ],
              'error': null,
            }),
            200,
          );
        }),
      );

      final users = await UserService(client).getUsers();

      expect(users, hasLength(1));
      expect(users.single.username, 'alanp');
      expect(users.single.toJson()['email'], 'alanp@example.com');
    });

    test('convierte un error HTTP del Gateway a ApiException', () async {
      final client = ApiClient(
        baseUrl: 'http://gateway.test:8080',
        httpClient: MockClient(
          (_) async => http.Response(
            jsonEncode({
              'success': false,
              'data': null,
              'error': {
                'code': 'VALIDATION_ERROR',
                'message': 'El campo username es obligatorio',
              },
            }),
            400,
          ),
        ),
      );

      expect(
        UserService(client).createUser(username: '', email: 'test@example.com'),
        throwsA(
          isA<ApiException>()
              .having((error) => error.statusCode, 'statusCode', 400)
              .having((error) => error.code, 'code', 'VALIDATION_ERROR'),
        ),
      );
    });
  });
}
