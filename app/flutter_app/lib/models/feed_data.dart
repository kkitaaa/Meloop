class FeedPost {
  final int id;
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
    required this.isAuthor,
    required this.isEdited,
    required this.within24Hours,
    this.imageUrl,
  });

  factory FeedPost.fromJson(Object? json) {
    if (json is! Map<String, dynamic> ||
        json['id'] is! int ||
        json['user'] is! String ||
        (json['user_id'] != null && json['user_id'] is! String) ||
        json['time'] is! String ||
        json['content'] is! String ||
        json['likes'] is! int ||
        json['comments'] is! int ||
        json['isAuthor'] is! bool ||
        json['isEdited'] is! bool ||
        json['within24Hours'] is! bool ||
        (json['song'] != null && json['song'] is! String) ||
        (json['song_id'] != null && json['song_id'] is! String) ||
        (json['spotify_track_id'] != null &&
            json['spotify_track_id'] is! String) ||
        (json['artist'] != null && json['artist'] is! String) ||
        (json['audio_url'] != null && json['audio_url'] is! String) ||
        (json['audio_name'] != null && json['audio_name'] is! String) ||
        (json['image_url'] != null && json['image_url'] is! String)) {
      throw const FormatException(
        'Una publicación del feed tiene datos inválidos.',
      );
    }

    return FeedPost(
      id: json['id'] as int,
      user: json['user'] as String,
      userId: json['user_id'] as String?,
      time: json['time'] as String,
      content: json['content'] as String,
      song: json['song'] as String?,
      songId: (json['song_id'] ?? json['spotify_track_id']) as String?,
      artist: json['artist'] as String?,
      audioUrl: json['audio_url'] as String?,
      audioName: json['audio_name'] as String?,
      likes: json['likes'] as int,
      comments: json['comments'] as int,
      isAuthor: json['isAuthor'] as bool,
      isEdited: json['isEdited'] as bool,
      within24Hours: json['within24Hours'] as bool,
      imageUrl: json['image_url'] as String?,
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
    if (json is! Map<String, dynamic> ||
        json['items'] is! List ||
        json['page'] is! int ||
        json['pageSize'] is! int ||
        json['hasMore'] is! bool) {
      throw const FormatException(
        'La respuesta del feed tiene datos inválidos.',
      );
    }

    return FeedPage(
      items: (json['items'] as List).map(FeedPost.fromJson).toList(),
      page: json['page'] as int,
      pageSize: json['pageSize'] as int,
      hasMore: json['hasMore'] as bool,
    );
  }
}
