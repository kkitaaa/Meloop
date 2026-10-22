import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_app/features/notifications/presentation/grouped_notification.dart';
import 'package:flutter_app/features/notifications/state/app_notification.dart';

void main() {
  test('groups same kind and post, naming actors and counting others', () {
    final groups = GroupedNotification.group([
      _notification(id: '1', actor: 'Pedro', kind: NotificationKind.like),
      _notification(id: '2', actor: 'Lucía', kind: NotificationKind.like),
      _notification(id: '3', actor: 'Ana', kind: NotificationKind.like),
      _notification(
        id: '4',
        actor: 'Juan',
        kind: NotificationKind.comment,
        postId: 22,
      ),
    ]);

    expect(groups, hasLength(2));
    expect(groups.first.count, 3);
    expect(
      groups.first.title,
      'A Pedro y 2 personas más les gustó tu publicación',
    );
    expect(groups.last.title, 'Juan comentó tu publicación');
  });

  test('honors aggregate counts returned by backend', () {
    final groups = GroupedNotification.group([
      _notification(
        id: '1',
        actor: 'Pedro',
        kind: NotificationKind.like,
        groupCount: 5,
      ),
    ]);

    expect(
      groups.single.title,
      'A Pedro y 4 personas más les gustó tu publicación',
    );
  });
}

AppNotification _notification({
  required String id,
  required String actor,
  required NotificationKind kind,
  int postId = 12,
  int groupCount = 1,
}) => AppNotification(
  id: id,
  kind: kind,
  title: 'Notificación de $actor',
  detail: '',
  createdAt: DateTime(2026, 10, 21),
  postId: postId,
  actorName: actor,
  groupCount: groupCount,
);
