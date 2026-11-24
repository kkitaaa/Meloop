import 'dart:async';
import 'dart:math' as math;

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

import '../core/navigation/artists_and_songs_screen.dart';
import '../core/navigation/desktop_navigation_shell.dart';
import '../features/gamification/presentation/gamification_demo_controller.dart';
import '../features/moderation/data/moderation_models.dart';
import '../features/moderation/presentation/moderation_queue_screen.dart';
import '../features/posts/presentation/post_editor_dialog.dart';
import '../features/profile/presentation/profile_screen.dart';
import '../features/search/presentation/global_search_route.dart';
import '../services/api_client.dart';
import '../services/auth_session.dart';
import '../theme/meloop_typography.dart';
import '../widgets/app_settings_menu.dart';
import '../widgets/desktop_app_header.dart';
import '../widgets/friends_sidebar.dart';
import '../widgets/meloop_pill_tabs.dart';
import '../widgets/mobile_app_header.dart';
import '../widgets/paper/stacked_paper_card.dart';
import '../widgets/post_card.dart';
import '../widgets/rumi_xp_notice.dart';
import '../widgets/states/empty_state.dart';
import '../widgets/states/error_state.dart';
import '../widgets/states/loading_state.dart';
import 'friends_screen.dart';

class FeedScreen extends StatefulWidget {
  final ApiClient? apiClient;
  final ModerationRole userRole;

  const FeedScreen({
    super.key,
    this.apiClient,
    this.userRole = ModerationRole.user,
  });

  @override
  State<FeedScreen> createState() => FeedScreenState();
}

class FeedScreenState extends State<FeedScreen> {
  static const int _pageSize = 5;
  static const _suggestedArtists = <({
    String name,
    String genres,
    String stats,
  })>[
    (
      name: 'Travis Scott',
      genres: 'Hip Hop · Pop Rap · Hip Hop experimental',
      stats: '100+ Canciones · 8 Álbumes',
    ),
    (
      name: 'JPEGMAFIA',
      genres: 'Hip Hop experimental · Hip Hop alternativo',
      stats: '100+ Canciones · 6 Álbumes',
    ),
    (
      name: 'Hideki Naganuma',
      genres: 'J-POP · Electrónica · Funk · Breakbeat',
      stats: '25+ Canciones · 5 Álbumes',
    ),
    (
      name: 'Nanahira',
      genres: 'J-POP · Electrónica · Denpa · Breakcore',
      stats: '100+ Canciones · 14 Álbumes',
    ),
  ];

  final Color _tealAccent = const Color(0xFF1ABC9C);

  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  final ScrollController _scrollController = ScrollController();
  final List<Map<String, dynamic>> _posts = [];

  bool _isInitialLoading = true;
  bool _isLoadingMore = false;
  bool _hasMore = true;
  String _selectedCategory = 'recent';
  int _requestGeneration = 0;
  int _currentPage = 1;
  bool _hasError = false;
  String? _feedError;
  int _errorCode = 0;

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ApiClient();
    _ownsApiClient = widget.apiClient == null;

    _fetchInitialPosts();
    _scrollController.addListener(_handleInfiniteScroll);
  }

  void _handleInfiniteScroll() {
    if (_scrollController.hasClients &&
        _scrollController.position.pixels >=
            _scrollController.position.maxScrollExtent - 280) {
      if (!_isInitialLoading && !_isLoadingMore && _hasMore && !_hasError) {
        _fetchMorePosts();
      }
    }
  }

  @override
  void dispose() {
    _scrollController.removeListener(_handleInfiniteScroll);
    _scrollController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  Future<void> _fetchInitialPosts({String? category}) async {
    if (category != null) _selectedCategory = category;
    final requestGeneration = ++_requestGeneration;
    setState(() {
      _isInitialLoading = true;
      _isLoadingMore = false;
      _hasError = false;
      _feedError = null;
      _hasMore = true;
    });

    try {
      final response = await _apiClient.getFeedPage(
        category: _selectedCategory,
        page: 1,
        pageSize: _pageSize,
      );
      final page = response.data;
      if (page == null) {
        throw const ApiException(message: 'El API no devolvió publicaciones.');
      }
      if (!mounted || requestGeneration != _requestGeneration) return;
      setState(() {
        _posts
          ..clear()
          ..addAll(page.items.map((post) => post.toPostCardData()));
        _isInitialLoading = false;
        _currentPage = page.page;
        _hasMore = page.hasMore;
      });
      if (_scrollController.hasClients) {
        _scrollController.jumpTo(0);
      }
    } on ApiException catch (error) {
      if (!mounted || requestGeneration != _requestGeneration) return;
      setState(() {
        _isInitialLoading = false;
        _hasError = true;
        _errorCode = error.statusCode ?? 0;
        _feedError = error.message;
      });
    }
  }

  Future<void> _fetchMorePosts() async {
    if (_isLoadingMore || !_hasMore || _isInitialLoading) return;
    final requestGeneration = _requestGeneration;
    setState(() => _isLoadingMore = true);

    try {
      final response = await _apiClient.getFeedPage(
        category: _selectedCategory,
        page: _currentPage + 1,
        pageSize: _pageSize,
      );
      final page = response.data;
      if (page == null) {
        throw const ApiException(message: 'El API no devolvió publicaciones.');
      }
      if (!mounted || requestGeneration != _requestGeneration) return;
      setState(() {
        _posts.addAll(page.items.map((post) => post.toPostCardData()));
        _currentPage = page.page;
        _hasMore = page.hasMore;
        _isLoadingMore = false;
      });
    } on ApiException catch (error) {
      if (!mounted || requestGeneration != _requestGeneration) return;
      setState(() => _isLoadingMore = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'No se pudieron cargar más publicaciones: ${error.message}',
          ),
        ),
      );
    }
  }

  Future<void> _createPost() async {
    final result = await showDialog<PostEditorResult>(
      context: context,
      builder: (context) =>
          PostEditorDialog(tealAccent: _tealAccent, apiClient: _apiClient),
    );
    if (!mounted || result == null) return;

    final post = result.publishedPost ?? const <String, dynamic>{};
    final serverImages = post['images'];
    final firstServerImage = serverImages is List && serverImages.isNotEmpty
        ? serverImages.first
        : null;
    final firstServerImageUrl = firstServerImage is String
        ? firstServerImage
        : firstServerImage is Map<String, dynamic>
        ? firstServerImage['url'] as String?
        : null;
    setState(() {
      _posts.insert(0, {
        "id": post['id'] ?? DateTime.now().microsecondsSinceEpoch,
        "user":
            post['user'] ??
            post['username'] ??
            AuthSession.shared.user?.username ??
            "Tú",
        "user_id": post['user_id'],
        "time": "Hace un momento",
        "content": post['content'] ?? result.content,
        "image_bytes": result.imageBytes,
        "image_name": result.imageName,
        "image_url": post['image_url'] ?? firstServerImageUrl,
        "images": serverImages ?? result.images,
        "song": post['song_title'] ?? result.songTitle,
        "artist": post['artist_name'] ?? result.artistName,
        "song_id": post['song_id'] ?? post['spotify_track_id'] ?? result.songId,
        "audio_url": post['audio_url'],
        "audio_bytes": result.audioBytes,
        "audio_name": result.audioName,
        "likes": 0,
        "comments": 0,
        "isAuthor": true,
        "isEdited": false,
        "within24Hours": true,
      });
    });

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: RumiXpNotice(xpAwarded: result.xpAwarded)),
    );
  }

  Future<void> createPost() => _createPost();

  void _openModerationQueue() {
    Navigator.push(
      context,
      MaterialPageRoute<void>(
        builder: (_) => ModerationQueueScreen(role: widget.userRole),
      ),
    );
  }

  void _openOwnProfile() {
    final shell = DesktopShellScope.maybeOf(context);
    if (shell != null) {
      shell.openProfile();
      return;
    }
    Navigator.push(
      context,
      MaterialPageRoute<void>(
        builder: (_) => ProfileScreen(apiClient: _apiClient),
      ),
    );
  }

  void _openArtistsSection() {
    final shell = DesktopShellScope.maybeOf(context);
    if (shell != null) {
      shell.openArtists();
      return;
    }
    navigateToDesktopSection(context, (_) => const ArtistsAndSongsScreen());
  }

  void _openFriendsSection() {
    final shell = DesktopShellScope.maybeOf(context);
    if (shell != null) {
      shell.openFriends();
      return;
    }
    navigateToDesktopSection(context, (_) => const FriendsScreen());
  }

  void _openUserProfile(String username, String? userId) {
    final shell = DesktopShellScope.maybeOf(context);
    if (shell != null) {
      shell.openProfile(username: username, userId: userId);
      return;
    }
    Navigator.push(
      context,
      MaterialPageRoute<void>(
        builder: (_) => ProfileScreen(
          username: username,
          userId: userId,
          apiClient: _apiClient,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.sizeOf(context).width;
    final bool isDesktop = screenWidth > 900;
    final bool insideDesktopShell = DesktopShellScope.maybeOf(context) != null;

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Column(
        children: [
          if (!insideDesktopShell) _buildTopBar(isDesktop),
          Expanded(
            child: isDesktop ? _buildDesktopLayout() : _buildMobileLayout(),
          ),
        ],
      ),
    );
  }

  Widget _buildTopBar(bool isDesktop) {
    final canModerate = widget.userRole != ModerationRole.user;
    if (isDesktop) {
      return DesktopAppHeader(
        activeSection: DesktopNavigationSection.home,
        onHomeTap: () {},
        onFriendsTap: () =>
            navigateToDesktopSection(context, (_) => const FriendsScreen()),
        onArtistsTap: _openArtistsSection,
        onProfileTap: _openOwnProfile,
        onModerationTap: canModerate ? _openModerationQueue : null,
        onSearchTap: () => openGlobalSearch(context),
        apiClient: _apiClient,
      );
    }

    if (defaultTargetPlatform == TargetPlatform.android) {
      return MobileAppHeader(
        onSearchTap: () => openGlobalSearch(context),
        onModerationTap: canModerate ? _openModerationQueue : null,
        apiClient: _apiClient,
      );
    }

    return Container(
      height: 55,
      color: _tealAccent,
      padding: const EdgeInsets.symmetric(horizontal: 24),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Image.asset(
            'assets/imagenes/meloop.png',
            height: 24,
            fit: BoxFit.contain,
          ),
          Row(
            children: [
              IconButton(
                tooltip: 'Búsqueda global',
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
                icon: const Icon(Icons.search, color: Colors.white, size: 20),
                onPressed: () => openGlobalSearch(context),
              ),
              const SizedBox(width: 16),
              AppSettingsMenu(apiClient: _apiClient),
              const SizedBox(width: 16),
              const Icon(Icons.notifications, color: Colors.white, size: 20),
              const SizedBox(width: 16),
              InkWell(
                onTap: _openOwnProfile,
                child: const CircleAvatar(
                  radius: 14,
                  backgroundColor: Colors.white,
                  child: Icon(Icons.person, color: Color(0xFF1ABC9C), size: 18),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildDesktopLayout() {
    return LayoutBuilder(
      builder: (context, constraints) {
        final bool showRightSidebar = constraints.maxWidth >= 1220;
        final double horizontalPadding =
            (constraints.maxWidth * 0.022).clamp(20.0, 42.0);
        final double contentWidth = math.min(
          constraints.maxWidth - horizontalPadding * 2,
          1840.0,
        );
        // Ancho simétrico en ambos paneles laterales para mantener el Feed
        // centrado debajo de la barra de navegación y con un ancho cómodo de
        // lectura en pantallas 1080p.
        final double sidebarWidth = showRightSidebar
            ? (contentWidth * 0.19).clamp(250.0, 345.0)
            : (contentWidth * 0.24).clamp(270.0, 360.0);
        final double columnGap = showRightSidebar
            ? (contentWidth * 0.022).clamp(22.0, 40.0)
            : (contentWidth * 0.03).clamp(26.0, 48.0);

        final noScrollbarBehavior = ScrollConfiguration.of(
          context,
        ).copyWith(scrollbars: false);

        return Center(
          child: SizedBox(
            width: contentWidth,
            child: Padding(
              padding: const EdgeInsets.only(top: 34, bottom: 20),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  SizedBox(
                    width: sidebarWidth,
                    child: ScrollConfiguration(
                      behavior: noScrollbarBehavior,
                      child: SingleChildScrollView(
                        clipBehavior: Clip.none,
                        padding: const EdgeInsets.only(top: 14, bottom: 16),
                        child: Column(
                          children: [
                            _buildPolaroidProfile(),
                            const SizedBox(height: 22),
                            _buildListeningTo(),
                            const SizedBox(height: 22),
                            _buildLevelCard(),
                            const SizedBox(height: 22),
                            _buildSuggestions(),
                          ],
                        ),
                      ),
                    ),
                  ),
                  SizedBox(width: columnGap),
                  Expanded(child: _buildFeedPaper()),
                  if (showRightSidebar) ...[
                    SizedBox(width: columnGap),
                    SizedBox(
                      width: sidebarWidth,
                      child: ScrollConfiguration(
                        behavior: noScrollbarBehavior,
                        child: SingleChildScrollView(
                          clipBehavior: Clip.none,
                          padding: const EdgeInsets.only(top: 2, bottom: 16),
                          child: DesktopFriendsSidebar(
                            onSeeMoreOnlineTap: _openFriendsSection,
                            onSeeMoreRecommendedTap: _openFriendsSection,
                            onProfileTap: _openUserProfile,
                          ),
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        );
      },
    );
  }

  Widget _buildMobileLayout() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        children: [
          const SizedBox(height: 12),
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 320),
            child: _buildPolaroidProfile(compactPhotoHeight: 160),
          ),
          const SizedBox(height: 16),
          _buildListeningTo(),
          const SizedBox(height: 20),
          SizedBox(height: 700, child: _buildFeedPaper(isMobile: true)),
        ],
      ),
    );
  }

  Widget _buildPolaroidProfile({double? compactPhotoHeight}) {
    final username =
        AuthSession.shared.user?.username ?? 'Nombre.de.usuario';
    final gamification = GamificationDemoController.instance;

    return AnimatedBuilder(
      animation: gamification,
      builder: (context, _) {
        return GestureDetector(
          onTap: _openOwnProfile,
          child: MouseRegion(
            cursor: SystemMouseCursors.click,
            child: Stack(
              alignment: Alignment.topCenter,
              clipBehavior: Clip.none,
              children: [
                StackedPaperCard(
                  textureAsset: 'assets/textures/photopaper.png',
                  padding: const EdgeInsets.fromLTRB(20, 26, 20, 22),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (compactPhotoHeight != null)
                        SizedBox(
                          height: compactPhotoHeight,
                          width: double.infinity,
                          child: const MeloopImagePlaceholder(
                            iconSize: 42,
                            backgroundColor: Color(0xFFE7ECEB),
                            iconColor: Color(0xFF7B8E8C),
                            borderColor: Color(0xFFD0DADA),
                          ),
                        )
                      else
                        const AspectRatio(
                          aspectRatio: 1.0,
                          child: MeloopImagePlaceholder(
                            iconSize: 48,
                            backgroundColor: Color(0xFFE7ECEB),
                            iconColor: Color(0xFF7B8E8C),
                            borderColor: Color(0xFFD0DADA),
                          ),
                        ),
                      const SizedBox(height: 18),
                      Transform.rotate(
                        angle: -2.0 * math.pi / 180.0,
                        child: Text(
                          username,
                          textAlign: TextAlign.center,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: MeloopTypography.username(
                            fontSize: 25,
                            fontWeight: FontWeight.w700,
                            color: const Color(0xFF1E2525),
                          ),
                        ),
                      ),
                      const SizedBox(height: 6),
                      Text(
                        'Me siento: Feliz',
                        textAlign: TextAlign.center,
                        style: MeloopTypography.lexend(
                          fontSize: 15,
                          color: const Color(0xFF6B7777),
                        ),
                      ),
                    ],
                  ),
                ),
                Positioned(
                  top: -22,
                  child: Transform.translate(
                    offset: const Offset(4, 0),
                    child: CustomizablePin(
                      color: gamification.activePinColor,
                      size: 54,
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildListeningTo() {
    return StackedPaperCard(
      tintColor: const Color(0xFFF6F1D5),
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Escuchando',
            textAlign: TextAlign.center,
            style: MeloopTypography.lexend(
              fontSize: 16,
              fontWeight: FontWeight.w700,
              color: const Color(0xFF1A2020),
            ),
          ),
          const SizedBox(height: 10),
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              gradient: const LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [Color(0xFF1D7E78), Color(0xFF135F5C)],
              ),
              borderRadius: BorderRadius.circular(10),
            ),
            child: Row(
              children: [
                const MeloopImagePlaceholder(
                  width: 48,
                  height: 48,
                  iconSize: 22,
                  borderRadius: BorderRadius.all(Radius.circular(6)),
                  backgroundColor: Color(0xFF215E5A),
                  iconColor: Color(0xFFA7D9D4),
                  borderColor: Color(0xFF2E7873),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        'Flashing Lights',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.musicPlayer(
                          color: Colors.white,
                          fontSize: 16,
                          height: 1.1,
                        ),
                      ),
                      const SizedBox(height: 3),
                      Row(
                        children: [
                          Flexible(
                            child: Text(
                              'Graduation',
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: MeloopTypography.musicPlayer(
                                color: Colors.white.withValues(alpha: 0.9),
                                fontSize: 13.5,
                              ),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Flexible(
                            child: Text(
                              'Kanye West',
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                              style: MeloopTypography.musicPlayer(
                                color: Colors.white.withValues(alpha: 0.68),
                                fontSize: 13,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: 6),
                Align(
                  alignment: Alignment.topRight,
                  child: Container(
                    width: 20,
                    height: 20,
                    decoration: BoxDecoration(
                      color: Colors.white.withValues(alpha: 0.2),
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(
                      Icons.graphic_eq_rounded,
                      color: Colors.white,
                      size: 13,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLevelCard() {
    final progressController = GamificationDemoController.instance;
    return AnimatedBuilder(
      animation: progressController,
      builder: (context, _) {
        final level = progressController.level;
        final xp = progressController.xp;
        final xpNeeded = progressController.experienceNeeded;
        final fraction = xpNeeded == 0
            ? 0.0
            : (xp / xpNeeded).clamp(0.0, 1.0).toDouble();

        return StackedPaperCard(
          tintColor: const Color(0xFFDDF5F7),
          padding: const EdgeInsets.fromLTRB(18, 14, 18, 14),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(
                    Icons.star_border_rounded,
                    color: Color(0xFF158A82),
                    size: 22,
                  ),
                  const SizedBox(width: 6),
                  Flexible(
                    child: Text(
                      'Nivel $level • ${_levelRankLabel(level)}',
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: MeloopTypography.lexend(
                        fontSize: 14.5,
                        fontWeight: FontWeight.w600,
                        color: const Color(0xFF1B2B2B),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              TweenAnimationBuilder<double>(
                tween: Tween<double>(end: fraction),
                duration: const Duration(milliseconds: 420),
                curve: Curves.easeOutCubic,
                builder: (context, value, _) {
                  return Container(
                    height: 20,
                    width: double.infinity,
                    decoration: BoxDecoration(
                      color: const Color(0xFF3E8A85),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Align(
                      alignment: Alignment.centerLeft,
                      child: FractionallySizedBox(
                        widthFactor: value.clamp(0.06, 1.0),
                        child: Container(
                          height: 20,
                          decoration: BoxDecoration(
                            color: const Color(0xFF4BE3C8),
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(
                              color: const Color(0xFF0E6C66),
                              width: 1.5,
                            ),
                          ),
                        ),
                      ),
                    ),
                  );
                },
              ),
              const SizedBox(height: 8),
              Text(
                '$xp / $xpNeeded EXP',
                style: MeloopTypography.lexend(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: const Color(0xFF5FAEA8),
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  static String _levelRankLabel(int level) {
    if (level <= 3) return 'Aun empezando';
    if (level <= 6) return 'En sintonía';
    if (level <= 9) return 'Melómano activo';
    return 'Leyenda musical';
  }

  Widget _buildSuggestions() {
    return StackedPaperCard(
      padding: const EdgeInsets.fromLTRB(18, 16, 18, 18),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Artistas que te podrían gustar',
            textAlign: TextAlign.center,
            style: MeloopTypography.lexend(
              fontSize: 15,
              fontWeight: FontWeight.w700,
              color: const Color(0xFF1A2020),
            ),
          ),
          const SizedBox(height: 14),
          for (final artist in _suggestedArtists)
            _artistSuggestionTile(
              name: artist.name,
              genres: artist.genres,
              stats: artist.stats,
            ),
          const SizedBox(height: 6),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF00B298),
              foregroundColor: Colors.white,
              elevation: 0,
              padding: const EdgeInsets.symmetric(horizontal: 32, vertical: 10),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(20),
              ),
            ),
            onPressed: _openArtistsSection,
            child: Text(
              'Ver más',
              style: MeloopTypography.lexend(
                fontSize: 13.5,
                fontWeight: FontWeight.w600,
                color: Colors.white,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _artistSuggestionTile({
    required String name,
    required String genres,
    required String stats,
  }) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.white.withValues(alpha: 0.72),
        border: Border.all(color: const Color(0xFFE6ECEB)),
      ),
      child: Row(
        children: [
          const MeloopImagePlaceholder(
            width: 46,
            height: 46,
            iconSize: 20,
            backgroundColor: Color(0xFFE7ECEB),
            iconColor: Color(0xFF7B8E8C),
            borderColor: Color(0xFFD0DADA),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 13.5,
                    fontWeight: FontWeight.w600,
                    color: const Color(0xFF1C2323),
                  ),
                ),
                const SizedBox(height: 2),
                Text(
                  genres,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 10.5,
                    color: const Color(0xFF7C8888),
                  ),
                ),
                Text(
                  stats,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.lexend(
                    fontSize: 10,
                    color: const Color(0xFF9AA5A5),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFeedPaper({bool isMobile = false}) {
    final paper = StackedPaperCard(
      padding: EdgeInsets.fromLTRB(
        isMobile ? 16 : 44,
        isMobile ? 24 : 34,
        isMobile ? 16 : 44,
        20,
      ),
      child: Column(
        children: [
          _buildFeedHeader(isMobile: isMobile),
          SizedBox(height: isMobile ? 16 : 24),

          // Estados de carga, error, vacío y feed con scroll infinito
          Expanded(
            child: _isInitialLoading
                ? const LoadingState()
                : _hasError
                ? ErrorState(
                    errorCode: _errorCode,
                    onRetry: _fetchInitialPosts,
                    tealAccent: _tealAccent,
                    message: _feedError,
                  )
                : _posts.isEmpty
                ? const EmptyState(
                    title: "Feed vacío",
                    message:
                        "Aún no hay publicaciones recientes. ¡Sé el primero en crear un blog!",
                    icon: Icons.post_add,
                  )
                : ListView.builder(
                    controller: _scrollController,
                    padding: EdgeInsets.only(
                      right: isMobile ? 0 : 8,
                      bottom: 24,
                    ),
                    itemCount: _posts.length + (_isLoadingMore ? 1 : 0),
                    itemBuilder: (context, index) {
                      if (index == _posts.length) {
                        return Padding(
                          padding: const EdgeInsets.symmetric(vertical: 24.0),
                          child: Center(
                            child: CircularProgressIndicator(
                              color: _tealAccent,
                            ),
                          ),
                        );
                      }
                      return PostCard(
                        postData: _posts[index],
                        tealAccent: _tealAccent,
                        apiClient: _apiClient,
                        onDelete: () {
                          setState(() {
                            _posts.removeAt(index);
                          });
                        },
                      );
                    },
                  ),
          ),
        ],
      ),
    );

    if (isMobile) {
      return TapeCornerOverlay(leftWidth: 112, rightWidth: 96, child: paper);
    }
    return TapeCornerOverlay(leftWidth: 175, rightWidth: 152, child: paper);
  }

  Widget _buildFeedHeader({required bool isMobile}) {
    final titleWidget = Transform.rotate(
      // En Figma +2.66° inclina hacia arriba a la derecha (CCW = negativo en Flutter)
      angle: -2.66 * math.pi / 180.0,
      child: Text(
        'Feed',
        style: MeloopTypography.momoDisplay(
          fontSize: isMobile ? 24 : 34,
          color: const Color(0xFF141919),
        ),
      ),
    );

    final filterBar = MeloopPillTabs<String>(
      items: const [
        MeloopPillItem(value: 'recent', label: 'Recientes'),
        MeloopPillItem(value: 'suggested', label: 'Tendencia'),
        MeloopPillItem(value: 'friends', label: 'Amigos'),
      ],
      selectedValue: _selectedCategory,
      disabled: _isInitialLoading,
      onSelected: (category) {
        if (category == _selectedCategory) return;
        _fetchInitialPosts(category: category);
      },
      height: isMobile ? 38 : 44,
      padding: 4,
      fontSize: isMobile ? 12.5 : 14,
      horizontalItemPadding: isMobile ? 14 : 20,
      fixedButtonWidth: isMobile ? 96 : 120,
      pillColor: const Color(0xFF288B83),
      equalWidths: true,
    );

    final writeButton = DecoratedBox(
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [Color(0xFF42D6BE), Color(0xFF0FB39A)],
        ),
        borderRadius: BorderRadius.circular(24),
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: _createPost,
          borderRadius: BorderRadius.circular(24),
          child: Padding(
            padding: EdgeInsets.symmetric(
              horizontal: isMobile ? 16 : 24,
              vertical: isMobile ? 9 : 11,
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  'Escribir blog',
                  style: MeloopTypography.lexend(
                    color: Colors.white,
                    fontSize: isMobile ? 13 : 15,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(width: 10),
                const Icon(
                  Icons.edit_note_rounded,
                  color: Colors.white,
                  size: 20,
                ),
              ],
            ),
          ),
        ),
      ),
    );

    if (isMobile) {
      return Wrap(
        spacing: 12,
        runSpacing: 12,
        alignment: WrapAlignment.center,
        crossAxisAlignment: WrapCrossAlignment.center,
        children: [titleWidget, filterBar, writeButton],
      );
    }

    return LayoutBuilder(
      builder: (context, constraints) {
        if (constraints.maxWidth < 700) {
          return Wrap(
            spacing: 16,
            runSpacing: 14,
            alignment: WrapAlignment.spaceBetween,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: [titleWidget, filterBar, writeButton],
          );
        }

        return SizedBox(
          height: 52,
          width: double.infinity,
          child: Stack(
            alignment: Alignment.center,
            children: [
              Align(alignment: Alignment.centerLeft, child: titleWidget),
              Center(child: filterBar),
              Align(alignment: Alignment.centerRight, child: writeButton),
            ],
          ),
        );
      },
    );
  }
}
