import 'package:flutter/material.dart';
import 'package:flutter_app/features/profile/presentation/profile_screen.dart';
import 'package:flutter_app/models/mutual_friends.dart';
import 'package:flutter_app/services/api_client.dart';

class MutualFriendsScreen extends StatefulWidget {
  final String targetUserId;
  final MutualFriendsPage? initialPage;
  final ApiClient? apiClient;

  const MutualFriendsScreen({
    super.key,
    required this.targetUserId,
    this.initialPage,
    this.apiClient,
  });

  @override
  State<MutualFriendsScreen> createState() => _MutualFriendsScreenState();
}

class _MutualFriendsScreenState extends State<MutualFriendsScreen> {
  final Color _tealAccent = const Color(0xFF1ABC9C);
  final ScrollController _scrollController = ScrollController();
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  final List<MutualFriend> _friends = [];
  final Set<String> _friendIds = {};

  MutualFriendsPage? _page;
  ApiException? _error;
  bool _isInitialLoading = true;
  bool _isLoadingMore = false;
  bool _isNotFound = false;

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    _scrollController.addListener(_onScroll);

    final initialPage = widget.initialPage;
    if (initialPage == null) {
      _loadInitialPage();
    } else {
      _page = initialPage;
      _appendUnique(initialPage.items);
      _isInitialLoading = false;
    }
  }

  @override
  void dispose() {
    _scrollController
      ..removeListener(_onScroll)
      ..dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  void _onScroll() {
    if (!_scrollController.hasClients || _isLoadingMore || _isInitialLoading) {
      return;
    }
    if (_scrollController.position.extentAfter < 240 &&
        (_page?.hasMore ?? false) &&
        !_isNotFound) {
      _loadMore();
    }
  }

  Future<void> _loadInitialPage() async {
    setState(() {
      _isInitialLoading = true;
      _error = null;
      _isNotFound = false;
    });

    try {
      final response = await _apiClient.getMutualFriends(
        widget.targetUserId,
      );
      if (!mounted) return;
      final page = response.data;
      if (page == null) {
        throw const ApiException(
          message: 'El servidor no devolvió los amigos en común.',
        );
      }
      setState(() {
        _page = page;
        _friends.clear();
        _friendIds.clear();
        _appendUnique(page.items);
        _isInitialLoading = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error;
        _isNotFound = error.statusCode == 404;
        _isInitialLoading = false;
      });
    }
  }

  Future<void> _loadMore() async {
    final cursor = _page?.nextCursor;
    if (cursor == null || cursor.isEmpty) return;

    setState(() {
      _isLoadingMore = true;
      _error = null;
    });
    try {
      final response = await _apiClient.getMutualFriends(
        widget.targetUserId,
        cursor: cursor,
      );
      if (!mounted) return;
      final page = response.data;
      if (page == null) {
        throw const ApiException(
          message: 'El servidor no devolvió los amigos en común.',
        );
      }
      setState(() {
        _page = page;
        _appendUnique(page.items);
        _isLoadingMore = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error;
        _isNotFound = error.statusCode == 404;
        _isLoadingMore = false;
      });
    }
  }

  void _appendUnique(List<MutualFriend> friends) {
    for (final friend in friends) {
      if (_friendIds.add(friend.id)) _friends.add(friend);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: _tealAccent,
        title: const Text(
          'Amigos en común',
          style: TextStyle(color: Colors.white, fontSize: 16),
        ),
        iconTheme: const IconThemeData(color: Colors.white),
        elevation: 0,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isInitialLoading) {
      return Center(child: CircularProgressIndicator(color: _tealAccent));
    }

    if (_isNotFound) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.person_off, size: 56, color: Colors.grey),
              SizedBox(height: 16),
              Text(
                'Perfil no disponible',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              SizedBox(height: 8),
              Text(
                'No se puede acceder a la información de esta persona en este momento.',
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      );
    }

    final page = _page;
    if (page?.hidden == true) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.visibility_off, size: 56, color: Colors.grey),
              const SizedBox(height: 16),
              Text(
                '${page!.totalCount} amigos en común',
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              const Text(
                'Este usuario oculta su lista de amigos',
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      );
    }

    if (_error != null && _friends.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(Icons.cloud_off, size: 48, color: Colors.grey),
              const SizedBox(height: 12),
              Text(_error!.message, textAlign: TextAlign.center),
              const SizedBox(height: 12),
              TextButton(
                onPressed: _loadInitialPage,
                child: const Text('Reintentar'),
              ),
            ],
          ),
        ),
      );
    }

    if (_friends.isEmpty) {
      return const Center(child: Text('No tienen amigos en común.'));
    }

    return ListView.builder(
      controller: _scrollController,
      itemCount: _friends.length + (_isLoadingMore || _error != null ? 1 : 0),
      itemBuilder: (context, index) {
        if (index >= _friends.length) {
          if (_isLoadingMore) {
            return Padding(
              padding: const EdgeInsets.all(16),
              child: Center(
                child: CircularProgressIndicator(color: _tealAccent),
              ),
            );
          }
          return Center(
            child: TextButton(
              onPressed: _loadMore,
              child: const Text('No se pudo cargar más. Reintentar'),
            ),
          );
        }

        final friend = _friends[index];
        return Semantics(
          button: true,
          label: 'Ver perfil de ${friend.username}',
          child: ListTile(
            leading: Semantics(
              image: true,
              label: 'Avatar de ${friend.username}',
              child: _FriendAvatar(friend: friend),
            ),
            title: Text(
              friend.username,
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
            ),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute<void>(
                  builder: (_) => ProfileScreen(
                    username: friend.username,
                    userId: friend.id,
                  ),
                ),
              );
            },
          ),
        );
      },
    );
  }
}

class _FriendAvatar extends StatelessWidget {
  final MutualFriend friend;

  const _FriendAvatar({required this.friend});

  @override
  Widget build(BuildContext context) {
    final avatarUrl = friend.avatarUrl;
    if (avatarUrl == null || avatarUrl.isEmpty) {
      return const CircleAvatar(
        backgroundColor: Color(0xFFE0E0E0),
        child: Icon(Icons.person, color: Colors.grey),
      );
    }

    return CircleAvatar(
      backgroundColor: const Color(0xFFE0E0E0),
      foregroundImage: NetworkImage(avatarUrl),
      onForegroundImageError: (_, _) {},
      child: const Icon(Icons.person, color: Colors.grey),
    );
  }
}
