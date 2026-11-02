import 'package:flutter/material.dart';
import 'package:flutter_app/models/music_track.dart';

class MusicTrackDetailScreen extends StatelessWidget {
  final MusicTrack track;

  const MusicTrackDetailScreen({super.key, required this.track});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Ficha musical')),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (track.artworkUrl != null)
                ClipRRect(
                  borderRadius: BorderRadius.circular(16),
                  child: Image.network(
                    track.artworkUrl!,
                    width: 220,
                    height: 220,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) =>
                        const Icon(Icons.album, size: 120),
                  ),
                )
              else
                const Icon(Icons.album, size: 120),
              const SizedBox(height: 20),
              Text(
                track.title,
                textAlign: TextAlign.center,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 8),
              Text(
                track.artist,
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
