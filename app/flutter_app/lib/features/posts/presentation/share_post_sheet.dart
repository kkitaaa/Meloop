import 'package:flutter/material.dart';

import '../../../models/global_search.dart';
import '../../../services/api_client.dart';

class SharePostSheet extends StatefulWidget {
  final ApiClient apiClient;
  final String postId;

  const SharePostSheet({
    super.key,
    required this.apiClient,
    required this.postId,
  });

  @override
  State<SharePostSheet> createState() => _SharePostSheetState();
}

class _SharePostSheetState extends State<SharePostSheet> {
  final TextEditingController _searchController = TextEditingController();
  List<SearchUser> _friends = const [];
  bool _isLoading = true;
  String? _error;
  String? _sendingId;
  String? _sentTo;

  @override
  void initState() {
    super.initState();
    _loadFriends();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadFriends() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });
    try {
      final response = await widget.apiClient.getFriends();
      if (!mounted) return;
      final page = response.data;
      if (page == null) {
        throw const ApiException(
          message: 'El servidor no devolvió tus amigos.',
        );
      }
      setState(() {
        _friends = page.items;
        _isLoading = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error.message;
        _isLoading = false;
      });
    }
  }

  Future<void> _shareWith(SearchUser friend) async {
    setState(() => _sendingId = friend.id);
    try {
      await widget.apiClient.sharePost(
        postId: widget.postId,
        recipientId: friend.id,
      );
      if (mounted) {
        setState(() {
          _sentTo = friend.name;
          _sendingId = null;
        });
      }
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _sendingId = null);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(error.message),
          backgroundColor: const Color(0xFF0D5C5E),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final query = _searchController.text.trim().toLowerCase();
    final visibleFriends = _friends
        .where((friend) => friend.name.toLowerCase().contains(query))
        .toList();
    return SafeArea(
      child: Padding(
        padding: EdgeInsets.only(
          left: 20,
          right: 20,
          top: 20,
          bottom: MediaQuery.viewInsetsOf(context).bottom + 16,
        ),
        child: SizedBox(
          height: MediaQuery.sizeOf(context).height * 0.65,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  const Expanded(
                    child: Text(
                      'Compartir',
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: Color(0xFF183D3B),
                      ),
                    ),
                  ),
                  IconButton(
                    tooltip: 'Cerrar',
                    onPressed: () => Navigator.pop(context),
                    icon: const Icon(Icons.close),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _searchController,
                onChanged: (_) => setState(() {}),
                decoration: InputDecoration(
                  hintText: 'Buscar amigos',
                  prefixIcon: const Icon(Icons.search),
                  filled: true,
                  fillColor: const Color(0xFFF2F7F5),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(14),
                    borderSide: BorderSide.none,
                  ),
                ),
              ),
              const SizedBox(height: 12),
              Expanded(child: _buildFriends(visibleFriends)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildFriends(List<SearchUser> friends) {
    final sentTo = _sentTo;
    if (sentTo != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(
              Icons.check_circle_rounded,
              color: Color(0xFF1ABC9C),
              size: 52,
            ),
            const SizedBox(height: 12),
            Text(
              'Enviada a @$sentTo',
              style: const TextStyle(fontWeight: FontWeight.w700),
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Listo'),
            ),
          ],
        ),
      );
    }
    if (_isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: Color(0xFF1ABC9C)),
      );
    }
    if (_error != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(_error!, textAlign: TextAlign.center),
            TextButton(
              onPressed: _loadFriends,
              child: const Text('Reintentar'),
            ),
          ],
        ),
      );
    }
    if (friends.isEmpty) {
      return Center(
        child: Text(
          _friends.isEmpty
              ? 'Todavía no tienes amigos para compartir.'
              : 'No encontramos amigos con ese nombre.',
          textAlign: TextAlign.center,
          style: const TextStyle(color: Colors.black54),
        ),
      );
    }
    return ListView.separated(
      itemCount: friends.length,
      separatorBuilder: (_, _) => const Divider(height: 1),
      itemBuilder: (context, index) {
        final friend = friends[index];
        final sending = _sendingId == friend.id;
        return ListTile(
          contentPadding: EdgeInsets.zero,
          leading: CircleAvatar(
            backgroundColor: const Color(0xFFE2F5EF),
            backgroundImage: friend.avatarUrl == null
                ? null
                : NetworkImage(friend.avatarUrl!),
            child: friend.avatarUrl == null
                ? const Icon(Icons.person, color: Color(0xFF0D7771))
                : null,
          ),
          title: Text(friend.name),
          subtitle: Text('Nivel ${friend.level}'),
          trailing: IconButton(
            tooltip: 'Enviar a ${friend.name}',
            onPressed: _sendingId == null ? () => _shareWith(friend) : null,
            icon: sending
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.send_rounded, color: Color(0xFF0D7771)),
          ),
        );
      },
    );
  }
}
