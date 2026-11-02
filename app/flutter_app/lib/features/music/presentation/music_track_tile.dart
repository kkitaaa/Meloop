import 'package:flutter/material.dart';
import 'package:flutter_app/models/music_track.dart';

class MusicTrackTile extends StatelessWidget {
  final MusicTrack track;
  final VoidCallback onTap;
  final Widget? trailing;

  const MusicTrackTile({
    super.key,
    required this.track,
    required this.onTap,
    this.trailing,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 5),
      elevation: 0,
      color: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: const BorderSide(color: Color(0xFFE7EEEC)),
      ),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 5),
        leading: ClipRRect(
          borderRadius: BorderRadius.circular(11),
          child: SizedBox(
            height: 56,
            width: 56,
            child: track.artworkUrl == null
                ? const ColoredBox(
                    color: Color(0xFFE8F7F3),
                    child: Icon(
                      Icons.music_note_rounded,
                      color: Color(0xFF0D7771),
                    ),
                  )
                : Image.network(
                    track.artworkUrl!,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) =>
                        const ColoredBox(
                          color: Color(0xFFE8F7F3),
                          child: Icon(
                            Icons.music_note_rounded,
                            color: Color(0xFF0D7771),
                          ),
                        ),
                  ),
          ),
        ),
        title: Text(
          track.title,
          style: const TextStyle(
            color: Color(0xFF183D3B),
            fontWeight: FontWeight.w700,
          ),
        ),
        subtitle: Text(
          track.artist.isEmpty
              ? switch (track.type) {
                  'artist' => 'Artista',
                  'album' => 'Álbum',
                  'genre' => 'Género',
                  _ => 'Canción',
                }
              : track.artist,
          style: const TextStyle(color: Color(0xFF788A86)),
        ),
        trailing: trailing ?? const Icon(Icons.chevron_right_rounded),
        onTap: onTap,
      ),
    );
  }
}
