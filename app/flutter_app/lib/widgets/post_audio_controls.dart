import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:flutter_app/services/post_audio_playback.dart';

import '../theme/meloop_typography.dart';
import 'paper/stacked_paper_card.dart';

class PostAudioControls extends StatefulWidget {
  final String postId;
  final String title;
  final String artist;
  final String? album;
  final Uint8List? audioBytes;
  final String? audioUrl;
  final String fileName;
  final Color accentColor;
  final PostAudioPlaybackController? playbackController;
  final VoidCallback? onTrackTap;

  const PostAudioControls({
    super.key,
    required this.postId,
    required this.title,
    required this.artist,
    this.album,
    required this.audioBytes,
    required this.audioUrl,
    required this.fileName,
    required this.accentColor,
    this.playbackController,
    this.onTrackTap,
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

        return LayoutBuilder(
          builder: (context, constraints) {
            final compactBar = constraints.maxWidth < 460;
            return Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
              decoration: BoxDecoration(
                gradient: const LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [Color(0xFF1A847D), Color(0xFF125E5B)],
                ),
                borderRadius: BorderRadius.circular(14),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.12),
                    blurRadius: 8,
                    offset: const Offset(0, 3),
                  ),
                ],
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Row(
                    children: [
                      // Botón circular blanco de reproducción
                      IconButton(
                        tooltip: !_hasAudio
                            ? 'Audio no disponible'
                            : isPlaying
                            ? 'Pausar audio'
                            : 'Reproducir audio',
                        onPressed: _hasAudio ? _togglePlayback : null,
                        constraints: const BoxConstraints.tightFor(
                          width: 40,
                          height: 40,
                        ),
                        padding: EdgeInsets.zero,
                        style: IconButton.styleFrom(
                          backgroundColor: Colors.white,
                          foregroundColor: const Color(0xFF136862),
                          disabledBackgroundColor: Colors.white.withValues(
                            alpha: 0.82,
                          ),
                          disabledForegroundColor: const Color(0xFF136862)
                              .withValues(alpha: 0.6),
                          shape: const CircleBorder(),
                        ),
                        icon: Icon(
                          isActive && _controller.isLoading
                              ? Icons.hourglass_top_rounded
                              : isPlaying
                              ? Icons.pause_rounded
                              : Icons.play_arrow_rounded,
                          size: 24,
                        ),
                      ),
                      const SizedBox(width: 12),

                      // Placeholder de carátula cuadrada
                      const MeloopImagePlaceholder(
                        width: 44,
                        height: 44,
                        iconSize: 20,
                        borderRadius: BorderRadius.all(Radius.circular(6)),
                        backgroundColor: Color(0xFF215E5A),
                        iconColor: Color(0xFFA7D9D4),
                        borderColor: Color(0xFF2E7873),
                      ),
                      const SizedBox(width: 12),

                      // Metadata + barra de progreso en Roboto Condensed Regular
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.end,
                              children: [
                                Expanded(
                                  child: InkWell(
                                    onTap: widget.onTrackTap,
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Text(
                                          widget.title,
                                          maxLines: 1,
                                          overflow: TextOverflow.ellipsis,
                                          style: MeloopTypography.musicPlayer(
                                            color: Colors.white,
                                            fontSize: 16,
                                            height: 1.1,
                                          ),
                                        ),
                                        const SizedBox(height: 2),
                                        Row(
                                          children: [
                                            if (widget.album != null &&
                                                widget.album!.isNotEmpty) ...[
                                              Flexible(
                                                child: Text(
                                                  widget.album!,
                                                  maxLines: 1,
                                                  overflow:
                                                      TextOverflow.ellipsis,
                                                  style:
                                                      MeloopTypography.musicPlayer(
                                                        color: Colors.white
                                                            .withValues(
                                                              alpha: 0.9,
                                                            ),
                                                        fontSize: 13.5,
                                                        height: 1.1,
                                                      ),
                                                ),
                                              ),
                                              const SizedBox(width: 10),
                                            ],
                                            Flexible(
                                              child: Text(
                                                widget.artist,
                                                maxLines: 1,
                                                overflow: TextOverflow.ellipsis,
                                                style:
                                                    MeloopTypography.musicPlayer(
                                                      color: Colors.white
                                                          .withValues(
                                                            alpha: 0.72,
                                                          ),
                                                      fontSize: 13.5,
                                                      height: 1.1,
                                                    ),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 8),
                                Flexible(
                                  child: Text(
                                    '${_formatDuration(position)} / ${_formatDuration(duration ?? Duration.zero)}',
                                    maxLines: 1,
                                    overflow: TextOverflow.ellipsis,
                                    textAlign: TextAlign.end,
                                    style: MeloopTypography.musicPlayer(
                                      color: Colors.white.withValues(
                                        alpha: 0.85,
                                      ),
                                      fontSize: 12.5,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 6),
                            if (_hasAudio)
                              SizedBox(
                                height: 14,
                                child: SliderTheme(
                                  data: SliderTheme.of(context).copyWith(
                                    trackHeight: 3.5,
                                    thumbShape: const RoundSliderThumbShape(
                                      enabledThumbRadius: 5,
                                      disabledThumbRadius: 4,
                                    ),
                                    overlayShape: const RoundSliderOverlayShape(
                                      overlayRadius: 10,
                                    ),
                                    activeTrackColor: const Color(0xFF66F0D8),
                                    inactiveTrackColor: Colors.white.withValues(
                                      alpha: 0.28,
                                    ),
                                    thumbColor: Colors.white,
                                    disabledThumbColor: Colors.white70,
                                    disabledActiveTrackColor: Colors.white70,
                                    disabledInactiveTrackColor: Colors.white24,
                                  ),
                                  child: Slider(
                                    min: 0,
                                    max: maxMilliseconds > 0
                                        ? maxMilliseconds.toDouble()
                                        : 1,
                                    value: value,
                                    onChanged: isActive && maxMilliseconds > 0
                                        ? (milliseconds) => _controller.seek(
                                            Duration(
                                              milliseconds: milliseconds
                                                  .round(),
                                            ),
                                          )
                                        : null,
                                  ),
                                ),
                              )
                            else
                              ClipRRect(
                                borderRadius: BorderRadius.circular(3),
                                child: Container(
                                  height: 3.5,
                                  width: double.infinity,
                                  color: Colors.white.withValues(alpha: 0.26),
                                ),
                              ),
                          ],
                        ),
                      ),

                      const SizedBox(width: 12),

                      // Pill "Escuchar en Spotify" (o icono compacto si hay poco ancho)
                      Material(
                        color: const Color(0xFF1EC47B),
                        borderRadius: BorderRadius.circular(20),
                        child: InkWell(
                          onTap: widget.onTrackTap,
                          borderRadius: BorderRadius.circular(20),
                          child: Padding(
                            padding: EdgeInsets.symmetric(
                              horizontal: compactBar ? 10 : 14,
                              vertical: 7,
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                if (!compactBar) ...[
                                  Text(
                                    'Escuchar en Spotify',
                                    style: MeloopTypography.lexend(
                                      color: Colors.white,
                                      fontSize: 12.5,
                                      fontWeight: FontWeight.w500,
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                ],
                                const Icon(
                                  Icons.graphic_eq,
                                  color: Colors.white,
                                  size: 18,
                                ),
                              ],
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                  if (isActive && _controller.errorMessage != null)
                    Padding(
                      padding: const EdgeInsets.only(top: 6),
                      child: Align(
                        alignment: Alignment.centerLeft,
                        child: Text(
                          _controller.errorMessage!,
                          style: MeloopTypography.musicPlayer(
                            color: Colors.white,
                            fontSize: 12,
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            );
          },
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
