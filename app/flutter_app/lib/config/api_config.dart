import 'package:flutter/foundation.dart';

class ApiConfig {
  ApiConfig._();

  static const _configuredBaseUrl = String.fromEnvironment('API_BASE_URL');
  static const currentUserId = String.fromEnvironment('CURRENT_USER_ID');

  static String get baseUrl {
    final configuredUrl = _configuredBaseUrl.trim();
    if (configuredUrl.isNotEmpty) {
      return configuredUrl.replaceFirst(RegExp(r'/+$'), '');
    }

    if (defaultTargetPlatform == TargetPlatform.android) {
      return 'http://10.0.2.2:8080';
    }

    return 'http://localhost:8080';
  }
}
