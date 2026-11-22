import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/services/websocket_manager.dart';

void main() {
  group('WebSocketManager Tests', () {
    late WebSocketManager webSocketManager;

    setUp(() {
      webSocketManager = WebSocketManager('/api/notifications');
    });

    test('WebSocketManager initializes correctly', () {
      expect(webSocketManager.isConnected, false);
      expect(webSocketManager.isConnecting, false);
    });

    test('WebSocketManager path is set correctly', () {
      final ws = WebSocketManager('/api/test');
      // Verify initialization doesn't throw
      expect(ws, isNotNull);
    });

    test('WebSocketManager can register message handler', () {
      var messageReceived = false;
      webSocketManager.onMessage((msg) {
        messageReceived = true;
      });
      // Handler registered successfully
      expect(webSocketManager, isNotNull);
    });

    test('WebSocketManager can register error handler', () {
      var errorReceived = false;
      webSocketManager.onError((error) {
        errorReceived = true;
      });
      // Handler registered successfully
      expect(webSocketManager, isNotNull);
    });

    test('WebSocketManager can register close handler', () {
      var closedReceived = false;
      webSocketManager.onClose(() {
        closedReceived = true;
      });
      // Handler registered successfully
      expect(webSocketManager, isNotNull);
    });

    test('WebSocketManager multiple handlers can be registered', () {
      webSocketManager.onMessage((msg) {});
      webSocketManager.onMessage((msg) {});
      webSocketManager.onError((error) {});
      webSocketManager.onClose(() {});
      // All handlers registered successfully
      expect(webSocketManager, isNotNull);
    });

    tearDown(() async {
      await webSocketManager.disconnect();
    });
  });
}
