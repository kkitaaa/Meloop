import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

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
import 'post_interactions/post_interaction.dart';
import '../services/post_audio_playback.dart';

class PostCard extends StatefulWidget {
  final Map<String, dynamic> postData;
  final Color tealAccent;
  final Color textColor;
  final bool compact;
  final PostAudioPlaybackController? audioPlaybackController;
  final ApiClient? apiClient;
  final PostInteractionService? interactionService;
  final VoidCallback? onUnsave;
  final VoidCallback?
  onDelete; // Callback para eliminar de la vista sin recargar

  const PostCard({
    super.key,
    required this.postData,
    required this.tealAccent,
    this.audioPlaybackController,
    this.apiClient,
    this.interactionService,
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
  late final PostInteractionService _interactionService;
  bool _isSaved = false;
  bool _isSaving = false;

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    _interactionService =
        widget.interactionService ?? ApiPostInteractionService(_apiClient);
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

  Future<void> _editPost() async {
    final PostEditorResult? result = await showDialog<PostEditorResult>(
      context: context,
      builder: (context) => PostEditorDialog(
        postData: widget.postData,
        tealAccent: widget.tealAccent,
      ),
    );
    if (result != null && mounted) {
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
  }

  Future<void> _showMobileOptionsSheet({
    required String postId,
    required String user,
    required bool isAuthor,
    required bool within24Hours,
  }) async {
    HapticFeedback.mediumImpact();
    final action = await showModalBottomSheet<String>(
      context: context,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(22)),
      ),
      builder: (sheetContext) => SafeArea(
        child: Padding(
          padding: const EdgeInsets.fromLTRB(8, 10, 8, 12),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                width: 40,
                height: 4,
                margin: const EdgeInsets.only(bottom: 10),
                decoration: BoxDecoration(
                  color: const Color(0xFFD5DFDD),
                  borderRadius: BorderRadius.circular(2),
                ),
              ),
              ListTile(
                leading: const Icon(
                  Icons.share_outlined,
                  color: Color(0xFF16877F),
                ),
                title: Text(
                  'Compartir publicación',
                  style: MeloopTypography.lexend(
                    fontSize: 14.5,
                    color: const Color(0xFF182222),
                  ),
                ),
                onTap: () => Navigator.pop(sheetContext, 'share'),
              ),
              ListTile(
                leading: Icon(
                  _isSaved
                      ? Icons.bookmark_rounded
                      : Icons.bookmark_border_rounded,
                  color: const Color(0xFF16877F),
                ),
                title: Text(
                  _isSaved ? 'Quitar de Guardados' : 'Guardar publicación',
                  style: MeloopTypography.lexend(
                    fontSize: 14.5,
                    color: const Color(0xFF182222),
                  ),
                ),
                onTap: () => Navigator.pop(sheetContext, 'save'),
              ),
              if (isAuthor) ...[
                if (within24Hours)
                  ListTile(
                    leading: const Icon(
                      Icons.edit_outlined,
                      color: Color(0xFF16877F),
                    ),
                    title: Text(
                      'Editar publicación',
                      style: MeloopTypography.lexend(
                        fontSize: 14.5,
                        color: const Color(0xFF182222),
                      ),
                    ),
                    onTap: () => Navigator.pop(sheetContext, 'edit'),
                  )
                else
                  ListTile(
                    enabled: false,
                    leading: const Icon(
                      Icons.edit_off_outlined,
                      color: Colors.grey,
                    ),
                    title: Text(
                      'Tiempo de edición agotado',
                      style: MeloopTypography.lexend(
                        fontSize: 14,
                        color: Colors.grey,
                      ),
                    ),
                  ),
                ListTile(
                  leading: const Icon(
                    Icons.delete_outline_rounded,
                    color: Colors.redAccent,
                  ),
                  title: Text(
                    'Eliminar',
                    style: MeloopTypography.lexend(
                      fontSize: 14.5,
                      color: Colors.redAccent,
                    ),
                  ),
                  onTap: () => Navigator.pop(sheetContext, 'delete'),
                ),
              ] else
                ListTile(
                  leading: const Icon(
                    Icons.flag_outlined,
                    color: Colors.redAccent,
                  ),
                  title: Text(
                    'Reportar',
                    style: MeloopTypography.lexend(
                      fontSize: 14.5,
                      color: Colors.redAccent,
                    ),
                  ),
                  onTap: () => Navigator.pop(sheetContext, 'report'),
                ),
            ],
          ),
        ),
      ),
    );

    if (!mounted || action == null) return;
    switch (action) {
      case 'share':
        await _sharePost(postId);
      case 'save':
        await _toggleSaved(postId);
      case 'edit':
        await _editPost();
      case 'delete':
        _showDeleteConfirmation(context);
      case 'report':
        await showReportSheet(
          context,
          targetType: ModerationTargetType.post,
          targetId: postId,
          targetUsername: user,
        );
    }
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

    return LayoutBuilder(
      builder: (context, cardConstraints) {
        final isMobileCard = cardConstraints.maxWidth < 480;
        final isCompactStyle = widget.compact || isMobileCard;

        final cardBody = Container(
          margin: EdgeInsets.only(bottom: isMobileCard ? 14 : (widget.compact ? 16 : 22)),
          padding: EdgeInsets.symmetric(
            horizontal: isMobileCard ? 14 : (widget.compact ? 16 : 24),
            vertical: isMobileCard ? 14 : (widget.compact ? 16 : 20),
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
                      width: isMobileCard ? 38 : (widget.compact ? 40 : 48),
                      height: isMobileCard ? 38 : (widget.compact ? 40 : 48),
                      iconSize: isMobileCard ? 18 : (widget.compact ? 19 : 22),
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
                  SizedBox(width: isMobileCard ? 10 : 14),
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
                              fontSize: isMobileCard
                                  ? 16.5
                                  : (widget.compact ? 17 : 20),
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
                              fontSize: isMobileCard ? 11 : 12.5,
                              color: secondaryTextColor,
                            ),
                          )
                        else if (isEdited)
                          Text(
                            'Editado',
                            style: MeloopTypography.lexend(
                              fontSize: isMobileCard ? 10.5 : 11.5,
                              color: secondaryTextColor,
                            ),
                          ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  Padding(
                    padding: EdgeInsets.only(top: isMobileCard ? 2 : 4),
                    child: Text(
                      isEdited ? '$time (editado)' : time,
                      style: MeloopTypography.lexend(
                        fontSize: isMobileCard ? 11.5 : 13,
                        color: secondaryTextColor,
                      ),
                    ),
                  ),
                  // En escritorio mostramos el menú de 3 puntos; en móvil está
                  // disponible manteniendo presionada la publicación (long press).
                  if (!widget.compact && !isMobileCard) ...[
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
                          await _editPost();
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

              SizedBox(height: isMobileCard ? 10 : 14),

              if (title != null && title.isNotEmpty) ...[
                Text(
                  title,
                  style: MeloopTypography.lexend(
                    fontSize: isMobileCard ? 15 : 18,
                    fontWeight: FontWeight.w700,
                    color: widget.textColor,
                    height: 1.3,
                  ),
                ),
                SizedBox(height: isMobileCard ? 6 : 8),
              ],

              Text(
                content,
                maxLines: widget.compact ? 4 : null,
                overflow: widget.compact ? TextOverflow.ellipsis : null,
                style: MeloopTypography.lexend(
                  fontSize: isMobileCard ? 13 : (widget.compact ? 14 : 15),
                  color: widget.textColor,
                  height: isMobileCard ? 1.42 : 1.48,
                ),
              ),
              SizedBox(height: isMobileCard ? 10 : 14),

              if (imageProvider != null) ...[
                ClipRRect(
                  borderRadius: BorderRadius.circular(6),
                  child: Image(
                    image: imageProvider,
                    height: isCompactStyle ? 150 : 240,
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
                SizedBox(height: isMobileCard ? 10 : 14),
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
                SizedBox(height: isMobileCard ? 12 : 16),
              ],

              // Barra inferior de acciones:
              // En móvil: fila única limpia (Likes, Comentarios, Guardar, Leer más)
              // tal como en el mockup; Compartir y el menú de opciones se abren
              // manteniendo presionada la tarjeta.
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
                      padding: EdgeInsets.symmetric(
                        horizontal: isMobileCard ? 2 : 4,
                        vertical: 4,
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.mode_comment_outlined,
                            color: const Color(0xFF00AF92),
                            size: isMobileCard ? 18 : 21,
                          ),
                          SizedBox(width: isMobileCard ? 4 : 6),
                          Flexible(
                            child: Text(
                              "$comments Comentarios",
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: MeloopTypography.lexend(
                                color: const Color(0xFF00AF92),
                                fontSize: isMobileCard ? 12 : 14,
                                fontWeight: FontWeight.w500,
                              ),
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
                        padding: isMobileCard ? const EdgeInsets.all(4) : null,
                        constraints: isMobileCard
                            ? const BoxConstraints(minWidth: 30, minHeight: 30)
                            : null,
                        onPressed: _isSaving ? null : () => _toggleSaved(postId),
                        icon: _isSaving
                            ? SizedBox(
                                width: 16,
                                height: 16,
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
                                size: isMobileCard ? 20 : 22,
                              ),
                      ),
                      if (!isMobileCard)
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
                        SizedBox(width: isMobileCard ? 6 : 8),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: const Color(0xFF00B298),
                            foregroundColor: Colors.white,
                            elevation: 0,
                            minimumSize: isMobileCard
                                ? const Size(0, 32)
                                : null,
                            padding: EdgeInsets.symmetric(
                              horizontal: isMobileCard ? 16 : 20,
                              vertical: isMobileCard ? 6 : 10,
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
                              fontSize: isMobileCard ? 12 : 13.5,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                      ],
                    ],
                  );

                  final likeAction = PostLikeButton(
                    postId: postId,
                    likes: likes,
                    initiallyLiked: widget.postData['is_liked'] == true,
                    service: _interactionService,
                    compact: isMobileCard,
                    onLikesChanged: (count) {
                      widget.postData['likes'] = count;
                    },
                    onLikedChanged: (liked) =>
                        widget.postData['is_liked'] = liked,
                  );

                  if (isMobileCard) {
                    return Row(
                      children: [
                        likeAction,
                        const SizedBox(width: 10),
                        Expanded(
                          child: Align(
                            alignment: Alignment.centerLeft,
                            child: commentsAction,
                          ),
                        ),
                        const SizedBox(width: 6),
                        trailingActions,
                      ],
                    );
                  }

                  if (actionConstraints.maxWidth < 480) {
                    return SizedBox(
                      width: double.infinity,
                      child: Wrap(
                        spacing: 10,
                        runSpacing: 8,
                        crossAxisAlignment: WrapCrossAlignment.center,
                        alignment: WrapAlignment.spaceBetween,
                        children: [
                          likeAction,
                          commentsAction,
                          trailingActions,
                        ],
                      ),
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

        if (isMobileCard && !widget.compact) {
          return GestureDetector(
            behavior: HitTestBehavior.opaque,
            onLongPress: () => _showMobileOptionsSheet(
              postId: postId,
              user: user,
              isAuthor: isAuthor,
              within24Hours: within24Hours,
            ),
            child: cardBody,
          );
        }
        return cardBody;
      },
    );
  }
}
