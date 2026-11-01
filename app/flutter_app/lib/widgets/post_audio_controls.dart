import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_app/services/post_audio_playback.dart';

class PostAudioControls extends StatefulWidget {
  final String postId;
  final String title;
  final String artist;
  final Uint8List? audioBytes;
  final String? audioUrl;
  final String fileName;
  final Color accentColor;
  final PostAudioPlaybackController? playbackController;

  const PostAudioControls({
    super.key,
    required this.postId,
    required this.title,
    required this.artist,
    required this.audioBytes,
    required this.audioUrl,
    required this.fileName,
    required this.accentColor,
    this.playbackController,
  });

  @override
  State<PostAudioControls> createState() => _PostAudioControlsState();
}

class _PostAudioControlsState extends State<PostAudioControls> {
  late PostAudioPlaybackController _controller;
  late bool _isAttached;

  String get _sourceKey => _sourceKeyFor(widget);

  bool get _hasAudio =>
      (widget.audioBytes?.isNotEmpty ?? false) ||
      _isPlayableUrl(widget.audioUrl);

  @override
  void initState() {
    super.initState();
    _controller =
        widget.playbackController ?? PostAudioPlaybackController.shared;
    _isAttached = _hasAudio;
    if (_isAttached) _controller.attachCard();
  }

  @override
  void didUpdateWidget(covariant PostAudioControls oldWidget) {
    super.didUpdateWidget(oldWidget);
    final controller =
        widget.playbackController ?? PostAudioPlaybackController.shared;
    final postChanged = oldWidget.postId != widget.postId;
    final sourceChanged = _sourceKeyFor(oldWidget) != _sourceKeyFor(widget);
    if (!identical(controller, _controller) || postChanged || sourceChanged) {
      if (_isAttached) _controller.releaseCard(oldWidget.postId);
      _controller = controller;
      _isAttached = _hasAudio;
      if (_isAttached) _controller.attachCard();
    } else if (_isAttached != _hasAudio) {
      if (_isAttached) {
        _controller.releaseCard(widget.postId);
      } else {
        _controller.attachCard();
      }
      _isAttached = _hasAudio;
    }
  }

  @override
  void dispose() {
    if (_isAttached) _controller.releaseCard(widget.postId);
    super.dispose();
  }

  Future<void> _togglePlayback() => _controller.toggle(
    postId: widget.postId,
    bytes: widget.audioBytes,
    audioUrl: widget.audioUrl,
    sourceKey: _sourceKey,
    fileName: widget.fileName,
  );

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, _) {
        final isActive = _controller.activePostId == widget.postId;
        final isPlaying = isActive && _controller.isPlaying;
        final duration = isActive ? _controller.duration : null;
        final position = isActive ? _controller.position : Duration.zero;
        final maxMilliseconds = duration?.inMilliseconds ?? 0;
        final value = maxMilliseconds == 0
            ? 0.0
            : position.inMilliseconds.clamp(0, maxMilliseconds).toDouble();

        return Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: widget.accentColor,
            borderRadius: BorderRadius.circular(6),
          ),
          child: Column(
            children: [
              Row(
                children: [
                  IconButton(
                    tooltip: !_hasAudio
                        ? 'Audio no disponible'
                        : isPlaying
                        ? 'Pausar audio'
                        : 'Reproducir audio',
                    onPressed: _hasAudio ? _togglePlayback : null,
                    style: IconButton.styleFrom(
                      backgroundColor: Colors.black87,
                      foregroundColor: Colors.white,
                      disabledBackgroundColor: Colors.black54,
                      disabledForegroundColor: Colors.white54,
                    ),
                    icon: Icon(
                      isActive && _controller.isLoading
                          ? Icons.hourglass_top
                          : isPlaying
                          ? Icons.pause
                          : Icons.play_arrow,
                      size: 20,
                    ),
                  ),
                  const SizedBox(width: 4),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          widget.title,
                          style: const TextStyle(
                            color: Colors.white,
                            fontWeight: FontWeight.bold,
                            fontSize: 12,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                        Text(
                          widget.artist,
                          style: const TextStyle(
                            color: Colors.white70,
                            fontSize: 10,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ],
                    ),
                  ),
                  Icon(
                    Icons.graphic_eq,
                    color: Colors.white.withValues(alpha: 0.9),
                    size: 18,
                  ),
                ],
              ),
              if (_hasAudio) ...[
                Row(
                  children: [
                    Text(
                      _formatDuration(position),
                      style: const TextStyle(color: Colors.white, fontSize: 10),
                    ),
                    Expanded(
                      child: Slider(
                        min: 0,
                        max: maxMilliseconds > 0
                            ? maxMilliseconds.toDouble()
                            : 1,
                        value: value,
                        onChanged: isActive && maxMilliseconds > 0
                            ? (milliseconds) => _controller.seek(
                                Duration(milliseconds: milliseconds.round()),
                              )
                            : null,
                        activeColor: Colors.white,
                        inactiveColor: Colors.white38,
                      ),
                    ),
                    Text(
                      _formatDuration(duration ?? Duration.zero),
                      style: const TextStyle(color: Colors.white, fontSize: 10),
                    ),
                  ],
                ),
              ] else
                const Padding(
                  padding: EdgeInsets.only(top: 4, bottom: 2),
                  child: Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      'Audio no disponible para reproducir',
                      style: TextStyle(color: Colors.white70, fontSize: 10),
                    ),
                  ),
                ),
              if (isActive && _controller.errorMessage != null)
                Align(
                  alignment: Alignment.centerLeft,
                  child: Text(
                    _controller.errorMessage!,
                    style: const TextStyle(color: Colors.white, fontSize: 11),
                  ),
                ),
            ],
          ),
        );
      },
    );
  }
}

String _sourceKeyFor(PostAudioControls controls) => [
  controls.fileName,
  controls.audioUrl ?? '',
  controls.audioBytes?.length ?? 0,
  controls.audioBytes == null ? 0 : identityHashCode(controls.audioBytes),
].join(':');

bool _isPlayableUrl(String? value) {
  if (value == null || value.isEmpty) return false;
  final uri = Uri.tryParse(value);
  return uri != null &&
      (uri.scheme == 'https' || uri.scheme == 'http') &&
      uri.host.isNotEmpty;
}

String _formatDuration(Duration duration) {
  final minutes = duration.inMinutes.remainder(60).toString().padLeft(2, '0');
  final seconds = duration.inSeconds.remainder(60).toString().padLeft(2, '0');
  return '$minutes:$seconds';
}
