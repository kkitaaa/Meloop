import 'dart:convert';

import 'package:flutter_app/config/api_config.dart';
import 'package:flutter_app/services/auth_session.dart';
import 'package:http/http.dart' as http;

import 'models/user_progress.dart';

abstract interface class GamificationRepository {
  Future<UserProgress> fetchUserProgress();

  void dispose();
}

class HttpGamificationRepository implements GamificationRepository {
  HttpGamificationRepository({
    Uri? endpoint,
    required this.accessToken,
    Future<String?> Function()? userId,
    http.Client? client,
  }) : endpoint =
           endpoint ??
           Uri.parse(
             '${ApiConfig.baseUrl}${ApiConfig.gamificationProgressEndpoint}',
           ),
       userId = userId ?? (() async => AuthSession.shared.user?.id),
       _client = client ?? http.Client(),
       _ownsClient = client == null;

  final Uri endpoint;
  final Future<String?> Function() accessToken;
  final Future<String?> Function() userId;
  final http.Client _client;
  final bool _ownsClient;

  @override
  Future<UserProgress> fetchUserProgress() async {
    final token = await accessToken();
    final currentUserId = await userId();
    final response = await _client.get(
      endpoint,
      headers: {
        'Accept': 'application/json',
        if (token != null && token.isNotEmpty) 'Authorization': 'Bearer $token',
        if (currentUserId != null && currentUserId.isNotEmpty)
          'X-User-ID': currentUserId,
      },
    );

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw http.ClientException(
        'No se pudo cargar el progreso del usuario (HTTP ${response.statusCode}).',
        endpoint,
      );
    }

    final decoded = jsonDecode(utf8.decode(response.bodyBytes));
    if (decoded is! Map) {
      throw const FormatException(
        'La respuesta de progreso debe contener un objeto.',
      );
    }

    Object? data = decoded;
    if (decoded.containsKey('data')) {
      data = decoded['data'];
    }
    if (data is! Map) {
      throw const FormatException(
        'La respuesta de progreso debe contener un objeto data.',
      );
    }

    final payload = Map<String, dynamic>.from(data);
    final nestedProgress = payload['progress'] ?? payload['progreso'];
    if (nestedProgress != null) {
      if (nestedProgress is! Map) {
        throw const FormatException(
          'El objeto progress tiene formato inválido.',
        );
      }
      final progress = Map<String, dynamic>.from(nestedProgress);
      progress['rewards'] ??=
          payload['rewards'] ??
          payload['recompensas'] ??
          payload['unlocked_rewards'];
      return UserProgress.fromJson(progress);
    }
    return UserProgress.fromJson(payload);
  }

  @override
  void dispose() {
    if (_ownsClient) _client.close();
  }
}
