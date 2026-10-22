import 'package:flutter/widgets.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/features/notifications/state/app_notification.dart';
import 'package:flutter_app/features/notifications/state/notification_controller.dart';
import 'package:flutter_app/features/notifications/state/notification_repository.dart';

void main() {
  test('mock repository emits one like/comment per minute', () async {
    var now = DateTime(2026, 10, 21);
    final repository = MockNotificationRepository(clock: () => now);

    expect(await repository.fetchNotifications(), isEmpty);
    now = now.add(const Duration(seconds: 59));
    expect(await repository.fetchNotifications(), isEmpty);

    now = now.add(const Duration(seconds: 1));
    final like = await repository.fetchNotifications();
    expect(like, hasLength(1));
    expect(like.single.kind, NotificationKind.like);

    now = now.add(const Duration(minutes: 1));
    final comment = await repository.fetchNotifications();
    expect(comment, hasLength(1));
    expect(comment.single.kind, NotificationKind.comment);
  });

  testWidgets('polling pauses in background and polls on resume', (
    tester,
  ) async {
    final repository = _RecordingRepository();
    final controller = NotificationController(
      repository: repository,
      pollInterval: const Duration(seconds: 5),
      playSound: false,
    );

    controller.start();
    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump();
    expect(repository.fetchCount, 1);

    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.paused);
    await tester.pump(const Duration(seconds: 10));
    expect(repository.fetchCount, 1);

    tester.binding.handleAppLifecycleStateChanged(AppLifecycleState.resumed);
    await tester.pump();
    expect(repository.fetchCount, 2);

    controller.dispose();
  });
}

class _RecordingRepository implements NotificationRepository {
  int fetchCount = 0;

  @override
  Future<List<AppNotification>> fetchNotifications({DateTime? since}) async {
    fetchCount++;
    return const [];
  }

  @override
  Future<void> updatePreference({
    required String type,
    required bool enabled,
  }) async {}

  @override
  void dispose() {}
}
