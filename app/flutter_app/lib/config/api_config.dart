import 'package:flutter/foundation.dart';

enum Environment { development, staging, production }

class ApiConfig {
  ApiConfig._();

  static const _configuredEnvironment = String.fromEnvironment(
    'FLUTTER_ENV',
    defaultValue: 'development',
  );
  static const _configuredBaseUrl = String.fromEnvironment('API_BASE_URL');
  static const _k8sApiGatewayUrl = String.fromEnvironment(
    'K8S_API_GATEWAY_URL',
  );
  static const _k8sIngressHost = String.fromEnvironment('K8S_INGRESS_HOST');
  static const webSocketPath = String.fromEnvironment('WEBSOCKET_PATH');
  static const currentUserId = String.fromEnvironment('CURRENT_USER_ID');

  static Environment get environment {
    return switch (_configuredEnvironment.toLowerCase()) {
      'staging' => Environment.staging,
      'production' => Environment.production,
      _ => Environment.development,
    };
  }

  static String get baseUrl {
    final configuredUrl = _configuredBaseUrl.trim();
    final gatewayUrl = _k8sApiGatewayUrl.trim();
    final resolved = configuredUrl.isNotEmpty
        ? _normalizeHttpUrl(configuredUrl)
        : gatewayUrl.isNotEmpty
        ? _normalizeHttpUrl(gatewayUrl)
        : switch (environment) {
            Environment.development => _developmentBaseUrl,
            Environment.staging => _configuredIngressUrl('staging'),
            Environment.production => _configuredIngressUrl('production'),
          };
    if (environment == Environment.production &&
        Uri.parse(resolved).scheme != 'https') {
      throw StateError('El entorno production requiere una API URL HTTPS.');
    }
    return resolved;
  }

  static String get webSocketUrl => webSocketUrlFor(baseUrl);

  @visibleForTesting
  static String webSocketUrlFor(String httpUrl) {
    final uri = Uri.parse(_normalizeHttpUrl(httpUrl));
    return uri.replace(scheme: uri.scheme == 'https' ? 'wss' : 'ws').toString();
  }

  static String get _developmentBaseUrl {
    if (defaultTargetPlatform == TargetPlatform.android) {
      return 'http://10.0.2.2:8080';
    }
    return 'http://localhost:8080';
  }

  static String _configuredIngressUrl(String environmentName) {
    final host = _k8sIngressHost.trim();
    if (host.isEmpty) {
      throw StateError(
        'Configura API_BASE_URL o K8S_API_GATEWAY_URL para el entorno '
        '$environmentName. No se usará un dominio de ejemplo.',
      );
    }
    return _normalizeHttpUrl(host.contains('://') ? host : 'https://$host');
  }

  static String _normalizeHttpUrl(String value) {
    final uri = Uri.tryParse(value);
    if (uri == null ||
        !{'http', 'https'}.contains(uri.scheme.toLowerCase()) ||
        uri.host.isEmpty ||
        uri.userInfo.isNotEmpty ||
        uri.hasQuery ||
        uri.hasFragment) {
      throw FormatException(
        'API URL debe ser una URL HTTP(S) absoluta sin credenciales, query o fragmento.',
        value,
      );
    }

    final normalizedPath = uri.path.replaceFirst(RegExp(r'/+$'), '');
    return uri.replace(path: normalizedPath).toString();
  }

  static const String healthCheckEndpoint = '/health';
  static const String apiPrefix = '/api';
  static const String authEndpoint = '/auth';
  static const String authLoginEndpoint = '$authEndpoint/login';
  static const String authRegisterEndpoint = '$authEndpoint/register';
  static const String authRefreshEndpoint = '$authEndpoint/refresh';
  static const String authLogoutEndpoint = '$authEndpoint/logout';
  static const String authValidateEndpoint = '$authEndpoint/validate';
  static const String demoEndpoint = '$apiPrefix/demo';
  static const String feedEndpoint = '$apiPrefix/feed';
  static const String usersEndpoint = '/users';
  static const String recommendationsEndpoint = '/recommendations';

  static const Duration connectTimeout = Duration(seconds: 10);
  static const Duration receiveTimeout = Duration(seconds: 10);
  static const Duration sendTimeout = Duration(seconds: 10);
  static const Duration sessionRefreshThreshold = Duration(minutes: 2);

  static const Duration webSocketReconnectDelay = Duration(seconds: 3);
  static const int webSocketMaxReconnectAttempts = 5;

  static bool get enableLogging => environment == Environment.development;
}
