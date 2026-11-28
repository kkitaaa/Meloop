import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_app/config/api_config.dart';

/// Manages HTTP client configuration and retry logic
class HttpClientManager {
  late final Dio dio;

  HttpClientManager() {
    dio = _configureDio();
  }

  Dio _configureDio() {
    final dioInstance = Dio(
      BaseOptions(
        baseUrl: ApiConfig.baseUrl,
        connectTimeout: ApiConfig.connectTimeout,
        receiveTimeout: ApiConfig.receiveTimeout,
        sendTimeout: ApiConfig.sendTimeout,
        headers: {
          'Accept': 'application/json',
          'Content-Type': 'application/json',
        },
        validateStatus: (status) => status != null && status < 500,
      ),
    );

    // Add logging interceptor in development
    if (ApiConfig.enableLogging) {
      dioInstance.interceptors.add(
        LogInterceptor(
          requestBody: true,
          responseBody: true,
          requestHeader: true,
          responseHeader: true,
          error: true,
          logPrint: (object) => debugPrint(object.toString()),
        ),
      );
    }

    // Add retry interceptor
    final retryInterceptor = _RetryInterceptor();
    dioInstance.interceptors.add(retryInterceptor);
    retryInterceptor.setDio(dioInstance);

    return dioInstance;
  }

  /// Update base URL dynamically (useful for K8s endpoint changes)
  void updateBaseUrl(String newUrl) {
    dio.options.baseUrl = newUrl.replaceFirst(RegExp(r'/+$'), '');
  }

  /// Set authorization token
  void setAuthToken(String token) {
    dio.options.headers['Authorization'] = 'Bearer $token';
  }

  /// Remove authorization token
  void clearAuthToken() {
    dio.options.headers.remove('Authorization');
  }

  /// Set user ID header
  void setUserId(String userId) {
    dio.options.headers['X-User-ID'] = userId;
  }

  void dispose() {
    dio.close();
  }
}

/// Retry interceptor with exponential backoff
class _RetryInterceptor extends Interceptor {
  static const int maxRetries = 3;
  static const Duration initialDelay = Duration(milliseconds: 100);

  late Dio _dio;

  void setDio(Dio dio) {
    _dio = dio;
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    // Only retry on specific error types
    if (_shouldRetry(err) && err.requestOptions.extra['retryCount'] == null) {
      err.requestOptions.extra['retryCount'] = 0;
    }

    int retryCount = err.requestOptions.extra['retryCount'] ?? 0;

    if (_shouldRetry(err) && retryCount < maxRetries) {
      retryCount++;
      err.requestOptions.extra['retryCount'] = retryCount;

      // Exponential backoff: 100ms, 200ms, 400ms
      final delay = initialDelay * (1 << (retryCount - 1));
      await Future.delayed(delay);

      try {
        final response = await _dio.fetch<dynamic>(err.requestOptions);
        return handler.resolve(response);
      } on DioException catch (e) {
        return handler.next(e);
      }
    }

    return handler.next(err);
  }

  bool _shouldRetry(DioException error) {
    // Retry on network errors and 5xx errors
    return error.type == DioExceptionType.connectionTimeout ||
        error.type == DioExceptionType.receiveTimeout ||
        error.type == DioExceptionType.sendTimeout ||
        (error.response?.statusCode ?? 0) >= 500;
  }
}
