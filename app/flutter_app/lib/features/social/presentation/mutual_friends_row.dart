import 'package:flutter/material.dart';
import 'package:flutter_app/models/mutual_friends.dart';
import 'package:flutter_app/services/api_client.dart';

import 'mutual_friends_screen.dart';

class MutualFriendsRow extends StatefulWidget {
  final String targetUserId;
  final ApiClient? apiClient;

  const MutualFriendsRow({
    super.key,
    required this.targetUserId,
    this.apiClient,
  });

  @override
  State<MutualFriendsRow> createState() => _MutualFriendsRowState();
}

class _MutualFriendsRowState extends State<MutualFriendsRow> {
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  MutualFriendsPage? _page;
  ApiException? _error;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    _loadPreview();
  }

  @override
  void didUpdateWidget(covariant MutualFriendsRow oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.targetUserId != widget.targetUserId) {
      _loadPreview();
    }
  }

  @override
  void dispose() {
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  Future<void> _loadPreview() async {
    setState(() {
      _isLoading = true;
      _error = null;
      _page = null;
    });

    try {
      final response = await _apiClient.getMutualFriends(
        widget.targetUserId,
        limit: 3,
      );
      final page = response.data;
      if (page == null) {
        throw const ApiException(
          message: 'El servidor no devolvió los amigos en común.',
        );
      }
      if (!mounted) return;
      setState(() {
        _page = page;
        _isLoading = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error;
        _isLoading = false;
      });
    }
  }

  void _openScreen() {
    Navigator.push(
      context,
      MaterialPageRoute<void>(
        builder: (_) => MutualFriendsScreen(
          targetUserId: widget.targetUserId,
          initialPage: _page,
          apiClient: _apiClient,
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const SizedBox(
        height: 48,
        child: Center(
          child: SizedBox(
            width: 18,
            height: 18,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
        ),
      );
    }

    final page = _page;
    if (page != null && page.totalCount == 0) return const SizedBox.shrink();

    final error = _error;
    if (error != null) {
      final isNotFound = error.statusCode == 404;
      return TextButton.icon(
        onPressed: isNotFound ? _openScreen : _loadPreview,
        icon: Icon(isNotFound ? Icons.person_off : Icons.refresh),
        label: Text(
          isNotFound
              ? 'Perfil no disponible'
              : 'No se pudieron cargar los amigos en común. Reintentar',
        ),
      );
    }

    if (page == null) return const SizedBox.shrink();

    final avatars = page.items.take(3).toList();
    return Semantics(
      button: true,
      label: 'Ver ${page.totalCount} amigos en común',
      child: InkWell(
        onTap: _openScreen,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 12),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              SizedBox(
                width: avatars.isEmpty ? 0 : 24 + (avatars.length - 1) * 17,
                height: 26,
                child: Stack(
                  children: [
                    for (var index = 0; index < avatars.length; index++)
                      Positioned(
                        left: index * 17,
                        child: _MutualFriendAvatar(friend: avatars[index]),
                      ),
                  ],
                ),
              ),
              if (avatars.isNotEmpty) const SizedBox(width: 8),
              Text(
                '${page.totalCount} amigos en común',
                style: const TextStyle(fontSize: 13, color: Colors.black87),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class _MutualFriendAvatar extends StatelessWidget {
  final MutualFriend friend;

  const _MutualFriendAvatar({required this.friend});

  @override
  Widget build(BuildContext context) {
    final avatarUrl = friend.avatarUrl;
    return Semantics(
      label: 'Avatar de ${friend.username}, amigo en común',
      image: true,
      child: Container(
        width: 26,
        height: 26,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          border: Border.all(color: Colors.white, width: 2),
        ),
        child: ClipOval(
          child: avatarUrl == null || avatarUrl.isEmpty
              ? const ColoredBox(
                  color: Color(0xFFE0E0E0),
                  child: Icon(Icons.person, size: 16, color: Colors.white),
                )
              : Image.network(
                  avatarUrl,
                  fit: BoxFit.cover,
                  errorBuilder: (_, _, _) => const ColoredBox(
                    color: Color(0xFFE0E0E0),
                    child: Icon(Icons.person, size: 16, color: Colors.white),
                  ),
                ),
        ),
      ),
    );
  }
}
