import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/models/api_response.dart';
import 'package:flutter_app/models/demo_data.dart';
import 'package:flutter_app/models/feed_data.dart';
import 'package:flutter_app/models/global_search.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/models/music_preferences.dart';
import 'package:flutter_app/models/mutual_friends.dart';
import 'package:flutter_app/services/post_publisher.dart';
import 'package:http/http.dart' as http;

class ApiException implements Exception {
  final int? statusCode;
  final String? errorCode;
  final String message;

  const ApiException({required this.message, this.statusCode, this.errorCode});

  @override
  String toString() => message;
}

class ApiClient implements PostPublisher {
  final http.Client _httpClient;
  final Dio _dio;
  final Duration _timeout;
  final bool _ownsClient;
  final bool _ownsDio;

  ApiClient({
    http.Client? httpClient,
    Dio? dio,
    this._timeout = const Duration(seconds: 10),
  }) : _httpClient = httpClient ?? http.Client(),
       _ownsClient = httpClient == null,
       _dio = dio ?? Dio(),
       _ownsDio = dio == null {
    _dio.options
      ..baseUrl = ApiConfig.baseUrl
      ..connectTimeout = _timeout
      ..receiveTimeout = _timeout
      ..headers['Accept'] = 'application/json';
    final userId = ApiConfig.currentUserId.trim();
    if (userId.isNotEmpty) {
      _dio.options.headers['X-User-ID'] = userId;
    }
  }

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

  Future<ApiResponse<FeedPage>> getFeedPage({
    required String category,
    required int page,
    int pageSize = 5,
  }) {
    final query = Uri(
      queryParameters: {
        'category': category,
        'page': '$page',
        'pageSize': '$pageSize',
      },
    ).query;
    return get('/api/feed?$query', FeedPage.fromJson);
  }

  Future<ApiResponse<MutualFriendsPage>> getMutualFriends(
    String targetUserId, {
    String? cursor,
    int limit = 20,
  }) {
    final queryParameters = <String, String>{'limit': '$limit'};
    if (cursor != null && cursor.isNotEmpty) {
      queryParameters['cursor'] = cursor;
    }
    final query = Uri(queryParameters: queryParameters).query;
    return get(
      '/v1/social/users/${Uri.encodeComponent(targetUserId)}/mutual-friends?$query',
      MutualFriendsPage.fromJson,
    );
  }

  Future<ApiResponse<MusicSearchPage>> searchMusicTracks(
    String query, {
    String type = 'track',
    int offset = 0,
    CancelToken? cancelToken,
  }) {
    return _searchGet(
      '/v1/music/search',
      {'q': query, 'type': type, 'offset': '$offset'},
      (json) => MusicSearchPage.fromJson(json, type: type),
      cancelToken: cancelToken,
    );
  }

  Future<ApiResponse<MusicPreferencesData>> getMusicPreferences() {
    return _musicPreferenceRequest(
      'GET',
      '/v1/music/preferences',
      MusicPreferencesData.fromJson,
    );
  }

  Future<ApiResponse<MusicPreferencesData>> addMusicPreference({
    required String type,
    required String itemId,
    required String name,
    String? artworkUrl,
  }) {
    return _musicPreferenceRequest(
      'POST',
      '/v1/music/preferences',
      MusicPreferencesData.fromJson,
      data: {
        'type': type,
        'item_id': itemId,
        'name': name,
        'artwork_url': ?artworkUrl,
      },
      allowEmptyData: true,
    );
  }

  Future<ApiResponse<MusicPreferencesData>> removeMusicPreference({
    required String type,
    required String itemId,
  }) {
    return _musicPreferenceRequest(
      'DELETE',
      '/v1/music/preferences/${Uri.encodeComponent(type)}/${Uri.encodeComponent(itemId)}',
      MusicPreferencesData.fromJson,
      allowEmptyData: true,
    );
  }

  Future<ApiResponse<T>> _musicPreferenceRequest<T>(
    String method,
    String path,
    T Function(Object? json) parseData, {
    Object? data,
    bool allowEmptyData = false,
  }) async {
    try {
      final response = switch (method) {
        'GET' => await _dio.get<Object?>(path),
        'POST' => await _dio.post<Object?>(path, data: data),
        'DELETE' => await _dio.delete<Object?>(path, data: data),
        _ => throw ArgumentError.value(method, 'method'),
      };
      final body = response.data;
      if (response.statusCode == 204 || body == null) {
        return ApiResponse<T>(success: true);
      }
      if (body is! Map<String, dynamic>) {
        throw const FormatException('Respuesta inválida del API Gateway.');
      }
      final payload = body['data'];
      if (allowEmptyData &&
          body['success'] == true &&
          payload is Map<String, dynamic> &&
          !payload.containsKey('preferences') &&
          !payload.containsKey('songs') &&
          !payload.containsKey('tracks') &&
          !payload.containsKey('artists') &&
          !payload.containsKey('genres')) {
        return ApiResponse<T>(success: true);
      }
      final result = ApiResponse<T>.fromJson(body, parseData);
      if (!result.success) {
        throw ApiException(
          statusCode: response.statusCode,
          errorCode: result.error?.code,
          message:
              result.error?.message ?? 'No se pudo actualizar la preferencia.',
        );
      }
      return result;
    } on ApiException {
      rethrow;
    } on DioException catch (error) {
      final body = error.response?.data;
      final errorBody = body is Map<String, dynamic> ? body['error'] : null;
      final code = errorBody is Map<String, dynamic>
          ? errorBody['code'] as String?
          : null;
      final message = errorBody is Map<String, dynamic>
          ? errorBody['message'] as String?
          : null;
      throw ApiException(
        statusCode: error.response?.statusCode,
        errorCode: code,
        message:
            message ??
            'No se pudo conectar con el servicio de preferencias musicales.',
      );
    } on FormatException catch (error) {
      throw ApiException(message: error.message);
    }
  }

  Future<ApiResponse<SearchUserPage>> searchUsers(
    String query, {
    CancelToken? cancelToken,
  }) {
    return _searchGet(
      '/v1/search/users',
      {'q': query},
      SearchUserPage.fromJson,
      cancelToken: cancelToken,
    );
  }

  Future<ApiResponse<SearchPostPage>> searchPosts(
    String query, {
    required String filter,
    CancelToken? cancelToken,
  }) {
    return _searchGet(
      '/v1/search/posts',
      {'q': query, 'filter': filter},
      SearchPostPage.fromJson,
      cancelToken: cancelToken,
    );
  }

  Future<ApiResponse<Map<String, dynamic>>> sendFriendRequest(
    String targetUserId,
  ) {
    return post('/v1/social/friends/requests', {'id_usuario': targetUserId}, (
      json,
    ) {
      if (json is! Map<String, dynamic>) {
        throw const FormatException(
          'La respuesta de solicitud de amistad es inválida.',
        );
      }
      return json;
    });
  }

  Future<ApiResponse<T>> _searchGet<T>(
    String path,
    Map<String, String> queryParameters,
    T Function(Object? json) parseData, {
    CancelToken? cancelToken,
  }) async {
    try {
      final response = await _dio.get<Object?>(
        path,
        queryParameters: queryParameters,
        cancelToken: cancelToken,
      );
      final body = response.data;
      if (body is! Map<String, dynamic>) {
        throw const FormatException('Respuesta inválida del API Gateway.');
      }

      final result = ApiResponse<T>.fromJson(body, parseData);
      if (response.statusCode == null ||
          response.statusCode! < 200 ||
          response.statusCode! >= 300 ||
          !result.success) {
        throw ApiException(
          statusCode: response.statusCode,
          errorCode: result.error?.code,
          message:
              result.error?.message ??
              'El servidor no pudo completar la búsqueda.',
        );
      }
      return result;
    } on DioException catch (error) {
      if (CancelToken.isCancel(error)) rethrow;
      final body = error.response?.data;
      final decodedBody = body is Map<String, dynamic> ? body : null;
      final errorBody = decodedBody?['error'];
      final code = errorBody is Map<String, dynamic>
          ? errorBody['code'] as String?
          : null;
      final message = errorBody is Map<String, dynamic>
          ? errorBody['message'] as String?
          : null;
      throw ApiException(
        statusCode: error.response?.statusCode,
        errorCode: code,
        message: message ?? 'No se pudo conectar con el servicio de búsqueda. Inténtalo de nuevo.',
      );
    } on FormatException catch (error) {
      throw ApiException(message: error.message);
    }
  }

  @override
  Future<PostPublishResult> publishPost({
    required String content,
    required List<PostImageUpload> images,
    Uint8List? audioBytes,
    String? audioName,
    MusicTrack? song,
    required CancelToken cancelToken,
    required void Function(int sent, int total) onSendProgress,
  }) async {
    final formData = FormData()..fields.add(MapEntry('content', content));
    for (final image in images) {
      formData.files.add(
        MapEntry(
          'images',
          MultipartFile.fromBytes(image.bytes, filename: image.name),
        ),
      );
    }
    if (audioBytes != null && audioName != null) {
      formData.files.add(
        MapEntry(
          'audio',
          MultipartFile.fromBytes(audioBytes, filename: audioName),
        ),
      );
    }
    if (song != null) {
      formData.fields
        ..add(MapEntry('song_id', song.id))
        ..add(MapEntry('song_title', song.title))
        ..add(MapEntry('artist_name', song.artist));
    }

    try {
      final response = await _dio.post<Object?>(
        '/v1/posts',
        data: formData,
        cancelToken: cancelToken,
        onSendProgress: onSendProgress,
      );
      return _parsePostPublishResponse(response.data, response.statusCode);
    } on DioException catch (error) {
      if (CancelToken.isCancel(error)) {
        throw const ApiException(message: 'Subida cancelada.');
      }

      final responseData = error.response?.data;
      final errorCode = _readErrorCode(responseData);
      throw ApiException(
        statusCode: error.response?.statusCode,
        errorCode: errorCode,
        message: _localizedPostError(
          errorCode,
          _readErrorMessage(responseData) ??
              'No se pudo publicar. Inténtalo de nuevo.',
        ),
      );
    } on FormatException catch (error) {
      throw ApiException(message: error.message);
    }
  }

  PostPublishResult _parsePostPublishResponse(Object? body, int? statusCode) {
    if (body is! Map<String, dynamic>) {
      throw const ApiException(
        message: 'El API Gateway devolvió una respuesta inválida.',
      );
    }

    if (statusCode == null || statusCode < 200 || statusCode >= 300) {
      final errorCode = _readErrorCode(body);
      throw ApiException(
        statusCode: statusCode,
        errorCode: errorCode,
        message: _localizedPostError(
          errorCode,
          _readErrorMessage(body) ?? 'No se pudo publicar.',
        ),
      );
    }

    final success = body['success'];
    final data = body['data'];
    if (success != true || data is! Map<String, dynamic>) {
      final errorCode = _readErrorCode(body);
      throw ApiException(
        statusCode: statusCode,
        errorCode: errorCode,
        message: _localizedPostError(
          errorCode,
          _readErrorMessage(body) ?? 'El servidor rechazó la publicación.',
        ),
      );
    }

    final rawPost = data['post'];
    final post = rawPost is Map<String, dynamic> ? rawPost : data;
    final xpAwarded = data['xp_awarded'];
    return PostPublishResult(
      post: post,
      xpAwarded: xpAwarded is int ? xpAwarded : null,
    );
  }

  String? _readErrorCode(Object? body) {
    if (body is! Map<String, dynamic>) return null;
    final error = body['error'];
    if (error is Map<String, dynamic> && error['code'] is String) {
      return error['code'] as String;
    }
    return null;
  }

  String? _readErrorMessage(Object? body) {
    if (body is! Map<String, dynamic>) return null;
    final error = body['error'];
    if (error is Map<String, dynamic> && error['message'] is String) {
      return error['message'] as String;
    }
    return null;
  }

  String _localizedPostError(String? code, String fallback) {
    return switch (code) {
      'MUSIC_NOT_FOUND' => 'No encontramos esa canción en Spotify.',
      'MUSIC_UNAVAILABLE' =>
        'No pudimos validar la canción ahora. Intenta más tarde.',
      'POSTING_RESTRICTED' => 'Tienes una restricción temporal para publicar.',
      _ => fallback,
    };
  }

  Map<String, String> get _headers {
    final headers = <String, String>{
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    };
    final userId = ApiConfig.currentUserId.trim();
    if (userId.isNotEmpty) {
      headers['X-User-ID'] = userId;
    }
    return headers;
  }

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
    if (_ownsDio) _dio.close();
  }
}
