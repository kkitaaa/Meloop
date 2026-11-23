import 'dart:async';
import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter_app/config/api_config.dart';

typedef WebSocketMessageHandler = void Function(dynamic message);
typedef WebSocketErrorHandler = void Function(dynamic error);
typedef WebSocketCloseHandler = void Function();

/// WebSocket connection manager for real-time features
class WebSocketManager {
  final String _path;
  final Map<String, String>? _headers;

  WebSocket? _webSocket;
  StreamSubscription<dynamic>? _subscription;

  bool _isConnected = false;
  bool _isConnecting = false;
  bool _shouldReconnect = true;
  int _reconnectAttempts = 0;

  final List<WebSocketMessageHandler> _messageHandlers = [];
  final List<WebSocketErrorHandler> _errorHandlers = [];
  final List<WebSocketCloseHandler> _closeHandlers = [];

  WebSocketManager(this._path, {this._headers});

  bool get isConnected => _isConnected;
  bool get isConnecting => _isConnecting;

  /// Connect to WebSocket
  Future<void> connect() async {
    if (_isConnected || _isConnecting) {
      return;
    }

    _isConnecting = true;
    _shouldReconnect = true;
    _reconnectAttempts = 0;

    try {
      await _attemptConnect();
    } catch (e) {
      _isConnecting = false;
      _notifyError(e);
      await _scheduleReconnect();
    }
  }

  Future<void> _attemptConnect() async {
    try {
      final baseUri = Uri.parse(ApiConfig.webSocketUrl);
      final basePath = baseUri.path.replaceFirst(RegExp(r'/+$'), '');
      final endpointPath = _path.replaceFirst(RegExp(r'^/+'), '');
      final fullUrl = baseUri
          .replace(path: '$basePath/$endpointPath')
          .toString();

      debugPrint('[WebSocket] Connecting to: $fullUrl');

      _webSocket = await WebSocket.connect(fullUrl, headers: _headers).timeout(
        const Duration(seconds: 10),
        onTimeout: () => throw TimeoutException('WebSocket connection timeout'),
      );

      _isConnected = true;
      _isConnecting = false;
      _reconnectAttempts = 0;

      debugPrint('[WebSocket] Connected successfully');

      // Listen to messages
      _subscription = _webSocket!.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
      );
    } catch (e) {
      debugPrint('[WebSocket] Connection error: ${e.toString()}');
      _isConnecting = false;
      _isConnected = false;
      rethrow;
    }
  }

  void _onMessage(dynamic message) {
    debugPrint('[WebSocket] Message received: $message');
    _notifyMessage(message);
  }

  void _onError(dynamic error) {
    debugPrint('[WebSocket] Error: $error');
    _isConnected = false;
    _notifyError(error);
    _scheduleReconnect();
  }

  void _onDone() {
    debugPrint('[WebSocket] Connection closed');
    _isConnected = false;
    _notifyClose();

    if (_shouldReconnect) {
      _scheduleReconnect();
    }
  }

  Future<void> _scheduleReconnect() async {
    if (!_shouldReconnect ||
        _reconnectAttempts >= ApiConfig.webSocketMaxReconnectAttempts) {
      debugPrint('[WebSocket] Max reconnection attempts reached');
      _shouldReconnect = false;
      return;
    }

    _reconnectAttempts++;
    final delay = ApiConfig.webSocketReconnectDelay * _reconnectAttempts;

    debugPrint(
      '[WebSocket] Reconnecting in ${delay.inSeconds}s (attempt $_reconnectAttempts/${ApiConfig.webSocketMaxReconnectAttempts})',
    );

    await Future.delayed(delay);

    if (_shouldReconnect) {
      await connect();
    }
  }

  /// Send message to WebSocket
  void send(dynamic message) {
    if (!_isConnected || _webSocket == null) {
      throw StateError('WebSocket is not connected');
    }

    _webSocket!.add(message);
    debugPrint('[WebSocket] Message sent: $message');
  }

  /// Register message handler
  void onMessage(WebSocketMessageHandler handler) {
    _messageHandlers.add(handler);
  }

  /// Register error handler
  void onError(WebSocketErrorHandler handler) {
    _errorHandlers.add(handler);
  }

  /// Register close handler
  void onClose(WebSocketCloseHandler handler) {
    _closeHandlers.add(handler);
  }

  /// Remove handlers
  void removeMessageHandler(WebSocketMessageHandler handler) {
    _messageHandlers.remove(handler);
  }

  void removeErrorHandler(WebSocketErrorHandler handler) {
    _errorHandlers.remove(handler);
  }

  void removeCloseHandler(WebSocketCloseHandler handler) {
    _closeHandlers.remove(handler);
  }

  void _notifyMessage(dynamic message) {
    for (final handler in _messageHandlers) {
      try {
        handler(message);
      } catch (e) {
        debugPrint('[WebSocket] Error in message handler: $e');
      }
    }
  }

  void _notifyError(dynamic error) {
    for (final handler in _errorHandlers) {
      try {
        handler(error);
      } catch (e) {
        debugPrint('[WebSocket] Error in error handler: $e');
      }
    }
  }

  void _notifyClose() {
    for (final handler in _closeHandlers) {
      try {
        handler();
      } catch (e) {
        debugPrint('[WebSocket] Error in close handler: $e');
      }
    }
  }

  /// Disconnect WebSocket
  Future<void> disconnect() async {
    _shouldReconnect = false;
    _isConnected = false;

    await _subscription?.cancel();
    await _webSocket?.close();

    _webSocket = null;
    _subscription = null;

    debugPrint('[WebSocket] Disconnected');
  }

  /// Dispose resources
  void dispose() {
    _messageHandlers.clear();
    _errorHandlers.clear();
    _closeHandlers.clear();
    disconnect();
  }
}
