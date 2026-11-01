import 'dart:async';

import 'package:flutter/material.dart';

import '../services/api_client.dart';
import 'friends_screen.dart';
import 'api_test_screen.dart';
import '../../features/profile/presentation/profile_screen.dart';
import '../features/posts/presentation/post_editor_dialog.dart';
import '../../widgets/post_card.dart';
import '../../widgets/rumi_xp_notice.dart';
import '../../widgets/states/loading_state.dart';
import '../../widgets/states/empty_state.dart';
import '../../widgets/states/error_state.dart';
import '../features/profile/presentation/blocked_users_screen.dart';

class FeedScreen extends StatefulWidget {
  final ApiClient? apiClient;

  const FeedScreen({super.key, this.apiClient});

  @override
  State<FeedScreen> createState() => _FeedScreenState();
}

class _FeedScreenState extends State<FeedScreen> {
  static const int _pageSize = 5;

  final Color _bgColor = const Color(0xFF0D5C5E);
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
    _scrollController.addListener(() {
      if (_scrollController.hasClients &&
          _scrollController.position.pixels >=
              _scrollController.position.maxScrollExtent - 200) {
        if (!_isInitialLoading && !_isLoadingMore && _hasMore && !_hasError) {
          _fetchMorePosts();
        }
      }
    });
  }

  @override
  void dispose() {
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
        "user": post['user'] ?? post['username'] ?? "Tú",
        "user_id": post['user_id'],
        "time": "Hace un momento",
        "content": post['content'] ?? result.content,
        "image_bytes": result.imageBytes,
        "image_name": result.imageName,
        "image_url": post['image_url'] ?? firstServerImageUrl,
        "images": serverImages ?? result.images,
        "song": post['song_title'] ?? result.songTitle,
        "artist": post['artist_name'] ?? result.artistName,
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

  @override
  Widget build(BuildContext context) {
    final double screenWidth = MediaQuery.of(context).size.width;
    final bool isDesktop = screenWidth > 900;

    return Scaffold(
      backgroundColor: _bgColor,
      body: Column(
        children: [
          _buildTopBar(isDesktop),
          Expanded(
            child: isDesktop ? _buildDesktopLayout() : _buildMobileLayout(),
          ),
        ],
      ),
    );
  }

  Widget _buildTopBar(bool isDesktop) {
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
          if (isDesktop) ...[
            Row(
              children: [
                _navLink('Inicio', isActive: true),
                _navLink(
                  'Amigos',
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const FriendsScreen(),
                      ),
                    );
                  },
                ),
                _navLink(
                  'Perfil',
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => const ProfileScreen(),
                      ),
                    );
                  },
                ),
              ],
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
              decoration: BoxDecoration(
                color: Colors.white.withValues(alpha: 0.2),
                borderRadius: BorderRadius.circular(20),
              ),
              child: Row(
                children: const [
                  Icon(Icons.play_circle_fill, color: Colors.white, size: 18),
                  SizedBox(width: 8),
                  Text(
                    "Playing... DJ Gouz",
                    style: TextStyle(color: Colors.white, fontSize: 12),
                  ),
                  SizedBox(width: 8),
                  Icon(Icons.graphic_eq, color: Colors.white, size: 16),
                ],
              ),
            ),
          ],
          Row(
            children: [
              IconButton(
                tooltip: 'Probar conexión con API',
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
                icon: const Icon(
                  Icons.cloud_outlined,
                  color: Colors.white,
                  size: 20,
                ),
                onPressed: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const ApiTestScreen(),
                    ),
                  );
                },
              ),
              const SizedBox(width: 16),
              // Botón de Ajustes para ir a la pantalla de bloqueados
              IconButton(
                padding: EdgeInsets.zero,
                constraints: const BoxConstraints(),
                icon: const Icon(Icons.settings, color: Colors.white, size: 20),
                onPressed: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const BlockedUsersScreen(),
                    ),
                  );
                },
              ),
              const SizedBox(width: 16),
              const Icon(Icons.notifications, color: Colors.white, size: 20),
              const SizedBox(width: 16),
              InkWell(
                onTap: () {
                  Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const ProfileScreen(),
                    ),
                  );
                },
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

  Widget _navLink(String text, {bool isActive = false, VoidCallback? onTap}) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(16),
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 4.0),
        padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 6.0),
        decoration: BoxDecoration(
          color: isActive ? Colors.white : Colors.transparent,
          borderRadius: BorderRadius.circular(16),
        ),
        child: Text(
          text,
          style: TextStyle(
            color: isActive ? _tealAccent : Colors.white,
            fontWeight: isActive ? FontWeight.bold : FontWeight.normal,
            fontSize: 12,
          ),
        ),
      ),
    );
  }

  Widget _buildDesktopLayout() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          SizedBox(
            width: 220,
            child: SingleChildScrollView(
              child: Column(
                children: [
                  _buildPolaroidProfile(),
                  const SizedBox(height: 16),
                  _buildListeningTo(),
                  const SizedBox(height: 16),
                  _buildLevelCard(),
                  const SizedBox(height: 16),
                  _buildSuggestions(),
                ],
              ),
            ),
          ),
          const SizedBox(width: 24),
          ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 650),
            child: _buildFeedPaper(),
          ),
        ],
      ),
    );
  }

  Widget _buildMobileLayout() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        children: [
          _buildPolaroidProfile(),
          const SizedBox(height: 16),
          _buildListeningTo(),
          const SizedBox(height: 24),
          SizedBox(height: 700, child: _buildFeedPaper()),
        ],
      ),
    );
  }

  Widget _buildPolaroidProfile() {
    return Stack(
      alignment: Alignment.topCenter,
      clipBehavior: Clip.none,
      children: [
        Transform.rotate(
          angle: -0.02,
          child: Container(
            margin: const EdgeInsets.only(top: 10),
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: Colors.white,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.1),
                  blurRadius: 8,
                ),
              ],
            ),
            child: Column(
              children: [
                Container(
                  height: 140,
                  width: double.infinity,
                  color: Colors.grey[200],
                  child: const Icon(Icons.pets, size: 50, color: Colors.grey),
                ),
                const SizedBox(height: 16),
                const Text(
                  "Nombre de usuario",
                  style: TextStyle(
                    fontFamily: 'Comic Sans MS',
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
                const SizedBox(height: 4),
                const Text(
                  "Me siento Feliz",
                  style: TextStyle(fontSize: 11, color: Colors.black54),
                ),
              ],
            ),
          ),
        ),
        Positioned(
          top: 0,
          child: Transform.rotate(
            angle: -0.08,
            child: Container(
              width: 50,
              height: 18,
              color: Colors.white.withValues(alpha: 0.7),
            ),
          ),
        ),
        Positioned(
          top: 8,
          child: CircleAvatar(radius: 5, backgroundColor: Colors.red[700]),
        ),
      ],
    );
  }

  Widget _buildListeningTo() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              color: Colors.purple[800],
              borderRadius: BorderRadius.circular(4),
            ),
            child: const Icon(Icons.music_note, color: Colors.white),
          ),
          const SizedBox(width: 12),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: const [
              Text(
                "Escuchando:",
                style: TextStyle(color: Colors.black54, fontSize: 10),
              ),
              Text(
                "Finding Urself",
                style: TextStyle(
                  color: Colors.black,
                  fontWeight: FontWeight.bold,
                  fontSize: 12,
                ),
              ),
              Text(
                "DJ GOUZ",
                style: TextStyle(color: Colors.black54, fontSize: 10),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildLevelCard() {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        children: [
          Row(
            children: [
              Icon(Icons.star, color: _tealAccent, size: 14),
              const SizedBox(width: 4),
              const Text(
                "Nivel 3 -> Buen progreso",
                style: TextStyle(
                  color: Colors.black87,
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          LinearProgressIndicator(
            value: 0.6,
            backgroundColor: Colors.grey[200],
            valueColor: AlwaysStoppedAnimation<Color>(_tealAccent),
            minHeight: 6,
            borderRadius: BorderRadius.circular(3),
          ),
        ],
      ),
    );
  }

  Widget _buildSuggestions() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            "Artistas que te podrían gustar",
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.bold,
              color: Colors.black87,
            ),
          ),
          const Divider(),
          _artistListTile("Kidd Voodoo"),
          _artistListTile("Tiago PZK"),
          _artistListTile("Cris MJ"),
        ],
      ),
    );
  }

  Widget _artistListTile(String name) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 10.0),
      child: Row(
        children: [
          const CircleAvatar(
            radius: 10,
            backgroundColor: Colors.black12,
            child: Icon(Icons.person, size: 12, color: Colors.black54),
          ),
          const SizedBox(width: 8),
          Text(
            name,
            style: const TextStyle(fontSize: 11, color: Colors.black87),
          ),
        ],
      ),
    );
  }

  Widget _buildFeedPaper() {
    return Container(
      decoration: BoxDecoration(
        color: const Color(0xFFFDFDFD),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.15),
            blurRadius: 20,
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32.0, vertical: 24.0),
        child: Column(
          children: [
            Wrap(
              alignment: WrapAlignment.spaceBetween,
              runSpacing: 12,
              crossAxisAlignment: WrapCrossAlignment.center,
              children: [
                Wrap(
                  crossAxisAlignment: WrapCrossAlignment.center,
                  children: [
                    const Text(
                      "Ultimos Blogs:",
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(width: 16),
                    _filterChip("recientes", category: 'recent'),
                    _filterChip("amigos", category: 'friends'),
                    _filterChip("sugeridos", category: 'suggested'),
                  ],
                ),
                OutlinedButton.icon(
                  style: OutlinedButton.styleFrom(
                    side: BorderSide(color: _tealAccent),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(20),
                    ),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 12,
                    ),
                  ),
                  onPressed: _createPost,
                  icon: Icon(Icons.edit, color: _tealAccent, size: 14),
                  label: Text(
                    "escribir blog",
                    style: TextStyle(
                      color: _tealAccent,
                      fontSize: 12,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            const Divider(height: 32, thickness: 1, color: Colors.black12),

            // --- IMPLEMENTACIÓN DE LOS ESTADOS DE CARGA, ERROR Y VACÍO ---
            Expanded(
              child: _isInitialLoading
                  ? const LoadingState() // Estado animado (Skeletons)
                  : _hasError
                  ? ErrorState(
                      errorCode: _errorCode,
                      onRetry:
                          _fetchInitialPosts, // Botón para volver a intentar
                      tealAccent: _tealAccent,
                      message: _feedError,
                    )
                  : _posts.isEmpty
                  ? const EmptyState(
                      title: "Feed vacío",
                      message: "Aún no hay publicaciones recientes. ¡Sé el primero en crear un blog!",
                      icon: Icons.post_add,
                    )
                  : ListView.builder(
                      controller: _scrollController,
                      padding: const EdgeInsets.only(right: 16),
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
                        // Uso de la tarjeta reutilizable
                        return PostCard(
                          postData: _posts[index],
                          tealAccent: _tealAccent,
                          onDelete: () {
                            setState(() {
                              _posts.removeAt(index);
                            });
                          },
                        );
                      },
                    ),
            ),

            if (!_isInitialLoading &&
                !_hasError &&
                _posts.isNotEmpty &&
                _hasMore)
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: _tealAccent,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20),
                  ),
                  elevation: 0,
                ),
                onPressed: _isLoadingMore ? null : _fetchMorePosts,
                child: _isLoadingMore
                    ? const SizedBox(
                        width: 16,
                        height: 16,
                        child: CircularProgressIndicator(
                          color: Colors.white,
                          strokeWidth: 2,
                        ),
                      )
                    : const Text(
                        "Cargar más blogs",
                        style: TextStyle(color: Colors.white, fontSize: 12),
                      ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _filterChip(String label, {required String category}) {
    final isActive = _selectedCategory == category;
    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(12),
          onTap: isActive || _isInitialLoading
              ? null
              : () => _fetchInitialPosts(category: category),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            decoration: BoxDecoration(
              color: isActive ? Colors.white : Colors.grey[200],
              border: isActive
                  ? Border.all(color: _tealAccent, width: 1)
                  : Border.all(color: Colors.transparent),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Text(
              label,
              style: TextStyle(
                color: isActive ? _tealAccent : Colors.grey[600],
                fontSize: 10,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ),
      ),
    );
  }
}
