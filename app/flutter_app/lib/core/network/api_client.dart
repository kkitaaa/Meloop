import 'dart:async';
import 'dart:convert';

import 'package:flutter_app/core/errors/api_exception.dart';
import 'package:flutter_app/core/network/api_config.dart';
import 'package:flutter_app/core/network/api_response.dart';
import 'package:http/http.dart' as http;

/// Cliente HTTP compartido para los recursos expuestos por el API Gateway.
class ApiClient {
  ApiClient({http.Client? httpClient, String? baseUrl})
      : _httpClient = httpClient ?? http.Client(),
        _baseUri = Uri.parse(baseUrl ?? ApiConfig.baseUrl);

  final http.Client _httpClient;
  final Uri _baseUri;

  Future<ApiResponse<T>> get<T>(
    String path, {
    Map<String, String>? queryParameters,
    required T Function(Object? json) fromData,
  }) =>
      _request<T>(
        'GET',
        path,
        queryParameters: queryParameters,
        fromData: fromData,
      );

  Future<ApiResponse<T>> post<T>(
    String path, {
    Object? body,
    required T Function(Object? json) fromData,
  }) =>
      _request<T>('POST', path, body: body, fromData: fromData);

  Future<ApiResponse<T>> put<T>(
    String path, {
    Object? body,
    required T Function(Object? json) fromData,
  }) =>
      _request<T>('PUT', path, body: body, fromData: fromData);

  Future<ApiResponse<T>> delete<T>(
    String path, {
    Object? body,
    required T Function(Object? json) fromData,
  }) =>
      _request<T>('DELETE', path, body: body, fromData: fromData);

  Future<ApiResponse<T>> _request<T>(
    String method,
    String path, {
    Map<String, String>? queryParameters,
    Object? body,
    required T Function(Object? json) fromData,
  }) async {
    final uri = _buildUri(path, queryParameters);
    final request = http.Request(method, uri)
      ..headers['Accept'] = 'application/json';

    if (body != null) {
      request.headers['Content-Type'] = 'application/json';
      request.body = jsonEncode(body);
    }

    try {
      final streamedResponse = await _httpClient
          .send(request)
          .timeout(ApiConfig.requestTimeout);
      final response = await http.Response.fromStream(streamedResponse);
      final json = _decodeJson(response.body);
      final apiResponse = ApiResponse<T>.fromJson(json, fromData);

      if (response.statusCode < 200 || response.statusCode >= 300) {
        final error = apiResponse.error;
        throw ApiException(
          message: error?.message ?? 'La solicitud no pudo completarse.',
          statusCode: response.statusCode,
          code: error?.code,
          details: error?.details,
        );
      }

      if (!apiResponse.success) {
        final error = apiResponse.error;
        throw ApiException(
          message: error?.message ?? 'El servidor devolvió una respuesta no exitosa.',
          code: error?.code,
          details: error?.details,
        );
      }

      return apiResponse;
    } on TimeoutException {
      throw const ApiException(
        message: 'La solicitud tardó demasiado tiempo.',
      );
    } on http.ClientException catch (error) {
      throw ApiException(
        message: 'No se pudo conectar al servidor.',
        details: error,
      );
    } on FormatException catch (error) {
      throw ApiException(
        message: 'El servidor devolvió una respuesta inválida.',
        details: error,
      );
    }
  }

  Uri _buildUri(String path, Map<String, String>? queryParameters) {
    final normalizedPath = path.startsWith('/') ? path.substring(1) : path;
    return _baseUri
        .resolve(normalizedPath)
        .replace(queryParameters: queryParameters);
  }

  Map<String, dynamic> _decodeJson(String body) {
    final decoded = jsonDecode(body);
    if (decoded is! Map<String, dynamic>) {
      throw const FormatException('La respuesta no es un objeto JSON.');
    }
    return decoded;
  }

  void close() => _httpClient.close();
}
