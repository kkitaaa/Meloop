import 'dart:async';
import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/models/account_settings.dart';
import 'package:flutter_app/models/api_response.dart';
import 'package:flutter_app/models/demo_data.dart';
import 'package:flutter_app/models/feed_data.dart';
import 'package:flutter_app/models/global_search.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/models/music_preferences.dart';
import 'package:flutter_app/models/music_detail.dart';
import 'package:flutter_app/models/mutual_friends.dart';
import 'package:flutter_app/models/saved_posts.dart';
import 'package:flutter_app/services/auth_session.dart';
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
  final AuthSession _session;
  final Environment _environment;
  Environment get environment => _environment;

  ApiClient({
    http.Client? httpClient,
    Dio? dio,
    AuthSession? session,
    Environment? environment,
    this._timeout = const Duration(seconds: 10),
  }) : _httpClient = httpClient ?? http.Client(),
       _ownsClient = httpClient == null,
       _dio = dio ?? Dio(),
       _ownsDio = dio == null,
       _session = session ?? AuthSession.shared,
       _environment = environment ?? ApiConfig.environment {
    _dio.options
      ..baseUrl = ApiConfig.baseUrl
      ..connectTimeout = _timeout
      ..receiveTimeout = _timeout
      ..headers['Accept'] = 'application/json';
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          try {
            await _refreshSessionIfNeeded();
          } on Exception catch (error) {
            handler.reject(DioException(requestOptions: options, error: error));
            return;
          }

          final token = _session.token;
          if (token != null && token.isNotEmpty) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          final userId = _session.user?.id ?? ApiConfig.currentUserId.trim();
          if (userId.isNotEmpty) {
            options.headers['X-User-ID'] = userId;
          }
          handler.next(options);
        },
      ),
    );
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

  Future<ApiResponse<T>> patch<T>(
    String path,
    Object body,
    T Function(Object? json) parseData,
  ) {
    return _send(
      () => _httpClient.patch(
        _uriFor(path),
        headers: _headers,
        body: jsonEncode(body),
      ),
      parseData,
    );
  }

  Future<Map<String, dynamic>> updatePrivacySetting(
    String field,
    String value,
  ) async {
    final response = await patch('/users/me/privacy', {
      field: value,
    }, _settingsObject);
    return response.data!;
  }

  Future<List<Map<String, dynamic>>> getNotificationSettings() async {
    final response = await get(
      '/users/me/notifications/settings',
      _notificationSettings,
    );
    return response.data!;
  }

  Future<void> updateNotificationSetting({
    required String type,
    required bool enabled,
  }) async {
    await patch(
      '/users/me/notifications/settings/${Uri.encodeComponent(type)}',
      {'habilitada': enabled},
      _settingsObject,
    );
  }

  Future<AuthUser> login({
    required String email,
    required String password,
  }) async {
    final response = await _send<_SessionResponse>(
      () => _httpClient.post(
        _uriFor(ApiConfig.authLoginEndpoint),
        headers: _headers,
        body: jsonEncode({'email': email.trim(), 'password': password}),
      ),
      _SessionResponse.fromJson,
      refreshSession: false,
      retryUnauthorized: false,
    );
    final data = response.data!;
    await _session.establish(
      sessionToken: data.sessionToken,
      expiresInSeconds: data.expiresIn,
      sessionUser: data.user,
    );
    return data.user;
  }

  Future<void> register({
    required String username,
    required String email,
    required String password,
  }) async {
    await _send<Object?>(
      () => _httpClient.post(
        _uriFor(ApiConfig.authRegisterEndpoint),
        headers: _headers,
        body: jsonEncode({
          'username': username.trim(),
          'email': email.trim(),
          'password': password,
        }),
      ),
      (_) => null,
      refreshSession: false,
      retryUnauthorized: false,
    );
  }

  Future<AuthUser> validateSession() async {
    final response = await get(
      ApiConfig.authValidateEndpoint,
      AuthUser.fromJson,
    );
    return response.data!;
  }

  Future<ApiResponse<Object?>> changePassword({
    required String currentPassword,
    required String newPassword,
  }) {
    return _send(
      () => _httpClient.post(
        _uriFor(ApiConfig.authChangePasswordEndpoint),
        headers: _headers,
        body: jsonEncode({
          'current_password': currentPassword,
          'new_password': newPassword,
        }),
      ),
      (json) => json,
    );
  }

  Future<ApiResponse<AccountDetails>> getAccountDetails() =>
      get(ApiConfig.currentAccountEndpoint, AccountDetails.fromJson);

  Future<ApiResponse<PrivacySettings>> getPrivacySettings() =>
      get(ApiConfig.currentPrivacyEndpoint, PrivacySettings.fromJson);

  Future<ApiResponse<PrivacySettings>> updatePrivacySettings(
    PrivacySettingsPatch changes,
  ) {
    final payload = changes.toJson();
    if (payload.isEmpty) {
      throw ArgumentError.value(changes, 'changes', 'No privacy fields given');
    }
    return _send(
      () => _httpClient.patch(
        _uriFor(ApiConfig.currentPrivacyEndpoint),
        headers: _headers,
        body: jsonEncode(payload),
      ),
      PrivacySettings.fromJson,
    );
  }

  Future<ApiResponse<List<NotificationPreference>>>
  getNotificationPreferences() => get(
    ApiConfig.notificationSettingsEndpoint,
    NotificationPreference.listFromJson,
  );

  Future<ApiResponse<NotificationPreference>> updateNotificationPreference({
    required String type,
    required bool enabled,
  }) {
    final endpoint =
        '${ApiConfig.notificationSettingsEndpoint}/${Uri.encodeComponent(type)}';
    return _send(
      () => _httpClient.patch(
        _uriFor(endpoint),
        headers: _headers,
        body: jsonEncode({'habilitada': enabled}),
      ),
      NotificationPreference.fromJson,
    );
  }

  Future<void> logout() async {
    try {
      if (_session.isAuthenticated) {
        await _send<Object?>(
          () => _httpClient.post(
            _uriFor(ApiConfig.authLogoutEndpoint),
            headers: _headers,
          ),
          (_) => null,
          refreshSession: false,
          retryUnauthorized: false,
        );
      }
    } finally {
      await _session.clear();
    }
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
    final isDevelopment = _environment == Environment.development;
    final queryParameters = <String, String>{
      'page': '$page',
      if (isDevelopment) 'category': category,
      if (isDevelopment) 'pageSize': '$pageSize',
      if (!isDevelopment) 'limit': '$pageSize',
    };

    return _send(
      () => _httpClient.get(
        _uriFor(ApiConfig.feedEndpointFor(_environment))
            .replace(queryParameters: queryParameters),
        headers: _headers,
      ),
      FeedPage.fromJson,
      allowUnwrappedResponse: !isDevelopment,
    );
  }

  Future<List<Map<String, dynamic>>> getAuthorPosts(
    String authorId, {
    int limit = 10,
    int offset = 0,
  }) async {
    if (authorId.trim().isEmpty) {
      throw const ApiException(
        message: 'No se pudo identificar al autor de las publicaciones.',
      );
    }

    try {
      final response = await _dio.get<Object?>(
        '/api/v1/posts/author/${Uri.encodeComponent(authorId)}',
        queryParameters: {'limit': limit, 'offset': offset},
      );
      final body = response.data;
      final posts = body is List
          ? body
          : body is Map<String, dynamic> && body['data'] is List
          ? body['data'] as List
          : null;
      if (posts == null) {
        throw ApiException(
          statusCode: response.statusCode,
          message: 'El servicio devolvió una lista de publicaciones inválida.',
        );
      }

      return posts.map((post) {
        if (post is! Map<String, dynamic>) {
          throw ApiException(
            statusCode: response.statusCode,
            message: 'El servicio devolvió una publicación inválida.',
          );
        }
        return post;
      }).toList();
    } on DioException catch (error) {
      throw ApiException(
        statusCode: error.response?.statusCode,
        errorCode: _readErrorCode(error.response?.data),
        message:
            _readErrorMessage(error.response?.data) ??
            'No se pudieron cargar las publicaciones del perfil.',
      );
    }
  }

  Future<void> setPostLike({
    required String postId,
    required bool liked,
  }) async {
    final normalizedPostId = postId.trim();
    if (normalizedPostId.isEmpty) {
      throw const ApiException(
        message: 'No se pudo identificar la publicación.',
      );
    }

    final path = '/v1/posts/${Uri.encodeComponent(normalizedPostId)}/like';
    try {
      if (liked) {
        await _dio.post<Object?>(path);
      } else {
        await _dio.delete<Object?>(path);
      }
    } on DioException catch (error) {
      final responseData = error.response?.data;
      throw ApiException(
        statusCode: error.response?.statusCode,
        errorCode: _readErrorCode(responseData),
        message:
            _readErrorMessage(responseData) ??
            (liked
                ? 'No se pudo guardar el like. Inténtalo de nuevo.'
                : 'No se pudo quitar el like. Inténtalo de nuevo.'),
      );
    }
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

  Future<ApiResponse<SavedPostPage>> getSavedPosts({
    required int offset,
    int limit = 20,
  }) {
    return _requestEnvelope(
      'GET',
      '/v1/me/saved',
      SavedPostPage.fromJson,
      queryParameters: {'offset': '$offset', 'limit': '$limit'},
    );
  }

  Future<void> savePost(String postId) async {
    await _requestEnvelope<Map<String, dynamic>>(
      'POST',
      '/v1/me/saved',
      _emptyObject,
      data: {'post_id': postId},
      allowEmptyData: true,
    );
  }

  Future<void> removeSavedPost(String postId) async {
    await _requestEnvelope<Map<String, dynamic>>(
      'DELETE',
      '/v1/me/saved/${Uri.encodeComponent(postId)}',
      _emptyObject,
      allowEmptyData: true,
    );
  }

  Future<ApiResponse<SearchUserPage>> getFriends({int limit = 100}) {
    return _requestEnvelope(
      'GET',
      '/v1/social/friends',
      SearchUserPage.fromJson,
      queryParameters: {'limit': '$limit'},
    );
  }

  Future<void> sharePost({
    required String postId,
    required String recipientId,
  }) async {
    try {
      await _requestEnvelope<Map<String, dynamic>>(
        'POST',
        '/v1/posts/${Uri.encodeComponent(postId)}/share',
        _emptyObject,
        data: {'recipient_id': recipientId},
        allowEmptyData: true,
      );
    } on ApiException catch (error) {
      if (error.errorCode == 'NOT_FRIENDS' || error.errorCode == 'BLOCKED') {
        throw const ApiException(
          statusCode: 403,
          errorCode: 'SHARE_RESTRICTED',
          message: 'No puedes compartir con esta persona.',
        );
      }
      rethrow;
    }
  }

  Map<String, dynamic> _emptyObject(Object? json) {
    if (json == null) return {};
    if (json is Map<String, dynamic>) return json;
    throw const FormatException(
      'El API Gateway devolvió una respuesta inválida.',
    );
  }

  Map<String, dynamic> _settingsObject(Object? json) {
    if (json is Map<String, dynamic>) return json;
    throw const FormatException(
      'El API Gateway devolvió una configuración inválida.',
    );
  }

  List<Map<String, dynamic>> _notificationSettings(Object? json) {
    if (json is! List) {
      throw const FormatException(
        'El API Gateway devolvió preferencias de notificación inválidas.',
      );
    }
    return json
        .map((item) {
          if (item is! Map<String, dynamic>) {
            throw const FormatException(
              'El API Gateway devolvió una preferencia inválida.',
            );
          }
          return item;
        })
        .toList(growable: false);
  }

  Future<ApiResponse<T>> _requestEnvelope<T>(
    String method,
    String path,
    T Function(Object? json) parseData, {
    Object? data,
    Map<String, String>? queryParameters,
    bool allowEmptyData = false,
  }) async {
    try {
      final response = await _dio.request<Object?>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: Options(method: method),
      );
      if (response.statusCode == 204 || response.data == null) {
        if (allowEmptyData) return ApiResponse<T>(success: true);
        throw const FormatException(
          'El API Gateway devolvió una respuesta vacía.',
        );
      }
      if (response.data is! Map<String, dynamic>) {
        throw const FormatException(
          'El API Gateway devolvió una respuesta inválida.',
        );
      }
      final result = ApiResponse<T>.fromJson(
        response.data as Map<String, dynamic>,
        parseData,
      );
      if (!result.success) {
        throw ApiException(
          statusCode: response.statusCode,
          errorCode: result.error?.code,
          message: result.error?.message ?? 'El servidor rechazó la solicitud.',
        );
      }
      if (allowEmptyData && result.data == null) {
        return ApiResponse<T>(success: true);
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
            message ?? 'No se pudo completar la solicitud. Inténtalo de nuevo.',
      );
    } on FormatException catch (error) {
      throw ApiException(message: error.message);
    }
  }

  Future<ApiResponse<MusicSearchPage>> searchMusicTracks(
    String query, {
    String type = 'track',
    int offset = 0,
    int limit = 10,
    CancelToken? cancelToken,
  }) {
    if (offset < 0 || offset > 1000) {
      throw RangeError.range(offset, 0, 1000, 'offset');
    }
    if (limit < 1 || limit > 20) {
      throw RangeError.range(limit, 1, 20, 'limit');
    }

    return _musicCatalogGet(
      '/v1/music/search',
      {'q': query.trim(), 'type': type, 'offset': '$offset', 'limit': '$limit'},
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

  Future<ApiResponse<MusicDetail>> getMusicDetail({
    required String type,
    required String spotifyId,
  }) {
    final resource = switch (type) {
      'track' || 'tracks' => 'tracks',
      'artist' || 'artists' => 'artists',
      'album' || 'albums' => 'albums',
      _ => throw ArgumentError.value(
        type,
        'type',
        'Tipo musical no compatible.',
      ),
    };
    final normalizedType = switch (resource) {
      'tracks' => 'track',
      'artists' => 'artist',
      _ => 'album',
    };
    return _musicCatalogGet(
      '/v1/music/$resource/${Uri.encodeComponent(spotifyId)}',
      const {},
      (json) => MusicDetail.fromJson(json, type: normalizedType),
    );
  }

  Future<ApiResponse<T>> _musicCatalogGet<T>(
    String path,
    Map<String, String> queryParameters,
    T Function(Object? json) parseData, {
    CancelToken? cancelToken,
  }) async {
    try {
      final response = await _dio.get<Object?>(
        path,
        queryParameters: queryParameters.isEmpty ? null : queryParameters,
        cancelToken: cancelToken,
      );
      final body = response.data;
      if (body is! Map<String, dynamic>) {
        throw const FormatException('Respuesta inválida del servicio musical.');
      }

      if (response.statusCode == null ||
          response.statusCode! < 200 ||
          response.statusCode! >= 300) {
        throw ApiException(
          statusCode: response.statusCode,
          errorCode: _readErrorCode(body),
          message:
              _readErrorMessage(body) ??
              'El servicio musical respondió con un error.',
        );
      }

      if (body.containsKey('success')) {
        final result = ApiResponse<T>.fromJson(body, parseData);
        if (!result.success) {
          throw ApiException(
            statusCode: response.statusCode,
            errorCode: result.error?.code,
            message:
                result.error?.message ??
                'El servicio musical no pudo completar la solicitud.',
          );
        }
        return result;
      }

      return ApiResponse<T>(success: true, data: parseData(body));
    } on ApiException {
      rethrow;
    } on DioException catch (error) {
      if (CancelToken.isCancel(error)) rethrow;
      final responseData = error.response?.data;
      throw ApiException(
        statusCode: error.response?.statusCode,
        errorCode: _readErrorCode(responseData),
        message:
            _readErrorMessage(responseData) ??
            'No se pudo conectar con el servicio musical. Inténtalo de nuevo.',
      );
    } on FormatException catch (error) {
      throw ApiException(message: error.message);
    }
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
    if (images.isNotEmpty ||
        audioBytes != null ||
        audioName != null ||
        song != null) {
      throw const ApiException(
        message: 'Por ahora solo se pueden publicar textos, sin imágenes, audio ni canciones.',
      );
    }

    try {
      final response = await _dio.post<Object?>(
        '/api/v1/posts',
        data: {'content': content, 'media_urls': <String>[], 'music_id': ''},
        options: Options(contentType: Headers.jsonContentType),
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
    if (success == true && data is Map<String, dynamic>) {
      final rawPost = data['post'];
      final post = rawPost is Map<String, dynamic> ? rawPost : data;
      final xpAwarded = data['xp_awarded'];
      return PostPublishResult(
        post: post,
        xpAwarded: xpAwarded is int ? xpAwarded : null,
      );
    }

    if (statusCode == 201 && body['id'] != null && body['content'] is String) {
      return PostPublishResult(post: body);
    }

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
    if (error is String) return error;
    if (error is Map<String, dynamic> && error['message'] is String) {
      return error['message'] as String;
    }
    return body['message'] is String ? body['message'] as String : null;
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

  Future<void> _refreshSessionIfNeeded() async {
    if (_session.shouldRefresh) await _refreshSession();
  }

  Future<void> _refreshSession({bool force = false}) async {
    final token = _session.token;
    if (token == null || (!force && !_session.shouldRefresh)) return;

    await _session.refreshOnce(() async {
      final currentToken = _session.token;
      if (currentToken == null || (!force && !_session.shouldRefresh)) {
        return;
      }

      try {
        final response = await _httpClient
            .post(
              _uriFor(ApiConfig.authRefreshEndpoint),
              headers: {..._headers, 'Authorization': 'Bearer $currentToken'},
            )
            .timeout(_timeout);
        final result = _decodeResponse(response, _SessionResponse.fromJson);
        final data = result.data!;
        await _session.establish(
          sessionToken: data.sessionToken,
          expiresInSeconds: data.expiresIn,
          sessionUser: data.user,
        );
      } on ApiException catch (error) {
        if (error.statusCode == 401) await _session.clear();
        rethrow;
      } on TimeoutException {
        throw const ApiException(
          message: 'El servidor tardó demasiado al renovar la sesión.',
        );
      } on http.ClientException {
        throw const ApiException(
          message: 'No se pudo conectar para renovar la sesión.',
        );
      }
    });
  }

  ApiResponse<T> _decodeResponse<T>(
    http.Response response,
    T Function(Object? json) parseData, {
    bool allowUnwrappedResponse = false,
  }) {
    Object? decoded;
    try {
      decoded = jsonDecode(response.body);
    } on FormatException {
      if (response.statusCode < 200 || response.statusCode >= 300) {
        throw ApiException(
          statusCode: response.statusCode,
          message:
              'El Gateway respondió con HTTP ${response.statusCode} y una respuesta no JSON.',
        );
      }
      rethrow;
    }
    if (decoded is! Map<String, dynamic>) {
      throw const FormatException(
        'La respuesta del servidor no es un JSON válido.',
      );
    }

    if (response.statusCode < 200 || response.statusCode >= 300) {
      if (decoded['success'] is! bool) {
        final error = decoded['error'];
        final message = error is String
            ? error
            : decoded['message'] is String
            ? decoded['message'] as String
            : 'El servidor respondió con el estado ${response.statusCode}.';
        throw ApiException(statusCode: response.statusCode, message: message);
      }
      final result = ApiResponse<T>.fromJson(decoded, parseData);
      throw ApiException(
        statusCode: response.statusCode,
        errorCode: result.error?.code,
        message:
            result.error?.message ??
            'El servidor respondió con el estado ${response.statusCode}.',
      );
    }

    if (allowUnwrappedResponse && decoded['success'] is! bool) {
      return ApiResponse<T>(success: true, data: parseData(decoded));
    }

    final result = ApiResponse<T>.fromJson(decoded, parseData);
    if (!result.success) {
      throw ApiException(
        statusCode: response.statusCode,
        errorCode: result.error?.code,
        message: result.error?.message ?? 'El servidor rechazó la solicitud.',
      );
    }
    return result;
  }

  Map<String, String> get _headers {
    final headers = <String, String>{
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    };
    final token = _session.token;
    if (token != null && token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    final userId = _session.user?.id ?? ApiConfig.currentUserId.trim();
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
    T Function(Object? json) parseData, {
    bool refreshSession = true,
    bool retryUnauthorized = true,
    bool allowUnwrappedResponse = false,
  }) async {
    try {
      if (refreshSession) await _refreshSessionIfNeeded();

      var response = await send().timeout(_timeout);
      if (retryUnauthorized &&
          response.statusCode == 401 &&
          _session.hasToken) {
        await _refreshSession(force: true);
        response = await send().timeout(_timeout);
      }
      return _decodeResponse(
        response,
        parseData,
        allowUnwrappedResponse: allowUnwrappedResponse,
      );
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

class _SessionResponse {
  const _SessionResponse({
    required this.sessionToken,
    required this.expiresIn,
    required this.user,
  });

  final String sessionToken;
  final int expiresIn;
  final AuthUser user;

  factory _SessionResponse.fromJson(Object? value) {
    if (value is! Map<String, dynamic>) {
      throw const FormatException('El servidor devolvió una sesión inválida.');
    }
    final token = value['session_token'];
    final expiresIn = value['expires_in'];
    if (token is! String || expiresIn is! int) {
      throw const FormatException(
        'El servidor devolvió una sesión incompleta.',
      );
    }
    return _SessionResponse(
      sessionToken: token,
      expiresIn: expiresIn,
      user: AuthUser.fromJson(value['user']),
    );
  }
}
