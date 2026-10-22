import 'dart:convert';

import 'package:http/http.dart' as http;

import 'app_notification.dart';

abstract interface class NotificationRepository {
  Future<List<AppNotification>> fetchNotifications({DateTime? since});

  Future<void> updatePreference({required String type, required bool enabled});

  void dispose();
}

class MockNotificationRepository implements NotificationRepository {
  MockNotificationRepository({
    this.interval = const Duration(minutes: 1),
    DateTime Function()? clock,
  }) : _clock = clock ?? DateTime.now {
    _nextNotificationAt = _clock().add(interval);
  }

  final Duration interval;
  final DateTime Function() _clock;
  late DateTime _nextNotificationAt;
  int _notificationCount = 0;

  @override
  Future<List<AppNotification>> fetchNotifications({DateTime? since}) async {
    final now = _clock();
    if (now.isBefore(_nextNotificationAt)) return const [];

    _notificationCount++;
    _nextNotificationAt = now.add(interval);
    final isLike = _notificationCount.isOdd;

    return [
      AppNotification(
        id: 'mock-$_notificationCount-${now.microsecondsSinceEpoch}',
        kind: isLike ? NotificationKind.like : NotificationKind.comment,
        title: isLike
            ? 'A Camila le gustó tu publicación'
            : 'Felipe comentó tu publicación',
        detail: isLike ? 'Hace un momento' : '“Qué buena canción”',
        createdAt: now,
        postId: isLike ? 1 : 2,
      ),
    ];
  }

  @override
  Future<void> updatePreference({
    required String type,
    required bool enabled,
  }) async {}

  @override
  void dispose() {}
}

class HttpNotificationRepository implements NotificationRepository {
  HttpNotificationRepository({
    required this.endpoint,
    required this.settingsEndpoint,
    required this.accessToken,
    http.Client? client,
  }) : _client = client ?? http.Client(),
       _ownsClient = client == null;

  final Uri endpoint;
  final Uri settingsEndpoint;
  final Future<String?> Function() accessToken;
  final http.Client _client;
  final bool _ownsClient;

  @override
  Future<List<AppNotification>> fetchNotifications({DateTime? since}) async {
    final queryParameters = Map<String, String>.from(endpoint.queryParameters);
    if (since != null) {
      queryParameters['since'] = since.toUtc().toIso8601String();
    }
    final uri = endpoint.replace(queryParameters: queryParameters);
    final token = await accessToken();
    final response = await _client.get(
      uri,
      headers: {
        'Accept': 'application/json',
        if (token != null && token.isNotEmpty) 'Authorization': 'Bearer $token',
      },
    );

    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw http.ClientException(
        'No se pudieron cargar las notificaciones (HTTP ${response.statusCode}).',
        uri,
      );
    }

    final decoded = jsonDecode(utf8.decode(response.bodyBytes));
    final entries = _notificationEntries(decoded);
    return entries.map(_parseNotification).toList(growable: false);
  }

  @override
  Future<void> updatePreference({
    required String type,
    required bool enabled,
  }) async {
    final token = await accessToken();
    final response = await _client.patch(
      settingsEndpoint,
      headers: {
        'Accept': 'application/json',
        'Content-Type': 'application/json',
        if (token != null && token.isNotEmpty) 'Authorization': 'Bearer $token',
      },
      body: jsonEncode({
        'tipo_notificacion': _backendPreferenceType(type),
        'habilitada': enabled,
      }),
    );
    if (response.statusCode < 200 || response.statusCode >= 300) {
      throw http.ClientException(
        'No se pudo actualizar la preferencia (HTTP ${response.statusCode}).',
        settingsEndpoint,
      );
    }
  }

  String _backendPreferenceType(String type) => switch (type) {
    'like' => 'LIKE',
    'comment' => 'COMENTARIO',
    'friend' => 'SOLICITUD_AMISTAD',
    _ => throw ArgumentError.value(
      type,
      'type',
      'Tipo de preferencia inválido',
    ),
  };

  @override
  void dispose() {
    if (_ownsClient) _client.close();
  }

  List<Map<String, dynamic>> _notificationEntries(Object? response) {
    Object? data = response;
    if (data is Map<String, dynamic> && data.containsKey('data')) {
      data = data['data'];
    }
    if (data is Map<String, dynamic>) {
      data = data['notifications'] ?? data['notificaciones'];
    }
    if (data is! List) {
      throw const FormatException(
        'La respuesta de notificaciones debe contener una lista.',
      );
    }

    return data
        .map((entry) {
          if (entry is! Map) {
            throw const FormatException('Notificación con formato inválido.');
          }
          return Map<String, dynamic>.from(entry);
        })
        .toList(growable: false);
  }

  AppNotification _parseNotification(Map<String, dynamic> json) {
    final id = _stringValue(json['id_notificacion'] ?? json['id']);
    final rawType = _stringValue(json['tipo'] ?? json['type']);
    if (id == null || rawType == null) {
      throw const FormatException(
        'Cada notificación necesita id_notificacion/id y tipo/type.',
      );
    }

    final kind = _parseKind(rawType);
    final data = _mapValue(json['datos'] ?? json['data']);
    final actor = _stringValue(
      json['actor_name'] ?? json['nombre_emisor'] ?? data['actorName'],
    );
    final title =
        _stringValue(json['title'] ?? json['titulo']) ??
        _defaultTitle(kind, actor);
    final detail =
        _stringValue(json['detail'] ?? json['detalle'] ?? data['detail']) ?? '';
    final rawCreatedAt =
        json['created_at'] ?? json['fecha_creacion'] ?? json['createdAt'];
    final createdAt = rawCreatedAt is String
        ? DateTime.tryParse(rawCreatedAt)
        : null;
    final rawPostId =
        json['post_id'] ?? json['id_publicacion'] ?? data['postId'];

    return AppNotification(
      id: id,
      kind: kind,
      title: title,
      detail: detail,
      createdAt: createdAt ?? DateTime.now(),
      postId: rawPostId is int
          ? rawPostId
          : rawPostId is String
          ? int.tryParse(rawPostId)
          : null,
      isRead: json['leida'] == true || json['read'] == true,
    );
  }

  NotificationKind _parseKind(String value) => switch (value.toLowerCase()) {
    'like' || 'post.liked' => NotificationKind.like,
    'comment' ||
    'comentario' ||
    'comment.created' ||
    'comment.commented' => NotificationKind.comment,
    'friend' ||
    'solicitud_amistad' ||
    'friend.requested' => NotificationKind.friendRequest,
    _ => throw FormatException('Tipo de notificación desconocido: $value'),
  };

  String _defaultTitle(NotificationKind kind, String? actor) {
    final name = actor ?? 'Alguien';
    return switch (kind) {
      NotificationKind.like => 'A $name le gustó tu publicación',
      NotificationKind.comment => '$name comentó tu publicación',
      NotificationKind.friendRequest => '$name te envió una solicitud',
    };
  }

  String? _stringValue(Object? value) =>
      value is String && value.isNotEmpty ? value : null;

  Map<String, dynamic> _mapValue(Object? value) {
    if (value is Map<String, dynamic>) return value;
    if (value is String) {
      final decoded = jsonDecode(value);
      if (decoded is Map<String, dynamic>) return decoded;
    }
    return const {};
  }
}
