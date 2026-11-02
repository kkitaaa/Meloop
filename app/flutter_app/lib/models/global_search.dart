class SearchUser {
  final String id;
  final String name;
  final String? avatarUrl;
  final int level;
  final String friendStatus;

  const SearchUser({
    required this.id,
    required this.name,
    this.avatarUrl,
    required this.level,
    required this.friendStatus,
  });

  factory SearchUser.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) {
      throw const FormatException('Un usuario tiene datos inválidos.');
    }

    final id = json['id_usuario'] ?? json['id'];
    final name = json['nombre'] ?? json['username'] ?? json['name'];
    final avatar = json['avatar_url'];
    final level = json['nivel'] ?? json['level'] ?? json['id_nivel'] ?? 0;
    final friendStatus = json['friend_status'] ?? 'none';
    if (id is! String ||
        id.isEmpty ||
        name is! String ||
        name.isEmpty ||
        (avatar != null && avatar is! String) ||
        level is! int ||
        friendStatus is! String) {
      throw const FormatException('Un usuario tiene datos inválidos.');
    }

    return SearchUser(
      id: id,
      name: name,
      avatarUrl: avatar as String?,
      level: level,
      friendStatus: friendStatus,
    );
  }
}

class SearchUserPage {
  final List<SearchUser> items;

  const SearchUserPage({required this.items});

  factory SearchUserPage.fromJson(Object? json) {
    final items = json is List
        ? json
        : json is Map<String, dynamic>
        ? json['items'] ?? json['users'] ?? json['results']
        : null;
    if (items is! List) {
      throw const FormatException(
        'La respuesta de búsqueda de usuarios es inválida.',
      );
    }
    return SearchUserPage(items: items.map(SearchUser.fromJson).toList());
  }
}

class SearchPost {
  final Map<String, dynamic> cardData;

  const SearchPost({required this.cardData});

  factory SearchPost.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) {
      throw const FormatException('Una publicación tiene datos inválidos.');
    }
    final author = json['user'] ?? json['username'] ?? json['author'];
    final content = json['content'] ?? '';
    if (author is! String || content is! String) {
      throw const FormatException('Una publicación tiene datos inválidos.');
    }

    return SearchPost(
      cardData: {
        ...json,
        'id': json['id'] ?? json['id_post'],
        'user': author,
        'user_id': json['user_id'] ?? json['id_usuario'],
        'time': json['time'] ?? json['created_at'] ?? '',
        'content': content,
        'song': json['song'] ?? json['song_title'],
        'artist': json['artist'] ?? json['artist_name'],
        'audio_url': json['audio_url'],
        'audio_name': json['audio_name'],
        'image_url': json['image_url'],
        'likes': json['likes'] ?? 0,
        'comments': json['comments'] ?? 0,
        'isAuthor': false,
        'isEdited': json['isEdited'] ?? false,
        'within24Hours': false,
      },
    );
  }
}

class SearchPostPage {
  final List<SearchPost> items;

  const SearchPostPage({required this.items});

  factory SearchPostPage.fromJson(Object? json) {
    final items = json is List
        ? json
        : json is Map<String, dynamic>
        ? json['items'] ?? json['posts'] ?? json['results']
        : null;
    if (items is! List) {
      throw const FormatException(
        'La respuesta de búsqueda de publicaciones es inválida.',
      );
    }
    return SearchPostPage(items: items.map(SearchPost.fromJson).toList());
  }
}
