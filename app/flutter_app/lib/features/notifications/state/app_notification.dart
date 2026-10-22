enum NotificationKind { like, comment, friendRequest, levelUp, reward }

class AppNotification {
  const AppNotification({
    required this.id,
    required this.kind,
    required this.title,
    required this.detail,
    required this.createdAt,
    this.postId,
    this.actorName,
    this.actorNames = const [],
    this.groupCount = 1,
    this.isRead = false,
  });

  final String id;
  final NotificationKind kind;
  final String title;
  final String detail;
  final DateTime createdAt;
  final int? postId;
  final String? actorName;
  final List<String> actorNames;
  final int groupCount;
  final bool isRead;

  AppNotification copyWith({bool? isRead}) => AppNotification(
    id: id,
    kind: kind,
    title: title,
    detail: detail,
    createdAt: createdAt,
    postId: postId,
    actorName: actorName,
    actorNames: actorNames,
    groupCount: groupCount,
    isRead: isRead ?? this.isRead,
  );

  static String preferenceKey(NotificationKind kind) => switch (kind) {
    NotificationKind.like => 'like',
    NotificationKind.comment => 'comment',
    NotificationKind.friendRequest => 'friend',
    NotificationKind.levelUp => 'level',
    NotificationKind.reward => 'reward',
  };
}
