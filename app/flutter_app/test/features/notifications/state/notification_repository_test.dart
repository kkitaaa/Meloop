import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/features/notifications/state/app_notification.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

void main() {
  test('parses grouped notification metadata from backend response', () async {
    final repository = HttpNotificationRepository(
      endpoint: Uri.parse('https://example.test/notifications'),
      settingsEndpoint: Uri.parse('https://example.test/settings'),
      accessToken: () async => null,
      client: MockClient(
        (_) async => http.Response(
          jsonEncode({
            'success': true,
            'data': [
              {
                'id': 'notification-1',
                'tipo': 'LIKE',
                'actor_name': 'Pedro',
                'actor_names': ['Pedro', 'Lucía'],
                'group_count': 2,
                'id_publicacion': 42,
                'fecha_creacion': '2026-10-21T12:00:00Z',
              },
            ],
          }),
          200,
          headers: {'content-type': 'application/json'},
        ),
      ),
    );

    final notifications = await repository.fetchNotifications();

    expect(notifications, hasLength(1));
    expect(notifications.single.kind, NotificationKind.like);
    expect(notifications.single.actorName, 'Pedro');
    expect(notifications.single.actorNames, ['Pedro', 'Lucía']);
    expect(notifications.single.groupCount, 2);
    expect(notifications.single.postId, 42);
    repository.dispose();
  });
}
