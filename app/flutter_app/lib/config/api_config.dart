import 'package:flutter/foundation.dart';

enum Environment { development, staging, production }

class ApiConfig {
  ApiConfig._();

  // Environment configuration
  static const _configuredEnvironment = String.fromEnvironment('FLUTTER_ENV', defaultValue: 'development');
  static const _configuredBaseUrl = String.fromEnvironment('API_BASE_URL');
  static const currentUserId = String.fromEnvironment('CURRENT_USER_ID');

  // Kubernetes configuration
  static const _k8sApiGatewayUrl = String.fromEnvironment('K8S_API_GATEWAY_URL');
  static const _k8sNamespace = String.fromEnvironment('K8S_NAMESPACE', defaultValue: 'meloop');
  static const _k8sIngressHost = String.fromEnvironment('K8S_INGRESS_HOST', defaultValue: 'meloop.local');

  static Environment get environment {
    return switch (_configuredEnvironment.toLowerCase()) {
      'staging' => Environment.staging,
      'production' => Environment.production,
      _ => Environment.development,
    };
  }

  /// Base URL for API calls
  static String get baseUrl {
    // 1. Check if explicit URL is configured
    final configuredUrl = _configuredBaseUrl.trim();
    if (configuredUrl.isNotEmpty) {
      return configuredUrl.replaceFirst(RegExp(r'/+$'), '');
    }

    // 2. Check if K8s API Gateway URL is configured
    final k8sUrl = _k8sApiGatewayUrl.trim();
    if (k8sUrl.isNotEmpty) {
      return k8sUrl.replaceFirst(RegExp(r'/+$'), '');
    }

    // 3. Use platform-specific defaults based on environment
    return switch (environment) {
      Environment.production => _productionBaseUrl,
      Environment.staging => _stagingBaseUrl,
      Environment.development => _developmentBaseUrl,
    };
  }

  /// WebSocket URL for real-time features
  static String get webSocketUrl {
    final baseUrl = ApiConfig.baseUrl;
    return baseUrl
        .replaceFirst(RegExp(r'^https?://'), 'ws://')
        .replaceFirst(RegExp(r'^wss?://'), 'ws://');
  }

  /// Development environment URLs
  static String get _developmentBaseUrl {
    if (defaultTargetPlatform == TargetPlatform.android) {
      // Android emulator: 10.0.2.2 is the special alias for host machine
      return 'http://10.0.2.2:8080';
    } else if (defaultTargetPlatform == TargetPlatform.windows) {
      // Windows desktop
      return 'http://localhost:8080';
    }
    // iOS, macOS, web, etc.
    return 'http://localhost:8080';
  }

  /// Staging environment URLs (K8s cluster)
  static String get _stagingBaseUrl {
    // For staging, use the K8s Ingress hostname
    final ingressHost = _k8sIngressHost.trim();
    if (ingressHost.isNotEmpty && ingressHost != 'meloop.local') {
      return 'http://$ingressHost';
    }
    // Fallback: try to use service DNS name within cluster
    return 'http://backend.$_k8sNamespace.svc.cluster.local:8080';
  }

  /// Production environment URLs
  static String get _productionBaseUrl {
    // Use configured K8s Gateway URL or Ingress hostname
    final k8sUrl = _k8sApiGatewayUrl.trim();
    if (k8sUrl.isNotEmpty) {
      return k8sUrl;
    }
    // Fallback to environment-based production URL
    return 'https://api.meloop.com';
  }

  /// API endpoints
  static const String healthCheckEndpoint = '/health';
  static const String apiPrefix = '/api';
  static const String authEndpoint = '$apiPrefix/auth';
  static const String demoEndpoint = '$apiPrefix/demo';
  static const String feedEndpoint = '$apiPrefix/feed';
  static const String postsEndpoint = '$apiPrefix/posts';
  static const String usersEndpoint = '$apiPrefix/users';
  static const String profileEndpoint = '$apiPrefix/profile';
  static const String notificationsEndpoint = '$apiPrefix/notifications';
  static const String musicEndpoint = '$apiPrefix/music';

  /// Request timeouts
  static const Duration connectTimeout = Duration(seconds: 10);
  static const Duration receiveTimeout = Duration(seconds: 10);
  static const Duration sendTimeout = Duration(seconds: 10);

  /// WebSocket configuration
  static const Duration webSocketReconnectDelay = Duration(seconds: 3);
  static const int webSocketMaxReconnectAttempts = 5;

  /// Debugging
  static bool get enableLogging => environment == Environment.development;
}
