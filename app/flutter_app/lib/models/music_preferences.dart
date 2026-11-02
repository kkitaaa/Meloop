import 'package:flutter_app/models/music_track.dart';

class MusicPreference {
  final String id;
  final String name;
  final String type;
  final String? artworkUrl;

  const MusicPreference({
    required this.id,
    required this.name,
    required this.type,
    this.artworkUrl,
  });

  factory MusicPreference.fromJson(Object? json, {required String type}) {
    if (json is! Map<String, dynamic>) {
      throw const FormatException('Una preferencia musical es inválida.');
    }

    final id = json['id'] ?? json['item_id'] ?? json['spotify_id'];
    final name = json['name'] ?? json['title'];
    final artworkUrl =
        json['artwork_url'] ?? json['image_url'] ?? json['album_image_url'];
    if (id is! String ||
        id.isEmpty ||
        name is! String ||
        name.isEmpty ||
        (artworkUrl != null && artworkUrl is! String)) {
      throw const FormatException('Una preferencia musical es inválida.');
    }

    return MusicPreference(
      id: id,
      name: name,
      type: type,
      artworkUrl: artworkUrl as String?,
    );
  }

  factory MusicPreference.fromTrack(MusicTrack track, {required String type}) {
    return MusicPreference(
      id: track.id,
      name: track.title,
      type: type,
      artworkUrl: track.artworkUrl,
    );
  }
}

class MusicPreferencesData {
  final List<MusicPreference> songs;
  final List<MusicPreference> artists;
  final List<MusicPreference> genres;
  final Map<String, int> limits;

  const MusicPreferencesData({
    required this.songs,
    required this.artists,
    required this.genres,
    required this.limits,
  });

  factory MusicPreferencesData.fromJson(Object? json) {
    if (json is! Map<String, dynamic>) {
      throw const FormatException('Las preferencias musicales son inválidas.');
    }

    final preferences = json['preferences'] is Map<String, dynamic>
        ? json['preferences'] as Map<String, dynamic>
        : json;
    final limits = json['limits'] is Map<String, dynamic>
        ? json['limits'] as Map<String, dynamic>
        : const <String, dynamic>{};

    return MusicPreferencesData(
      songs: _parseItems(
        preferences['songs'] ??
            preferences['tracks'] ??
            preferences['tracks_preferences'],
        type: 'track',
      ),
      artists: _parseItems(
        preferences['artists'] ?? preferences['artist_preferences'],
        type: 'artist',
      ),
      genres: _parseItems(
        preferences['genres'] ?? preferences['genre_preferences'],
        type: 'genre',
      ),
      limits: {
        'track': _readLimit(limits, 'track', 'songs'),
        'artist': _readLimit(limits, 'artist', 'artists'),
        'genre': _readLimit(limits, 'genre', 'genres'),
      },
    );
  }

  static List<MusicPreference> _parseItems(
    Object? json, {
    required String type,
  }) {
    if (json == null) return const [];
    if (json is! List) {
      throw const FormatException('La lista de preferencias es inválida.');
    }
    return json
        .map((item) => MusicPreference.fromJson(item, type: type))
        .toList();
  }

  static int _readLimit(
    Map<String, dynamic> limits,
    String type,
    String plural,
  ) {
    final limit = limits[type] ?? limits[plural];
    if (limit == null) return 30;
    if (limit is! int || limit < 0) {
      throw const FormatException('El límite de preferencias es inválido.');
    }
    return limit;
  }
}
