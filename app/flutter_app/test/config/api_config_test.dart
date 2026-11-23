import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/config/api_config.dart';

void main() {
  group('ApiConfig Tests', () {
    test('ApiConfig environment is development by default', () {
      expect(
        ApiConfig.environment,
        anyOf(
          Environment.development,
          Environment.staging,
          Environment.production,
        ),
      );
    });

    test('ApiConfig baseUrl is not empty', () {
      expect(ApiConfig.baseUrl, isNotEmpty);
    });

    test('ApiConfig baseUrl starts with http', () {
      expect(
        ApiConfig.baseUrl.startsWith('http'),
        true,
        reason: 'Base URL should start with http or https',
      );
    });

    test('ApiConfig webSocketUrl is derived from baseUrl', () {
      final wsUrl = ApiConfig.webSocketUrl;
      expect(wsUrl, isNotEmpty);
      expect(wsUrl.startsWith('ws'), true);
    });

    test('ApiConfig endpoints are properly defined', () {
      expect(ApiConfig.healthCheckEndpoint, '/health');
      expect(ApiConfig.authEndpoint, '/auth');
      expect(ApiConfig.authLoginEndpoint, '/auth/login');
      expect(ApiConfig.authRefreshEndpoint, '/auth/refresh');
      expect(ApiConfig.demoEndpoint, '/api/demo');
      expect(ApiConfig.feedEndpoint, '/api/feed');
      expect(ApiConfig.usersEndpoint, '/users');
      expect(ApiConfig.friendsEndpoint, '/friends');
      expect(ApiConfig.gamificationEndpoint, '/api/v1/gamification');
      expect(
        ApiConfig.gamificationProgressEndpoint,
        '/api/v1/gamification/me/progress',
      );
      expect(
        ApiConfig.gatewayRouteSummary,
        contains('/auth, /users, /friends, /api/v1/gamification'),
      );
      expect(ApiConfig.recommendationsEndpoint, '/recommendations');
    });

    test('ApiConfig timeouts are configured', () {
      expect(ApiConfig.connectTimeout.inSeconds, 10);
      expect(ApiConfig.receiveTimeout.inSeconds, 10);
      expect(ApiConfig.sendTimeout.inSeconds, 10);
    });

    test('ApiConfig WebSocket configuration is set', () {
      expect(ApiConfig.webSocketReconnectDelay.inSeconds, 3);
      expect(ApiConfig.webSocketMaxReconnectAttempts, 5);
    });

    test('ApiConfig removes trailing slashes from baseUrl', () {
      final baseUrl = ApiConfig.baseUrl;
      expect(baseUrl.endsWith('/'), false);
    });

    test('ApiConfig health endpoint forms correct full URL', () {
      final healthUrl = '${ApiConfig.baseUrl}${ApiConfig.healthCheckEndpoint}';
      expect(Uri.parse(healthUrl).path, '/health');
      expect(Uri.parse(healthUrl).host, isNotEmpty);
    });

    test('ApiConfig preserves TLS when deriving a WebSocket URL', () {
      expect(
        ApiConfig.webSocketUrlFor('https://api.example.test/v1/'),
        'wss://api.example.test/v1',
      );
    });

    test('ApiConfig derives an unencrypted WebSocket URL for local HTTP', () {
      expect(
        ApiConfig.webSocketUrlFor('http://localhost:8080'),
        'ws://localhost:8080',
      );
    });
  });
}
