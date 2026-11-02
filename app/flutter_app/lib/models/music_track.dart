class MusicTrack {
  final String id;
  final String title;
  final String artist;
  final String? artworkUrl;

  const MusicTrack({
    required this.id,
    required this.title,
    required this.artist,
    this.artworkUrl,
  });

  factory MusicTrack.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) {
      throw const FormatException('Una canción tiene datos inválidos.');
    }

    final id = json['id'] ?? json['spotify_id'];
    final title = json['title'] ?? json['name'];
    final artist = json['artist'] ?? json['artist_name'];
    final artworkUrl =
        json['artwork_url'] ?? json['image_url'] ?? json['album_image_url'];
    if (id is! String ||
        id.isEmpty ||
        title is! String ||
        title.isEmpty ||
        artist is! String ||
        (artworkUrl != null && artworkUrl is! String)) {
      throw const FormatException('Una canción tiene datos inválidos.');
    }

    return MusicTrack(
      id: id,
      title: title,
      artist: artist,
      artworkUrl: artworkUrl as String?,
    );
  }
}

class MusicSearchPage {
  final List<MusicTrack> items;

  const MusicSearchPage({required this.items});

  factory MusicSearchPage.fromJson(Object? json) {
    final items = json is List
        ? json
        : json is Map<String, dynamic>
        ? json['items'] ?? json['tracks'] ?? json['results']
        : null;
    if (items is! List) {
      throw const FormatException(
        'La respuesta de búsqueda musical tiene datos inválidos.',
      );
    }
    return MusicSearchPage(items: items.map(MusicTrack.fromJson).toList());
  }
}
