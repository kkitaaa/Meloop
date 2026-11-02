import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/music/presentation/music_track_detail_screen.dart';
import 'package:flutter_app/features/music/presentation/music_track_tile.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';

enum MusicSearchMode { pickForPost, pickForPreference, browse }

class MusicSearchScreen extends StatefulWidget {
  final MusicSearchMode mode;
  final ApiClient? apiClient;
  final String initialType;

  const MusicSearchScreen({
    super.key,
    required this.mode,
    this.apiClient,
    this.initialType = 'track',
  });

  @override
  State<MusicSearchScreen> createState() => _MusicSearchScreenState();
}

class _MusicSearchScreenState extends State<MusicSearchScreen> {
  static const _searchDelay = Duration(milliseconds: 400);
  static const _minimumQueryLength = 2;
  static const _types = <_MusicSearchType>[
    _MusicSearchType('Canciones', 'track', Icons.music_note_rounded),
    _MusicSearchType('Artistas', 'artist', Icons.people_alt_outlined),
    _MusicSearchType('Álbumes', 'album', Icons.album_rounded),
    _MusicSearchType('Géneros', 'genre', Icons.category_outlined),
  ];

  List<_MusicSearchType> get _availableTypes =>
      widget.mode == MusicSearchMode.pickForPreference
      ? _types
            .where(
              (type) => type.apiType == 'track' || type.apiType == 'artist',
            )
            .toList()
      : _types;

  final TextEditingController _queryController = TextEditingController();
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  Timer? _debounce;
  CancelToken? _activeRequest;
  int _requestId = 0;
  int _selectedTypeIndex = 0;
  List<MusicTrack> _items = const [];
  int _nextOffset = 0;
  bool _isLoading = false;
  bool _isLoadingMore = false;
  bool _hasMore = false;
  bool _hasSearched = false;
  String? _errorMessage;

  _MusicSearchType get _selectedType => _availableTypes[_selectedTypeIndex];

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ApiClient();
    _ownsApiClient = widget.apiClient == null;
    final initialIndex = _availableTypes.indexWhere(
      (type) => type.apiType == widget.initialType,
    );
    _selectedTypeIndex = initialIndex < 0 ? 0 : initialIndex;
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _activeRequest?.cancel('La pantalla de búsqueda se cerró.');
    _queryController.dispose();
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  void _onQueryChanged(String query) {
    _debounce?.cancel();
    _invalidateRequest();
    setState(() {
      _items = const [];
      _nextOffset = 0;
      _isLoading = false;
      _isLoadingMore = false;
      _hasMore = false;
      _hasSearched = false;
      _errorMessage = null;
    });
    if (query.trim().length < _minimumQueryLength) return;
    _scheduleSearch();
  }

  void _onTypeSelected(int index) {
    if (index == _selectedTypeIndex) return;
    _debounce?.cancel();
    _invalidateRequest();
    setState(() {
      _selectedTypeIndex = index;
      _items = const [];
      _nextOffset = 0;
      _isLoading = false;
      _isLoadingMore = false;
      _hasMore = false;
      _hasSearched = false;
      _errorMessage = null;
    });
    if (_queryController.text.trim().length >= _minimumQueryLength) {
      _scheduleSearch();
    }
  }

  void _scheduleSearch() {
    _debounce = Timer(_searchDelay, () => _loadResults(reset: true));
  }

  void _invalidateRequest() {
    _requestId++;
    _activeRequest?.cancel('La búsqueda cambió.');
    _activeRequest = null;
  }

  Future<void> _loadResults({required bool reset}) async {
    final query = _queryController.text.trim();
    if (query.length < _minimumQueryLength ||
        _isLoading ||
        _isLoadingMore ||
        (!reset && !_hasMore)) {
      return;
    }

    _activeRequest?.cancel('Se inició una búsqueda más reciente.');
    final token = CancelToken();
    _activeRequest = token;
    final requestId = ++_requestId;
    final requestedOffset = reset ? 0 : _nextOffset;
    final selectedType = _selectedType;
    setState(() {
      _isLoading = reset;
      _isLoadingMore = !reset;
      _hasSearched = true;
      _errorMessage = null;
      if (reset) {
        _items = const [];
        _nextOffset = 0;
        _hasMore = false;
      }
    });

    try {
      final response = await _apiClient.searchMusicTracks(
        query,
        type: selectedType.apiType,
        offset: requestedOffset,
        cancelToken: token,
      );
      final page = response.data;
      if (page == null) {
        throw const ApiException(
          message: 'El servicio de música no devolvió resultados válidos.',
        );
      }
      if (!mounted || token.isCancelled || requestId != _requestId) return;

      final existingIds = _items.map((item) => item.id).toSet();
      final uniqueItems = page.items
          .where((item) => existingIds.add(item.id))
          .toList();
      final pageOffset = page.offset < requestedOffset
          ? requestedOffset
          : page.offset;
      setState(() {
        _items = [..._items, ...uniqueItems];
        _nextOffset = pageOffset + page.items.length;
        _hasMore = page.hasMore;
        _isLoading = false;
        _isLoadingMore = false;
        _errorMessage = null;
      });
    } on DioException catch (error) {
      if (CancelToken.isCancel(error) || !mounted || requestId != _requestId) {
        return;
      }
      _setError(
        error.message ?? 'No pudimos conectar con Spotify. Intenta nuevamente en unos minutos.',
      );
    } on ApiException catch (error) {
      if (!mounted || requestId != _requestId) return;
      _setError(
        error.errorCode == 'MUSIC_UNAVAILABLE'
            ? 'No pudimos conectar con Spotify. Intenta nuevamente en unos minutos.'
            : error.message,
      );
    } finally {
      if (mounted && requestId == _requestId) {
        setState(() {
          _isLoading = false;
          _isLoadingMore = false;
        });
      }
    }
  }

  void _setError(String message) {
    setState(() {
      _isLoading = false;
      _isLoadingMore = false;
      _errorMessage = message;
    });
  }

  String get _appBarTitle => switch (widget.mode) {
    MusicSearchMode.pickForPost => 'Elegir canción',
    MusicSearchMode.pickForPreference => 'Agregar preferencia',
    MusicSearchMode.browse => 'Explorar música',
  };

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
            Text(
              _appBarTitle,
              style: const TextStyle(
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
        onChanged: _onQueryChanged,
        decoration: InputDecoration(
          hintText: 'Canciones, artistas, álbumes o géneros…',
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
                  onPressed: () {
                    _queryController.clear();
                    _onQueryChanged('');
                  },
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
        isScrollable: true,
        tabAlignment: TabAlignment.start,
        onTap: _onTypeSelected,
        dividerColor: Colors.transparent,
        indicatorSize: TabBarIndicatorSize.tab,
        indicator: BoxDecoration(
          color: const Color(0xFF1ABC9C),
          borderRadius: BorderRadius.circular(11),
        ),
        labelColor: Colors.white,
        unselectedLabelColor: const Color(0xFF667C78),
        labelStyle: const TextStyle(fontSize: 12, fontWeight: FontWeight.w700),
        tabs: [
          for (final type in _availableTypes)
            Tab(icon: Icon(type.icon, size: 18), text: type.label),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: _availableTypes.length,
      initialIndex: _selectedTypeIndex,
      child: Scaffold(
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
                    if (_isLoading)
                      const LinearProgressIndicator(color: Color(0xFF1ABC9C)),
                    Expanded(child: _buildResults(context)),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildResults(BuildContext context) {
    if (_isLoading && _items.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (!_hasSearched) {
      final hasShortQuery = _queryController.text.trim().isNotEmpty;
      return _MessageState(
        icon: Icons.search_rounded,
        title: hasShortQuery ? 'Escribe un poco más' : '¿Qué quieres escuchar?',
        message: hasShortQuery
            ? 'Escribe al menos 2 caracteres para buscar.'
            : 'Busca canciones, artistas, álbumes y géneros en Meloop.',
      );
    }

    if (_errorMessage != null && _items.isEmpty) {
      return _MessageState(
        icon: Icons.cloud_off_rounded,
        title: 'No pudimos conectar',
        message: _errorMessage!,
        actionLabel: 'Reintentar',
        onAction: () => _loadResults(reset: true),
        isError: true,
      );
    }

    if (!_isLoading && _items.isEmpty) {
      return _MessageState(
        icon: Icons.music_off_rounded,
        title: 'Sin resultados',
        message: 'Sin resultados para «${_queryController.text.trim()}»',
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.only(top: 8, bottom: 20),
      itemCount: _items.length + (_hasMore || _errorMessage != null ? 1 : 0),
      itemBuilder: (context, index) {
        if (index == _items.length) {
          if (_errorMessage != null) {
            return _MessageState(
              icon: Icons.cloud_off_rounded,
              title: 'No pudimos cargar más resultados',
              message: _errorMessage!,
              actionLabel: 'Reintentar',
              onAction: () => _loadResults(reset: false),
              isError: true,
            );
          }
          return Padding(
            padding: const EdgeInsets.symmetric(vertical: 8),
            child: Center(
              child: _isLoadingMore
                  ? const CircularProgressIndicator()
                  : OutlinedButton(
                      onPressed: () => _loadResults(reset: false),
                      child: const Text('Cargar más'),
                    ),
            ),
          );
        }

        final item = _items[index];
        final isSelectionMode = widget.mode != MusicSearchMode.browse;
        return MusicTrackTile(
          track: item,
          trailing: isSelectionMode
              ? const Icon(Icons.add_circle_outline_rounded)
              : null,
          onTap: () {
            if (isSelectionMode) {
              Navigator.pop(context, item);
              return;
            }
            Navigator.push(
              context,
              MaterialPageRoute<void>(
                builder: (_) => MusicTrackDetailScreen(track: item),
              ),
            );
          },
        );
      },
    );
  }
}

class _MusicSearchType {
  final String label;
  final String apiType;
  final IconData icon;

  const _MusicSearchType(this.label, this.apiType, this.icon);
}

class _MessageState extends StatelessWidget {
  final IconData icon;
  final String? title;
  final String message;
  final String? actionLabel;
  final VoidCallback? onAction;
  final bool isError;

  const _MessageState({
    required this.icon,
    required this.message,
    this.title,
    this.actionLabel,
    this.onAction,
    this.isError = false,
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
                child: Icon(
                  icon,
                  color: isError
                      ? const Color(0xFFB65353)
                      : const Color(0xFF0D7771),
                  size: 30,
                ),
              ),
              const SizedBox(height: 12),
              if (title != null) ...[
                Text(
                  title!,
                  textAlign: TextAlign.center,
                  style: const TextStyle(
                    color: Color(0xFF183D3B),
                    fontSize: 16,
                    fontWeight: FontWeight.w700,
                  ),
                ),
                const SizedBox(height: 5),
              ],
              Text(
                message,
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: Color(0xFF788A86),
                  fontSize: 13,
                  height: 1.45,
                ),
              ),
              if (actionLabel != null && onAction != null) ...[
                const SizedBox(height: 12),
                OutlinedButton.icon(
                  onPressed: onAction,
                  icon: const Icon(Icons.refresh_rounded, size: 18),
                  label: Text(actionLabel!),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: const Color(0xFF0D5C5E),
                    side: const BorderSide(color: Color(0xFFBDE9E1)),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
