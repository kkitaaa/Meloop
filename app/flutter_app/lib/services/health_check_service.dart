import 'dart:async';
import 'dart:io';

import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_app/config/api_config.dart';

/// Models for health check response
class HealthCheckResponse {
  final String status;
  final String version;
  final Map<String, dynamic> services;
  final DateTime timestamp;

  HealthCheckResponse({
    required this.status,
    required this.version,
    required this.services,
    required this.timestamp,
  });

  factory HealthCheckResponse.fromJson(Map<String, dynamic> json) {
    return HealthCheckResponse(
      status: json['status'] ?? 'unknown',
      version: json['version'] ?? 'unknown',
      services: (json['services'] as Map<String, dynamic>?) ?? {},
      timestamp: json['timestamp'] != null
          ? DateTime.parse(json['timestamp'] as String)
          : DateTime.now(),
    );
  }

  bool get isHealthy =>
      status.toLowerCase() == 'ok' || status.toLowerCase() == 'healthy';
}

class ServiceHealth {
  final String name;
  final bool isHealthy;
  final String? message;
  final Duration? responseTime;

  ServiceHealth({
    required this.name,
    required this.isHealthy,
    this.message,
    this.responseTime,
  });

  factory ServiceHealth.fromJson(String name, dynamic data) {
    if (data is bool) {
      return ServiceHealth(name: name, isHealthy: data);
    }
    if (data is Map<String, dynamic>) {
      return ServiceHealth(
        name: name,
        isHealthy:
            (data['status'] ?? 'unknown').toString().toLowerCase() == 'ok',
        message: data['message']?.toString(),
        responseTime: data['responseTime'] != null
            ? Duration(milliseconds: data['responseTime'] as int)
            : null,
      );
    }
    return ServiceHealth(name: name, isHealthy: false);
  }
}

/// Service for checking backend health and connectivity
class HealthCheckService {
  final Dio _dio;

  HealthCheckService({Dio? dio})
    : _dio =
          dio ??
          Dio(
            BaseOptions(
              baseUrl: ApiConfig.baseUrl,
              connectTimeout: ApiConfig.connectTimeout,
              receiveTimeout: ApiConfig.receiveTimeout,
            ),
          );

  /// Check overall backend health
  Future<HealthCheckResponse> checkHealth() async {
    try {
      final stopwatch = Stopwatch()..start();

      final response = await _dio
          .get<Map<String, dynamic>>(
            '${ApiConfig.baseUrl}${ApiConfig.healthCheckEndpoint}',
          )
          .timeout(
            const Duration(seconds: 5),
            onTimeout: () => throw TimeoutException('Health check timeout'),
          );

      stopwatch.stop();

      if (response.statusCode == 200 && response.data != null) {
        final health = HealthCheckResponse.fromJson(response.data!);
        debugPrint(
          '[HealthCheck] Backend is healthy (${stopwatch.elapsedMilliseconds}ms)',
        );
        return health;
      } else {
        throw HealthCheckException(
          'Health check returned status ${response.statusCode}',
          response.statusCode,
        );
      }
    } on DioException catch (e) {
      debugPrint('[HealthCheck] DioException: ${e.message}');
      throw HealthCheckException(
        'Failed to connect to backend: ${e.message}',
        e.response?.statusCode,
      );
    } on TimeoutException catch (e) {
      debugPrint('[HealthCheck] Timeout: ${e.message}');
      throw HealthCheckException(e.message ?? 'Timeout');
    } catch (e) {
      debugPrint('[HealthCheck] Error: $e');
      throw HealthCheckException('Health check failed: $e');
    }
  }

  /// Check health of a specific API endpoint
  Future<bool> checkEndpoint(String endpoint) async {
    try {
      final response = await _dio
          .get<dynamic>(endpoint, options: Options(validateStatus: (_) => true))
          .timeout(const Duration(seconds: 5));
      return response.statusCode != null &&
          response.statusCode! >= 200 &&
          response.statusCode! < 300;
    } catch (e) {
      debugPrint('[HealthCheck] Endpoint check failed for $endpoint: $e');
      return false;
    }
  }

  /// Verify all major endpoints are accessible
  Future<Map<String, bool>> verifyEndpoints() async {
    final endpoints = <(String, Future<bool> Function())>[
      ('health', () => checkEndpoint(ApiConfig.healthCheckEndpoint)),
      ('demo', () => checkEndpoint(ApiConfig.demoEndpoint)),
      ('feed', () => checkEndpoint(ApiConfig.feedEndpoint)),
      ('recommendations', _checkRecommendationEndpoint),
      ('auth', _checkAuthEndpoint),
    ];

    final results = <String, bool>{};

    for (final entry in endpoints) {
      results[entry.$1] = await entry.$2();
      await Future.delayed(const Duration(milliseconds: 100)); // Rate limiting
    }

    return results;
  }

  Future<bool> _checkAuthEndpoint() async {
    try {
      final response = await _dio
          .post<dynamic>(
            ApiConfig.authLoginEndpoint,
            data: const <String, Object?>{},
            options: Options(validateStatus: (_) => true),
          )
          .timeout(const Duration(seconds: 5));
      return response.statusCode == 400;
    } catch (e) {
      debugPrint('[HealthCheck] Auth gateway check failed: $e');
      return false;
    }
  }

  Future<bool> _checkRecommendationEndpoint() async {
    try {
      final response = await _dio
          .get<dynamic>(
            '${ApiConfig.recommendationsEndpoint}/1/music',
            options: Options(validateStatus: (_) => true),
          )
          .timeout(const Duration(seconds: 5));
      return response.statusCode != null &&
          response.statusCode! >= 200 &&
          response.statusCode! < 300;
    } catch (e) {
      debugPrint('[HealthCheck] Recommendation endpoint check failed: $e');
      return false;
    }
  }

  /// Test WebSocket connectivity
  Future<bool> testWebSocket() async {
    if (ApiConfig.webSocketPath.trim().isEmpty) return false;
    try {
      final baseUri = Uri.parse(ApiConfig.webSocketUrl);
      final endpointPath = ApiConfig.webSocketPath.replaceFirst(
        RegExp(r'^/+'),
        '',
      );
      final wsUrl = baseUri
          .replace(
            path:
                '${baseUri.path.replaceFirst(RegExp(r'/+$'), '')}/$endpointPath',
          )
          .toString();
      debugPrint('[HealthCheck] Testing WebSocket at: $wsUrl');

      final ws = await WebSocket.connect(wsUrl).timeout(
        const Duration(seconds: 5),
        onTimeout: () => throw TimeoutException('WebSocket connection timeout'),
      );

      await ws.close();
      debugPrint('[HealthCheck] WebSocket is accessible');
      return true;
    } catch (e) {
      debugPrint('[HealthCheck] WebSocket test failed: $e');
      return false;
    }
  }

  /// Comprehensive connectivity check
  Future<ConnectivityReport> runDiagnostics() async {
    final report = ConnectivityReport();

    // Check base URL
    report.baseUrl = ApiConfig.baseUrl;
    report.environment = ApiConfig.environment.toString();

    // Check health endpoint
    try {
      final health = await checkHealth();
      report.backendHealthy = health.isHealthy;
      report.backendVersion = health.version;
      report.services = health.services;
    } catch (e) {
      report.backendHealthy = false;
      report.error = e.toString();
    }

    // Check endpoints
    report.endpoints = await verifyEndpoints();

    // Check WebSocket
    report.webSocketHealthy = ApiConfig.webSocketPath.trim().isEmpty
        ? null
        : await testWebSocket();

    return report;
  }
}

/// Exception thrown during health checks
class HealthCheckException implements Exception {
  final String message;
  final int? statusCode;

  HealthCheckException(this.message, [this.statusCode]);

  @override
  String toString() => 'HealthCheckException: $message (HTTP $statusCode)';
}

/// Report of connectivity and health status
class ConnectivityReport {
  String baseUrl = '';
  String environment = '';
  bool backendHealthy = false;
  String backendVersion = '';
  Map<String, dynamic> services = {};
  Map<String, bool> endpoints = {};
  bool? webSocketHealthy;
  String? error;

  bool get isFullyHealthy =>
      backendHealthy &&
      endpoints.values.every((v) => v) &&
      (webSocketHealthy == null || webSocketHealthy == true);

  @override
  String toString() =>
      '''
ConnectivityReport:
  Base URL: $baseUrl
  Environment: $environment
  Backend Healthy: $backendHealthy
  Backend Version: $backendVersion
  Endpoints: $endpoints
  WebSocket: ${webSocketHealthy == null
          ? 'NOT CONFIGURED'
          : webSocketHealthy!
          ? 'HEALTHY'
          : 'UNHEALTHY'}
  Overall Status: ${isFullyHealthy ? 'HEALTHY' : 'UNHEALTHY'}
  Error: ${error ?? 'None'}
  Services: $services
''';
}
