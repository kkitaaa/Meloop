import 'dart:async';
import 'dart:io';
import 'package:dio/dio.dart';
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

  bool get isHealthy => status.toLowerCase() == 'ok' || status.toLowerCase() == 'healthy';
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
        isHealthy: (data['status'] ?? 'unknown').toString().toLowerCase() == 'ok',
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

  HealthCheckService({Dio? dio}) : _dio = dio ?? Dio();

  /// Check overall backend health
  Future<HealthCheckResponse> checkHealth() async {
    try {
      final stopwatch = Stopwatch()..start();

      final response = await _dio.get<Map<String, dynamic>>(
        '${ApiConfig.baseUrl}${ApiConfig.healthCheckEndpoint}',
      ).timeout(
        const Duration(seconds: 5),
        onTimeout: () => throw TimeoutException('Health check timeout'),
      );

      stopwatch.stop();

      if (response.statusCode == 200 && response.data != null) {
        final health = HealthCheckResponse.fromJson(response.data!);
        print('[HealthCheck] Backend is healthy (${stopwatch.elapsedMilliseconds}ms)');
        return health;
      } else {
        throw HealthCheckException(
          'Health check returned status ${response.statusCode}',
          response.statusCode,
        );
      }
    } on DioException catch (e) {
      print('[HealthCheck] DioException: ${e.message}');
      throw HealthCheckException(
        'Failed to connect to backend: ${e.message}',
        e.response?.statusCode,
      );
    } on TimeoutException catch (e) {
      print('[HealthCheck] Timeout: ${e.message}');
      throw HealthCheckException(e.message ?? 'Timeout');
    } catch (e) {
      print('[HealthCheck] Error: $e');
      throw HealthCheckException('Health check failed: $e');
    }
  }

  /// Check health of a specific API endpoint
  Future<bool> checkEndpoint(String endpoint) async {
    try {
      final response = await _dio.head<dynamic>(
        '${ApiConfig.baseUrl}$endpoint',
      ).timeout(const Duration(seconds: 5));

      return response.statusCode == 200 || response.statusCode == 204;
    } catch (e) {
      print('[HealthCheck] Endpoint check failed for $endpoint: $e');
      return false;
    }
  }

  /// Verify all major endpoints are accessible
  Future<Map<String, bool>> verifyEndpoints() async {
    final endpoints = {
      'health': ApiConfig.healthCheckEndpoint,
      'auth': ApiConfig.authEndpoint,
      'feed': ApiConfig.feedEndpoint,
      'posts': ApiConfig.postsEndpoint,
      'users': ApiConfig.usersEndpoint,
      'notifications': ApiConfig.notificationsEndpoint,
      'music': ApiConfig.musicEndpoint,
    };

    final results = <String, bool>{};

    for (final entry in endpoints.entries) {
      results[entry.key] = await checkEndpoint(entry.value);
      await Future.delayed(const Duration(milliseconds: 100)); // Rate limiting
    }

    return results;
  }

  /// Test WebSocket connectivity
  Future<bool> testWebSocket() async {
    try {
      final wsUrl = ApiConfig.webSocketUrl;
      print('[HealthCheck] Testing WebSocket at: $wsUrl');

      final ws = await WebSocket.connect(wsUrl).timeout(
        const Duration(seconds: 5),
        onTimeout: () => throw TimeoutException('WebSocket connection timeout'),
      );

      await ws.close();
      print('[HealthCheck] WebSocket is accessible');
      return true;
    } catch (e) {
      print('[HealthCheck] WebSocket test failed: $e');
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
    report.webSocketHealthy = await testWebSocket();

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
  bool webSocketHealthy = false;
  String? error;

  bool get isFullyHealthy => 
      backendHealthy && 
      endpoints.values.every((v) => v) && 
      webSocketHealthy;

  @override
  String toString() => '''
ConnectivityReport:
  Base URL: $baseUrl
  Environment: $environment
  Backend Healthy: $backendHealthy
  Backend Version: $backendVersion
  Endpoints: $endpoints
  WebSocket Healthy: $webSocketHealthy
  Overall Status: ${isFullyHealthy ? 'HEALTHY' : 'UNHEALTHY'}
  Error: ${error ?? 'None'}
  Services: $services
''';
}
