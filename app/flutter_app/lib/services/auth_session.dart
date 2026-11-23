import 'dart:convert';

import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class AuthUser {
  const AuthUser({
    required this.id,
    required this.username,
    required this.email,
  });

  final String id;
  final String username;
  final String email;

  factory AuthUser.fromJson(Object? value) {
    if (value is! Map<String, dynamic>) {
      throw const FormatException('El servidor devolvió un usuario inválido.');
    }

    final id = value['id'];
    final username = value['username'];
    final email = value['email'];
    if (id is! String || username is! String || email is! String) {
      throw const FormatException(
        'El servidor devolvió un usuario incompleto.',
      );
    }

    return AuthUser(id: id, username: username, email: email);
  }

  Map<String, String> toJson() => {
    'id': id,
    'username': username,
    'email': email,
  };
}

class AuthSession {
  AuthSession({bool persistent = true, FlutterSecureStorage? storage})
    : _storage = persistent ? (storage ?? FlutterSecureStorage()) : null;

  static final AuthSession shared = AuthSession();

  static const _tokenKey = 'meloop.session.token';
  static const _expiryKey = 'meloop.session.expiry';
  static const _userKey = 'meloop.session.user';

  final FlutterSecureStorage? _storage;

  String? token;
  DateTime? expiresAt;
  AuthUser? user;
  Future<void>? _refreshInProgress;

  bool get hasToken => token != null && token!.isNotEmpty;

  bool get isAuthenticated =>
      hasToken && expiresAt != null && expiresAt!.isAfter(DateTime.now());

  bool get shouldRefresh {
    final expiry = expiresAt;
    return hasToken &&
        expiry != null &&
        expiry.difference(DateTime.now()) <= const Duration(minutes: 2);
  }

  Future<void> restore() async {
    final storage = _storage;
    if (storage == null) return;

    final storedToken = await storage.read(key: _tokenKey);
    final storedExpiry = await storage.read(key: _expiryKey);
    final storedUser = await storage.read(key: _userKey);
    if (storedToken == null || storedExpiry == null || storedUser == null) {
      await clear();
      return;
    }

    try {
      final parsedExpiry = DateTime.parse(storedExpiry);
      final parsedUser = AuthUser.fromJson(jsonDecode(storedUser));
      if (storedToken.isEmpty || !parsedExpiry.isAfter(DateTime.now())) {
        await clear();
        return;
      }
      token = storedToken;
      expiresAt = parsedExpiry;
      user = parsedUser;
    } on FormatException {
      await clear();
    }
  }

  Future<void> establish({
    required String sessionToken,
    required int expiresInSeconds,
    required AuthUser sessionUser,
  }) async {
    if (sessionToken.isEmpty || expiresInSeconds <= 0) {
      throw const FormatException('El servidor devolvió una sesión inválida.');
    }

    final expiry = DateTime.now().add(Duration(seconds: expiresInSeconds));
    final storage = _storage;
    if (storage != null) {
      await storage.write(key: _tokenKey, value: sessionToken);
      await storage.write(key: _expiryKey, value: expiry.toIso8601String());
      await storage.write(
        key: _userKey,
        value: jsonEncode(sessionUser.toJson()),
      );
    }

    token = sessionToken;
    expiresAt = expiry;
    user = sessionUser;
  }

  Future<void> refreshOnce(Future<void> Function() refresh) {
    final current = _refreshInProgress;
    if (current != null) return current;

    final operation = refresh();
    _refreshInProgress = operation;
    return operation.whenComplete(() {
      if (identical(_refreshInProgress, operation)) {
        _refreshInProgress = null;
      }
    });
  }

  Future<void> clear() async {
    token = null;
    expiresAt = null;
    user = null;
    final storage = _storage;
    if (storage != null) {
      await storage.delete(key: _tokenKey);
      await storage.delete(key: _expiryKey);
      await storage.delete(key: _userKey);
    }
  }
}
