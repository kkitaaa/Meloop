import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_app/features/music/presentation/music_search_screen.dart';
import 'package:flutter_app/models/music_preferences.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';

class MusicPreferencesScreen extends StatefulWidget {
  final ApiClient? apiClient;

  const MusicPreferencesScreen({super.key, this.apiClient});

  @override
  State<MusicPreferencesScreen> createState() => _MusicPreferencesScreenState();
}

class _MusicPreferencesScreenState extends State<MusicPreferencesScreen> {
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  MusicPreferencesData? _preferences;
  bool _isLoading = true;
  String? _errorMessage;
  final Set<String> _pendingChanges = {};

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ApiClient();
    _ownsApiClient = widget.apiClient == null;
    _loadPreferences();
  }

  @override
  void dispose() {
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  Future<void> _loadPreferences() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final response = await _apiClient.getMusicPreferences();
      final preferences = response.data;
      if (preferences == null) {
        throw const ApiException(
          message: 'El servicio no devolvió tus preferencias musicales.',
        );
      }
      if (!mounted) return;
      setState(() {
        _preferences = preferences;
        _isLoading = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _errorMessage = error.message;
        _isLoading = false;
      });
    }
  }

  Future<void> _openMusicSearch(String type) async {
    final selected = await Navigator.of(context).push<MusicTrack>(
      MaterialPageRoute(
        builder: (_) => MusicSearchScreen(
          mode: MusicSearchMode.pickForPreference,
          initialType: type,
          apiClient: _apiClient,
        ),
      ),
    );
    if (selected == null || !mounted) return;
    await _addPreference(MusicPreference.fromTrack(selected, type: type));
  }

  Future<void> _openGenrePicker() async {
    final selected = await showDialog<MusicTrack>(
      context: context,
      builder: (_) => _GenreAutocompleteDialog(apiClient: _apiClient),
    );
    if (selected == null || !mounted) return;
    await _addPreference(MusicPreference.fromTrack(selected, type: 'genre'));
  }

  Future<void> _addPreference(MusicPreference preference) async {
    final key = '${preference.type}:${preference.id}';
    if (_pendingChanges.contains(key)) return;
    setState(() => _pendingChanges.add(key));
    try {
      final response = await _apiClient.addMusicPreference(
        type: preference.type,
        itemId: preference.id,
        name: preference.name,
        artworkUrl: preference.artworkUrl,
      );
      if (!mounted) return;
      final updated = response.data;
      setState(() {
        if (updated != null) {
          _preferences = updated;
        } else {
          _preferences = _withPreference(preference, add: true);
        }
        _pendingChanges.remove(key);
      });
      _showMessage('Preferencia agregada.');
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _pendingChanges.remove(key));
      if (error.errorCode == 'ALREADY_EXISTS') {
        _showMessage('Ya está en tus preferencias');
      } else if (error.errorCode == 'PREFERENCE_LIMIT') {
        _showMessage('Llegaste al límite de 30');
        await _loadPreferences();
      } else {
        _showMessage(error.message);
      }
    }
  }

  Future<void> _removePreference(MusicPreference preference) async {
    final shouldRemove = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Quitar preferencia'),
        content: Text(
          '¿Quieres quitar «${preference.name}» de tus preferencias?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Quitar'),
          ),
        ],
      ),
    );
    if (shouldRemove != true || !mounted) return;

    final key = '${preference.type}:${preference.id}';
    setState(() => _pendingChanges.add(key));
    try {
      final response = await _apiClient.removeMusicPreference(
        type: preference.type,
        itemId: preference.id,
      );
      if (!mounted) return;
      setState(() {
        if (response.data != null) {
          _preferences = response.data;
        } else {
          _preferences = _withPreference(preference, add: false);
        }
        _pendingChanges.remove(key);
      });
      _showMessage('Preferencia quitada.');
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _pendingChanges.remove(key));
      _showMessage(error.message);
    }
  }

  MusicPreferencesData _withPreference(
    MusicPreference preference, {
    required bool add,
  }) {
    final current = _preferences;
    if (current == null) return _emptyPreferences();

    List<MusicPreference> update(List<MusicPreference> items) {
      if (add) {
        if (items.any((item) => item.id == preference.id)) return items;
        return [...items, preference];
      }
      return items.where((item) => item.id != preference.id).toList();
    }

    return MusicPreferencesData(
      songs: preference.type == 'track' ? update(current.songs) : current.songs,
      artists: preference.type == 'artist'
          ? update(current.artists)
          : current.artists,
      genres: preference.type == 'genre'
          ? update(current.genres)
          : current.genres,
      limits: current.limits,
    );
  }

  MusicPreferencesData _emptyPreferences() {
    return const MusicPreferencesData(
      songs: [],
      artists: [],
      genres: [],
      limits: {'track': 30, 'artist': 30, 'genre': 30},
    );
  }

  int _limitFor(String type) => _preferences?.limits[type] ?? 30;

  List<MusicPreference> _itemsFor(String type) {
    final preferences = _preferences;
    if (preferences == null) return const [];
    return switch (type) {
      'track' => preferences.songs,
      'artist' => preferences.artists,
      'genre' => preferences.genres,
      _ => const [],
    };
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
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
              'Preferencias musicales',
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0D5C5E),
      appBar: _buildBrandedAppBar(),
      body: Center(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: ConstrainedBox(
            constraints: const BoxConstraints(maxWidth: 920),
            child: Container(
              color: const Color(0xFFF5F8F7),
              child: _isLoading && _preferences == null
                  ? const Center(
                      child: CircularProgressIndicator(
                        color: Color(0xFF1ABC9C),
                      ),
                    )
                  : _errorMessage != null && _preferences == null
                  ? _buildLoadError()
                  : RefreshIndicator(
                      onRefresh: _loadPreferences,
                      color: const Color(0xFF0D7771),
                      child: ListView(
                        physics: const AlwaysScrollableScrollPhysics(),
                        padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
                        children: [
                          _buildIntro(),
                          if (_errorMessage != null) ...[
                            const SizedBox(height: 12),
                            _buildRefreshError(),
                          ],
                          const SizedBox(height: 16),
                          _buildSection(
                            title: 'Canciones',
                            type: 'track',
                            icon: Icons.music_note_rounded,
                          ),
                          _buildSection(
                            title: 'Artistas',
                            type: 'artist',
                            icon: Icons.people_alt_outlined,
                          ),
                          _buildSection(
                            title: 'Géneros',
                            type: 'genre',
                            icon: Icons.category_outlined,
                          ),
                        ],
                      ),
                    ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildIntro() {
    return Container(
      padding: const EdgeInsets.all(18),
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFFE8F7F3), Color(0xFFF7FBFA)],
        ),
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: const Color(0xFFDCEEE9)),
      ),
      child: const Row(
        children: [
          Icon(Icons.auto_awesome_rounded, color: Color(0xFF0D7771), size: 24),
          SizedBox(width: 12),
          Expanded(
            child: Text(
              'Estas preferencias se usan para sugerirte música y personas afines.',
              style: TextStyle(
                color: Color(0xFF315550),
                fontSize: 13,
                height: 1.4,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRefreshError() {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: const Color(0xFFFFF5F2),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFF2D7CE)),
      ),
      child: Row(
        children: [
          const Icon(Icons.cloud_off_rounded, color: Color(0xFFB65353)),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              _errorMessage!,
              style: const TextStyle(color: Color(0xFF714B43), fontSize: 13),
            ),
          ),
          IconButton(
            tooltip: 'Reintentar',
            onPressed: _loadPreferences,
            icon: const Icon(Icons.refresh_rounded),
          ),
        ],
      ),
    );
  }

  Widget _buildSection({
    required String title,
    required String type,
    required IconData icon,
  }) {
    final items = _itemsFor(type);
    final limit = _limitFor(type);
    final atLimit = items.length >= limit;
    final isGenre = type == 'genre';
    return Card(
      margin: const EdgeInsets.only(bottom: 14),
      elevation: 0,
      color: Colors.white,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(18),
        side: const BorderSide(color: Color(0xFFE4ECEA)),
      ),
      child: Padding(
        padding: const EdgeInsets.fromLTRB(16, 14, 16, 10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  height: 38,
                  width: 38,
                  decoration: BoxDecoration(
                    color: const Color(0xFFE8F7F3),
                    borderRadius: BorderRadius.circular(13),
                  ),
                  child: Icon(icon, color: const Color(0xFF0D7771), size: 21),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Text(
                    title,
                    style: const TextStyle(
                      color: Color(0xFF183D3B),
                      fontSize: 16,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                ),
                Text(
                  '${items.length} / $limit',
                  style: const TextStyle(
                    color: Color(0xFF70827F),
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (items.isEmpty)
              Padding(
                padding: const EdgeInsets.fromLTRB(2, 8, 2, 12),
                child: Text(
                  'Aún no agregas ${_lowercaseTitle(title)}. Añade tu primera preferencia.',
                  style: const TextStyle(
                    color: Color(0xFF788A86),
                    fontSize: 13,
                  ),
                ),
              )
            else
              for (final item in items) _buildPreferenceTile(item),
            Align(
              alignment: Alignment.centerRight,
              child: OutlinedButton.icon(
                onPressed: atLimit || _isLoading
                    ? null
                    : () =>
                          isGenre ? _openGenrePicker() : _openMusicSearch(type),
                icon: const Icon(Icons.add_rounded, size: 18),
                label: Text(atLimit ? 'Límite alcanzado' : 'Agregar'),
                style: OutlinedButton.styleFrom(
                  foregroundColor: const Color(0xFF0D5C5E),
                  backgroundColor: const Color(0xFFE9F8F5),
                  disabledForegroundColor: const Color(0xFF889894),
                  side: const BorderSide(color: Color(0xFFBDE9E1)),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPreferenceTile(MusicPreference preference) {
    final key = '${preference.type}:${preference.id}';
    final isPending = _pendingChanges.contains(key);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        children: [
          if (preference.artworkUrl != null && preference.type != 'genre')
            ClipRRect(
              borderRadius: BorderRadius.circular(9),
              child: Image.network(
                preference.artworkUrl!,
                width: 38,
                height: 38,
                fit: BoxFit.cover,
                errorBuilder: (_, _, _) => _preferenceFallback(preference.type),
              ),
            )
          else
            _preferenceFallback(preference.type),
          const SizedBox(width: 10),
          Expanded(
            child: Text(
              preference.name,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: const TextStyle(
                color: Color(0xFF294441),
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          IconButton(
            tooltip: 'Quitar ${preference.name}',
            onPressed: isPending ? null : () => _removePreference(preference),
            icon: isPending
                ? const SizedBox(
                    width: 18,
                    height: 18,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.close_rounded, size: 19),
            color: const Color(0xFF82918E),
          ),
        ],
      ),
    );
  }

  Widget _preferenceFallback(String type) {
    final icon = switch (type) {
      'artist' => Icons.person_rounded,
      'genre' => Icons.category_outlined,
      _ => Icons.music_note_rounded,
    };
    return Container(
      width: 38,
      height: 38,
      decoration: BoxDecoration(
        color: const Color(0xFFE8F7F3),
        borderRadius: BorderRadius.circular(9),
      ),
      child: Icon(icon, color: const Color(0xFF0D7771), size: 19),
    );
  }

  String _lowercaseTitle(String title) => switch (title) {
    'Canciones' => 'canciones',
    'Artistas' => 'artistas',
    _ => 'géneros',
  };

  Widget _buildLoadError() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(
              Icons.cloud_off_rounded,
              color: Color(0xFF0D7771),
              size: 44,
            ),
            const SizedBox(height: 12),
            const Text(
              'No pudimos cargar tus preferencias',
              style: TextStyle(
                color: Color(0xFF183D3B),
                fontSize: 16,
                fontWeight: FontWeight.w700,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 6),
            Text(
              _errorMessage ?? 'Intenta nuevamente.',
              style: const TextStyle(color: Color(0xFF788A86), fontSize: 13),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 12),
            OutlinedButton.icon(
              onPressed: _loadPreferences,
              icon: const Icon(Icons.refresh_rounded),
              label: const Text('Reintentar'),
            ),
          ],
        ),
      ),
    );
  }
}

class _GenreAutocompleteDialog extends StatefulWidget {
  final ApiClient apiClient;

  const _GenreAutocompleteDialog({required this.apiClient});

  @override
  State<_GenreAutocompleteDialog> createState() =>
      _GenreAutocompleteDialogState();
}

class _GenreAutocompleteDialogState extends State<_GenreAutocompleteDialog> {
  final TextEditingController _controller = TextEditingController();
  Timer? _debounce;
  CancelToken? _cancelToken;
  List<MusicTrack> _genres = const [];
  bool _isLoading = false;
  String? _errorMessage;

  @override
  void dispose() {
    _debounce?.cancel();
    _cancelToken?.cancel('El selector de géneros se cerró.');
    _controller.dispose();
    super.dispose();
  }

  void _onQueryChanged(String query) {
    _debounce?.cancel();
    _cancelToken?.cancel('Cambió la búsqueda de géneros.');
    if (query.trim().length < 2) {
      setState(() {
        _genres = const [];
        _isLoading = false;
        _errorMessage = null;
      });
      return;
    }
    _debounce = Timer(
      const Duration(milliseconds: 300),
      () => _searchGenres(query.trim()),
    );
  }

  Future<void> _searchGenres(String query) async {
    final token = CancelToken();
    _cancelToken = token;
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });
    try {
      final response = await widget.apiClient.searchMusicTracks(
        query,
        type: 'genre',
        cancelToken: token,
      );
      if (!mounted || token.isCancelled) return;
      setState(() {
        _genres = response.data?.items ?? const [];
        _isLoading = false;
      });
    } on DioException catch (error) {
      if (CancelToken.isCancel(error) || !mounted) return;
      setState(() {
        _errorMessage = error.message ?? 'No se pudieron cargar los géneros.';
        _isLoading = false;
      });
    } on ApiException catch (error) {
      if (!mounted || token.isCancelled) return;
      setState(() {
        _errorMessage = error.message;
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Agregar género'),
      content: SizedBox(
        width: 420,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: _controller,
              autofocus: true,
              onChanged: _onQueryChanged,
              decoration: InputDecoration(
                hintText: 'Escribe para buscar géneros',
                prefixIcon: const Icon(Icons.search_rounded),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(14),
                ),
              ),
            ),
            const SizedBox(height: 8),
            if (_isLoading) const LinearProgressIndicator(),
            if (_errorMessage != null)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Text(
                  _errorMessage!,
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              )
            else if (_controller.text.trim().length < 2)
              const Padding(
                padding: EdgeInsets.all(12),
                child: Text('Escribe al menos 2 caracteres.'),
              )
            else if (!_isLoading && _genres.isEmpty)
              const Padding(
                padding: EdgeInsets.all(12),
                child: Text('No encontramos géneros con ese nombre.'),
              )
            else
              Flexible(
                child: ListView.builder(
                  shrinkWrap: true,
                  itemCount: _genres.length,
                  itemBuilder: (context, index) {
                    final genre = _genres[index];
                    return ListTile(
                      leading: const Icon(
                        Icons.category_outlined,
                        color: Color(0xFF0D7771),
                      ),
                      title: Text(genre.title),
                      onTap: () => Navigator.pop(context, genre),
                    );
                  },
                ),
              ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Cancelar'),
        ),
      ],
    );
  }
}
