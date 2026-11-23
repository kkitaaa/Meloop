import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/services/http_client_manager.dart';
import 'package:flutter_app/config/api_config.dart';

void main() {
  group('HttpClientManager Tests', () {
    late HttpClientManager httpClientManager;

    setUp(() {
      httpClientManager = HttpClientManager();
    });

    test('HttpClientManager initializes with correct base URL', () {
      expect(httpClientManager.dio.options.baseUrl, isNotEmpty);
      expect(httpClientManager.dio.options.baseUrl.contains('http'), true);
    });

    test('HttpClientManager sets auth token correctly', () {
      const testToken = 'test-token-12345';
      httpClientManager.setAuthToken(testToken);

      final authHeader = httpClientManager.dio.options.headers['Authorization'];
      expect(authHeader, 'Bearer $testToken');
    });

    test('HttpClientManager clears auth token', () {
      const testToken = 'test-token-12345';
      httpClientManager.setAuthToken(testToken);

      var authHeader = httpClientManager.dio.options.headers['Authorization'];
      expect(authHeader, isNotNull);

      httpClientManager.clearAuthToken();
      authHeader = httpClientManager.dio.options.headers['Authorization'];
      expect(authHeader, isNull);
    });

    test('HttpClientManager sets user ID header', () {
      const userId = 'user-123';
      httpClientManager.setUserId(userId);

      final userIdHeader = httpClientManager.dio.options.headers['X-User-ID'];
      expect(userIdHeader, userId);
    });

    test('HttpClientManager has correct timeout configuration', () {
      expect(
        httpClientManager.dio.options.connectTimeout,
        ApiConfig.connectTimeout,
      );
      expect(
        httpClientManager.dio.options.receiveTimeout,
        ApiConfig.receiveTimeout,
      );
      expect(httpClientManager.dio.options.sendTimeout, ApiConfig.sendTimeout);
    });

    test('HttpClientManager updates base URL dynamically', () {
      const newUrl = 'http://new-backend:8080';
      httpClientManager.updateBaseUrl(newUrl);

      expect(httpClientManager.dio.options.baseUrl, newUrl);
    });

    test('HttpClientManager has interceptors registered', () {
      final interceptorCount = httpClientManager.dio.interceptors.length;
      expect(interceptorCount, greaterThan(0));
    });

    tearDown(() {
      httpClientManager.dispose();
    });
  });
}
