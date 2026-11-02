import 'dart:async';

import 'package:flutter/material.dart';

import '../../../core/navigation/artists_and_songs_screen.dart';
import '../../../features/profile/presentation/profile_screen.dart';
import '../../../screens/friends_screen.dart';
import '../../../widgets/desktop_app_header.dart';
import '../data/global_search_models.dart';
import '../data/global_search_repository.dart';

class GlobalSearchScreen extends StatefulWidget {
  const GlobalSearchScreen({super.key, GlobalSearchRepository? repository})
    : repository = repository ?? const DemoGlobalSearchRepository();

  final GlobalSearchRepository repository;

  @override
  State<GlobalSearchScreen> createState() => _GlobalSearchScreenState();
}

class _GlobalSearchScreenState extends State<GlobalSearchScreen> {
  static const _pageSize = 8;
  final _queryController = TextEditingController();
  final _scrollController = ScrollController();
  Timer? _debounce;

  GlobalSearchCategory _category = GlobalSearchCategory.users;
  List<GlobalSearchResult> _results = [];
  bool _isLoading = false;
  bool _isLoadingMore = false;
  bool _hasMore = false;
  String? _errorMessage;
  int _page = 1;
  int _requestId = 0;

  @override
  void initState() {
    super.initState();
    _queryController.addListener(_scheduleSearch);
    _scrollController.addListener(_loadMoreWhenNearEnd);
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _queryController
      ..removeListener(_scheduleSearch)
      ..dispose();
    _scrollController
      ..removeListener(_loadMoreWhenNearEnd)
      ..dispose();
    super.dispose();
  }

  void _scheduleSearch() {
    _debounce?.cancel();
    final query = _queryController.text.trim();
    if (query.isEmpty) {
      _requestId++;
      setState(() {
        _results = [];
        _isLoading = false;
        _isLoadingMore = false;
        _hasMore = false;
        _errorMessage = null;
      });
      return;
    }

    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    _debounce = Timer(
      const Duration(milliseconds: 250),
      () => _fetchResults(reset: true),
    );
  }

  Future<void> _fetchResults({required bool reset}) async {
    final query = _queryController.text.trim();
    if (query.isEmpty || (!reset && (!_hasMore || _isLoadingMore))) return;
    final requestId = ++_requestId;
    final requestedPage = reset ? 1 : _page + 1;

    setState(() {
      if (reset) {
        _isLoading = true;
        _isLoadingMore = false;
        _errorMessage = null;
      } else {
        _isLoadingMore = true;
      }
    });

    try {
      final resultPage = await widget.repository.search(
        query: query,
        category: _category,
        page: requestedPage,
        pageSize: _pageSize,
      );
      if (!mounted || requestId != _requestId) return;
      setState(() {
        _results = reset
            ? resultPage.results
            : [..._results, ...resultPage.results];
        _page = requestedPage;
        _hasMore = resultPage.hasMore;
        _isLoading = false;
        _isLoadingMore = false;
        _errorMessage = null;
      });
    } catch (error, stackTrace) {
      debugPrint('Error al buscar contenido: $error\n$stackTrace');
      if (!mounted || requestId != _requestId) return;
      setState(() {
        _isLoading = false;
        _isLoadingMore = false;
        _errorMessage = 'No se pudo completar la búsqueda. Inténtalo otra vez.';
      });
    }
  }

  void _loadMoreWhenNearEnd() {
    if (!_scrollController.hasClients ||
        _scrollController.position.extentAfter > 240) {
      return;
    }
    unawaited(_fetchResults(reset: false));
  }

  void _selectCategory(GlobalSearchCategory category) {
    if (_category == category) return;
    setState(() {
      _category = category;
      _results = [];
      _page = 1;
      _hasMore = false;
      _errorMessage = null;
    });
    _debounce?.cancel();
    if (_queryController.text.trim().isNotEmpty) {
      unawaited(_fetchResults(reset: true));
    }
  }

  void _openSection(WidgetBuilder builder) {
    Navigator.of(context).pushAndRemoveUntil(
      PageRouteBuilder<void>(
        pageBuilder: (context, animation, secondaryAnimation) =>
            builder(context),
        transitionDuration: Duration.zero,
        reverseTransitionDuration: Duration.zero,
      ),
      (route) => route.isFirst,
    );
  }

  @override
  Widget build(BuildContext context) {
    final isDesktop = MediaQuery.sizeOf(context).width > 900;
    return Scaffold(
      backgroundColor: Colors.transparent,
      body: Column(
        children: [
          if (isDesktop)
            DesktopAppHeader(
              activeSection: DesktopNavigationSection.none,
              searchOpen: true,
              searchController: _queryController,
              onSearchBack: Navigator.of(context).pop,
              onHomeTap: () => navigateToDesktopHome(context),
              onFriendsTap: () => _openSection((_) => const FriendsScreen()),
              onArtistsTap: () =>
                  _openSection((_) => const ArtistsAndSongsScreen()),
              onProfileTap: () => _openSection((_) => const ProfileScreen()),
            )
          else
            _buildMobileSearchHeader(),
          Expanded(
            child: Column(
              children: [
                _buildCategoryTabs(),
                Expanded(child: _buildResults()),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMobileSearchHeader() {
    return Container(
      height: 55,
      color: const Color(0xFF1ABC9C),
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: Row(
        children: [
          IconButton(
            tooltip: 'Volver',
            onPressed: Navigator.of(context).pop,
            icon: const Icon(Icons.arrow_back, color: Colors.white),
          ),
          Expanded(
            child: TextField(
              controller: _queryController,
              autofocus: true,
              textInputAction: TextInputAction.search,
              style: const TextStyle(color: Colors.white),
              decoration: const InputDecoration(
                hintText: 'Buscar en Meloop',
                hintStyle: TextStyle(color: Colors.white70),
                border: InputBorder.none,
              ),
            ),
          ),
          if (_queryController.text.isNotEmpty)
            IconButton(
              tooltip: 'Borrar búsqueda',
              onPressed: _queryController.clear,
              icon: const Icon(Icons.close, color: Colors.white),
            ),
        ],
      ),
    );
  }

  Widget _buildCategoryTabs() {
    const categories = GlobalSearchCategory.values;
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
      child: Wrap(
        spacing: 8,
        children: categories.map((category) {
          final isSelected = category == _category;
          return ChoiceChip(
            label: Text(_categoryLabel(category)),
            selected: isSelected,
            onSelected: (_) => _selectCategory(category),
            selectedColor: const Color(0xFF00AF92),
            backgroundColor: Colors.white.withValues(alpha: 0.9),
            labelStyle: TextStyle(
              color: isSelected ? Colors.white : const Color(0xFF174B4B),
              fontWeight: FontWeight.w600,
            ),
            side: BorderSide.none,
          );
        }).toList(),
      ),
    );
  }

  Widget _buildResults() {
    final query = _queryController.text.trim();
    if (query.isEmpty) {
      return const Center(
        child: Text(
          'Busca personas, publicaciones o música.',
          style: TextStyle(color: Colors.white, fontSize: 16),
          textAlign: TextAlign.center,
        ),
      );
    }
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }
    if (_errorMessage != null) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              _errorMessage!,
              style: const TextStyle(color: Colors.white),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            OutlinedButton(
              onPressed: () => _fetchResults(reset: true),
              child: const Text('Reintentar'),
            ),
          ],
        ),
      );
    }
    if (_results.isEmpty) {
      return const Center(
        child: Text(
          'No encontramos resultados para esta búsqueda.',
          style: TextStyle(color: Colors.white),
          textAlign: TextAlign.center,
        ),
      );
    }

    return Center(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 760),
        child: ListView.separated(
          controller: _scrollController,
          padding: const EdgeInsets.fromLTRB(16, 8, 16, 24),
          itemCount: _results.length + (_isLoadingMore ? 1 : 0),
          separatorBuilder: (_, _) => const SizedBox(height: 10),
          itemBuilder: (context, index) {
            if (index == _results.length) {
              return const Padding(
                padding: EdgeInsets.all(16),
                child: Center(child: CircularProgressIndicator()),
              );
            }
            return _SearchResultCard(result: _results[index]);
          },
        ),
      ),
    );
  }

  String _categoryLabel(GlobalSearchCategory category) => switch (category) {
    GlobalSearchCategory.users => 'Usuarios',
    GlobalSearchCategory.posts => 'Publicaciones',
    GlobalSearchCategory.music => 'Música',
  };
}

class _SearchResultCard extends StatelessWidget {
  const _SearchResultCard({required this.result});

  final GlobalSearchResult result;

  @override
  Widget build(BuildContext context) {
    final icon = switch (result.category) {
      GlobalSearchCategory.users => Icons.person_outline,
      GlobalSearchCategory.posts => Icons.article_outlined,
      GlobalSearchCategory.music => Icons.music_note,
    };

    return Card(
      margin: EdgeInsets.zero,
      color: Colors.white.withValues(alpha: 0.96),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: const Color(0xFF00AF92).withValues(alpha: 0.14),
          child: Icon(icon, color: const Color(0xFF008F79)),
        ),
        title: Text(
          result.title,
          style: const TextStyle(fontWeight: FontWeight.w600),
        ),
        subtitle: Padding(
          padding: const EdgeInsets.only(top: 3),
          child: Text('${result.subtitle}\n${result.description}'),
        ),
        isThreeLine: true,
      ),
    );
  }
}
