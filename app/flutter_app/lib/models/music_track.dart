class MusicTrack {
  final String id;
  final String title;
  final String artist;
  final String? artworkUrl;
  final String type;

  const MusicTrack({
    required this.id,
    required this.title,
    required this.artist,
    this.artworkUrl,
    this.type = 'track',
  });

  factory MusicTrack.fromJson(Object? json, {String type = 'track'}) {
    if (json is String) {
      return MusicTrack(id: json, title: json, artist: '', type: type);
    }
    if (json is! Map<String, dynamic>) {
      throw const FormatException('Una canción tiene datos inválidos.');
    }

    final id = json['id'] ?? json['spotify_id'] ?? json['name'];
    final title = json['title'] ?? json['name'];
    final artist = _artistLabel(
      json['artist'] ?? json['artist_name'] ?? json['artists'],
    );
    final artworkUrl =
        json['artwork_url'] ??
        json['image_url'] ??
        json['album_image_url'] ??
        _firstImageUrl(json['images']) ??
        _albumImageUrl(json['album']);
    if (id is! String ||
        id.isEmpty ||
        title is! String ||
        title.isEmpty ||
        (artworkUrl != null && artworkUrl is! String)) {
      throw const FormatException('Una canción tiene datos inválidos.');
    }

    return MusicTrack(
      id: id,
      title: title,
      artist: artist,
      artworkUrl: artworkUrl as String?,
      type: type,
    );
  }

  static String _artistLabel(Object? value) {
    if (value is String) return value;
    if (value is List) {
      return value
          .map((artist) {
            if (artist is String) return artist;
            if (artist is Map<String, dynamic>) {
              final name = artist['name'];
              return name is String ? name : '';
            }
            return '';
          })
          .where((name) => name.isNotEmpty)
          .join(', ');
    }
    if (value is Map<String, dynamic>) {
      final name = value['name'];
      return name is String ? name : '';
    }
    return '';
  }

  static String? _firstImageUrl(Object? value) {
    if (value is List && value.isNotEmpty) {
      final first = value.first;
      if (first is Map<String, dynamic>) {
        final url = first['url'];
        return url is String ? url : null;
      }
    }
    return null;
  }

  static String? _albumImageUrl(Object? value) {
    if (value is Map<String, dynamic>) {
      final artworkUrl = value['artwork_url'];
      return _firstImageUrl(value['images']) ??
          (artworkUrl is String ? artworkUrl : null);
    }
    return null;
  }
}

class MusicSearchPage {
  final List<MusicTrack> items;
  final int offset;
  final int? total;
  final bool hasMore;

  const MusicSearchPage({
    required this.items,
    this.offset = 0,
    this.total,
    this.hasMore = false,
  });

  factory MusicSearchPage.fromJson(Object? json, {String type = 'track'}) {
    final payload = _payloadForType(json, type);
    final items = payload is List
        ? payload
        : payload is Map<String, dynamic>
        ? payload['items'] ?? payload['results']
        : null;
    if (items is! List) {
      throw const FormatException(
        'La respuesta de búsqueda musical tiene datos inválidos.',
      );
    }
    final pageMeta = payload is Map<String, dynamic>
        ? payload
        : json is Map<String, dynamic>
        ? json
        : const <String, dynamic>{};
    final offset = pageMeta['offset'] is int ? pageMeta['offset'] as int : 0;
    final total = pageMeta['total'] is int ? pageMeta['total'] as int : null;
    final explicitHasMore =
        pageMeta['has_more'] ?? pageMeta['hasMore'] ?? pageMeta['hasNext'];
    final nextOffset = pageMeta['next_offset'] ?? pageMeta['nextOffset'];
    final hasMore = explicitHasMore is bool
        ? explicitHasMore
        : nextOffset is int
        ? nextOffset > offset
        : total != null
        ? offset + items.length < total
        : items.isNotEmpty;
    return MusicSearchPage(
      items: items
          .map((item) => MusicTrack.fromJson(item, type: type))
          .toList(),
      offset: offset,
      total: total,
      hasMore: hasMore,
    );
  }

  static Object? _payloadForType(Object? json, String type) {
    if (json is List) return json;
    if (json is! Map<String, dynamic>) return null;
    final payload =
        json['items'] ??
        json[type] ??
        switch (type) {
          'track' => json['tracks'],
          'artist' => json['artists'],
          'album' => json['albums'],
          'genre' => json['genres'],
          _ => null,
        } ??
        json['results'];
    if (payload is Map<String, dynamic> &&
        (payload['items'] is List || payload['results'] is List)) {
      return payload;
    }
    return payload;
  }
}
