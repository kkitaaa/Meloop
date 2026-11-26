import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/core/navigation/artists_and_songs_screen.dart';
import 'package:flutter_app/core/navigation/desktop_navigation_shell.dart';
import 'package:flutter_app/core/navigation/responsive_navigation_shell.dart';
import 'package:flutter_app/features/gamification/presentation/gamification_demo_controller.dart';
import 'package:flutter_app/features/profile/presentation/profile_screen.dart';
import 'package:flutter_app/models/global_search.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/screens/friends_screen.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/theme/meloop_typography.dart';
import 'package:flutter_app/widgets/desktop_app_header.dart';
import 'package:flutter_app/widgets/meloop_pill_tabs.dart';
import 'package:flutter_app/widgets/meloop_toast_pill.dart';
import 'package:flutter_app/widgets/paper/stacked_paper_card.dart';
import 'package:flutter_app/widgets/post_card.dart';

enum _SearchSource { users, posts, music }

enum _DesktopLeftTab { all, posts, music }

enum _CompactSearchTab { posts, music, users }

class GlobalSearchScreen extends StatefulWidget {
  final ApiClient? apiClient;
  final TextEditingController? externalQueryController;
  final Listenable? externalSubmitTrigger;
  final MobileShellController? mobileShellController;

  const GlobalSearchScreen({
    super.key,
    this.apiClient,
    this.externalQueryController,
    this.externalSubmitTrigger,
    this.mobileShellController,
  });

  @override
  State<GlobalSearchScreen> createState() => _GlobalSearchScreenState();
}

class _GlobalSearchScreenState extends State<GlobalSearchScreen> {
  static const Color _tealAccent = Color(0xFF00AF92);
  static const Color _pillTrackColor = Color(0xFF288B83);
  static const Color _inkDark = Color(0xFF182828);
  static const Color _inkMuted = Color(0xFF6B807D);

  static const _filters = <String, String>{
    'Texto': 'text',
    'Canción': 'song',
    'Artista': 'artist',
    'Género': 'genre',
  };

  late final TextEditingController _queryController;
  late final bool _ownsQueryController;
  final List<String> _recentSearches = [];
  final Map<_SearchSource, CancelToken> _cancelTokens = {};
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;

  _DesktopLeftTab _desktopLeftTab = _DesktopLeftTab.all;
  _CompactSearchTab _compactTab = _CompactSearchTab.posts;

  List<SearchUser> _users = const [];
  List<SearchPost> _posts = const [];
  List<MusicTrack> _tracks = const [];
  final Map<_SearchSource, String?> _errors = {
    for (final source in _SearchSource.values) source: null,
  };
  final Map<_SearchSource, bool> _loading = {
    for (final source in _SearchSource.values) source: false,
  };
  final Map<_SearchSource, bool> _hasResults = {
    for (final source in _SearchSource.values) source: false,
  };
  final Set<String> _sendingRequests = {};
  String _postFilter = 'text';
  String _activeQuery = '';
  bool _showMinLengthHint = false;

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ApiClient();
    _ownsApiClient = widget.apiClient == null;
    _ownsQueryController = widget.externalQueryController == null;
    _queryController =
        widget.externalQueryController ?? TextEditingController();
    _queryController.addListener(_onQueryChanged);
    widget.externalSubmitTrigger?.addListener(_onExternalSubmit);

    final initialQuery = _queryController.text.trim();
    if (initialQuery.length >= 2) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _queryController.text.trim() == initialQuery) {
          _searchAll(initialQuery);
        }
      });
    }
  }

  @override
  void didUpdateWidget(covariant GlobalSearchScreen oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.externalSubmitTrigger != widget.externalSubmitTrigger) {
      oldWidget.externalSubmitTrigger?.removeListener(_onExternalSubmit);
      widget.externalSubmitTrigger?.addListener(_onExternalSubmit);
    }
  }

  @override
  void dispose() {
    for (final token in _cancelTokens.values) {
      token.cancel('La búsqueda cambió.');
    }
    widget.externalSubmitTrigger?.removeListener(_onExternalSubmit);
    _queryController.removeListener(_onQueryChanged);
    if (_ownsQueryController) _queryController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  void _onExternalSubmit() {
    _submitSearch();
  }

  void _onQueryChanged() {
    final query = _queryController.text.trim();
    if (query.isEmpty) {
      _cancelPendingRequests();
      setState(() {
        _activeQuery = '';
        _showMinLengthHint = false;
        _clearResults();
      });
      return;
    }
    if (_showMinLengthHint && query.length >= 2) {
      setState(() => _showMinLengthHint = false);
      return;
    }
    setState(() {});
  }

  void _submitSearch([String? rawQuery]) {
    final query = (rawQuery ?? _queryController.text).trim();
    if (query.length < 2) {
      _cancelPendingRequests();
      setState(() {
        _activeQuery = '';
        _showMinLengthHint = query.isNotEmpty;
        _clearResults();
      });
      return;
    }
    _cancelPendingRequests();
    setState(() => _showMinLengthHint = false);
    _searchAll(query);
  }

  void _cancelPendingRequests() {
    for (final token in _cancelTokens.values) {
      token.cancel('La búsqueda cambió.');
    }
    _cancelTokens.clear();
  }

  void _clearResults() {
    _users = const [];
    _posts = const [];
    _tracks = const [];
    for (final source in _SearchSource.values) {
      _errors[source] = null;
      _loading[source] = false;
      _hasResults[source] = false;
    }
  }

  void _searchAll(String query) {
    _activeQuery = query;
    _rememberSearch(query);
    setState(() {
      _users = const [];
      _posts = const [];
      _tracks = const [];
      for (final source in _SearchSource.values) {
        _errors[source] = null;
        _loading[source] = true;
        _hasResults[source] = true;
        _cancelTokens[source] = CancelToken();
      }
    });
    _searchUsers(query, _cancelTokens[_SearchSource.users]!);
    _searchPosts(query, _cancelTokens[_SearchSource.posts]!);
    _searchMusic(query, _cancelTokens[_SearchSource.music]!);
  }

  void _rememberSearch(String query) {
    _recentSearches.remove(query);
    _recentSearches.insert(0, query);
    if (_recentSearches.length > 5) _recentSearches.removeLast();
  }

  void _applyRecentQuery(String query) {
    _queryController.text = query;
    _queryController.selection = TextSelection.collapsed(offset: query.length);
    _submitSearch(query);
  }

  Future<void> _searchUsers(String query, CancelToken token) async {
    try {
      final response = await _apiClient.searchUsers(query, cancelToken: token);
      if (!mounted || token.isCancelled) return;
      setState(() {
        _users = response.data?.items ?? const [];
        _loading[_SearchSource.users] = false;
        _errors[_SearchSource.users] = response.data == null
            ? 'El servicio no devolvió resultados.'
            : null;
      });
    } on DioException catch (error) {
      if (CancelToken.isCancel(error) || !mounted) return;
      _setSourceError(_SearchSource.users, error.message);
    } on ApiException catch (error) {
      if (!mounted || token.isCancelled) return;
      _setSourceError(_SearchSource.users, error.message);
    }
  }

  Future<void> _searchPosts(String query, CancelToken token) async {
    try {
      final response = await _apiClient.searchPosts(
        query,
        filter: _postFilter,
        cancelToken: token,
      );
      if (!mounted || token.isCancelled) return;
      setState(() {
        _posts = response.data?.items ?? const [];
        _loading[_SearchSource.posts] = false;
        _errors[_SearchSource.posts] = response.data == null
            ? 'El servicio no devolvió resultados.'
            : null;
      });
    } on DioException catch (error) {
      if (CancelToken.isCancel(error) || !mounted) return;
      _setSourceError(_SearchSource.posts, error.message);
    } on ApiException catch (error) {
      if (!mounted || token.isCancelled) return;
      _setSourceError(_SearchSource.posts, error.message);
    }
  }

  Future<void> _searchMusic(String query, CancelToken token) async {
    try {
      final response = await _apiClient.searchMusicTracks(
        query,
        cancelToken: token,
      );
      if (!mounted || token.isCancelled) return;
      setState(() {
        _tracks = response.data?.items ?? const [];
        _loading[_SearchSource.music] = false;
        _errors[_SearchSource.music] = response.data == null
            ? 'El servicio de música no devolvió resultados.'
            : null;
      });
    } on DioException catch (error) {
      if (CancelToken.isCancel(error) || !mounted) return;
      _setSourceError(
        _SearchSource.music,
        error.message ?? 'Spotify no está disponible. Inténtalo de nuevo.',
      );
    } on ApiException catch (error) {
      if (!mounted || token.isCancelled) return;
      _setSourceError(_SearchSource.music, error.message);
    }
  }

  void _setSourceError(_SearchSource source, String? message) {
    setState(() {
      _loading[source] = false;
      _errors[source] = message ?? 'No se pudo completar la búsqueda.';
    });
  }

  Future<void> _retry(_SearchSource source) async {
    final query = _activeQuery;
    if (query.length < 2) return;
    _cancelTokens[source]?.cancel('Se reintentó la búsqueda.');
    final token = CancelToken();
    setState(() {
      _cancelTokens[source] = token;
      _errors[source] = null;
      _loading[source] = true;
    });
    switch (source) {
      case _SearchSource.users:
        await _searchUsers(query, token);
      case _SearchSource.posts:
        await _searchPosts(query, token);
      case _SearchSource.music:
        await _searchMusic(query, token);
    }
  }

  Future<void> _changePostFilter(String filter) async {
    if (filter == _postFilter) return;
    setState(() => _postFilter = filter);
    if (_activeQuery.length < 2) return;
    _cancelTokens[_SearchSource.posts]?.cancel('Cambió el filtro.');
    final token = CancelToken();
    setState(() {
      _cancelTokens[_SearchSource.posts] = token;
      _errors[_SearchSource.posts] = null;
      _loading[_SearchSource.posts] = true;
    });
    await _searchPosts(_activeQuery, token);
  }

  Future<void> _sendFriendRequest(SearchUser user) async {
    if (_sendingRequests.contains(user.id)) return;
    setState(() => _sendingRequests.add(user.id));
    try {
      await _apiClient.sendFriendRequest(user.id);
      if (!mounted) return;
      setState(() {
        _users = _users
            .map(
              (result) => result.id == user.id
                  ? SearchUser(
                      id: result.id,
                      name: result.name,
                      avatarUrl: result.avatarUrl,
                      level: result.level,
                      friendStatus: 'pending',
                    )
                  : result,
            )
            .toList();
        _sendingRequests.remove(user.id);
      });
      showMeloopToastPill(
        context,
        message: 'Solicitud enviada a ${user.name}',
        icon: Icons.person_add_alt_1_rounded,
      );
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _sendingRequests.remove(user.id));
      showMeloopToastPill(
        context,
        message: 'No se pudo enviar la solicitud: ${error.message}',
        icon: Icons.error_outline_rounded,
      );
    }
  }

  /// Abre la pestaña de "Artistas y canciones" con el término buscado para ver
  /// el detalle completo del álbum, canciones y acciones.
  void _openInArtistsTab(String? query) {
    final normalized = query?.trim();
    final desktopShell = DesktopShellScope.maybeOf(context);
    if (desktopShell != null) {
      desktopShell.openArtists(initialQuery: normalized);
      return;
    }
    final mobileShell =
        widget.mobileShellController ?? MobileShellScope.maybeOf(context);
    if (mobileShell != null) {
      Navigator.of(context).maybePop();
      mobileShell.openArtists(initialQuery: normalized);
      return;
    }
    Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (_) => ArtistsAndSongsScreen(initialQuery: normalized),
      ),
    );
  }

  /// Abre la pestaña de "Amigos" con el nombre buscado para gestionar amistad
  /// o ver el detalle de ese usuario.
  void _openInFriendsTab(String? query) {
    final normalized = query?.trim();
    final desktopShell = DesktopShellScope.maybeOf(context);
    if (desktopShell != null) {
      desktopShell.openFriends(initialQuery: normalized);
      return;
    }
    final mobileShell =
        widget.mobileShellController ?? MobileShellScope.maybeOf(context);
    if (mobileShell != null) {
      Navigator.of(context).maybePop();
      mobileShell.openFriends(initialQuery: normalized);
      return;
    }
    Navigator.of(context).push<void>(
      MaterialPageRoute<void>(
        builder: (_) => FriendsScreen(initialQuery: normalized),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final isDesktop = MediaQuery.sizeOf(context).width > 900;
    final insideDesktopShell = DesktopShellScope.maybeOf(context) != null;

    return Scaffold(
      backgroundColor: Colors.transparent,
      appBar: insideDesktopShell ? null : _buildBrandedAppBar(isDesktop),
      body: SafeArea(
        top: false,
        child: isDesktop
            ? _buildDesktopLayout(insideDesktopShell: insideDesktopShell)
            : _buildCompactLayout(),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // ESCRITORIO WIDESCREEN: 2 hojas consistentes con Inicio / Amigos / Artistas
  // ---------------------------------------------------------------------------
  Widget _buildDesktopLayout({required bool insideDesktopShell}) {
    final gamification = GamificationDemoController.instance;
    final hasSubmittedQuery = _activeQuery.isNotEmpty;

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 1460),
        child: Padding(
          padding: const EdgeInsets.fromLTRB(28, 22, 28, 24),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Expanded(
                flex: 64,
                child: TapeCornerOverlay(
                  leftWidth: 148,
                  rightWidth: 132,
                  child: StackedPaperCard(
                    textureAsset: 'assets/textures/paper.png',
                    backAngleDegrees: -1.15,
                    padding: const EdgeInsets.fromLTRB(26, 22, 26, 18),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        _buildDesktopLeftHeader(
                          hasSubmittedQuery: hasSubmittedQuery,
                        ),
                        if (hasSubmittedQuery &&
                            (_desktopLeftTab == _DesktopLeftTab.all ||
                                _desktopLeftTab == _DesktopLeftTab.posts)) ...[
                          const SizedBox(height: 12),
                          _buildPostFilterChips(),
                        ],
                        const SizedBox(height: 14),
                        const Divider(height: 1, color: Color(0xFFE0E8E6)),
                        const SizedBox(height: 10),
                        Expanded(
                          child: hasSubmittedQuery
                              ? _buildDesktopLeftContent()
                              : _buildRecentSearches(isDesktop: true),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 28),
              Expanded(
                flex: 36,
                child: AnimatedBuilder(
                  animation: gamification,
                  builder: (context, _) => Stack(
                    clipBehavior: Clip.none,
                    children: [
                      Positioned.fill(
                        child: StackedPaperCard(
                          textureAsset: 'assets/textures/paper.png',
                          tintColor: const Color(0xFFF2FAF8),
                          backAngleDegrees: 1.25,
                          padding: const EdgeInsets.fromLTRB(22, 26, 22, 18),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.stretch,
                            children: [
                              _buildDesktopFriendsHeader(
                                hasSubmittedQuery: hasSubmittedQuery,
                              ),
                              const SizedBox(height: 12),
                              const Divider(
                                height: 1,
                                color: Color(0xFFD5E6E3),
                              ),
                              const SizedBox(height: 10),
                              Expanded(
                                child: hasSubmittedQuery
                                    ? _buildUsersList(isDesktop: true)
                                    : _buildFriendsIdleCard(),
                              ),
                            ],
                          ),
                        ),
                      ),
                      Positioned(
                        top: -18,
                        left: 0,
                        right: 0,
                        child: Center(
                          child: CustomizablePin(
                            color: gamification.activePinColor,
                            size: 46,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildDesktopLeftHeader({required bool hasSubmittedQuery}) {
    final postsCount = _posts.length;
    final musicCount = _tracks.length;
    final typed = _queryController.text.trim();

    return Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Transform.rotate(
                angle: -0.025,
                alignment: Alignment.centerLeft,
                child: Text(
                  'Publicaciones y música',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: MeloopTypography.momoDisplay(
                    fontSize: 27,
                    color: _inkDark,
                    letterSpacing: -0.3,
                  ),
                ),
              ),
              const SizedBox(height: 2),
              Text(
                hasSubmittedQuery
                    ? 'Resultados para «$_activeQuery» · $postsCount publicaciones · $musicCount canciones'
                    : _showMinLengthHint
                    ? 'Escribe al menos 2 caracteres y presiona Enter para buscar'
                    : typed.length >= 2
                    ? 'Presiona Enter para buscar «$typed»'
                    : 'Búsqueda rápida · Escribe y presiona Enter para buscar',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: MeloopTypography.lexend(
                  fontSize: 12.5,
                  color: _showMinLengthHint
                      ? const Color(0xFFB4690E)
                      : _inkMuted,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(width: 14),
        MeloopPillTabs<_DesktopLeftTab>(
          items: const [
            MeloopPillItem(value: _DesktopLeftTab.all, label: 'Todo'),
            MeloopPillItem(
              value: _DesktopLeftTab.posts,
              label: 'Publicaciones',
            ),
            MeloopPillItem(value: _DesktopLeftTab.music, label: 'Música'),
          ],
          selectedValue: _desktopLeftTab,
          onSelected: (tab) => setState(() => _desktopLeftTab = tab),
          height: 40,
          padding: 3.5,
          fontSize: 12.5,
          horizontalItemPadding: 14,
          fixedButtonWidth: 112,
          pillColor: _pillTrackColor,
        ),
      ],
    );
  }

  Widget _buildDesktopFriendsHeader({required bool hasSubmittedQuery}) {
    return Row(
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Amigos y personas',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: MeloopTypography.lexend(
                  fontSize: 19,
                  fontWeight: FontWeight.w700,
                  color: _inkDark,
                ),
              ),
              const SizedBox(height: 2),
              Text(
                hasSubmittedQuery
                    ? '${_users.length} perfiles encontrados'
                    : 'Toca un perfil para verlo en la pestaña Amigos',
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
                style: MeloopTypography.lexend(fontSize: 12, color: _inkMuted),
              ),
            ],
          ),
        ),
        TextButton(
          onPressed: () => _openInFriendsTab(
            _activeQuery.isNotEmpty ? _activeQuery : null,
          ),
          style: TextButton.styleFrom(
            foregroundColor: const Color(0xFF16877F),
            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                'Ir a Amigos',
                style: MeloopTypography.lexend(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(width: 4),
              const Icon(Icons.arrow_forward_rounded, size: 15),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildDesktopLeftContent() {
    switch (_desktopLeftTab) {
      case _DesktopLeftTab.posts:
        return _buildPostsList();
      case _DesktopLeftTab.music:
        return _buildMusicList();
      case _DesktopLeftTab.all:
        return ListView(
          padding: const EdgeInsets.only(top: 4, bottom: 20),
          children: [
            _buildCombinedSectionHeader(
              icon: Icons.album_rounded,
              title: 'Canciones y artistas',
              trailingLabel: 'Ver todo en Artistas',
              onTrailingTap: () => _openInArtistsTab(_activeQuery),
            ),
            const SizedBox(height: 8),
            _buildInlineMusicSection(),
            const SizedBox(height: 20),
            _buildCombinedSectionHeader(
              icon: Icons.article_outlined,
              title: 'Publicaciones',
              subtitle: 'Puedes dar me gusta, comentar y leer aquí mismo',
            ),
            const SizedBox(height: 10),
            _buildInlinePostsSection(),
          ],
        );
    }
  }

  Widget _buildCombinedSectionHeader({
    required IconData icon,
    required String title,
    String? subtitle,
    String? trailingLabel,
    VoidCallback? onTrailingTap,
  }) {
    return Row(
      children: [
        Icon(icon, size: 18, color: const Color(0xFF16877F)),
        const SizedBox(width: 8),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                title,
                style: MeloopTypography.lexend(
                  fontSize: 15,
                  fontWeight: FontWeight.w700,
                  color: _inkDark,
                ),
              ),
              if (subtitle != null)
                Text(
                  subtitle,
                  style: MeloopTypography.lexend(
                    fontSize: 11.5,
                    color: _inkMuted,
                  ),
                ),
            ],
          ),
        ),
        if (trailingLabel != null && onTrailingTap != null)
          TextButton(
            onPressed: onTrailingTap,
            style: TextButton.styleFrom(
              foregroundColor: const Color(0xFF16877F),
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  trailingLabel,
                  style: MeloopTypography.lexend(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(width: 4),
                const Icon(Icons.arrow_forward_rounded, size: 14),
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildInlineMusicSection() {
    final error = _errors[_SearchSource.music];
    if (_loading[_SearchSource.music] == true) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 20),
        child: Center(
          child: SizedBox(
            width: 24,
            height: 24,
            child: CircularProgressIndicator(
              strokeWidth: 2.4,
              color: _tealAccent,
            ),
          ),
        ),
      );
    }
    if (error != null) {
      return _buildCompactErrorBanner(
        message: error,
        onRetry: () => _retry(_SearchSource.music),
      );
    }
    if (_tracks.isEmpty) {
      return _buildCompactEmptyBanner(
        'No encontramos canciones para «$_activeQuery».',
        actionLabel: 'Buscar en Artistas y canciones',
        onAction: () => _openInArtistsTab(_activeQuery),
      );
    }
    final preview = _tracks.take(4).toList();
    return Column(
      children: [
        for (final track in preview)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: _buildMusicQuickCard(track),
          ),
      ],
    );
  }

  Widget _buildInlinePostsSection() {
    final error = _errors[_SearchSource.posts];
    if (_loading[_SearchSource.posts] == true) {
      return const Padding(
        padding: EdgeInsets.symmetric(vertical: 28),
        child: Center(
          child: CircularProgressIndicator(color: _tealAccent),
        ),
      );
    }
    if (error != null) {
      return _buildCompactErrorBanner(
        message: error,
        onRetry: () => _retry(_SearchSource.posts),
      );
    }
    if (_posts.isEmpty) {
      return _buildCompactEmptyBanner(
        'Sin publicaciones para «$_activeQuery». Prueba cambiar el filtro.',
      );
    }
    return Column(
      children: [
        for (final post in _posts)
          Padding(
            padding: const EdgeInsets.only(bottom: 14),
            child: _buildInteractivePostCard(post),
          ),
      ],
    );
  }

  // ---------------------------------------------------------------------------
  // COMPACTO / ANDROID: hoja estilo papel sin alterar contratos de test
  // ---------------------------------------------------------------------------
  Widget _buildCompactLayout() {
    final hasSubmittedQuery = _activeQuery.isNotEmpty;

    return Center(
      child: Padding(
        padding: const EdgeInsets.fromLTRB(14, 14, 14, 14),
        child: ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 880),
          child: DecoratedBox(
            decoration: BoxDecoration(
              color: const Color(0xFFFAFBF9),
              border: Border.all(color: const Color(0xFFDCE7E5)),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.14),
                  blurRadius: 14,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: Column(
              children: [
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 16, 16, 10),
                  child: _buildSearchInput(),
                ),
                if (_showMinLengthHint)
                  Padding(
                    padding: const EdgeInsets.fromLTRB(20, 0, 20, 8),
                    child: Align(
                      alignment: Alignment.centerLeft,
                      child: Text(
                        'Escribe al menos 2 caracteres para buscar.',
                        style: MeloopTypography.lexend(
                          fontSize: 12,
                          color: const Color(0xFFB4690E),
                        ),
                      ),
                    ),
                  ),
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16),
                  child: Center(
                    child: SingleChildScrollView(
                      scrollDirection: Axis.horizontal,
                      child: MeloopPillTabs<_CompactSearchTab>(
                        items: const [
                          MeloopPillItem(
                            value: _CompactSearchTab.posts,
                            label: 'Publicaciones',
                          ),
                          MeloopPillItem(
                            value: _CompactSearchTab.music,
                            label: 'Música',
                          ),
                          MeloopPillItem(
                            value: _CompactSearchTab.users,
                            label: 'Usuarios',
                          ),
                        ],
                        selectedValue: _compactTab,
                        onSelected: (tab) => setState(() => _compactTab = tab),
                        height: 40,
                        padding: 3.5,
                        fontSize: 12.5,
                        horizontalItemPadding: 14,
                        fixedButtonWidth: 112,
                        pillColor: _pillTrackColor,
                      ),
                    ),
                  ),
                ),
                if (hasSubmittedQuery && _compactTab == _CompactSearchTab.posts)
                  Padding(
                    padding: const EdgeInsets.fromLTRB(16, 10, 16, 0),
                    child: _buildPostFilterChips(),
                  ),
                const SizedBox(height: 10),
                const Divider(height: 1, color: Color(0xFFE3ECEA)),
                Expanded(
                  child: hasSubmittedQuery
                      ? switch (_compactTab) {
                          _CompactSearchTab.posts => _buildPostsList(),
                          _CompactSearchTab.music => _buildMusicList(),
                          _CompactSearchTab.users => _buildUsersList(
                            isDesktop: false,
                          ),
                        }
                      : _buildRecentSearches(isDesktop: false),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // ---------------------------------------------------------------------------
  // LISTAS DE RESULTADOS
  // ---------------------------------------------------------------------------
  Widget _buildPostsList() {
    final error = _errors[_SearchSource.posts];
    if (_loading[_SearchSource.posts] == true) {
      return const Center(child: CircularProgressIndicator(color: _tealAccent));
    }
    if (error != null) {
      return Center(
        child: _StateCard(
          icon: Icons.cloud_off_rounded,
          title: 'No pudimos conectar',
          message: error,
          action: OutlinedButton.icon(
            onPressed: () => _retry(_SearchSource.posts),
            icon: const Icon(Icons.refresh),
            label: const Text('Reintentar'),
          ),
        ),
      );
    }
    if (_hasResults[_SearchSource.posts] != true) {
      return const _StateCard(
        icon: Icons.travel_explore_rounded,
        title: 'Todo empieza con una búsqueda',
        message: 'Encuentra personas, publicaciones y música en Meloop.',
      );
    }
    if (_posts.isEmpty) return _emptyResults();

    return ListView.builder(
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 24),
      itemCount: _posts.length,
      itemBuilder: (context, index) => Padding(
        padding: const EdgeInsets.only(bottom: 14),
        child: _buildInteractivePostCard(_posts[index]),
      ),
    );
  }

  Widget _buildInteractivePostCard(SearchPost post) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: const Color(0xFFE1EAE8)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.05),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      padding: const EdgeInsets.all(14),
      child: PostCard(
        postData: post.cardData,
        tealAccent: _tealAccent,
        apiClient: _apiClient,
        compact: true,
      ),
    );
  }

  Widget _buildMusicList() {
    final error = _errors[_SearchSource.music];
    if (_loading[_SearchSource.music] == true) {
      return const Center(child: CircularProgressIndicator(color: _tealAccent));
    }
    if (error != null) {
      return Center(
        child: _StateCard(
          icon: Icons.cloud_off_rounded,
          title: 'No pudimos conectar',
          message: error,
          action: OutlinedButton.icon(
            onPressed: () => _retry(_SearchSource.music),
            icon: const Icon(Icons.refresh),
            label: const Text('Reintentar'),
          ),
        ),
      );
    }
    if (_hasResults[_SearchSource.music] != true) {
      return const _StateCard(
        icon: Icons.music_note_rounded,
        title: 'Encuentra música al instante',
        message: 'Toca una canción o artista para abrirlo en Artistas y canciones.',
      );
    }
    if (_tracks.isEmpty) {
      return Center(
        child: _StateCard(
          icon: Icons.search_off_rounded,
          title: 'Sin resultados para «$_activeQuery»',
          message: 'Prueba buscarlo directamente en Artistas y canciones.',
          action: OutlinedButton.icon(
            onPressed: () => _openInArtistsTab(_activeQuery),
            icon: const Icon(Icons.graphic_eq_rounded, size: 18),
            label: const Text('Abrir Artistas y canciones'),
          ),
        ),
      );
    }

    return ListView.separated(
      padding: const EdgeInsets.fromLTRB(12, 12, 12, 24),
      itemCount: _tracks.length,
      separatorBuilder: (_, _) => const SizedBox(height: 8),
      itemBuilder: (context, index) => _buildMusicQuickCard(_tracks[index]),
    );
  }

  Widget _buildMusicQuickCard(MusicTrack track) {
    final subtitleParts = <String>[
      if (track.artist.trim().isNotEmpty) track.artist.trim(),
      if (track.type.trim().isNotEmpty && track.type != 'track')
        track.type.trim(),
    ];
    final targetQuery = track.title.trim().isNotEmpty
        ? track.title.trim()
        : track.artist.trim();

    return Material(
      color: Colors.white,
      child: InkWell(
        onTap: () => _openInArtistsTab(targetQuery),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            border: Border.all(color: const Color(0xFFE0EAE8)),
          ),
          child: Row(
            children: [
              if (track.artworkUrl != null && track.artworkUrl!.isNotEmpty)
                ClipRRect(
                  borderRadius: BorderRadius.circular(2),
                  child: Image.network(
                    track.artworkUrl!,
                    width: 46,
                    height: 46,
                    fit: BoxFit.cover,
                    errorBuilder: (_, _, _) => const MeloopImagePlaceholder(
                      width: 46,
                      height: 46,
                      iconSize: 20,
                      borderColor: null,
                    ),
                  ),
                )
              else
                const MeloopImagePlaceholder(
                  width: 46,
                  height: 46,
                  iconSize: 20,
                  borderColor: null,
                ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      track.title,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: MeloopTypography.lexend(
                        fontSize: 14,
                        fontWeight: FontWeight.w700,
                        color: _inkDark,
                      ),
                    ),
                    if (subtitleParts.isNotEmpty) ...[
                      const SizedBox(height: 2),
                      Text(
                        subtitleParts.join(' · '),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: 12,
                          color: _inkMuted,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(width: 10),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 5),
                decoration: BoxDecoration(
                  color: const Color(0xFFEAF7F4),
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      'Ver en música',
                      style: MeloopTypography.lexend(
                        fontSize: 11,
                        fontWeight: FontWeight.w600,
                        color: const Color(0xFF0D5C5E),
                      ),
                    ),
                    const SizedBox(width: 3),
                    const Icon(
                      Icons.arrow_forward_rounded,
                      size: 13,
                      color: Color(0xFF0D5C5E),
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

  Widget _buildUsersList({required bool isDesktop}) {
    final error = _errors[_SearchSource.users];
    if (_loading[_SearchSource.users] == true) {
      return const Center(child: CircularProgressIndicator(color: _tealAccent));
    }
    if (error != null) {
      return Center(
        child: _StateCard(
          icon: Icons.cloud_off_rounded,
          title: 'No pudimos conectar',
          message: error,
          action: OutlinedButton.icon(
            onPressed: () => _retry(_SearchSource.users),
            icon: const Icon(Icons.refresh),
            label: const Text('Reintentar'),
          ),
        ),
      );
    }
    if (_hasResults[_SearchSource.users] != true) {
      return _buildFriendsIdleCard();
    }
    if (_users.isEmpty) {
      return Center(
        child: _StateCard(
          icon: Icons.person_search_rounded,
          title: 'Sin personas para «$_activeQuery»',
          message: 'Explora sugerencias y solicitudes en la pestaña Amigos.',
          action: OutlinedButton.icon(
            onPressed: () => _openInFriendsTab(_activeQuery),
            icon: const Icon(Icons.people_outline_rounded, size: 18),
            label: const Text('Ir a Amigos'),
          ),
        ),
      );
    }

    return ListView.separated(
      padding: EdgeInsets.fromLTRB(isDesktop ? 0 : 12, 8, isDesktop ? 0 : 12, 24),
      itemCount: _users.length,
      separatorBuilder: (_, _) => const SizedBox(height: 8),
      itemBuilder: (context, index) => _buildUserCard(_users[index]),
    );
  }

  Widget _buildUserCard(SearchUser user) {
    final status = user.friendStatus.toLowerCase();
    final isFriend = status == 'friends' || status == 'accepted';
    final isPending = status == 'pending' || status == 'pendiente';
    final isSending = _sendingRequests.contains(user.id);

    return Material(
      color: Colors.white,
      child: InkWell(
        onTap: () => _openInFriendsTab(user.name),
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
          decoration: BoxDecoration(
            border: Border.all(color: const Color(0xFFDCEAE7)),
          ),
          child: Row(
            children: [
              if (user.avatarUrl != null && user.avatarUrl!.isNotEmpty)
                Container(
                  width: 44,
                  height: 44,
                  decoration: BoxDecoration(
                    border: Border.all(color: const Color(0xFFD0DFDC)),
                  ),
                  child: Image.network(
                    user.avatarUrl!,
                    fit: BoxFit.cover,
                    errorBuilder: (_, _, _) => const MeloopImagePlaceholder(
                      width: 44,
                      height: 44,
                      iconSize: 20,
                    ),
                  ),
                )
              else
                const MeloopImagePlaceholder(
                  width: 44,
                  height: 44,
                  iconSize: 20,
                ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      user.name,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: MeloopTypography.lexend(
                        fontSize: 14.5,
                        fontWeight: FontWeight.w700,
                        color: _inkDark,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Row(
                      children: [
                        const Icon(
                          Icons.stars_rounded,
                          size: 14,
                          color: Color(0xFFD49D26),
                        ),
                        const SizedBox(width: 4),
                        Flexible(
                          child: Text(
                            'Nivel ${user.level}',
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: MeloopTypography.lexend(
                              fontSize: 12,
                              color: _inkMuted,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              OutlinedButton(
                style: OutlinedButton.styleFrom(
                  foregroundColor: isFriend || isPending
                      ? const Color(0xFF56716D)
                      : const Color(0xFF0D5C5E),
                  backgroundColor: isFriend || isPending
                      ? const Color(0xFFF0F5F4)
                      : const Color(0xFFE9F8F5),
                  side: BorderSide(
                    color: isFriend || isPending
                        ? const Color(0xFFDDE7E5)
                        : const Color(0xFFBDE9E1),
                  ),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(18),
                  ),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 12,
                    vertical: 6,
                  ),
                  minimumSize: const Size(86, 34),
                ),
                onPressed: isFriend || isPending || isSending
                    ? null
                    : () => _sendFriendRequest(user),
                child: isSending
                    ? const SizedBox(
                        width: 15,
                        height: 15,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Text(
                        isFriend
                            ? 'Amigos'
                            : isPending
                            ? 'Pendiente'
                            : 'Agregar',
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                        style: MeloopTypography.lexend(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFriendsIdleCard() {
    return Center(
      child: _StateCard(
        icon: Icons.people_alt_outlined,
        title: 'Busca amigos rápidamente',
        message:
            'Escribe un nombre para encontrar personas o abre tu red completa en Amigos.',
        action: OutlinedButton.icon(
          onPressed: () => _openInFriendsTab(null),
          icon: const Icon(Icons.arrow_forward_rounded, size: 16),
          label: const Text('Explorar Amigos'),
        ),
      ),
    );
  }

  Widget _emptyResults() => Center(
    child: _StateCard(
      icon: Icons.search_off_rounded,
      title: 'Sin resultados para «$_activeQuery»',
      message: 'Prueba con otras palabras o cambia el filtro de búsqueda.',
    ),
  );

  Widget _buildCompactErrorBanner({
    required String message,
    required VoidCallback onRetry,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF8EE),
        border: Border.all(color: const Color(0xFFF3DEC0)),
      ),
      child: Row(
        children: [
          const Icon(
            Icons.cloud_off_rounded,
            size: 18,
            color: Color(0xFFB4690E),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: MeloopTypography.lexend(
                fontSize: 12.5,
                color: const Color(0xFF6B4210),
              ),
            ),
          ),
          TextButton(onPressed: onRetry, child: const Text('Reintentar')),
        ],
      ),
    );
  }

  Widget _buildCompactEmptyBanner(
    String message, {
    String? actionLabel,
    VoidCallback? onAction,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFFF6FAF9),
        border: Border.all(color: const Color(0xFFE0ECEA)),
      ),
      child: Row(
        children: [
          const Icon(Icons.search_off_rounded, size: 18, color: _inkMuted),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              message,
              style: MeloopTypography.lexend(fontSize: 12.5, color: _inkMuted),
            ),
          ),
          if (actionLabel != null && onAction != null)
            TextButton(onPressed: onAction, child: Text(actionLabel)),
        ],
      ),
    );
  }

  Widget _buildRecentSearches({required bool isDesktop}) {
    if (_recentSearches.isEmpty) {
      return const _StateCard(
        icon: Icons.manage_search_rounded,
        title: '¿Qué tienes en mente?',
        message: 'Tus búsquedas recientes aparecerán aquí.',
      );
    }

    return ListView(
      padding: EdgeInsets.fromLTRB(isDesktop ? 0 : 16, 8, isDesktop ? 0 : 16, 20),
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(2, 4, 2, 10),
          child: Row(
            children: [
              const Icon(
                Icons.history_rounded,
                size: 18,
                color: Color(0xFF0D5C5E),
              ),
              const SizedBox(width: 8),
              Text(
                'Búsquedas recientes',
                style: MeloopTypography.lexend(
                  color: _inkDark,
                  fontSize: 14.5,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const Spacer(),
              Text(
                '${_recentSearches.length}/5',
                style: MeloopTypography.lexend(color: _inkMuted, fontSize: 12),
              ),
            ],
          ),
        ),
        for (final query in _recentSearches)
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Material(
              color: Colors.white,
              shape: const RoundedRectangleBorder(
                side: BorderSide(color: Color(0xFFE2ECEA)),
              ),
              child: ListTile(
                dense: true,
                leading: const Icon(
                  Icons.history_rounded,
                  size: 18,
                  color: Color(0xFF79918D),
                ),
                title: Text(
                  query,
                  style: MeloopTypography.lexend(
                    color: _inkDark,
                    fontSize: 13.5,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                onTap: () => _applyRecentQuery(query),
                trailing: IconButton(
                  tooltip: 'Quitar búsqueda reciente',
                  icon: const Icon(Icons.close, size: 17),
                  color: _inkMuted,
                  onPressed: () =>
                      setState(() => _recentSearches.remove(query)),
                ),
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildPostFilterChips() {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: [
          Text(
            'Filtrar blogs:',
            style: MeloopTypography.lexend(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: _inkMuted,
            ),
          ),
          const SizedBox(width: 8),
          for (final entry in _filters.entries) ...[
            Padding(
              padding: const EdgeInsets.only(right: 6),
              child: ChoiceChip(
                label: Text(
                  entry.key,
                  style: MeloopTypography.lexend(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: _postFilter == entry.value ? Colors.white : _inkDark,
                  ),
                ),
                selected: _postFilter == entry.value,
                selectedColor: _tealAccent,
                backgroundColor: const Color(0xFFF0F6F5),
                showCheckmark: false,
                side: BorderSide(
                  color: _postFilter == entry.value
                      ? _tealAccent
                      : const Color(0xFFD5E4E1),
                ),
                visualDensity: VisualDensity.compact,
                onSelected: (_) => _changePostFilter(entry.value),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildSearchInput() {
    return Row(
      children: [
        Expanded(
          child: Container(
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              border: Border.all(color: const Color(0xFFD8E6E4)),
            ),
            child: TextField(
              controller: _queryController,
              textInputAction: TextInputAction.search,
              onSubmitted: _submitSearch,
              style: MeloopTypography.lexend(fontSize: 14, color: _inkDark),
              decoration: InputDecoration(
                hintText: 'Publicaciones, canciones o amigos…',
                hintStyle: MeloopTypography.lexend(
                  color: _inkMuted,
                  fontSize: 13.5,
                ),
                prefixIcon: const Icon(
                  Icons.search_rounded,
                  color: Color(0xFF16877F),
                  size: 20,
                ),
                suffixIcon: _queryController.text.isEmpty
                    ? null
                    : IconButton(
                        tooltip: 'Limpiar búsqueda',
                        icon: const Icon(Icons.close_rounded, size: 19),
                        color: _inkMuted,
                        onPressed: () => _queryController.clear(),
                      ),
                border: InputBorder.none,
                contentPadding: const EdgeInsets.symmetric(
                  horizontal: 14,
                  vertical: 12,
                ),
              ),
            ),
          ),
        ),
        const SizedBox(width: 8),
        FilledButton(
          onPressed: _submitSearch,
          style: FilledButton.styleFrom(
            backgroundColor: _tealAccent,
            foregroundColor: Colors.white,
            elevation: 0,
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 13),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(22),
            ),
          ),
          child: Text(
            'Buscar',
            style: MeloopTypography.lexend(
              fontSize: 13,
              fontWeight: FontWeight.w600,
              color: Colors.white,
            ),
          ),
        ),
      ],
    );
  }

  PreferredSizeWidget _buildBrandedAppBar(bool isDesktop) {
    if (isDesktop) {
      return PreferredSize(
        preferredSize: const Size.fromHeight(84),
        child: DesktopAppHeader(
          activeSection: DesktopNavigationSection.none,
          onHomeTap: () => navigateToDesktopHome(context),
          onFriendsTap: () =>
              navigateToDesktopSection(context, (_) => const FriendsScreen()),
          onArtistsTap: () => navigateToDesktopSection(
            context,
            (_) => const ArtistsAndSongsScreen(),
          ),
          onProfileTap: () => navigateToDesktopSection(
            context,
            (_) => ProfileScreen(apiClient: _apiClient),
          ),
          searchOpen: true,
          searchController: _queryController,
          onSearchSubmitted: _submitSearch,
          onSearchBack: () => navigateToDesktopHome(context),
        ),
      );
    }

    if (Theme.of(context).platform == TargetPlatform.android) {
      final topInset = MediaQuery.viewPaddingOf(context).top;
      return PreferredSize(
        preferredSize: Size.fromHeight(55 + topInset),
        child: Container(
          color: const Color(0xFF1ABC9C),
          child: SafeArea(
            bottom: false,
            child: Container(
              height: 55,
              color: const Color(0xFF1ABC9C),
              padding: const EdgeInsets.symmetric(horizontal: 8),
              child: Row(
                children: [
                  IconButton(
                    tooltip: 'Volver',
                    onPressed: () => Navigator.maybePop(context),
                    icon: const Icon(Icons.arrow_back, color: Colors.white),
                  ),
                  Expanded(
                    child: Center(
                      child: Image.asset(
                        'assets/imagenes/meloop.png',
                        height: 24,
                        fit: BoxFit.contain,
                      ),
                    ),
                  ),
                  const SizedBox(width: 48),
                ],
              ),
            ),
          ),
        ),
      );
    }

    return PreferredSize(
      preferredSize: const Size.fromHeight(55),
      child: Container(
        height: 55,
        color: const Color(0xFF1ABC9C),
        padding: const EdgeInsets.symmetric(horizontal: 20),
        child: Row(
          children: [
            IconButton(
              tooltip: 'Volver',
              padding: EdgeInsets.zero,
              constraints: const BoxConstraints(),
              onPressed: () => Navigator.maybePop(context),
              icon: const Icon(
                Icons.arrow_back_rounded,
                color: Colors.white,
                size: 20,
              ),
            ),
            const SizedBox(width: 16),
            Image.asset(
              'assets/imagenes/meloop.png',
              height: 24,
              fit: BoxFit.contain,
            ),
            const SizedBox(width: 18),
            Text(
              'Búsqueda rápida',
              style: MeloopTypography.lexend(
                color: Colors.white,
                fontSize: 13.5,
                fontWeight: FontWeight.w600,
              ),
            ),
            const Spacer(),
            const Icon(Icons.graphic_eq_rounded, color: Colors.white, size: 20),
          ],
        ),
      ),
    );
  }
}

class _StateCard extends StatelessWidget {
  final IconData icon;
  final String title;
  final String message;
  final Widget? action;

  const _StateCard({
    required this.icon,
    required this.title,
    required this.message,
    this.action,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Container(
          constraints: const BoxConstraints(maxWidth: 360),
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 18),
          decoration: BoxDecoration(
            color: Colors.white,
            border: Border.all(color: const Color(0xFFE0EAE8)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                height: 48,
                width: 48,
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F7F3),
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Icon(icon, color: const Color(0xFF0D7771), size: 26),
              ),
              const SizedBox(height: 12),
              Text(
                title,
                textAlign: TextAlign.center,
                style: MeloopTypography.lexend(
                  color: const Color(0xFF183D3B),
                  fontSize: 15,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 5),
              Text(
                message,
                textAlign: TextAlign.center,
                style: MeloopTypography.lexend(
                  color: const Color(0xFF788A86),
                  fontSize: 12.5,
                  height: 1.45,
                ),
              ),
              if (action != null) ...[const SizedBox(height: 12), action!],
            ],
          ),
        ),
      ),
    );
  }
}
