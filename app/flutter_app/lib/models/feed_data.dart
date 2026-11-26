class FeedPost {
  final Object id;
  final String user;
  final String? userId;
  final String time;
  final String content;
  final String? song;
  final String? songId;
  final String? artist;
  final String? audioUrl;
  final String? audioName;
  final int likes;
  final int comments;
  final bool isLiked;
  final bool isAuthor;
  final bool isEdited;
  final bool within24Hours;
  final String? imageUrl;

  const FeedPost({
    required this.id,
    required this.user,
    this.userId,
    required this.time,
    required this.content,
    this.song,
    this.songId,
    this.artist,
    this.audioUrl,
    this.audioName,
    required this.likes,
    required this.comments,
    this.isLiked = false,
    required this.isAuthor,
    required this.isEdited,
    required this.within24Hours,
    this.imageUrl,
  });

  factory FeedPost.fromJson(Object? json) {
    final id = json is Map<String, dynamic> ? json['id'] : null;
    final user = json is Map<String, dynamic>
        ? json['user'] ?? json['author_username']
        : null;
    final userId = json is Map<String, dynamic>
        ? json['user_id'] ?? json['author_id']
        : null;
    final rawTime = json is Map<String, dynamic> ? json['time'] : null;
    final createdAt = json is Map<String, dynamic> ? json['created_at'] : null;
    final time = rawTime is String ? rawTime : _formatFeedTime(createdAt);
    final likes = json is Map<String, dynamic>
        ? json['likes'] ?? json['likes_count'] ?? 0
        : null;
    final comments = json is Map<String, dynamic>
        ? json['comments'] ?? json['comments_count'] ?? 0
        : null;
    final isLiked = json is Map<String, dynamic>
        ? json['is_liked'] ?? false
        : null;
    final imageUrl = json is Map<String, dynamic>
        ? json['image_url'] ?? json['media_url']
        : null;

    if (json is! Map<String, dynamic> ||
        (id is! int && (id is! String || id.isEmpty)) ||
        user is! String ||
        (userId != null && userId is! String) ||
        time == null ||
        json['content'] is! String ||
        likes is! int ||
        comments is! int ||
        isLiked is! bool ||
        (json['isAuthor'] != null && json['isAuthor'] is! bool) ||
        (json['isEdited'] != null && json['isEdited'] is! bool) ||
        (json['within24Hours'] != null && json['within24Hours'] is! bool) ||
        (json['song'] != null && json['song'] is! String) ||
        (json['song_id'] != null && json['song_id'] is! String) ||
        (json['spotify_track_id'] != null &&
            json['spotify_track_id'] is! String) ||
        (json['artist'] != null && json['artist'] is! String) ||
        (json['audio_url'] != null && json['audio_url'] is! String) ||
        (json['audio_name'] != null && json['audio_name'] is! String) ||
        (imageUrl != null && imageUrl is! String)) {
      throw const FormatException(
        'Una publicación del feed tiene datos inválidos.',
      );
    }

    return FeedPost(
      id: id,
      user: user,
      userId: userId as String?,
      time: time,
      content: json['content'] as String,
      song: json['song'] as String?,
      songId: (json['song_id'] ?? json['spotify_track_id']) as String?,
      artist: json['artist'] as String?,
      audioUrl: json['audio_url'] as String?,
      audioName: json['audio_name'] as String?,
      likes: likes,
      comments: comments,
      isLiked: isLiked,
      isAuthor: json['isAuthor'] as bool? ?? false,
      isEdited: json['isEdited'] as bool? ?? false,
      within24Hours: json['within24Hours'] as bool? ?? true,
      imageUrl: imageUrl as String?,
    );
  }

  Map<String, dynamic> toPostCardData() => {
    'id': id,
    'user': user,
    'user_id': userId,
    'time': time,
    'content': content,
    'song': song,
    'song_id': songId,
    'artist': artist,
    'audio_url': audioUrl,
    'audio_name': audioName,
    'likes': likes,
    'comments': comments,
    'is_liked': isLiked,
    'isAuthor': isAuthor,
    'isEdited': isEdited,
    'within24Hours': within24Hours,
    'image_url': imageUrl,
  };
}

class FeedPage {
  final List<FeedPost> items;
  final int page;
  final int pageSize;
  final bool hasMore;

  const FeedPage({
    required this.items,
    required this.page,
    required this.pageSize,
    required this.hasMore,
  });

  factory FeedPage.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) {
      throw const FormatException(
        'La respuesta del feed tiene datos inválidos.',
      );
    }

    final page = json['page'];
    final items = json['items'];
    final pageSize = json['pageSize'];
    final hasMore = json['hasMore'];
    if (items is List && page is int && pageSize is int && hasMore is bool) {
      return FeedPage(
        items: items.map(FeedPost.fromJson).toList(),
        page: page,
        pageSize: pageSize,
        hasMore: hasMore,
      );
    }

    final data = json['data'];
    final limit = json['limit'];
    if (data is! List || page is! int || limit is! int || limit <= 0) {
      throw const FormatException(
        'La respuesta del feed tiene datos inválidos.',
      );
    }

    final feedItems = data.map(FeedPost.fromJson).toList();
    return FeedPage(
      items: feedItems,
      page: page,
      pageSize: limit,
      hasMore: feedItems.length >= limit,
    );
  }
}

String? _formatFeedTime(Object? value) {
  if (value is! String) return null;
  final createdAt = DateTime.tryParse(value);
  if (createdAt == null) return null;

  final elapsed = DateTime.now().difference(createdAt.toLocal());
  if (elapsed.isNegative || elapsed.inMinutes < 1) return 'Ahora';
  if (elapsed.inHours < 1) return 'Hace ${elapsed.inMinutes} min';
  if (elapsed.inDays < 1) return 'Hace ${elapsed.inHours} h';
  return 'Hace ${elapsed.inDays} días';
}
