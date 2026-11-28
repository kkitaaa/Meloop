class SavedPost {
  final String id;
  final bool isAvailable;
  final Map<String, dynamic> cardData;

  const SavedPost({
    required this.id,
    required this.isAvailable,
    required this.cardData,
  });

  factory SavedPost.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) {
      throw const FormatException(
        'Una publicación guardada tiene datos inválidos.',
      );
    }

    final rawPost = json['post'];
    final post = rawPost is Map<String, dynamic> ? rawPost : json;
    final hasFlatPostData =
        json.containsKey('content') ||
        json.containsKey('username') ||
        json.containsKey('user');
    final rawId =
        json['post_id'] ?? json['id_post'] ?? post['id'] ?? post['id_post'];
    final id = rawId is String || rawId is int ? '$rawId' : null;
    if (id == null || id.isEmpty) {
      throw const FormatException(
        'Una publicación guardada no tiene identificador.',
      );
    }

    final available = json['is_available'] ?? json['available'];
    final deleted = json['deleted'] ?? json['is_deleted'];
    final isAvailable =
        (rawPost is Map<String, dynamic> || hasFlatPostData) &&
        available != false &&
        deleted != true;

    return SavedPost(
      id: id,
      isAvailable: isAvailable,
      cardData: {
        ...post,
        'id': post['id'] ?? int.tryParse(id) ?? id,
        'user': post['user'] ?? post['username'] ?? post['author'] ?? 'Usuario',
        'user_id': post['user_id'] ?? post['id_usuario'],
        'time': post['time'] ?? post['created_at'] ?? '',
        'content': post['content'] ?? '',
        'song': post['song'] ?? post['song_title'],
        'artist': post['artist'] ?? post['artist_name'],
        'image_url': post['image_url'],
        'audio_url': post['audio_url'],
        'audio_name': post['audio_name'],
        'likes': post['likes'] ?? 0,
        'comments': post['comments'] ?? 0,
        'isAuthor': false,
        'isEdited': post['isEdited'] ?? false,
        'within24Hours': false,
        'is_saved': true,
      },
    );
  }
}

class SavedPostPage {
  final List<SavedPost> items;
  final bool hasMore;

  const SavedPostPage({required this.items, required this.hasMore});

  factory SavedPostPage.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) {
      throw const FormatException('La respuesta de guardados es inválida.');
    }
    final rawItems = json['items'] ?? json['saved'] ?? json['posts'];
    if (rawItems is! List) {
      throw const FormatException('La respuesta de guardados es inválida.');
    }
    final hasMore = json['hasMore'] ?? json['has_more'];
    if (hasMore != null && hasMore is! bool) {
      throw const FormatException('La paginación de guardados es inválida.');
    }
    return SavedPostPage(
      items: rawItems.map(SavedPost.fromJson).toList(),
      hasMore: hasMore as bool? ?? false,
    );
  }
}
