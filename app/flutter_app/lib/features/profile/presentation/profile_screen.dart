import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import '../../../core/navigation/artists_and_songs_screen.dart';
import '../../../core/navigation/desktop_navigation_shell.dart';
import '../../../core/navigation/responsive_navigation_shell.dart';
import '../../../models/feed_data.dart';
import '../../../screens/friends_screen.dart';
import '../../../services/api_client.dart';
import '../../../services/auth_session.dart';
import '../../../theme/meloop_theme_controller.dart';
import '../../../theme/meloop_typography.dart';
import '../../../widgets/desktop_app_header.dart';
import '../../../widgets/meloop_pill_tabs.dart';
import '../../../widgets/meloop_toast_pill.dart';
import '../../../widgets/mobile_app_header.dart';
import '../../../widgets/paper/stacked_paper_card.dart';
import '../../../widgets/post_card.dart';
import '../../gamification/presentation/gamification_demo_controller.dart';
import '../../posts/presentation/post_editor_dialog.dart';
import '../../posts/presentation/saved_posts_screen.dart';
import '../../search/presentation/global_search_route.dart';
import '../../social/presentation/mutual_friends_row.dart';
import '../models/profile_customization.dart';
import '../widgets/profile_editor_panel.dart';
import '../widgets/profile_stickers.dart';

enum _ProfileBlogsFilter { recent, trending }

class ProfileScreen extends StatefulWidget {
  final String? username;
  final String? userId;
  final ApiClient? apiClient;
  final bool openInventory;
  final bool insideDesktopShell;

  /// Personalización remota opcional cuando se visita el perfil de otro usuario
  /// cargado desde el backend (`GET /users/:id/profile`).
  final ProfileCustomizationConfig? initialCustomization;

  const ProfileScreen({
    super.key,
    this.username,
    this.userId,
    this.apiClient,
    this.openInventory = false,
    this.insideDesktopShell = false,
    this.initialCustomization,
  });

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen>
    with SingleTickerProviderStateMixin {
  final ScrollController _scrollController = ScrollController();
  final GlobalKey _stickerCanvasKey = GlobalKey();
  final List<Map<String, dynamic>> _posts = [];
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  late final AnimationController _rgbBorderController;
  Offset _stickerDragOffset = Offset.zero;

  bool _isInitialLoading = true;
  bool _isLoadingMore = false;
  bool _hasMorePosts = true;
  String? _postsError;
  static const int _postsPageSize = 10;

  // Estados del perfil: own, public, private, blocked
  late String _profileStatus;
  late String _displayUsername;
  String _biography =
      'Hola gente, me gusta mucho la música ambiental y la música rapidísima '
      'con mucho bajo, también el jazz fusión y otras cosas más.\n'
      'Al igual que muchos soy re fan de los gatos.';

  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _bioController = TextEditingController();

  _ProfileBlogsFilter _blogsFilter = _ProfileBlogsFilter.recent;
  bool _isEditing = false;
  ProfileEditorTab _editorInitialTab = ProfileEditorTab.style;
  String? _selectedStickerId;
  bool _songPlaying = false;

  // TODO(backend): Reemplazar artistas y amigos de muestra con los endpoints
  // reales de music-service (`GET /v1/music/preferences`) y social-service
  // (`GET /v1/social/friends`).
  static const List<({String name, String genres, String stats})>
  _favoriteArtists = [
    (
      name: 'Travis Scott',
      genres: 'Hip Hop · Trap · Experimental',
      stats: '600+ Canciones · 6 Álbumes',
    ),
    (
      name: 'JPEGMAFIA',
      genres: 'Hip Hop Experimental · Alternativo',
      stats: '100+ Canciones · 6 Álbumes',
    ),
    (
      name: 'Hideki Naganuma',
      genres: 'J-Pop · Electrónica · Funk · Breakbeat',
      stats: '25+ Canciones · 5 Álbumes',
    ),
    (
      name: 'Casiopea',
      genres: 'Jazz Fusión · Ambiente',
      stats: '25+ Canciones · 5 Álbumes',
    ),
    (
      name: 'Daft Punk',
      genres: 'Electrónica · French House · Funk',
      stats: '90+ Canciones · 4 Álbumes',
    ),
    (
      name: 'Kidd Voodoo',
      genres: 'Urbano · Reggaetón · Pop',
      stats: '60+ Canciones · 3 Álbumes',
    ),
  ];

  static const List<({String id, String name, bool online})> _profileFriends = [
    (id: '1', name: 'Vicente', online: true),
    (id: '2', name: 'Andrea', online: true),
    (id: '3', name: 'Felipe', online: true),
    (id: '4', name: 'Valentina', online: false),
  ];

  ProfileCustomizationConfig get _activeConfig => _profileStatus == 'own'
      ? MeloopThemeController.instance.ownConfig
      : (widget.initialCustomization ?? const ProfileCustomizationConfig());

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    _rgbBorderController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 4),
    );
    _displayUsername =
        widget.username ??
        AuthSession.shared.user?.username ??
        'Nombre.de.usuario';
    _usernameController.text = _displayUsername;
    _bioController.text = _biography;

    _determineProfileStatus();

    // Si se visita un perfil ajeno, mostramos su tema (o el tema por defecto
    // mientras no venga del backend) sin alterar la personalización propia.
    if (_profileStatus != 'own') {
      final otherConfig =
          widget.initialCustomization ?? const ProfileCustomizationConfig();
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        MeloopThemeController.instance.setViewingOverride(this, otherConfig);
      });
    }

    if (_profileStatus == 'own' && widget.openInventory) {
      _isEditing = true;
      _editorInitialTab = ProfileEditorTab.decorations;
    }

    if (_profileStatus == 'own' || _profileStatus == 'public') {
      _fetchInitialPosts();
    } else {
      _isInitialLoading = false;
    }

    _scrollController.addListener(_handleProfileScroll);
    _syncRgbAnimation();
  }

  void _syncRgbAnimation() {
    if (_activeConfig.borderStyle == ProfileBorderStyle.rainbowRgb) {
      if (!_rgbBorderController.isAnimating) {
        _rgbBorderController.repeat();
      }
    } else if (_rgbBorderController.isAnimating) {
      _rgbBorderController.stop();
    }
  }

  void _determineProfileStatus() {
    if (widget.userId != null && widget.userId == AuthSession.shared.user?.id) {
      _profileStatus = 'own';
    } else if (widget.username == 'UsuarioPrivado') {
      _profileStatus = 'private';
    } else if (widget.username == 'UsuarioBloqueado') {
      _profileStatus = 'blocked';
    } else if (widget.username == null && widget.userId == null) {
      _profileStatus = 'own';
    } else if (widget.username == 'MiUsuario') {
      _profileStatus = 'own';
    } else {
      _profileStatus = 'public';
    }
  }

  String? get _profileAuthorId {
    final authorId =
        widget.userId ??
        (_profileStatus == 'own' ? AuthSession.shared.user?.id : null);
    return authorId?.trim().isNotEmpty == true ? authorId : null;
  }

  @override
  void dispose() {
    MeloopThemeController.instance.clearViewingOverride(this);
    _rgbBorderController.dispose();
    _scrollController
      ..removeListener(_handleProfileScroll)
      ..dispose();
    if (_ownsApiClient) _apiClient.close();
    _usernameController.dispose();
    _bioController.dispose();
    super.dispose();
  }

  Future<void> _fetchInitialPosts() async {
    setState(() {
      _isInitialLoading = true;
      _isLoadingMore = false;
      _hasMorePosts = true;
      _postsError = null;
      _posts.clear();
    });

    final authorId = _profileAuthorId;
    if (authorId == null) {
      setState(() {
        _isInitialLoading = false;
        _postsError = 'No se pudo identificar el perfil para cargar sus blogs.';
      });
      return;
    }

    try {
      final posts = await _apiClient.getAuthorPosts(
        authorId,
        limit: _postsPageSize,
      );
      final postCards = posts.map(_toPostCardData).toList();
      if (!mounted) return;
      setState(() {
        _posts.addAll(postCards);
        _hasMorePosts = posts.length == _postsPageSize;
        _isInitialLoading = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _isInitialLoading = false;
        _postsError = error.statusCode == null
            ? error.message
            : 'HTTP ${error.statusCode}: ${error.message}';
      });
    } on FormatException catch (error) {
      if (!mounted) return;
      setState(() {
        _isInitialLoading = false;
        _postsError = error.message;
      });
    }
  }

  Future<void> _fetchMorePosts() async {
    final authorId = _profileAuthorId;
    if (_isInitialLoading ||
        _isLoadingMore ||
        !_hasMorePosts ||
        authorId == null) {
      return;
    }

    setState(() => _isLoadingMore = true);
    try {
      final posts = await _apiClient.getAuthorPosts(
        authorId,
        limit: _postsPageSize,
        offset: _posts.length,
      );
      final postCards = posts.map(_toPostCardData).toList();
      if (!mounted) return;
      setState(() {
        _posts.addAll(postCards);
        _hasMorePosts = posts.length == _postsPageSize;
        _isLoadingMore = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _isLoadingMore = false);
      showMeloopToastPill(
        context,
        message: 'No se pudieron cargar más blogs: ${error.message}',
        icon: Icons.error_outline_rounded,
      );
    } on FormatException catch (error) {
      if (!mounted) return;
      setState(() => _isLoadingMore = false);
      showMeloopToastPill(
        context,
        message: 'No se pudieron mostrar los blogs: ${error.message}',
        icon: Icons.error_outline_rounded,
      );
    }
  }

  void _handleProfileScroll() {
    if (!_scrollController.hasClients) return;
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 220) {
      _fetchMorePosts();
    }
  }

  Map<String, dynamic> _toPostCardData(Map<String, dynamic> post) {
    final authorId = post['author_id'] ?? post['user_id'];
    final username =
        post['author_username'] ?? post['user'] ?? _displayUsername;
    final content = post['content'];
    final postId = post['id'];
    final createdAt = post['created_at'] ?? post['time'];
    if ((authorId is! String) ||
        (username is! String) ||
        (content is! String) ||
        postId == null) {
      throw const FormatException(
        'El servicio devolvió datos incompletos para un blog.',
      );
    }

    final feedPost = FeedPost.fromJson({
      ...post,
      'user': username,
      'user_id': authorId,
      'time': _formatPostTime(createdAt),
      'image_url':
          post['image_url'] ?? post['url_multimedia'] ?? post['media_url'],
      'song_id': post['song_id'] ?? post['music_id'],
      'likes': post['likes'] ?? post['likes_count'] ?? 0,
      'comments': post['comments'] ?? post['comments_count'] ?? 0,
      'isAuthor': authorId == AuthSession.shared.user?.id,
      'isEdited': post['isEdited'] ?? false,
      'within24Hours': post['within24Hours'] ?? _isWithinLastDay(createdAt),
    });
    return feedPost.toPostCardData();
  }

  String _formatPostTime(Object? value) {
    if (value is! String) {
      throw const FormatException('Un blog no contiene una fecha válida.');
    }
    final createdAt = DateTime.tryParse(value);
    if (createdAt == null) {
      throw const FormatException('Un blog contiene una fecha inválida.');
    }

    final elapsed = DateTime.now().difference(createdAt.toLocal());
    if (elapsed.isNegative || elapsed.inMinutes < 1) return 'Ahora';
    if (elapsed.inHours < 1) return 'Hace ${elapsed.inMinutes} min';
    if (elapsed.inDays < 1) return 'Hace ${elapsed.inHours} h';
    return 'Hace ${elapsed.inDays} días';
  }

  bool _isWithinLastDay(Object? value) {
    if (value is! String) return false;
    final createdAt = DateTime.tryParse(value);
    return createdAt != null &&
        DateTime.now().difference(createdAt.toLocal()) <=
            const Duration(hours: 24);
  }

  // ---------------------------------------------------------------------------
  // ACCIONES DE PERSONALIZACIÓN
  // ---------------------------------------------------------------------------
  void _updateOwnCustomization(ProfileCustomizationConfig next) {
    if (_profileStatus != 'own') return;
    MeloopThemeController.instance.updateConfig(next);
    _syncRgbAnimation();
    setState(() {});
  }

  void _addStickerToCanvas(String stickerId) {
    final config = _activeConfig;
    final count = config.stickers.length;
    final placement = ProfileStickerPlacement(
      id: 'stk-${DateTime.now().microsecondsSinceEpoch}',
      stickerId: stickerId,
      // Posiciones iniciales variadas cerca de las esquinas de las tarjetas
      x: (0.20 + (count * 0.19) % 0.56).clamp(0.08, 0.84),
      y: (0.22 + (count * 0.14) % 0.42).clamp(0.08, 0.78),
      scale: 1.0,
      rotation: count.isEven ? -0.09 : 0.11,
    );
    _updateOwnCustomization(
      config.copyWith(stickers: [...config.stickers, placement]),
    );
    setState(() => _selectedStickerId = placement.id);
  }

  void _updateStickerPlacement(ProfileStickerPlacement updated) {
    final config = _activeConfig;
    final nextList = config.stickers
        .map((item) => item.id == updated.id ? updated : item)
        .toList(growable: false);
    _updateOwnCustomization(config.copyWith(stickers: nextList));
  }

  void _removeStickerById(String id) {
    final config = _activeConfig;
    final nextList = config.stickers
        .where((item) => item.id != id)
        .toList(growable: false);
    _updateOwnCustomization(config.copyWith(stickers: nextList));
    if (_selectedStickerId == id) {
      setState(() => _selectedStickerId = null);
    }
  }

  void _demoLevelUp() {
    final gamification = GamificationDemoController.instance;
    final remaining = gamification.experienceNeeded - gamification.xp;
    final steps = math.max(1, (remaining / 10).ceil());
    for (var i = 0; i < steps; i++) {
      gamification.grantLikeXp();
    }
    showMeloopToastPill(
      context,
      message:
          'Nivel ${gamification.level} alcanzado · Revisa nuevos desbloqueos',
      icon: Icons.auto_awesome_rounded,
    );
    setState(() {});
  }

  void _toggleEditMode({ProfileEditorTab tab = ProfileEditorTab.style}) {
    final isDesktop = MediaQuery.sizeOf(context).width > 960;
    if (isDesktop) {
      setState(() {
        _editorInitialTab = tab;
        _isEditing = !_isEditing;
        if (!_isEditing) _selectedStickerId = null;
      });
      return;
    }

    // En móvil abrimos el estudio como hoja inferior para conservar toda la vista
    setState(() => _isEditing = true);
    showModalBottomSheet<void>(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (sheetContext) => FractionallySizedBox(
        heightFactor: 0.78,
        child: AnimatedBuilder(
          animation: Listenable.merge([
            MeloopThemeController.instance,
            GamificationDemoController.instance,
          ]),
          builder: (context, _) => ProfileEditorPanel(
            config: _activeConfig,
            userLevel: GamificationDemoController.instance.level,
            rgbFrameUnlocked:
                GamificationDemoController.instance.rgbFrameUnlocked,
            usernameController: _usernameController,
            bioController: _bioController,
            selectedStickerId: _selectedStickerId,
            initialTab: tab,
            onConfigChanged: _updateOwnCustomization,
            onProfileTextChanged: _syncProfileTextFields,
            onAddSticker: _addStickerToCanvas,
            onUpdateSelectedSticker: _updateStickerPlacement,
            onRemoveSelectedSticker: () {
              if (_selectedStickerId != null) {
                _removeStickerById(_selectedStickerId!);
              }
            },
            onDemoLevelUp: _demoLevelUp,
            onResetDefault: () {
              MeloopThemeController.instance.resetToDefault();
              setState(() => _selectedStickerId = null);
            },
            onClose: () => Navigator.of(sheetContext).pop(),
          ),
        ),
      ),
    ).whenComplete(() {
      if (mounted) {
        setState(() {
          _isEditing = false;
          _selectedStickerId = null;
        });
      }
    });
  }

  void _syncProfileTextFields() {
    setState(() {
      final trimmedName = _usernameController.text.trim();
      if (trimmedName.isNotEmpty) {
        _displayUsername = trimmedName;
      }
      _biography = _bioController.text.trim();
    });
  }

  Future<void> _openCreateBlogDialog() async {
    final created = await showDialog<PostEditorResult>(
      context: context,
      builder: (_) => PostEditorDialog(
        tealAccent: _activeConfig.colorTheme.primary,
        apiClient: _apiClient,
      ),
    );
    if (!mounted || created == null) return;
    setState(() {
      _posts.insert(0, {
        'id': 'local-${DateTime.now().millisecondsSinceEpoch}',
        'user': _displayUsername,
        'user_id': _profileAuthorId ?? 'local-user',
        'time': 'Ahora',
        'content': created.content,
        'mood': _activeConfig.mood,
        'song': created.songTitle,
        'artist': created.artistName,
        'likes': 0,
        'comments': 0,
        'isAuthor': true,
        'isEdited': false,
        'within24Hours': true,
      });
    });
    showMeloopToastPill(
      context,
      message: 'Blog publicado',
      xpAwarded: created.xpAwarded ?? 15,
      icon: Icons.check_circle_rounded,
    );
  }

  void _openFriendsSection() {
    final desktopShell = DesktopShellScope.maybeOf(context);
    if (desktopShell != null) {
      desktopShell.openFriends();
      return;
    }
    final mobileShell = MobileShellScope.maybeOf(context);
    if (mobileShell != null) {
      mobileShell.openFriends();
      return;
    }
    Navigator.of(
      context,
    ).push<void>(MaterialPageRoute<void>(builder: (_) => const FriendsScreen()));
  }

  void _openArtistsSection({String? query}) {
    final desktopShell = DesktopShellScope.maybeOf(context);
    if (desktopShell != null) {
      desktopShell.openArtists(initialQuery: query);
      return;
    }
    final mobileShell = MobileShellScope.maybeOf(context);
    if (mobileShell != null) {
      mobileShell.openArtists(initialQuery: query);
      return;
    }
    Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (_) => ArtistsAndSongsScreen(initialQuery: query),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.sizeOf(context).width;
    final bool isDesktop = screenWidth > 960;
    final bool insideDesktopShell =
        widget.insideDesktopShell || DesktopShellScope.maybeOf(context) != null;
    final bool insideMobileShell = MobileShellScope.maybeOf(context) != null;

    return AnimatedBuilder(
      animation: Listenable.merge([
        MeloopThemeController.instance,
        GamificationDemoController.instance,
      ]),
      builder: (context, _) {
        final config = _activeConfig;
        final accent = config.colorTheme.primary;

        return Scaffold(
          backgroundColor: Colors.transparent,
          body: Column(
            children: [
              if (!insideDesktopShell && !insideMobileShell)
                _buildTopBar(isDesktop, context),
              if (_isEditing && _profileStatus == 'own')
                _buildEditModeBanner(accent),
              Expanded(
                child: isDesktop
                    ? _buildDesktopWithOptionalEditor(config)
                    : _buildMobileLayout(config),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _buildEditModeBanner(Color accent) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
      decoration: BoxDecoration(
        color: accent,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.14),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Row(
        children: [
          const Icon(Icons.auto_awesome_rounded, color: Colors.white, size: 17),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              'Modo personalización activo · Arrastra tus stickers sobre cualquier parte del perfil',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: MeloopTypography.lexend(
                color: Colors.white,
                fontSize: 12.5,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          TextButton(
            onPressed: () => setState(() {
              _isEditing = false;
              _selectedStickerId = null;
            }),
            style: TextButton.styleFrom(
              foregroundColor: Colors.white,
              backgroundColor: Colors.white.withValues(alpha: 0.2),
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(18),
              ),
            ),
            child: Text(
              'Listo',
              style: MeloopTypography.lexend(
                color: Colors.white,
                fontSize: 12,
                fontWeight: FontWeight.w700,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDesktopWithOptionalEditor(ProfileCustomizationConfig config) {
    final gamification = GamificationDemoController.instance;
    return Row(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Expanded(child: _buildDesktopProfileCanvas(config)),
        if (_isEditing && _profileStatus == 'own')
          SizedBox(
            width: 390,
            child: Padding(
              padding: const EdgeInsets.fromLTRB(0, 14, 18, 18),
              child: ProfileEditorPanel(
                config: config,
                userLevel: gamification.level,
                rgbFrameUnlocked: gamification.rgbFrameUnlocked,
                usernameController: _usernameController,
                bioController: _bioController,
                selectedStickerId: _selectedStickerId,
                initialTab: _editorInitialTab,
                onConfigChanged: _updateOwnCustomization,
                onProfileTextChanged: _syncProfileTextFields,
                onAddSticker: _addStickerToCanvas,
                onUpdateSelectedSticker: _updateStickerPlacement,
                onRemoveSelectedSticker: () {
                  if (_selectedStickerId != null) {
                    _removeStickerById(_selectedStickerId!);
                  }
                },
                onDemoLevelUp: _demoLevelUp,
                onResetDefault: () {
                  MeloopThemeController.instance.resetToDefault();
                  setState(() => _selectedStickerId = null);
                },
                onClose: () => setState(() {
                  _isEditing = false;
                  _selectedStickerId = null;
                }),
              ),
            ),
          ),
      ],
    );
  }

  // ---------------------------------------------------------------------------
  // LIENZO PRINCIPAL DE PERFIL (ESCRITORIO) + CAPA LIBRE DE STICKERS
  // ---------------------------------------------------------------------------
  Widget _buildDesktopProfileCanvas(ProfileCustomizationConfig config) {
    return SingleChildScrollView(
      controller: _scrollController,
      padding: const EdgeInsets.fromLTRB(28, 22, 28, 42),
      child: Center(
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 1320),
          child: LayoutBuilder(
            builder: (context, constraints) {
              return Stack(
                key: _stickerCanvasKey,
                clipBehavior: Clip.none,
                children: [
                  _buildDesktopStructuredContent(config),
                  if (_profileStatus != 'blocked')
                    Positioned.fill(
                      child: _buildStickersOverlay(
                        config: config,
                        canvasWidth: constraints.maxWidth,
                      ),
                    ),
                ],
              );
            },
          ),
        ),
      ),
    );
  }

  Widget _buildDesktopStructuredContent(ProfileCustomizationConfig config) {
    if (_profileStatus == 'blocked') {
      return Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(width: 248, child: _buildPolaroidProfile(config)),
          const SizedBox(width: 26),
          Expanded(child: _buildBlockedState(config)),
        ],
      );
    }

    final layout = config.layout;
    final hasRightColumn =
        _profileStatus != 'private' &&
        layout.preset != ProfileLayoutPreset.compactFocus &&
        (layout.showFavoriteSong || layout.showFriends);

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Columna izquierda: Polaroid + Nivel (+ Canción/Amigos si preset Enfoque blogs)
        SizedBox(
          width: 248,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _buildPolaroidProfile(config),
              const SizedBox(height: 18),
              _buildLevelCard(config),
              if (layout.preset == ProfileLayoutPreset.compactFocus &&
                  _profileStatus != 'private' &&
                  layout.showFavoriteSong) ...[
                const SizedBox(height: 18),
                _buildFavoriteSongCard(config),
              ],
              if (layout.preset == ProfileLayoutPreset.compactFocus &&
                  _profileStatus != 'private' &&
                  layout.showFriends) ...[
                const SizedBox(height: 18),
                _buildFriendsCard(config),
              ],
            ],
          ),
        ),
        const SizedBox(width: 26),

        // Área derecha principal:
        // 1) Barra superior de píldoras y acciones alineada arriba
        // 2) Bloques principales (Sobre mí + Artistas | Canción + Amigos) en la misma línea base
        // 3) Sección inferior de Blogs
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              _buildStatsAndActionsBar(config),
              if (_profileStatus == 'public' &&
                  widget.userId != null &&
                  widget.userId!.isNotEmpty) ...[
                const SizedBox(height: 10),
                MutualFriendsRow(targetUserId: widget.userId!),
              ],
              const SizedBox(height: 16),
              if (_profileStatus == 'private')
                _buildPrivateState(config)
              else ...[
                Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Expanded(
                      flex: hasRightColumn
                          ? (layout.preset == ProfileLayoutPreset.showcase
                                ? 68
                                : 64)
                          : 100,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.stretch,
                        children: _buildCenterInfoBlocks(config),
                      ),
                    ),
                    if (hasRightColumn) ...[
                      const SizedBox(width: 22),
                      Expanded(
                        flex: layout.preset == ProfileLayoutPreset.showcase
                            ? 32
                            : 36,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            if (layout.showFavoriteSong) ...[
                              _buildFavoriteSongCard(config),
                              if (layout.showFriends)
                                const SizedBox(height: 18),
                            ],
                            if (layout.showFriends) _buildFriendsCard(config),
                          ],
                        ),
                      ),
                    ],
                  ],
                ),
                const SizedBox(height: 24),
                _buildBlogsSection(config),
              ],
            ],
          ),
        ),
      ],
    );
  }

  List<Widget> _buildCenterInfoBlocks(ProfileCustomizationConfig config) {
    final bio = _buildAboutMeCard(config);
    final artists = config.layout.showFavoriteArtists
        ? _buildFavoriteArtistsCard(config)
        : null;

    if (artists == null) return [bio];
    if (config.layout.artistsBeforeBio) {
      return [artists, const SizedBox(height: 18), bio];
    }
    return [bio, const SizedBox(height: 18), artists];
  }

  // ---------------------------------------------------------------------------
  // MÓVIL / ANDROID
  // ---------------------------------------------------------------------------
  Widget _buildMobileLayout(ProfileCustomizationConfig config) {
    return SingleChildScrollView(
      controller: _scrollController,
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 148),
      child: LayoutBuilder(
        builder: (context, constraints) => Stack(
          key: _stickerCanvasKey,
          clipBehavior: Clip.none,
          children: [
            Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Center(
                  child: SizedBox(
                    width: 250,
                    child: _buildPolaroidProfile(config),
                  ),
                ),
                const SizedBox(height: 16),
                if (_profileStatus == 'blocked')
                  _buildBlockedState(config)
                else ...[
                  _buildLevelCard(config),
                  const SizedBox(height: 14),
                  _buildStatsAndActionsBar(config, isMobile: true),
                  if (_profileStatus == 'public' &&
                      widget.userId != null &&
                      widget.userId!.isNotEmpty) ...[
                    const SizedBox(height: 10),
                    MutualFriendsRow(targetUserId: widget.userId!),
                  ],
                  const SizedBox(height: 16),
                  if (_profileStatus == 'private')
                    _buildPrivateState(config)
                  else ...[
                    ..._buildCenterInfoBlocks(config),
                    if (config.layout.showFavoriteSong) ...[
                      const SizedBox(height: 16),
                      _buildFavoriteSongCard(config),
                    ],
                    if (config.layout.showFriends) ...[
                      const SizedBox(height: 16),
                      _buildFriendsCard(config),
                    ],
                    const SizedBox(height: 20),
                    _buildBlogsSection(config),
                  ],
                ],
              ],
            ),
            if (_profileStatus != 'blocked')
              Positioned.fill(
                child: _buildStickersOverlay(
                  config: config,
                  canvasWidth: constraints.maxWidth,
                ),
              ),
          ],
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // CAPA INTERACTIVA DE STICKERS LIBRES (ARRASTRABLES SOBRE CUALQUIER TARJETA)
  // ---------------------------------------------------------------------------
  Widget _buildStickersOverlay({
    required ProfileCustomizationConfig config,
    required double canvasWidth,
  }) {
    if (config.stickers.isEmpty) return const SizedBox.shrink();

    return LayoutBuilder(
      builder: (context, box) {
        final width = box.maxWidth > 0 ? box.maxWidth : canvasWidth;
        final height = box.maxHeight > 0 ? box.maxHeight : 900.0;

        return Stack(
          clipBehavior: Clip.none,
          children: [
            for (final sticker in config.stickers)
              _buildSinglePlacedSticker(
                sticker: sticker,
                canvasWidth: width,
                canvasHeight: height,
                accent: config.colorTheme.primary,
              ),
          ],
        );
      },
    );
  }

  Widget _buildSinglePlacedSticker({
    required ProfileStickerPlacement sticker,
    required double canvasWidth,
    required double canvasHeight,
    required Color accent,
  }) {
    const baseSize = 86.0;
    final displaySize = baseSize * sticker.scale;
    final double left = (sticker.x * canvasWidth - displaySize / 2)
        .clamp(0.0, math.max(0.0, canvasWidth - displaySize))
        .toDouble();
    final double top = (sticker.y * canvasHeight - displaySize / 2)
        .clamp(0.0, math.max(0.0, canvasHeight - displaySize))
        .toDouble();
    final isSelected = _isEditing && _selectedStickerId == sticker.id;

    final artwork = Transform.rotate(
      angle: sticker.rotation,
      child: ProfileStickerArtwork(
        stickerId: sticker.stickerId,
        size: displaySize,
      ),
    );

    if (!_isEditing || _profileStatus != 'own') {
      return Positioned(
        left: left,
        top: top,
        child: IgnorePointer(child: artwork),
      );
    }

    return Positioned(
      left: left,
      top: top,
      child: MouseRegion(
        cursor: SystemMouseCursors.grab,
        child: GestureDetector(
          behavior: HitTestBehavior.opaque,
          onTap: () => setState(() => _selectedStickerId = sticker.id),
          onPanStart: (details) {
            setState(() => _selectedStickerId = sticker.id);
            final box =
                _stickerCanvasKey.currentContext?.findRenderObject()
                    as RenderBox?;
            if (box != null && box.hasSize) {
              final pointerInCanvas = box.globalToLocal(details.globalPosition);
              final centerInCanvas = Offset(
                sticker.x * box.size.width,
                sticker.y * box.size.height,
              );
              _stickerDragOffset = centerInCanvas - pointerInCanvas;
            } else {
              _stickerDragOffset = Offset.zero;
            }
          },
          onPanUpdate: (details) {
            final box =
                _stickerCanvasKey.currentContext?.findRenderObject()
                    as RenderBox?;
            if (box != null &&
                box.hasSize &&
                box.size.width > 0 &&
                box.size.height > 0) {
              final pointerInCanvas = box.globalToLocal(details.globalPosition);
              final nextCenter = pointerInCanvas + _stickerDragOffset;
              final nextX = (nextCenter.dx / box.size.width).clamp(0.02, 0.94);
              final nextY = (nextCenter.dy / box.size.height).clamp(0.02, 0.92);
              _updateStickerPlacement(sticker.copyWith(x: nextX, y: nextY));
            } else {
              final nextX =
                  ((left + details.delta.dx + displaySize / 2) / canvasWidth)
                      .clamp(0.02, 0.94);
              final nextY =
                  ((top + details.delta.dy + displaySize / 2) / canvasHeight)
                      .clamp(0.02, 0.92);
              _updateStickerPlacement(sticker.copyWith(x: nextX, y: nextY));
            }
          },
          child: Stack(
            clipBehavior: Clip.none,
            children: [
              Container(
                padding: const EdgeInsets.all(3),
                decoration: isSelected
                    ? BoxDecoration(
                        borderRadius: BorderRadius.circular(14),
                        border: Border.all(color: accent, width: 2),
                        color: accent.withValues(alpha: 0.08),
                      )
                    : null,
                child: artwork,
              ),
              if (isSelected)
                Positioned(
                  top: -10,
                  right: -10,
                  child: GestureDetector(
                    onTap: () => _removeStickerById(sticker.id),
                    child: Container(
                      width: 24,
                      height: 24,
                      decoration: const BoxDecoration(
                        color: Color(0xFFE04848),
                        shape: BoxShape.circle,
                      ),
                      child: const Icon(
                        Icons.close_rounded,
                        size: 14,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // TARJETAS DEL PERFIL (COHERENTES CON MOCKUP 1 Y MOCKUP 2)
  // ---------------------------------------------------------------------------
  Widget _buildPolaroidProfile(ProfileCustomizationConfig config) {
    final pinColor = config.activePinOption.color;
    final borderStyle = config.borderStyle;
    final isDark = config.appStyle == MeloopAppStyle.neonNight;
    final textColor = isDark ? Colors.white : const Color(0xFF1A2424);
    final mutedColor = isDark
        ? Colors.white.withValues(alpha: 0.72)
        : const Color(0xFF6B7E7C);

    final polaroidBody = StackedPaperCard(
      textureAsset: 'assets/textures/photopaper.png',
      tintColor: config.appStyle == MeloopAppStyle.paper
          ? null
          : config.colorTheme.softCardTint,
      backAngleDegrees: -1.6,
      padding: const EdgeInsets.fromLTRB(14, 18, 14, 16),
      child: Column(
        children: [
          _buildFramedAvatar(borderStyle, config.colorTheme.primary),
          const SizedBox(height: 14),
          Text(
            _displayUsername,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            textAlign: TextAlign.center,
            // Regla tipográfica de Meloop: tu propio nombre bajo la polaroid
            // usa Kalam (fuente manuscrita).
            style: MeloopTypography.username(
              fontSize: 18,
              fontWeight: FontWeight.w700,
              color: textColor,
            ),
          ),
          if (_profileStatus != 'blocked' && config.layout.showMood) ...[
            const SizedBox(height: 4),
            Text(
              'Me siento: ${config.mood}',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: MeloopTypography.lexend(fontSize: 12.5, color: mutedColor),
            ),
          ],
        ],
      ),
    );

    return Stack(
      alignment: Alignment.topCenter,
      clipBehavior: Clip.none,
      children: [
        Padding(padding: const EdgeInsets.only(top: 14), child: polaroidBody),
        Positioned(
          top: -6,
          child: CustomizablePin(color: pinColor, size: 46),
        ),
      ],
    );
  }

  Widget _buildFramedAvatar(ProfileBorderStyle border, Color accent) {
    const photo = AspectRatio(
      aspectRatio: 1.0,
      child: MeloopImagePlaceholder(
        iconSize: 48,
        backgroundColor: Color(0xFFE7EEEC),
        iconColor: Color(0xFF7B908D),
        borderColor: null,
      ),
    );

    switch (border) {
      case ProfileBorderStyle.none:
        return photo;
      case ProfileBorderStyle.pastelRibbon:
        return Container(
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            border: Border.all(color: const Color(0xFFFF8EC4), width: 3),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFFFF8EC4).withValues(alpha: 0.35),
                blurRadius: 10,
              ),
            ],
          ),
          child: photo,
        );
      case ProfileBorderStyle.goldenFrame:
        return Container(
          padding: const EdgeInsets.all(4),
          decoration: const BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [Color(0xFFFFE082), Color(0xFFD49A1A), Color(0xFFFFE082)],
            ),
          ),
          child: photo,
        );
      case ProfileBorderStyle.neonCyan:
        return Container(
          padding: const EdgeInsets.all(3.5),
          decoration: BoxDecoration(
            border: Border.all(color: const Color(0xFF26F7DF), width: 3),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFF26F7DF).withValues(alpha: 0.55),
                blurRadius: 14,
              ),
            ],
          ),
          child: photo,
        );
      case ProfileBorderStyle.aeroGlass:
        return Container(
          padding: const EdgeInsets.all(4.5),
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [Color(0xFFE8FAFF), Color(0xFF39B7E8), Color(0xFFB5EEFF)],
            ),
            borderRadius: BorderRadius.circular(10),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFF39B7E8).withValues(alpha: 0.45),
                blurRadius: 12,
              ),
            ],
          ),
          child: ClipRRect(borderRadius: BorderRadius.circular(6), child: photo),
        );
      case ProfileBorderStyle.rainbowRgb:
        return AnimatedBuilder(
          animation: _rgbBorderController,
          builder: (context, child) {
            final angle = _rgbBorderController.value * 2 * math.pi;
            return Container(
              padding: const EdgeInsets.all(4),
              decoration: BoxDecoration(
                gradient: SweepGradient(
                  transform: GradientRotation(angle),
                  colors: const [
                    Color(0xFFFF4D8D),
                    Color(0xFFFFC93C),
                    Color(0xFF00E5A3),
                    Color(0xFF38BDF8),
                    Color(0xFFA855F7),
                    Color(0xFFFF4D8D),
                  ],
                ),
                boxShadow: [
                  BoxShadow(
                    color: accent.withValues(alpha: 0.35),
                    blurRadius: 12,
                  ),
                ],
              ),
              child: child,
            );
          },
          child: photo,
        );
    }
  }

  Widget _buildLevelCard(ProfileCustomizationConfig config) {
    final gamification = GamificationDemoController.instance;
    final accent = config.colorTheme.primary;
    final progressRatio = (gamification.xp / gamification.experienceNeeded)
        .clamp(0.0, 1.0);
    final rankLabel = gamification.level >= 15
        ? 'Coleccionista'
        : gamification.level >= 10
        ? 'Melómano'
        : gamification.level >= 5
        ? 'Novato'
        : 'Aún empezando';
    final isDark = config.appStyle == MeloopAppStyle.neonNight;

    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      tintColor: config.colorTheme.xpCardTint,
      backAngleDegrees: 1.35,
      padding: const EdgeInsets.fromLTRB(14, 14, 14, 12),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.star_border_rounded, size: 17, color: accent),
              const SizedBox(width: 5),
              Flexible(
                child: Text(
                  'Nivel ${gamification.level} • $rankLabel',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 12.5,
                    fontWeight: FontWeight.w700,
                    color: isDark ? Colors.white : const Color(0xFF1A2929),
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Container(
            height: 14,
            decoration: BoxDecoration(
              color: Colors.white.withValues(alpha: isDark ? 0.16 : 0.9),
              borderRadius: BorderRadius.circular(10),
              border: Border.all(
                color: accent.withValues(alpha: 0.45),
                width: 1.2,
              ),
            ),
            padding: const EdgeInsets.all(2),
            child: Align(
              alignment: Alignment.centerLeft,
              child: FractionallySizedBox(
                widthFactor: math.max(0.06, progressRatio),
                child: Container(
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [config.colorTheme.headerTop, accent],
                    ),
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(height: 7),
          Text(
            '${gamification.xp} / ${gamification.experienceNeeded} EXP',
            style: MeloopTypography.lexend(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: isDark
                  ? Colors.white.withValues(alpha: 0.78)
                  : const Color(0xFF5D7675),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatsAndActionsBar(
    ProfileCustomizationConfig config, {
    bool isMobile = false,
  }) {
    final accent = config.colorTheme.primary;
    final level = GamificationDemoController.instance.level;
    final maxStickers = ProfileCustomizationConfig.maxStickersForLevel(level);
    final placedStickers = config.stickers.length;
    final decorationsCount = config.equippedDecorationsCount;

    final pills = <Widget>[
      _statCapsule(
        'Amigos: ${_profileFriends.length}',
        config,
        onTap: _openFriendsSection,
      ),
      _statCapsule(
        'Stickers: $placedStickers/$maxStickers',
        config,
        onTap: _profileStatus == 'own'
            ? () => _toggleEditMode(tab: ProfileEditorTab.stickers)
            : null,
      ),
      _statCapsule(
        'Decoraciones: $decorationsCount',
        config,
        onTap: _profileStatus == 'own'
            ? () => _toggleEditMode(tab: ProfileEditorTab.decorations)
            : null,
      ),
    ];

    final actions = <Widget>[
      if (_profileStatus == 'own') ...[
        FilledButton(
          style: FilledButton.styleFrom(
            backgroundColor: accent,
            foregroundColor: Colors.white,
            elevation: 2,
            padding: const EdgeInsets.symmetric(horizontal: 18, vertical: 10),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(22),
            ),
          ),
          onPressed: () => _toggleEditMode(tab: ProfileEditorTab.style),
          child: Text(
            'Editar perfil',
            style: MeloopTypography.lexend(
              fontSize: 12.5,
              fontWeight: FontWeight.w700,
              color: Colors.white,
            ),
          ),
        ),
        OutlinedButton.icon(
          style: OutlinedButton.styleFrom(
            backgroundColor: Colors.white.withValues(alpha: 0.94),
            foregroundColor: accent,
            side: BorderSide(color: accent.withValues(alpha: 0.6)),
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(22),
            ),
          ),
          onPressed: () => Navigator.push(
            context,
            MaterialPageRoute<void>(
              builder: (_) => SavedPostsScreen(apiClient: widget.apiClient),
            ),
          ),
          icon: Icon(Icons.bookmark_border_rounded, color: accent, size: 16),
          label: Text(
            'Guardados',
            style: MeloopTypography.lexend(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: accent,
            ),
          ),
        ),
      ] else
        FilledButton.icon(
          style: FilledButton.styleFrom(
            backgroundColor: accent,
            foregroundColor: Colors.white,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(22),
            ),
          ),
          onPressed: () {
            showMeloopToastPill(
              context,
              message: 'Solicitud enviada a $_displayUsername',
              icon: Icons.person_add_alt_1_rounded,
            );
          },
          icon: const Icon(Icons.person_add_alt_1_rounded, size: 16),
          label: Text(
            'Añadir amigo',
            style: MeloopTypography.lexend(
              fontSize: 12.5,
              fontWeight: FontWeight.w700,
              color: Colors.white,
            ),
          ),
        ),
    ];

    return Wrap(
      spacing: 10,
      runSpacing: 8,
      alignment: isMobile ? WrapAlignment.center : WrapAlignment.start,
      crossAxisAlignment: WrapCrossAlignment.center,
      children: [...pills, ...actions],
    );
  }

  Widget _statCapsule(
    String text,
    ProfileCustomizationConfig config, {
    VoidCallback? onTap,
  }) {
    final isDark = config.appStyle == MeloopAppStyle.neonNight;
    return Material(
      color: isDark ? const Color(0xFF1D2635) : Colors.white,
      borderRadius: BorderRadius.circular(22),
      elevation: 2,
      shadowColor: Colors.black.withValues(alpha: 0.14),
      child: InkWell(
        borderRadius: BorderRadius.circular(22),
        onTap: onTap,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 9),
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(22),
            border: Border.all(
              color: config.colorTheme.primary.withValues(
                alpha: isDark ? 0.5 : 0.22,
              ),
            ),
          ),
          child: Text(
            text,
            style: MeloopTypography.lexend(
              fontSize: 12.5,
              fontWeight: FontWeight.w700,
              color: isDark ? Colors.white : const Color(0xFF1A2828),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildAboutMeCard(ProfileCustomizationConfig config) {
    final isDark = config.appStyle == MeloopAppStyle.neonNight;
    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      tintColor: config.appStyle == MeloopAppStyle.paper
          ? null
          : config.colorTheme.softCardTint,
      backAngleDegrees: -0.95,
      padding: const EdgeInsets.fromLTRB(22, 18, 22, 20),
      child: SizedBox(
        width: double.infinity,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Transform.rotate(
              angle: -0.02,
              alignment: Alignment.centerLeft,
              child: Text(
                'Sobre mí:',
                style: MeloopTypography.momoDisplay(
                  fontSize: 22,
                  color: isDark ? Colors.white : const Color(0xFF182424),
                ),
              ),
            ),
            const SizedBox(height: 10),
            Text(
              _biography.isEmpty
                  ? 'Aún no has escrito tu biografía. Presiona "Editar perfil" para contar qué música te apasiona.'
                  : _biography,
              style: MeloopTypography.lexend(
                fontSize: 13.5,
                height: 1.55,
                color: isDark
                    ? Colors.white.withValues(alpha: 0.88)
                    : const Color(0xFF283737),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFavoriteArtistsCard(ProfileCustomizationConfig config) {
    final accent = config.colorTheme.primary;
    final isDark = config.appStyle == MeloopAppStyle.neonNight;

    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      tintColor: config.appStyle == MeloopAppStyle.paper
          ? null
          : config.colorTheme.softCardTint,
      backAngleDegrees: 1.05,
      padding: const EdgeInsets.fromLTRB(18, 16, 18, 14),
      child: SizedBox(
        width: double.infinity,
        child: Column(
          children: [
          Text(
            'Artistas que me gustan',
            style: MeloopTypography.lexend(
              fontSize: 15,
              fontWeight: FontWeight.w700,
              color: isDark ? Colors.white : const Color(0xFF1A2626),
            ),
          ),
          const SizedBox(height: 12),
          LayoutBuilder(
            builder: (context, constraints) {
              final twoCols = constraints.maxWidth >= 430;
              if (!twoCols) {
                return Column(
                  children: [
                    for (final artist in _favoriteArtists.take(4))
                      Padding(
                        padding: const EdgeInsets.only(bottom: 8),
                        child: _buildArtistMiniTile(artist, config),
                      ),
                  ],
                );
              }
              final leftCol = <Widget>[];
              final rightCol = <Widget>[];
              for (var i = 0; i < _favoriteArtists.length; i++) {
                final tile = Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: _buildArtistMiniTile(_favoriteArtists[i], config),
                );
                if (i.isEven) {
                  leftCol.add(tile);
                } else {
                  rightCol.add(tile);
                }
              }
              return Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(child: Column(children: leftCol)),
                  const SizedBox(width: 10),
                  Expanded(child: Column(children: rightCol)),
                ],
              );
            },
          ),
          const SizedBox(height: 6),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: accent,
              foregroundColor: Colors.white,
              elevation: 0,
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 6),
              minimumSize: const Size(0, 32),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(18),
              ),
            ),
            onPressed: () => _openArtistsSection(),
            child: Text(
              'Ver más',
              style: MeloopTypography.lexend(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: Colors.white,
              ),
            ),
          ),
        ],
        ),
      ),
    );
  }

  Widget _buildArtistMiniTile(
    ({String name, String genres, String stats}) artist,
    ProfileCustomizationConfig config,
  ) {
    final isDark = config.appStyle == MeloopAppStyle.neonNight;
    return Material(
      color: isDark
          ? Colors.white.withValues(alpha: 0.06)
          : Colors.white.withValues(alpha: 0.92),
      child: InkWell(
        onTap: () => _openArtistsSection(query: artist.name),
        child: Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            border: Border.all(
              color: isDark
                  ? Colors.white.withValues(alpha: 0.12)
                  : const Color(0xFFE2ECEA),
            ),
          ),
          child: Row(
            children: [
              const MeloopImagePlaceholder(
                width: 40,
                height: 40,
                iconSize: 18,
                borderColor: null,
              ),
              const SizedBox(width: 10),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      artist.name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: MeloopTypography.lexend(
                        fontSize: 12.5,
                        fontWeight: FontWeight.w700,
                        color: isDark ? Colors.white : const Color(0xFF1A2525),
                      ),
                    ),
                    Text(
                      artist.genres,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: MeloopTypography.lexend(
                        fontSize: 10,
                        color: isDark
                            ? Colors.white.withValues(alpha: 0.65)
                            : const Color(0xFF6B807D),
                      ),
                    ),
                    Text(
                      artist.stats,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: MeloopTypography.lexend(
                        fontSize: 9.5,
                        color: isDark
                            ? Colors.white.withValues(alpha: 0.5)
                            : const Color(0xFF889A97),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFavoriteSongCard(ProfileCustomizationConfig config) {
    final accent = config.colorTheme.primary;
    final isDark = config.appStyle == MeloopAppStyle.neonNight;

    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      tintColor: config.colorTheme.songCardTint,
      backAngleDegrees: 1.4,
      padding: const EdgeInsets.fromLTRB(14, 12, 14, 14),
      child: Column(
        children: [
          Text(
            'Canción favorita',
            style: MeloopTypography.lexend(
              fontSize: 14,
              fontWeight: FontWeight.w700,
              color: isDark ? Colors.white : const Color(0xFF182828),
            ),
          ),
          const SizedBox(height: 10),
          InkWell(
            borderRadius: BorderRadius.circular(12),
            onTap: () => setState(() => _songPlaying = !_songPlaying),
            child: Container(
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                  colors: [config.colorTheme.headerTop, accent],
                ),
                borderRadius: BorderRadius.circular(12),
                boxShadow: [
                  BoxShadow(
                    color: accent.withValues(alpha: 0.25),
                    blurRadius: 8,
                    offset: const Offset(0, 3),
                  ),
                ],
              ),
              child: Row(
                children: [
                  Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: Colors.black.withValues(alpha: 0.22),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Icon(
                      _songPlaying
                          ? Icons.pause_rounded
                          : Icons.play_arrow_rounded,
                      color: Colors.white,
                      size: 24,
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          config.favoriteSongTitle,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: MeloopTypography.musicPlayer(
                            fontSize: 15,
                            color: Colors.white,
                            height: 1.1,
                          ),
                        ),
                        Text(
                          '${config.favoriteSongAlbum} · ${config.favoriteSongArtist}',
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: MeloopTypography.musicPlayer(
                            fontSize: 12.5,
                            color: Colors.white.withValues(alpha: 0.88),
                          ),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 6),
                  Icon(
                    _songPlaying
                        ? Icons.graphic_eq_rounded
                        : Icons.music_note_rounded,
                    color: Colors.white,
                    size: 18,
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFriendsCard(ProfileCustomizationConfig config) {
    final accent = config.colorTheme.primary;
    final isDark = config.appStyle == MeloopAppStyle.neonNight;

    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      tintColor: config.colorTheme.softCardTint,
      backAngleDegrees: -1.2,
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
      child: Column(
        children: [
          Text(
            'Amigos:',
            style: MeloopTypography.lexend(
              fontSize: 15,
              fontWeight: FontWeight.w700,
              color: isDark ? Colors.white : const Color(0xFF1A2626),
            ),
          ),
          const SizedBox(height: 10),
          for (final friend in _profileFriends)
            Padding(
              padding: const EdgeInsets.only(bottom: 8),
              child: Material(
                color: isDark
                    ? Colors.white.withValues(alpha: 0.06)
                    : Colors.white,
                child: InkWell(
                  onTap: _openFriendsSection,
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 10,
                      vertical: 8,
                    ),
                    decoration: BoxDecoration(
                      border: Border.all(
                        color: isDark
                            ? Colors.white.withValues(alpha: 0.12)
                            : const Color(0xFFE0ECE9),
                      ),
                    ),
                    child: Row(
                      children: [
                        const MeloopImagePlaceholder(
                          width: 34,
                          height: 34,
                          iconSize: 16,
                        ),
                        const SizedBox(width: 10),
                        Expanded(
                          child: Text(
                            friend.name,
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            // Regla tipográfica: usuarios en listas usan Lexend Bold
                            style: MeloopTypography.lexend(
                              fontSize: 13,
                              fontWeight: FontWeight.w700,
                              color: isDark
                                  ? Colors.white
                                  : const Color(0xFF1A2525),
                            ),
                          ),
                        ),
                        Container(
                          width: 9,
                          height: 9,
                          decoration: BoxDecoration(
                            color: friend.online
                                ? const Color(0xFF22C55E)
                                : const Color(0xFF9CA3AF),
                            shape: BoxShape.circle,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          const SizedBox(height: 4),
          FilledButton(
            style: FilledButton.styleFrom(
              backgroundColor: accent,
              foregroundColor: Colors.white,
              elevation: 0,
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 6),
              minimumSize: const Size(0, 32),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(18),
              ),
            ),
            onPressed: _openFriendsSection,
            child: Text(
              'Ver más',
              style: MeloopTypography.lexend(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: Colors.white,
              ),
            ),
          ),
        ],
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // SECCIÓN INFERIOR: TUS BLOGS (CONSERVA PERSISTENCIA REAL DEL BACKEND)
  // ---------------------------------------------------------------------------
  Widget _buildBlogsSection(ProfileCustomizationConfig config) {
    final accent = config.colorTheme.primary;
    final isDark = config.appStyle == MeloopAppStyle.neonNight;
    final displayedPosts = List<Map<String, dynamic>>.from(_posts);
    if (_blogsFilter == _ProfileBlogsFilter.trending) {
      displayedPosts.sort(
        (a, b) => ((b['likes'] as num?) ?? 0).compareTo(
          (a['likes'] as num?) ?? 0,
        ),
      );
    }

    return TapeCornerOverlay(
      leftWidth: 138,
      rightWidth: 122,
      child: StackedPaperCard(
        textureAsset: 'assets/textures/paper.png',
        backAngleDegrees: 0.95,
        padding: const EdgeInsets.fromLTRB(26, 22, 26, 24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Wrap(
              spacing: 12,
              runSpacing: 10,
              alignment: WrapAlignment.spaceBetween,
              crossAxisAlignment: WrapCrossAlignment.center,
              children: [
                Transform.rotate(
                  angle: -0.025,
                  alignment: Alignment.centerLeft,
                  child: Text(
                    _profileStatus == 'own'
                        ? 'Tus Blogs:'
                        : 'Blogs de $_displayUsername:',
                    style: MeloopTypography.momoDisplay(
                      fontSize: 24,
                      color: isDark ? Colors.white : const Color(0xFF182424),
                    ),
                  ),
                ),
                MeloopPillTabs<_ProfileBlogsFilter>(
                  items: const [
                    MeloopPillItem(
                      value: _ProfileBlogsFilter.recent,
                      label: 'Recientes',
                    ),
                    MeloopPillItem(
                      value: _ProfileBlogsFilter.trending,
                      label: 'Tendencia',
                    ),
                  ],
                  selectedValue: _blogsFilter,
                  onSelected: (next) => setState(() => _blogsFilter = next),
                  height: 38,
                  padding: 3.5,
                  fontSize: 12.5,
                  fixedButtonWidth: 104,
                  pillColor: config.colorTheme.pillTrack,
                ),
                if (_profileStatus == 'own')
                  FilledButton.icon(
                    style: FilledButton.styleFrom(
                      backgroundColor: accent,
                      foregroundColor: Colors.white,
                      elevation: 0,
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 10,
                      ),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(20),
                      ),
                    ),
                    onPressed: _openCreateBlogDialog,
                    icon: const Icon(Icons.edit_note_rounded, size: 18),
                    label: Text(
                      'Escribir blog',
                      style: MeloopTypography.lexend(
                        fontSize: 12.5,
                        fontWeight: FontWeight.w600,
                        color: Colors.white,
                      ),
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 18),
            const Divider(height: 1, color: Color(0xFFE0E8E6)),
            const SizedBox(height: 18),
            if (_isInitialLoading)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 42),
                child: Center(child: CircularProgressIndicator(color: accent)),
              )
            else if (_postsError != null)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 32),
                child: Column(
                  children: [
                    Text(
                      _postsError!,
                      textAlign: TextAlign.center,
                      style: MeloopTypography.lexend(
                        fontSize: 13,
                        color: isDark
                            ? Colors.white70
                            : const Color(0xFF6B7E7C),
                      ),
                    ),
                    const SizedBox(height: 12),
                    OutlinedButton(
                      onPressed: _fetchInitialPosts,
                      child: const Text('Reintentar'),
                    ),
                  ],
                ),
              )
            else if (displayedPosts.isEmpty)
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 36),
                child: Column(
                  children: [
                    Icon(
                      Icons.post_add_rounded,
                      size: 42,
                      color: isDark ? Colors.white38 : Colors.grey[400],
                    ),
                    const SizedBox(height: 10),
                    Text(
                      'Aún no hay publicaciones',
                      style: MeloopTypography.lexend(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: isDark
                            ? Colors.white70
                            : const Color(0xFF6B7E7C),
                      ),
                    ),
                  ],
                ),
              )
            else ...[
              for (var i = 0; i < displayedPosts.length; i++)
                Padding(
                  padding: const EdgeInsets.only(bottom: 16),
                  child: PostCard(
                    postData: displayedPosts[i],
                    tealAccent: accent,
                    textColor: isDark ? Colors.white : const Color(0xFF1A2525),
                    apiClient: _apiClient,
                    onDelete: () {
                      final target = displayedPosts[i];
                      setState(() => _posts.remove(target));
                    },
                  ),
                ),
              if (_isLoadingMore)
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  child: Center(child: CircularProgressIndicator(color: accent)),
                )
              else
                Padding(
                  padding: const EdgeInsets.only(top: 6),
                  child: Center(
                    child: Text(
                      'No has publicado más blogs',
                      style: MeloopTypography.lexend(
                        fontSize: 12.5,
                        color: isDark
                            ? Colors.white54
                            : const Color(0xFF7D908E),
                      ),
                    ),
                  ),
                ),
            ],
          ],
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // ESTADOS ESPECIALES (PRIVADO / BLOQUEADO)
  // ---------------------------------------------------------------------------
  Widget _buildBlockedState(ProfileCustomizationConfig config) {
    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      padding: const EdgeInsets.all(42),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.no_accounts_rounded, size: 58, color: Colors.grey[400]),
          const SizedBox(height: 16),
          Text(
            'Esta cuenta no está disponible',
            style: MeloopTypography.lexend(
              fontSize: 18,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'El perfil que intentas buscar no existe o ha restringido tu acceso.',
            textAlign: TextAlign.center,
            style: MeloopTypography.lexend(
              fontSize: 13,
              color: const Color(0xFF6B7E7C),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPrivateState(ProfileCustomizationConfig config) {
    return StackedPaperCard(
      textureAsset: 'assets/textures/paper.png',
      padding: const EdgeInsets.all(42),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.lock_outline_rounded, size: 58, color: Colors.grey[400]),
          const SizedBox(height: 16),
          Text(
            'Este perfil es privado',
            style: MeloopTypography.lexend(
              fontSize: 18,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Añade a este usuario como amigo para ver sus publicaciones, artistas favoritos y más detalles.',
            textAlign: TextAlign.center,
            style: MeloopTypography.lexend(
              fontSize: 13,
              color: const Color(0xFF6B7E7C),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTopBar(bool isDesktop, BuildContext context) {
    if (!isDesktop && defaultTargetPlatform == TargetPlatform.android) {
      return MobileAppHeader(
        onSearchTap: () => openGlobalSearch(context),
        apiClient: widget.apiClient,
      );
    }

    if (isDesktop) {
      return DesktopAppHeader(
        activeSection: DesktopNavigationSection.none,
        onHomeTap: () => navigateToDesktopHome(context),
        onFriendsTap: () =>
            navigateToDesktopSection(context, (_) => const FriendsScreen()),
        onArtistsTap: () => navigateToDesktopSection(
          context,
          (_) => const ArtistsAndSongsScreen(),
        ),
        onProfileTap: () {},
        onSearchTap: () => openGlobalSearch(context),
        apiClient: widget.apiClient,
      );
    }

    return MobileAppHeader(
      onSearchTap: () => openGlobalSearch(context),
      apiClient: widget.apiClient,
    );
  }
}
