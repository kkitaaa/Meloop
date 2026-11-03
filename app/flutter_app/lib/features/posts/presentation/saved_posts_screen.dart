import 'package:flutter/material.dart';

import '../../../models/saved_posts.dart';
import '../../../services/api_client.dart';
import '../../../widgets/post_card.dart';

class SavedPostsScreen extends StatefulWidget {
  final ApiClient? apiClient;

  const SavedPostsScreen({super.key, this.apiClient});

  @override
  State<SavedPostsScreen> createState() => _SavedPostsScreenState();
}

class _SavedPostsScreenState extends State<SavedPostsScreen> {
  static const int _pageSize = 20;
  final ScrollController _scrollController = ScrollController();
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  final List<SavedPost> _posts = [];
  bool _isLoading = true;
  bool _isLoadingMore = false;
  bool _hasMore = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _ownsApiClient = widget.apiClient == null;
    _apiClient = widget.apiClient ?? ApiClient();
    _scrollController.addListener(_onScroll);
    _loadPage(reset: true);
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
    if (_scrollController.hasClients &&
        _scrollController.position.extentAfter < 240 &&
        !_isLoading &&
        !_isLoadingMore &&
        _hasMore &&
        _error == null) {
      _loadPage();
    }
  }

  Future<void> _loadPage({bool reset = false}) async {
    if (reset) {
      setState(() {
        _isLoading = true;
        _error = null;
        _hasMore = true;
      });
    } else {
      setState(() {
        _isLoadingMore = true;
        _error = null;
      });
    }
    try {
      final response = await _apiClient.getSavedPosts(
        offset: reset ? 0 : _posts.length,
        limit: _pageSize,
      );
      if (!mounted) return;
      final page = response.data;
      if (page == null) {
        throw const ApiException(
          message: 'El servidor no devolvió tus guardados.',
        );
      }
      setState(() {
        if (reset) _posts.clear();
        _posts.addAll(page.items);
        _hasMore = page.hasMore;
        _isLoading = false;
        _isLoadingMore = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _error = error.message;
        _isLoading = false;
        _isLoadingMore = false;
      });
    }
  }

  Future<void> _removeSaved(SavedPost saved) async {
    try {
      await _apiClient.removeSavedPost(saved.id);
      if (!mounted) return;
      setState(() => _posts.removeWhere((post) => post.id == saved.id));
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Publicación quitada de Guardados.')),
      );
    } on ApiException catch (error) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(error.message)));
    }
  }

  void _removeSavedLocally(SavedPost saved) {
    setState(() => _posts.removeWhere((post) => post.id == saved.id));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF5F8F7),
      appBar: AppBar(
        backgroundColor: const Color(0xFF1ABC9C),
        foregroundColor: Colors.white,
        title: const Text('Guardados'),
        elevation: 0,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    if (_isLoading) {
      return const Center(
        child: CircularProgressIndicator(color: Color(0xFF1ABC9C)),
      );
    }
    if (_error != null && _posts.isEmpty) {
      return _MessageState(
        icon: Icons.cloud_off_outlined,
        message: _error!,
        actionLabel: 'Reintentar',
        onPressed: () => _loadPage(reset: true),
      );
    }
    if (_posts.isEmpty) {
      return const _MessageState(
        icon: Icons.bookmark_border_rounded,
        message: 'Las publicaciones que guardes aparecerán aquí.',
      );
    }
    return RefreshIndicator(
      color: const Color(0xFF1ABC9C),
      onRefresh: () => _loadPage(reset: true),
      child: ListView.separated(
        controller: _scrollController,
        padding: const EdgeInsets.all(16),
        itemCount: _posts.length + (_isLoadingMore || _error != null ? 1 : 0),
        separatorBuilder: (_, _) => const SizedBox(height: 12),
        itemBuilder: (context, index) {
          if (index == _posts.length) {
            if (_error != null) {
              return TextButton(
                onPressed: _loadPage,
                child: Text('No se pudieron cargar más. Reintentar: $_error'),
              );
            }
            return const Padding(
              padding: EdgeInsets.all(16),
              child: Center(
                child: CircularProgressIndicator(color: Color(0xFF1ABC9C)),
              ),
            );
          }
          final saved = _posts[index];
          return Card(
            color: Colors.white,
            elevation: 1,
            clipBehavior: Clip.antiAlias,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
              side: const BorderSide(color: Color(0xFFE7EEEC)),
            ),
            child: saved.isAvailable
                ? PostCard(
                    postData: saved.cardData,
                    tealAccent: const Color(0xFF1ABC9C),
                    apiClient: _apiClient,
                    onUnsave: () => _removeSavedLocally(saved),
                  )
                : _UnavailableSavedPost(onRemove: () => _removeSaved(saved)),
          );
        },
      ),
    );
  }
}

class _UnavailableSavedPost extends StatelessWidget {
  final VoidCallback onRemove;

  const _UnavailableSavedPost({required this.onRemove});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 18, vertical: 8),
      leading: const CircleAvatar(
        backgroundColor: Color(0xFFF0F2F1),
        child: Icon(Icons.hide_image_outlined, color: Colors.grey),
      ),
      title: const Text(
        'Ya no está disponible',
        style: TextStyle(fontWeight: FontWeight.w700),
      ),
      subtitle: const Text('Esta publicación fue eliminada.'),
      trailing: TextButton(onPressed: onRemove, child: const Text('Quitar')),
    );
  }
}

class _MessageState extends StatelessWidget {
  final IconData icon;
  final String message;
  final String? actionLabel;
  final VoidCallback? onPressed;

  const _MessageState({
    required this.icon,
    required this.message,
    this.actionLabel,
    this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 52, color: const Color(0xFF0D7771)),
            const SizedBox(height: 12),
            Text(message, textAlign: TextAlign.center),
            if (actionLabel != null)
              TextButton(onPressed: onPressed, child: Text(actionLabel!)),
          ],
        ),
      ),
    );
  }
}
