import 'package:flutter_app/models/music_track.dart';

class MusicDetail {
  final String id;
  final String type;
  final String name;
  final String? artist;
  final String? album;
  final String? artworkUrl;
  final List<String> genres;
  final List<MusicTrack> tracks;
  final String? externalUrl;

  const MusicDetail({
    required this.id,
    required this.type,
    required this.name,
    this.artist,
    this.album,
    this.artworkUrl,
    this.genres = const [],
    this.tracks = const [],
    this.externalUrl,
  });

  factory MusicDetail.fromJson(Object? json, {required String type}) {
    if (json is Map<String, dynamic>) {
      final nested =
          json['item'] ??
          json[type] ??
          switch (type) {
            'track' => json['tracks'],
            'artist' => json['artists'],
            'album' => json['albums'],
            _ => null,
          };
      if (nested is Map<String, dynamic>) json = nested;
    }
    if (json is! Map<String, dynamic>) {
      throw const FormatException(
        'El detalle musical tiene un formato inválido.',
      );
    }

    final id = json['id'] ?? json['spotify_id'];
    final name = json['name'] ?? json['title'];
    final rawGenres = json['genres'];
    final rawTracks = json['tracks'];
    final rawAlbum = json['album'];
    final rawArtist = json['artist'] ?? json['artists'] ?? json['artist_name'];
    final externalUrl =
        json['external_url'] ??
        (json['external_urls'] is Map<String, dynamic>
            ? (json['external_urls'] as Map<String, dynamic>)['spotify']
            : null);
    final artworkUrl =
        json['artwork_url'] ??
        json['image_url'] ??
        json['album_image_url'] ??
        _imageUrl(json['images']) ??
        (rawAlbum is Map<String, dynamic>
            ? _imageUrl(rawAlbum['images'])
            : null);

    if (id is! String ||
        id.isEmpty ||
        name is! String ||
        name.isEmpty ||
        (externalUrl != null && externalUrl is! String) ||
        (artworkUrl != null && artworkUrl is! String)) {
      throw const FormatException('El detalle musical tiene datos inválidos.');
    }

    final genres = rawGenres is List
        ? rawGenres.whereType<String>().toList()
        : const <String>[];
    final tracks = _parseTracks(rawTracks);
    return MusicDetail(
      id: id,
      type: type,
      name: name,
      artist: _artistName(rawArtist),
      album: rawAlbum is String
          ? rawAlbum
          : rawAlbum is Map<String, dynamic>
          ? _optionalString(rawAlbum['name'])
          : _optionalString(json['album_name']),
      artworkUrl: artworkUrl as String?,
      genres: genres,
      tracks: tracks,
      externalUrl: externalUrl as String?,
    );
  }

  static List<MusicTrack> _parseTracks(Object? value) {
    if (value is List) {
      return value
          .map((item) => MusicTrack.fromJson(item, type: 'track'))
          .toList();
    }
    if (value is Map<String, dynamic>) {
      final items = value['items'] ?? value['results'];
      if (items is List) {
        return items
            .map((item) => MusicTrack.fromJson(item, type: 'track'))
            .toList();
      }
    }
    return const [];
  }

  static String? _artistName(Object? value) {
    if (value is String) return value;
    if (value is Map<String, dynamic>) {
      final name = value['name'];
      return name is String ? name : null;
    }
    if (value is List) {
      final names = value
          .map((item) {
            if (item is String) return item;
            if (item is Map<String, dynamic>) {
              final name = item['name'];
              return name is String ? name : '';
            }
            return '';
          })
          .where((name) => name.isNotEmpty)
          .toList();
      return names.isEmpty ? null : names.join(', ');
    }
    return null;
  }

  static String? _imageUrl(Object? value) {
    if (value is List && value.isNotEmpty) {
      final first = value.first;
      if (first is Map<String, dynamic>) {
        final url = first['url'];
        return url is String ? url : null;
      }
    }
    return null;
  }

  static String? _optionalString(Object? value) =>
      value is String ? value : null;
}
