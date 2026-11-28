import '../state/app_notification.dart';

class GroupedNotification {
  GroupedNotification._(this.notifications);

  final List<AppNotification> notifications;

  AppNotification get latest => notifications.first;

  int get count {
    final reportedCount = notifications.fold(
      0,
      (total, item) => total + item.groupCount,
    );
    return reportedCount > actorNames.length
        ? reportedCount
        : actorNames.length;
  }

  Set<String> get actorNames {
    final names = <String>{};
    for (final item in notifications) {
      names.addAll(item.actorNames.where((name) => name.isNotEmpty));
      if (item.actorName case final String name when name.isNotEmpty) {
        names.add(name);
      }
    }
    return names;
  }

  bool get isRead => notifications.every((item) => item.isRead);

  String get title {
    final names = actorNames;
    if (names.isEmpty) return latest.title;

    final firstName = names.first;
    final additionalPeople = count - 1;
    final others = additionalPeople > 0
        ? ' y $additionalPeople ${additionalPeople == 1 ? 'persona' : 'personas'} más'
        : '';

    return switch (latest.kind) {
      NotificationKind.like =>
        additionalPeople == 0
            ? 'A $firstName le gustó tu publicación'
            : 'A $firstName$others les gustó tu publicación',
      NotificationKind.comment =>
        additionalPeople == 0
            ? '$firstName comentó tu publicación'
            : '$firstName$others comentaron tu publicación',
      NotificationKind.friendRequest =>
        additionalPeople == 0
            ? '$firstName te envió una solicitud de amistad'
            : '$firstName$others te enviaron solicitudes de amistad',
      NotificationKind.levelUp || NotificationKind.reward => latest.title,
    };
  }

  static List<GroupedNotification> group(
    Iterable<AppNotification> notifications,
  ) {
    final groups = <String, List<AppNotification>>{};
    for (final notification in notifications) {
      final key = _groupKey(notification);
      groups.putIfAbsent(key, () => []).add(notification);
    }
    return groups.values
        .map((items) => GroupedNotification._(List.unmodifiable(items)))
        .toList(growable: false);
  }

  static String _groupKey(AppNotification notification) =>
      switch (notification.kind) {
        NotificationKind.like || NotificationKind.comment =>
          '${notification.kind.name}:${notification.postId ?? 'unknown'}',
        NotificationKind.friendRequest => notification.kind.name,
        NotificationKind.levelUp || NotificationKind.reward =>
          '${notification.kind.name}:${notification.id}',
      };
}
