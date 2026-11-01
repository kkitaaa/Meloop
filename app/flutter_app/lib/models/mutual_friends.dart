class MutualFriend {
  final String id;
  final String username;
  final String? avatarUrl;

  const MutualFriend({
    required this.id,
    required this.username,
    this.avatarUrl,
  });

  factory MutualFriend.fromJson(Object? json) {
    if (json is! Map<String, dynamic> ||
        json['id_usuario'] is! String ||
        json['username'] is! String ||
        (json['avatar_url'] != null && json['avatar_url'] is! String)) {
      throw const FormatException('Un amigo en común tiene datos inválidos.');
    }

    return MutualFriend(
      id: json['id_usuario'] as String,
      username: json['username'] as String,
      avatarUrl: json['avatar_url'] as String?,
    );
  }
}

class MutualFriendsPage {
  final List<MutualFriend> items;
  final int totalCount;
  final bool hidden;
  final String? nextCursor;
  final bool hasMore;

  const MutualFriendsPage({
    required this.items,
    required this.totalCount,
    required this.hidden,
    required this.nextCursor,
    required this.hasMore,
  });

  factory MutualFriendsPage.fromJson(Object? json) {
    if (json is! Map<String, dynamic> ||
        json['items'] is! List ||
        json['totalCount'] is! int ||
        json['hidden'] is! bool ||
        json['hasMore'] is! bool ||
        (json['nextCursor'] != null && json['nextCursor'] is! String)) {
      throw const FormatException(
        'La respuesta de amigos en común tiene datos inválidos.',
      );
    }

    final totalCount = json['totalCount'] as int;
    if (totalCount < 0) {
      throw const FormatException('El conteo de amigos no puede ser negativo.');
    }

    return MutualFriendsPage(
      items: (json['items'] as List).map(MutualFriend.fromJson).toList(),
      totalCount: totalCount,
      hidden: json['hidden'] as bool,
      nextCursor: json['nextCursor'] as String?,
      hasMore: json['hasMore'] as bool,
    );
  }
}
