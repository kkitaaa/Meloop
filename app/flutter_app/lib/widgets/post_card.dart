import 'dart:typed_data';

import 'package:flutter/material.dart';

import '../../features/comments/presentation/comments_screen.dart';
import '../../features/moderation/data/moderation_models.dart';
import '../../features/moderation/presentation/report_sheet.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../../features/music/presentation/music_track_detail_screen.dart';
import '../features/posts/post_detail_screen.dart';
import '../features/posts/presentation/post_editor_dialog.dart';
import '../features/posts/presentation/share_post_sheet.dart';
import '../services/api_client.dart';
import '../theme/meloop_typography.dart';
import 'paper/stacked_paper_card.dart';
import 'post_audio_controls.dart';
import 'post_interactions/post_interaction_demo.dart';
import '../services/post_audio_playback.dart';

class PostCard extends StatefulWidget {
  final Map<String, dynamic> postData;
  final Color tealAccent;
  final Color textColor;
  final bool compact;
  final PostAudioPlaybackController? audioPlaybackController;
  final ApiClient? apiClient;
  final VoidCallback? onUnsave;
  final VoidCallback?
  onDelete; // Callback para eliminar de la vista sin recargar

  const PostCard({
    super.key,
    required this.postData,
    required this.tealAccent,
    this.audioPlaybackController,
    this.apiClient,
    this.onUnsave,
    this.compact = false,
    this.textColor = Colors.black87,
    this.onDelete,
  });

  @override
  State<PostCard> createState() => _PostCardState();
}

class _PostCardState extends State<PostCard> {
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  bool _isSaved = false;
  bool _isSaving = false;

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    _isSaved = widget.postData['is_saved'] == true;
  }

  @override
  void dispose() {
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  @override
  void didUpdateWidget(covariant PostCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.postData != widget.postData) {
      _isSaved = widget.postData['is_saved'] == true;
    }
  }

  Future<void> _toggleSaved(String postId) async {
    setState(() => _isSaving = true);
    try {
      if (_isSaved) {
        await _apiClient.removeSavedPost(postId);
      } else {
        await _apiClient.savePost(postId);
      }
      if (!mounted) return;
      final saved = !_isSaved;
      setState(() {
        _isSaved = saved;
        _isSaving = false;
        widget.postData['is_saved'] = saved;
      });
      if (!saved) widget.onUnsave?.call();
      _showActionMessage(
        saved ? 'Publicación guardada.' : 'Publicación quitada de Guardados.',
      );
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _isSaving = false);
      _showActionMessage(error.message);
    }
  }

  Future<void> _sharePost(String postId) async {
    await showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
      ),
      builder: (_) => SharePostSheet(apiClient: _apiClient, postId: postId),
    );
  }

  void _openAuthorProfile(String user, String? authorId) {
    Navigator.push(
      context,
      MaterialPageRoute<void>(
        builder: (context) => ProfileScreen(
          username: user,
          userId: authorId,
          apiClient: _apiClient,
        ),
      ),
    );
  }

  void _openMusicDetail(String spotifyId) {
    Navigator.push(
      context,
      MaterialPageRoute<void>(
        builder: (_) => MusicDetailScreen(
          type: 'track',
          spotifyId: spotifyId,
          apiClient: _apiClient,
        ),
      ),
    );
  }

  void _showActionMessage(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), backgroundColor: widget.tealAccent),
    );
  }

  void _showDeleteConfirmation(BuildContext context) {
    showDialog(
      context: context,
      builder: (BuildContext dialogContext) {
        bool isDeleting = false;

        return StatefulBuilder(
          builder: (context, setState) {
            return AlertDialog(
              backgroundColor: Colors.white,
              title: const Text(
                "Eliminar publicación",
                style: TextStyle(fontWeight: FontWeight.bold),
              ),
              content: const Text(
                "¿Estás seguro de que deseas eliminar esta publicación? Esta acción no se puede deshacer y se borrará del servidor.",
                style: TextStyle(fontSize: 14),
              ),
              actions: [
                TextButton(
                  onPressed: isDeleting
                      ? null
                      : () => Navigator.pop(dialogContext),
                  child: const Text(
                    "Cancelar",
                    style: TextStyle(color: Colors.grey),
                  ),
                ),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.redAccent,
                    elevation: 0,
                  ),
                  onPressed: isDeleting
                      ? null
                      : () async {
                          setState(() => isDeleting = true);

                          // Simulación de petición HTTP DELETE al API Gateway
                          await Future.delayed(const Duration(seconds: 1));

                          if (!context.mounted) return;
                          Navigator.pop(dialogContext); // Cierra el modal

                          ScaffoldMessenger.of(context).showSnackBar(
                            const SnackBar(
                              content: Text(
                                "Publicación eliminada correctamente.",
                              ),
                              backgroundColor: Colors.redAccent,
                              duration: Duration(seconds: 2),
                            ),
                          );

                          // Ejecuta el callback para quitar la card de la lista local
                          if (widget.onDelete != null) {
                            widget.onDelete!();
                          }
                        },
                  child: isDeleting
                      ? const SizedBox(
                          width: 16,
                          height: 16,
                          child: CircularProgressIndicator(
                            color: Colors.white,
                            strokeWidth: 2,
                          ),
                        )
                      : const Text(
                          "Eliminar",
                          style: TextStyle(color: Colors.white),
                        ),
                ),
              ],
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final String user = widget.postData["user"] ?? "Usuario";
    final Object? authorIdData = widget.postData["user_id"];
    final String? authorId = authorIdData is String ? authorIdData : null;
    final String time = widget.postData["time"] ?? "Hace un momento";
    final String? mood = widget.postData["mood"] is String
        ? widget.postData["mood"] as String
        : null;
    final String? title = widget.postData["title"] is String
        ? widget.postData["title"] as String
        : null;
    final String content = widget.postData["content"] ?? "";
    final String? imageUrl = widget.postData["image_url"];
    final Object? imageBytesData = widget.postData["image_bytes"];
    final Uint8List? imageBytes = imageBytesData is Uint8List
        ? imageBytesData
        : null;
    final ImageProvider? imageProvider = imageBytes != null
        ? MemoryImage(imageBytes)
        : imageUrl != null && imageUrl.isNotEmpty
        ? NetworkImage(imageUrl)
        : null;
    final String? audioTitle = widget.postData["song"];
    final String? artistName = widget.postData["artist"];
    final String? albumName = widget.postData["album"] is String
        ? widget.postData["album"] as String
        : null;
    final Object? songIdData =
        widget.postData["song_id"] ??
        widget.postData["spotify_track_id"] ??
        widget.postData["track_id"];
    final String? songSpotifyId = songIdData is String && songIdData.isNotEmpty
        ? songIdData
        : null;
    final Object? audioBytesData = widget.postData["audio_bytes"];
    final Uint8List? audioBytes = audioBytesData is Uint8List
        ? audioBytesData
        : null;
    final String? audioUrl = widget.postData["audio_url"] is String
        ? widget.postData["audio_url"] as String
        : null;
    final String audioName = widget.postData["audio_name"] is String
        ? widget.postData["audio_name"] as String
        : '';
    final bool hasAudioAttachment =
        (audioBytes?.isNotEmpty ?? false) || (audioUrl?.isNotEmpty ?? false);
    final String? audioDisplayTitle = audioTitle?.isNotEmpty == true
        ? audioTitle
        : audioName.isNotEmpty
        ? audioName
        : hasAudioAttachment
        ? "Audio adjunto"
        : null;
    final int likes = widget.postData["likes"] ?? 0;
    final int comments = widget.postData["comments"] ?? 0;

    // --- Variables de lógica de negocio (Roles y Permisos) ---
    final Object? postIdData = widget.postData["id"];
    final String postId = postIdData is int || postIdData is String
        ? '$postIdData'
        : '${identityHashCode(widget.postData)}';
    final int postNumericId =
        int.tryParse(postId) ?? identityHashCode(widget.postData);
    final bool isAuthor =
        widget.postData["isAuthor"] ?? (postNumericId % 2 == 0);
    final bool isEdited = widget.postData["isEdited"] ?? false;
    final bool within24Hours = widget.postData["within24Hours"] ?? true;

    final isDarkSurface = widget.textColor.computeLuminance() > 0.5;
    final cardSurfaceColor = isDarkSurface
        ? Colors.white.withValues(alpha: 0.08)
        : const Color(0xFFFEFEFE);
    final secondaryTextColor = isDarkSurface
        ? Colors.white70
        : const Color(0xFF7B8787);

    return Container(
      margin: EdgeInsets.only(bottom: widget.compact ? 16 : 22),
      padding: EdgeInsets.symmetric(
        horizontal: widget.compact ? 16 : 24,
        vertical: widget.compact ? 16 : 20,
      ),
      decoration: BoxDecoration(
        color: cardSurfaceColor,
        borderRadius: BorderRadius.zero,
        border: Border.all(
          color: isDarkSurface
              ? Colors.white.withValues(alpha: 0.12)
              : const Color(0xFFE6ECEB),
        ),
        boxShadow: isDarkSurface
            ? null
            : [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.07),
                  blurRadius: 12,
                  offset: const Offset(0, 3),
                ),
              ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Cabecera: Avatar cuadrado placeholder + Usuario manuscrito + Hora a la derecha
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              InkWell(
                onTap: () => _openAuthorProfile(user, authorId),
                child: MeloopImagePlaceholder(
                  width: widget.compact ? 40 : 48,
                  height: widget.compact ? 40 : 48,
                  iconSize: widget.compact ? 19 : 22,
                  backgroundColor: isDarkSurface
                      ? Colors.white12
                      : const Color(0xFFE7ECEB),
                  iconColor: isDarkSurface
                      ? Colors.white60
                      : const Color(0xFF839492),
                  borderColor: isDarkSurface
                      ? Colors.white24
                      : const Color(0xFFD2DCDC),
                ),
              ),
              const SizedBox(width: 14),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    InkWell(
                      onTap: () => _openAuthorProfile(user, authorId),
                      child: Text(
                        user,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.username(
                          fontSize: widget.compact ? 17 : 20,
                          fontWeight: FontWeight.w700,
                          color: widget.textColor,
                          height: 1.15,
                        ),
                      ),
                    ),
                    if (mood != null && mood.isNotEmpty)
                      Text(
                        'Me siento: $mood',
                        style: MeloopTypography.lexend(
                          fontSize: 12.5,
                          color: secondaryTextColor,
                        ),
                      )
                    else if (isEdited)
                      Text(
                        'Editado',
                        style: MeloopTypography.lexend(
                          fontSize: 11.5,
                          color: secondaryTextColor,
                        ),
                      ),
                  ],
                ),
              ),
              const SizedBox(width: 10),
              Padding(
                padding: const EdgeInsets.only(top: 4),
                child: Text(
                  isEdited ? '$time (editado)' : time,
                  style: MeloopTypography.lexend(
                    fontSize: 13,
                    color: secondaryTextColor,
                  ),
                ),
              ),
              if (!widget.compact) ...[
                const SizedBox(width: 4),
                PopupMenuButton<String>(
                  icon: Icon(
                    Icons.more_horiz,
                    color: secondaryTextColor,
                    size: 20,
                  ),
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(
                    minWidth: 32,
                    minHeight: 32,
                  ),
                  color: Colors.white,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(10),
                  ),
                  onSelected: (value) async {
                    if (value == 'edit') {
                      final PostEditorResult? result =
                          await showDialog<PostEditorResult>(
                            context: context,
                            builder: (context) => PostEditorDialog(
                              postData: widget.postData,
                              tealAccent: widget.tealAccent,
                            ),
                          );
                      if (result != null) {
                        setState(() {
                          widget.postData["content"] = result.content;
                          widget.postData["image_bytes"] = result.imageBytes;
                          widget.postData["image_name"] = result.imageName;
                          widget.postData["image_url"] = result.imageUrl;
                          widget.postData["audio_bytes"] = result.audioBytes;
                          widget.postData["audio_name"] = result.audioName;
                          widget.postData["audio_url"] = null;
                          widget.postData["song"] = result.songTitle;
                          widget.postData["artist"] = result.artistName;
                          widget.postData["isEdited"] = true;
                        });
                      }
                    } else if (value == 'delete') {
                      _showDeleteConfirmation(context);
                    } else if (value == 'save') {
                      _toggleSaved(postId);
                    } else if (value == 'report') {
                      await showReportSheet(
                        context,
                        targetType: ModerationTargetType.post,
                        targetId: postId,
                        targetUsername: user,
                      );
                    }
                  },
                  itemBuilder: (context) {
                    if (isAuthor) {
                      return [
                        if (within24Hours)
                          const PopupMenuItem(
                            value: 'edit',
                            child: Text(
                              "Editar publicación",
                              style: TextStyle(fontSize: 13),
                            ),
                          )
                        else
                          const PopupMenuItem(
                            value: 'expired',
                            enabled: false,
                            child: Text(
                              "Tiempo de edición agotado",
                              style: TextStyle(
                                fontSize: 13,
                                color: Colors.grey,
                              ),
                            ),
                          ),
                        const PopupMenuItem(
                          value: 'delete',
                          child: Text(
                            "Eliminar",
                            style: TextStyle(
                              fontSize: 13,
                              color: Colors.redAccent,
                            ),
                          ),
                        ),
                      ];
                    } else {
                      return [
                        PopupMenuItem(
                          value: 'save',
                          child: Text(
                            _isSaved
                                ? 'Quitar de Guardados'
                                : 'Guardar publicación',
                            style: const TextStyle(fontSize: 13),
                          ),
                        ),
                        const PopupMenuItem(
                          value: 'report',
                          child: Text(
                            "Reportar",
                            style: TextStyle(
                              fontSize: 13,
                              color: Colors.redAccent,
                            ),
                          ),
                        ),
                      ];
                    }
                  },
                ),
              ],
            ],
          ),

          const SizedBox(height: 14),

          if (title != null && title.isNotEmpty) ...[
            Text(
              title,
              style: MeloopTypography.lexend(
                fontSize: 18,
                fontWeight: FontWeight.w700,
                color: widget.textColor,
                height: 1.3,
              ),
            ),
            const SizedBox(height: 8),
          ],

          Text(
            content,
            maxLines: widget.compact ? 4 : null,
            overflow: widget.compact ? TextOverflow.ellipsis : null,
            style: MeloopTypography.lexend(
              fontSize: widget.compact ? 14 : 15,
              color: widget.textColor,
              height: 1.48,
            ),
          ),
          const SizedBox(height: 14),

          if (imageProvider != null) ...[
            ClipRRect(
              borderRadius: BorderRadius.circular(6),
              child: Image(
                image: imageProvider,
                height: widget.compact ? 140 : 240,
                width: double.infinity,
                fit: BoxFit.cover,
                errorBuilder: (context, error, stackTrace) => Container(
                  height: 150,
                  color: Colors.grey[200],
                  child: const Center(
                    child: Icon(Icons.broken_image, color: Colors.grey),
                  ),
                ),
              ),
            ),
            const SizedBox(height: 14),
          ],

          if (audioDisplayTitle != null && audioDisplayTitle.isNotEmpty) ...[
            PostAudioControls(
              postId: postId,
              title: audioDisplayTitle,
              artist: artistName ?? "Artista desconocido",
              album: albumName,
              audioBytes: audioBytes,
              audioUrl: audioUrl,
              fileName: audioName,
              accentColor: widget.tealAccent,
              playbackController: widget.audioPlaybackController,
              onTrackTap: songSpotifyId != null
                  ? () => _openMusicDetail(songSpotifyId)
                  : null,
            ),
            const SizedBox(height: 16),
          ],

          // Barra inferior de acciones: Likes, Comentarios, Guardar, Compartir y "Leer más"
          LayoutBuilder(
            builder: (context, actionConstraints) {
              final commentsAction = InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute<void>(
                      builder: (context) =>
                          CommentsScreen(postData: widget.postData),
                    ),
                  );
                },
                borderRadius: BorderRadius.circular(8),
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 4,
                    vertical: 4,
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(
                        Icons.mode_comment_outlined,
                        color: Color(0xFF00AF92),
                        size: 21,
                      ),
                      const SizedBox(width: 6),
                      Text(
                        "$comments Comentarios",
                        style: MeloopTypography.lexend(
                          color: const Color(0xFF00AF92),
                          fontSize: 14,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  ),
                ),
              );

              final trailingActions = Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IconButton(
                    tooltip: _isSaved
                        ? 'Quitar de Guardados'
                        : 'Guardar publicación',
                    visualDensity: VisualDensity.compact,
                    onPressed: _isSaving ? null : () => _toggleSaved(postId),
                    icon: _isSaving
                        ? SizedBox(
                            width: 18,
                            height: 18,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: widget.tealAccent,
                            ),
                          )
                        : Icon(
                            _isSaved
                                ? Icons.bookmark_rounded
                                : Icons.bookmark_border_rounded,
                            color: _isSaved
                                ? const Color(0xFF00A98F)
                                : const Color(0xFF16877F),
                            size: 22,
                          ),
                  ),
                  IconButton(
                    tooltip: 'Compartir publicación',
                    visualDensity: VisualDensity.compact,
                    onPressed: () => _sharePost(postId),
                    icon: Icon(
                      Icons.share_outlined,
                      color: secondaryTextColor,
                      size: 20,
                    ),
                  ),
                  if (!widget.compact) ...[
                    const SizedBox(width: 8),
                    ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF00B298),
                        foregroundColor: Colors.white,
                        elevation: 0,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 20,
                          vertical: 10,
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(20),
                        ),
                      ),
                      onPressed: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute<void>(
                            builder: (_) =>
                                PostDetailScreen(postId: postNumericId),
                          ),
                        );
                      },
                      child: Text(
                        'Leer más',
                        style: MeloopTypography.lexend(
                          color: Colors.white,
                          fontSize: 13.5,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ],
              );

              final likeAction = DemoLikeButton(
                postId: postId,
                likes: likes,
                initiallyLiked: widget.postData['is_liked'] == true,
                onLikesChanged: (count) {
                  widget.postData['likes'] = count;
                },
                onLikedChanged: (liked) => widget.postData['is_liked'] = liked,
              );

              if (actionConstraints.maxWidth < 540) {
                return Wrap(
                  spacing: 12,
                  runSpacing: 8,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  alignment: WrapAlignment.spaceBetween,
                  children: [
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        likeAction,
                        const SizedBox(width: 12),
                        commentsAction,
                      ],
                    ),
                    trailingActions,
                  ],
                );
              }

              return Row(
                children: [
                  likeAction,
                  const SizedBox(width: 20),
                  commentsAction,
                  const Spacer(),
                  trailingActions,
                ],
              );
            },
          ),
        ],
      ),
    );
  }
}
