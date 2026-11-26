import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/services/health_check_service.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test('staging diagnostics check the deployed service routes', () async {
    final paths = <String>[];
    final dio = Dio(BaseOptions(baseUrl: ApiConfig.baseUrl));
    dio.httpClientAdapter = _StatusAdapter((options) {
      paths.add(options.uri.path);
      return switch (options.uri.path) {
        '/health' => 200,
        '/auth/login' => 400,
        '/users/me' ||
        '/friends' ||
        '/api/v1/gamification/me/progress' ||
        '/v1/feed' => 401,
        _ => 404,
      };
    });
    final service = HealthCheckService(dio: dio);

    final results = await service.verifyEndpoints(
      environment: Environment.staging,
    );

    expect(results, {
      'health': true,
      'auth': true,
      'users': true,
      'friends': true,
      'gamification': true,
      'feed': true,
    });
    expect(
      paths,
      containsAll([
        '/health',
        '/auth/login',
        '/users/me',
        '/friends',
        '/api/v1/gamification/me/progress',
        '/v1/feed',
      ]),
    );
    dio.close(force: true);
  });

  test('a missing route is not reported as reachable', () async {
    final dio = Dio(BaseOptions(baseUrl: ApiConfig.baseUrl));
    dio.httpClientAdapter = _StatusAdapter((_) => 404);
    final service = HealthCheckService(dio: dio);

    expect(await service.checkEndpoint('/missing'), isFalse);

    dio.close(force: true);
  });
}

class _StatusAdapter implements HttpClientAdapter {
  _StatusAdapter(this.statusCodeFor);

  final int Function(RequestOptions options) statusCodeFor;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    return ResponseBody.fromString(
      '{}',
      statusCodeFor(options),
      headers: {
        Headers.contentTypeHeader: ['application/json'],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}
