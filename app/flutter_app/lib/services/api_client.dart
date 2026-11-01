import 'dart:async';
import 'dart:convert';

import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/models/api_response.dart';
import 'package:flutter_app/models/demo_data.dart';
import 'package:http/http.dart' as http;

class ApiException implements Exception {
  final int? statusCode;
  final String message;

  const ApiException({required this.message, this.statusCode});

  @override
  String toString() => message;
}

class ApiClient {
  final http.Client _httpClient;
  final Duration _timeout;
  final bool _ownsClient;

  ApiClient({
    http.Client? httpClient,
    this._timeout = const Duration(seconds: 10),
  }) : _httpClient = httpClient ?? http.Client(),
       _ownsClient = httpClient == null;

  Future<ApiResponse<T>> get<T>(
    String path,
    T Function(Object? json) parseData,
  ) {
    return _send(
      () => _httpClient.get(_uriFor(path), headers: _headers),
      parseData,
    );
  }

  Future<ApiResponse<T>> post<T>(
    String path,
    Object body,
    T Function(Object? json) parseData,
  ) {
    return _send(
      () => _httpClient.post(
        _uriFor(path),
        headers: _headers,
        body: jsonEncode(body),
      ),
      parseData,
    );
  }

  Future<ApiResponse<DemoData>> getDemo() =>
      get('/api/demo', DemoData.fromJson);

  Future<ApiResponse<DemoPostData>> postDemo(String message) =>
      post('/api/demo', {'message': message}, DemoPostData.fromJson);

  Map<String, String> get _headers => const {
    'Accept': 'application/json',
    'Content-Type': 'application/json',
  };

  Uri _uriFor(String path) {
    final normalizedPath = path.startsWith('/') ? path.substring(1) : path;
    return Uri.parse('${ApiConfig.baseUrl}/').resolve(normalizedPath);
  }

  Future<ApiResponse<T>> _send<T>(
    Future<http.Response> Function() send,
    T Function(Object? json) parseData,
  ) async {
    try {
      final response = await send().timeout(_timeout);
      final decoded = jsonDecode(response.body);
      if (decoded is! Map<String, dynamic>) {
        throw const FormatException(
          'La respuesta del servidor no es un JSON válido.',
        );
      }

      final result = ApiResponse<T>.fromJson(decoded, parseData);
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw ApiException(
          statusCode: response.statusCode,
          message:
              result.error?.message ??
              'El servidor respondió con el estado ${response.statusCode}.',
        );
      }
      if (!result.success) {
        throw ApiException(
          statusCode: response.statusCode,
          message: result.error?.message ?? 'El servidor rechazó la solicitud.',
        );
      }

      return result;
    } on ApiException {
      rethrow;
    } on TimeoutException {
      throw const ApiException(
        message: 'El servidor tardó demasiado en responder.',
      );
    } on http.ClientException {
      throw const ApiException(
        message: 'No se pudo conectar con el API Gateway. Revisa la URL y que esté activo.',
      );
    } on FormatException {
      throw const ApiException(
        message: 'El API Gateway devolvió una respuesta inválida.',
      );
    }
  }

  void close() {
    if (_ownsClient) {
      _httpClient.close();
    }
  }
}
