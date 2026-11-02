import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/music/presentation/music_track_detail_screen.dart';
import 'package:flutter_app/features/music/presentation/music_track_tile.dart';
import 'package:flutter_app/features/profile/presentation/profile_screen.dart';
import 'package:flutter_app/models/global_search.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:flutter_app/widgets/post_card.dart';

enum _SearchSource { users, posts, music }

class GlobalSearchScreen extends StatefulWidget {
  final ApiClient? apiClient;

  const GlobalSearchScreen({super.key, this.apiClient});

  @override
  State<GlobalSearchScreen> createState() => _GlobalSearchScreenState();
}

class _GlobalSearchScreenState extends State<GlobalSearchScreen>
    with SingleTickerProviderStateMixin {
  static const _filters = <String, String>{
    'Texto': 'text',
    'Canción': 'song',
    'Artista': 'artist',
    'Género': 'genre',
  };

  final TextEditingController _queryController = TextEditingController();
  final List<String> _recentSearches = [];
  final Map<_SearchSource, CancelToken> _cancelTokens = {};
  late final TabController _tabController;
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  Timer? _debounce;

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

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ApiClient();
    _ownsApiClient = widget.apiClient == null;
    _tabController = TabController(length: 3, vsync: this);
    _tabController.addListener(_onTabChanged);
    _queryController.addListener(_onQueryChanged);
  }

  void _onTabChanged() {
    if (mounted) setState(() {});
  }

  @override
  void dispose() {
    _debounce?.cancel();
    for (final token in _cancelTokens.values) {
      token.cancel('La búsqueda cambió.');
    }
    _queryController
      ..removeListener(_onQueryChanged)
      ..dispose();
    _tabController.removeListener(_onTabChanged);
    _tabController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  void _onQueryChanged() {
    final query = _queryController.text.trim();
    _debounce?.cancel();
    _cancelPendingRequests();
    if (query.length < 2) {
      setState(() {
        _activeQuery = '';
        _clearResults();
      });
      return;
    }
    setState(() {
      _activeQuery = '';
      _clearResults();
    });

    _debounce = Timer(const Duration(milliseconds: 400), () {
      if (mounted && query == _queryController.text.trim()) {
        _searchAll(query);
      }
    });
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

  Future<void> _changePostFilter(String? filter) async {
    if (filter == null || filter == _postFilter) return;
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
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _sendingRequests.remove(user.id));
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('No se pudo enviar la solicitud: ${error.message}'),
        ),
      );
    }
  }

  Widget _buildSourceTab(_SearchSource source) {
    final error = _errors[source];
    if (_loading[source] == true) {
      return const Center(
        child: CircularProgressIndicator(color: Color(0xFF1ABC9C)),
      );
    }
    if (error != null) {
      return Center(
        child: _StateCard(
          icon: Icons.cloud_off_rounded,
          title: 'No pudimos conectar',
          message: error,
          action: OutlinedButton.icon(
            onPressed: () => _retry(source),
            icon: const Icon(Icons.refresh),
            label: const Text('Reintentar'),
          ),
        ),
      );
    }
    if (_hasResults[source] != true) {
      return const _StateCard(
        icon: Icons.travel_explore_rounded,
        title: 'Todo empieza con una búsqueda',
        message: 'Encuentra personas, publicaciones y música en Meloop.',
      );
    }

    switch (source) {
      case _SearchSource.users:
        if (_users.isEmpty) return _emptyResults();
        return ListView.builder(
          padding: const EdgeInsets.symmetric(vertical: 8),
          itemCount: _users.length,
          itemBuilder: (context, index) => _buildUserTile(_users[index]),
        );
      case _SearchSource.posts:
        if (_posts.isEmpty) return _emptyResults();
        return ListView.builder(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
          itemCount: _posts.length,
          itemBuilder: (context, index) => Card(
            margin: const EdgeInsets.only(bottom: 16),
            color: Colors.white,
            elevation: 1,
            shadowColor: const Color(0x140D5C5E),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(18),
              side: const BorderSide(color: Color(0xFFE7EEEC)),
            ),
            child: Padding(
              padding: const EdgeInsets.all(12),
              child: PostCard(
                postData: _posts[index].cardData,
                tealAccent: const Color(0xFF1ABC9C),
                apiClient: _apiClient,
                compact: true,
              ),
            ),
          ),
        );
      case _SearchSource.music:
        if (_tracks.isEmpty) return _emptyResults();
        return ListView.builder(
          itemCount: _tracks.length,
          itemBuilder: (context, index) {
            final track = _tracks[index];
            return MusicTrackTile(
              track: track,
              onTap: () => Navigator.push(
                context,
                MaterialPageRoute<void>(
                  builder: (_) => MusicDetailScreen(
                    type: track.type,
                    spotifyId: track.id,
                    apiClient: _apiClient,
                  ),
                ),
              ),
            );
          },
        );
    }
  }

  Widget _emptyResults() => Center(
    child: _StateCard(
      icon: Icons.search_off_rounded,
      title: 'Sin resultados para «$_activeQuery»',
      message: 'Prueba con otras palabras o cambia el filtro de búsqueda.',
    ),
  );

  Widget _buildUserTile(SearchUser user) {
    final status = user.friendStatus.toLowerCase();
    final isFriend = status == 'friends' || status == 'accepted';
    final isPending = status == 'pending' || status == 'pendiente';
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
      elevation: 0,
      color: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
        side: const BorderSide(color: Color(0xFFE7EEEC)),
      ),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: 14, vertical: 5),
        leading: Container(
          padding: const EdgeInsets.all(2),
          decoration: const BoxDecoration(
            shape: BoxShape.circle,
            gradient: LinearGradient(
              colors: [Color(0xFF1ABC9C), Color(0xFF80DED0)],
            ),
          ),
          child: CircleAvatar(
            radius: 23,
            backgroundColor: const Color(0xFFEAF7F4),
            backgroundImage: user.avatarUrl == null
                ? null
                : NetworkImage(user.avatarUrl!),
            child: user.avatarUrl == null
                ? const Icon(Icons.person_rounded, color: Color(0xFF0D5C5E))
                : null,
          ),
        ),
        title: Text(
          user.name,
          style: const TextStyle(
            color: Color(0xFF183D3B),
            fontWeight: FontWeight.w700,
          ),
        ),
        subtitle: Row(
          children: [
            const Icon(Icons.stars_rounded, size: 15, color: Color(0xFFD49D26)),
            const SizedBox(width: 4),
            Text(
              'Nivel ${user.level}',
              style: const TextStyle(color: Color(0xFF70827F), fontSize: 12),
            ),
          ],
        ),
        onTap: () => Navigator.push(
          context,
          MaterialPageRoute<void>(
            builder: (_) => ProfileScreen(username: user.name, userId: user.id),
          ),
        ),
        trailing: SizedBox(
          width: 112,
          child: OutlinedButton(
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
                borderRadius: BorderRadius.circular(12),
              ),
              padding: const EdgeInsets.symmetric(horizontal: 8),
            ),
            onPressed:
                isFriend || isPending || _sendingRequests.contains(user.id)
                ? null
                : () => _sendFriendRequest(user),
            child: _sendingRequests.contains(user.id)
                ? const SizedBox(
                    width: 16,
                    height: 16,
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
                  ),
          ),
        ),
      ),
    );
  }

  Widget _buildRecentSearches() {
    if (_recentSearches.isEmpty) {
      return const _StateCard(
        icon: Icons.manage_search_rounded,
        title: '¿Qué tienes en mente?',
        message: 'Tus búsquedas recientes aparecerán aquí.',
      );
    }
    return ListView(
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(4, 8, 4, 12),
          child: Row(
            children: [
              const Icon(Icons.history_rounded, color: Color(0xFF0D5C5E)),
              const SizedBox(width: 8),
              const Text(
                'Búsquedas recientes',
                style: TextStyle(
                  color: Color(0xFF183D3B),
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const Spacer(),
              Text(
                '${_recentSearches.length}/5',
                style: const TextStyle(color: Color(0xFF80918E), fontSize: 12),
              ),
            ],
          ),
        ),
        for (final query in _recentSearches)
          Card(
            elevation: 0,
            color: Colors.white,
            margin: const EdgeInsets.only(bottom: 8),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(14),
              side: const BorderSide(color: Color(0xFFE7EEEC)),
            ),
            child: ListTile(
              leading: const Icon(
                Icons.history_rounded,
                color: Color(0xFF79918D),
              ),
              title: Text(
                query,
                style: const TextStyle(
                  color: Color(0xFF294441),
                  fontWeight: FontWeight.w600,
                ),
              ),
              onTap: () {
                _queryController.text = query;
                _queryController.selection = TextSelection.collapsed(
                  offset: query.length,
                );
              },
              trailing: IconButton(
                tooltip: 'Quitar búsqueda reciente',
                icon: const Icon(Icons.close, size: 18),
                color: const Color(0xFF80918E),
                onPressed: () => setState(() => _recentSearches.remove(query)),
              ),
            ),
          ),
      ],
    );
  }

  PreferredSizeWidget _buildBrandedAppBar() {
    return PreferredSize(
      preferredSize: const Size.fromHeight(55),
      child: Container(
        height: 55,
        color: const Color(0xFF1ABC9C),
        padding: const EdgeInsets.symmetric(horizontal: 24),
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
            const SizedBox(width: 24),
            const Text(
              'Búsqueda global',
              style: TextStyle(
                color: Colors.white,
                fontSize: 13,
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

  Widget _buildSearchInput() {
    return Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFE7EEEC)),
        boxShadow: const [
          BoxShadow(
            color: Color(0x120D5C5E),
            blurRadius: 14,
            offset: Offset(0, 4),
          ),
        ],
      ),
      child: TextField(
        controller: _queryController,
        textInputAction: TextInputAction.search,
        decoration: InputDecoration(
          hintText: 'Personas, publicaciones, canciones…',
          hintStyle: const TextStyle(color: Color(0xFF8A9B98), fontSize: 14),
          prefixIcon: const Icon(
            Icons.search_rounded,
            color: Color(0xFF0D5C5E),
          ),
          suffixIcon: _queryController.text.isEmpty
              ? null
              : IconButton(
                  tooltip: 'Limpiar búsqueda',
                  icon: const Icon(Icons.close_rounded),
                  color: const Color(0xFF718480),
                  onPressed: () => _queryController.clear(),
                ),
          filled: true,
          fillColor: const Color(0xFFF5F9F8),
          contentPadding: const EdgeInsets.symmetric(vertical: 14),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none,
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: BorderSide.none,
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(12),
            borderSide: const BorderSide(color: Color(0xFF1ABC9C), width: 1.4),
          ),
        ),
      ),
    );
  }

  Widget _buildTabs() {
    return Container(
      padding: const EdgeInsets.all(5),
      decoration: BoxDecoration(
        color: const Color(0xFFF5F8F7),
        borderRadius: BorderRadius.circular(14),
      ),
      child: TabBar(
        controller: _tabController,
        dividerColor: Colors.transparent,
        indicatorSize: TabBarIndicatorSize.tab,
        indicator: BoxDecoration(
          color: const Color(0xFF1ABC9C),
          borderRadius: BorderRadius.circular(11),
        ),
        labelColor: Colors.white,
        unselectedLabelColor: const Color(0xFF667C78),
        labelStyle: const TextStyle(fontSize: 12, fontWeight: FontWeight.w700),
        tabs: const [
          Tab(
            icon: Icon(Icons.people_alt_outlined, size: 18),
            text: 'Usuarios',
          ),
          Tab(
            icon: Icon(Icons.article_outlined, size: 18),
            text: 'Publicaciones',
          ),
          Tab(icon: Icon(Icons.music_note_rounded, size: 18), text: 'Música'),
        ],
      ),
    );
  }

  Widget _buildPostFilter() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 2),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFE2ECE9)),
      ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<String>(
          value: _postFilter,
          icon: const Icon(Icons.keyboard_arrow_down_rounded),
          borderRadius: BorderRadius.circular(12),
          items: _filters.entries
              .map(
                (entry) => DropdownMenuItem(
                  value: entry.value,
                  child: Text(entry.key),
                ),
              )
              .toList(),
          onChanged: _changePostFilter,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final queryTooShort = _queryController.text.trim().length < 2;
    return Scaffold(
      backgroundColor: const Color(0xFF0D5C5E),
      appBar: _buildBrandedAppBar(),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 920),
            child: Container(
              color: Colors.white,
              child: Column(
                children: [
                  Padding(
                    padding: const EdgeInsets.fromLTRB(24, 20, 24, 12),
                    child: _buildSearchInput(),
                  ),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 24),
                    child: _buildTabs(),
                  ),
                  if (!queryTooShort &&
                      _tabController.index == 1 &&
                      _activeQuery.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.fromLTRB(24, 10, 24, 0),
                      child: Align(
                        alignment: Alignment.centerLeft,
                        child: _buildPostFilter(),
                      ),
                    ),
                  Expanded(
                    child: queryTooShort
                        ? _buildRecentSearches()
                        : TabBarView(
                            controller: _tabController,
                            children: [
                              _buildSourceTab(_SearchSource.users),
                              _buildSourceTab(_SearchSource.posts),
                              _buildSourceTab(_SearchSource.music),
                            ],
                          ),
                  ),
                ],
              ),
            ),
          ),
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
        padding: const EdgeInsets.all(24),
        child: Container(
          constraints: const BoxConstraints(maxWidth: 380),
          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 18),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(22),
            border: Border.all(color: const Color(0xFFE4ECEA)),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Container(
                height: 52,
                width: 52,
                decoration: BoxDecoration(
                  color: const Color(0xFFE8F7F3),
                  borderRadius: BorderRadius.circular(20),
                ),
                child: Icon(icon, color: const Color(0xFF0D7771), size: 30),
              ),
              const SizedBox(height: 12),
              Text(
                title,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: Color(0xFF183D3B),
                  fontSize: 16,
                  fontWeight: FontWeight.w700,
                ),
              ),
              const SizedBox(height: 5),
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: Color(0xFF788A86),
                  fontSize: 13,
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
