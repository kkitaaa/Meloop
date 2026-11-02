import 'package:flutter/material.dart';
import 'package:flutter_app/features/music/presentation/music_track_tile.dart';
import 'package:flutter_app/models/music_detail.dart';
import 'package:flutter_app/models/music_track.dart';
import 'package:flutter_app/services/api_client.dart';
import 'package:url_launcher/url_launcher.dart';

class MusicDetailScreen extends StatefulWidget {
  final String type;
  final String spotifyId;
  final ApiClient? apiClient;

  const MusicDetailScreen({
    super.key,
    required this.type,
    required this.spotifyId,
    this.apiClient,
  });

  @override
  State<MusicDetailScreen> createState() => _MusicDetailScreenState();
}

class _MusicDetailScreenState extends State<MusicDetailScreen> {
  late final ApiClient _apiClient;
  late final bool _ownsApiClient;
  MusicDetail? _detail;
  bool _isLoading = true;
  bool _isCheckingPreference = false;
  bool _isAddingPreference = false;
  bool _isInPreferences = false;
  String? _errorMessage;
  String? _preferenceError;

  String get _normalizedType => switch (widget.type) {
    'track' || 'tracks' => 'track',
    'artist' || 'artists' => 'artist',
    'album' || 'albums' => 'album',
    _ => widget.type,
  };

  bool get _supportsPreferences =>
      _normalizedType == 'track' || _normalizedType == 'artist';

  @override
  void initState() {
    super.initState();
    _apiClient = widget.apiClient ?? ApiClient();
    _ownsApiClient = widget.apiClient == null;
    _loadDetail();
  }

  @override
  void dispose() {
    if (_ownsApiClient) _apiClient.close();
    super.dispose();
  }

  Future<void> _loadDetail() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
      _preferenceError = null;
    });
    try {
      final response = await _apiClient.getMusicDetail(
        type: _normalizedType,
        spotifyId: widget.spotifyId,
      );
      final detail = response.data;
      if (detail == null) {
        throw const ApiException(
          message: 'El servicio de música no devolvió el detalle.',
        );
      }
      if (!mounted) return;
      setState(() {
        _detail = detail;
        _isLoading = false;
      });
      if (_supportsPreferences) await _loadPreferenceStatus();
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _errorMessage = _detailErrorMessage(error);
        _isLoading = false;
      });
    }
  }

  String _detailErrorMessage(ApiException error) {
    if (error.statusCode == 404) return 'Este contenido ya no está disponible';
    if (error.statusCode == 503 || error.errorCode == 'MUSIC_UNAVAILABLE') {
      return 'Spotify no está disponible. Intenta nuevamente en unos minutos.';
    }
    return error.message;
  }

  Future<void> _loadPreferenceStatus() async {
    setState(() {
      _isCheckingPreference = true;
      _preferenceError = null;
    });
    try {
      final response = await _apiClient.getMusicPreferences();
      final data = response.data;
      if (data == null) {
        throw const ApiException(
          message: 'No se pudo consultar el estado de tus preferencias.',
        );
      }
      final isInPreferences = switch (_normalizedType) {
        'track' => data.songs.any((item) => item.id == widget.spotifyId),
        'artist' => data.artists.any((item) => item.id == widget.spotifyId),
        _ => false,
      };
      if (!mounted) return;
      setState(() {
        _isInPreferences = isInPreferences;
        _isCheckingPreference = false;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() {
        _preferenceError = error.message;
        _isCheckingPreference = false;
      });
    }
  }

  Future<void> _addToPreferences() async {
    final detail = _detail;
    if (detail == null || !_supportsPreferences || _isAddingPreference) return;
    setState(() => _isAddingPreference = true);
    try {
      await _apiClient.addMusicPreference(
        type: _normalizedType,
        itemId: detail.id,
        name: detail.name,
        artworkUrl: detail.artworkUrl,
      );
      if (!mounted) return;
      setState(() {
        _isInPreferences = true;
        _isAddingPreference = false;
        _preferenceError = null;
      });
    } on ApiException catch (error) {
      if (!mounted) return;
      setState(() => _isAddingPreference = false);
      if (error.errorCode == 'ALREADY_EXISTS') {
        setState(() => _isInPreferences = true);
        _showMessage('Ya está en tus preferencias');
      } else if (error.errorCode == 'PREFERENCE_LIMIT') {
        _showMessage('Llegaste al límite de 30');
      } else {
        _showMessage(error.message);
      }
    }
  }

  Future<void> _openInSpotify() async {
    final url = _detail?.externalUrl;
    final uri = url == null ? null : Uri.tryParse(url);
    if (uri == null || !_isSupportedSpotifyUrl(uri)) {
      _showMessage('No hay un enlace válido de Spotify para este contenido.');
      return;
    }
    try {
      final launched = await launchUrl(
        uri,
        mode: LaunchMode.externalApplication,
      );
      if (!launched && mounted) {
        _showMessage('No se pudo abrir Spotify en este dispositivo.');
      }
    } on Exception catch (error) {
      if (mounted) _showMessage('No se pudo abrir Spotify: $error');
    }
  }

  bool _isSupportedSpotifyUrl(Uri uri) {
    final host = uri.host.toLowerCase();
    return uri.scheme == 'https' &&
        (host == 'spotify.com' ||
            host.endsWith('.spotify.com') ||
            host == 'spotify.link');
  }

  void _showMessage(String message) {
    ScaffoldMessenger.of(context)
      ..hideCurrentSnackBar()
      ..showSnackBar(SnackBar(content: Text(message)));
  }

  String get _title => switch (_normalizedType) {
    'artist' => 'Ficha del artista',
    'album' => 'Ficha del álbum',
    _ => 'Ficha musical',
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
              _title,
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
              child: _isLoading
                  ? _buildSkeleton()
                  : _errorMessage != null
                  ? _buildError()
                  : _buildContent(_detail!),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSkeleton() {
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Align(
          child: Container(
            width: 220,
            height: 220,
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(22),
            ),
          ),
        ),
        const SizedBox(height: 24),
        Container(
          height: 22,
          width: 210,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
          ),
        ),
        const SizedBox(height: 12),
        Container(
          height: 16,
          width: 150,
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(10),
          ),
        ),
        const SizedBox(height: 30),
        Container(
          height: 54,
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
          ),
        ),
      ],
    );
  }

  Widget _buildError() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: _StateCard(
          icon: Icons.cloud_off_rounded,
          title: 'No pudimos cargar la ficha',
          message: _errorMessage!,
          action: OutlinedButton.icon(
            onPressed: _loadDetail,
            icon: const Icon(Icons.refresh_rounded),
            label: const Text('Reintentar'),
          ),
        ),
      ),
    );
  }

  Widget _buildContent(MusicDetail detail) {
    return ListView(
      physics: const AlwaysScrollableScrollPhysics(),
      padding: const EdgeInsets.fromLTRB(20, 24, 20, 32),
      children: [
        Center(child: _buildArtwork(detail)),
        const SizedBox(height: 22),
        Text(
          detail.name,
          textAlign: TextAlign.center,
          style: const TextStyle(
            color: Color(0xFF183D3B),
            fontSize: 24,
            fontWeight: FontWeight.w800,
          ),
        ),
        if (detail.artist?.isNotEmpty == true) ...[
          const SizedBox(height: 7),
          Text(
            detail.artist!,
            textAlign: TextAlign.center,
            style: const TextStyle(color: Color(0xFF70827F), fontSize: 15),
          ),
        ],
        if (detail.album?.isNotEmpty == true) ...[
          const SizedBox(height: 5),
          Text(
            detail.album!,
            textAlign: TextAlign.center,
            style: const TextStyle(color: Color(0xFF8A9B98), fontSize: 13),
          ),
        ],
        if (detail.genres.isNotEmpty) ...[
          const SizedBox(height: 18),
          Wrap(
            alignment: WrapAlignment.center,
            spacing: 8,
            runSpacing: 8,
            children: [
              for (final genre in detail.genres)
                Chip(
                  label: Text(genre),
                  backgroundColor: const Color(0xFFE8F7F3),
                  side: const BorderSide(color: Color(0xFFDCEEE9)),
                  labelStyle: const TextStyle(
                    color: Color(0xFF0D5C5E),
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                  ),
                ),
            ],
          ),
        ],
        if (_supportsPreferences) ...[
          const SizedBox(height: 24),
          if (_preferenceError != null)
            Padding(
              padding: const EdgeInsets.only(bottom: 10),
              child: TextButton.icon(
                onPressed: _loadPreferenceStatus,
                icon: const Icon(Icons.refresh_rounded, size: 18),
                label: Text(
                  'No pudimos consultar tus preferencias. Reintentar',
                ),
              ),
            ),
          SizedBox(
            height: 50,
            child: FilledButton.icon(
              onPressed:
                  _isInPreferences ||
                      _isCheckingPreference ||
                      _isAddingPreference
                  ? null
                  : _addToPreferences,
              icon: _isCheckingPreference || _isAddingPreference
                  ? const SizedBox(
                      width: 18,
                      height: 18,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        color: Colors.white,
                      ),
                    )
                  : Icon(
                      _isInPreferences
                          ? Icons.check_rounded
                          : Icons.library_add_rounded,
                    ),
              label: Text(
                _isInPreferences
                    ? 'En tus preferencias'
                    : 'Agregar a mis preferencias',
              ),
              style: FilledButton.styleFrom(
                backgroundColor: const Color(0xFF0D7771),
                disabledBackgroundColor: const Color(0xFF9BC8C0),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(15),
                ),
              ),
            ),
          ),
        ],
        const SizedBox(height: 12),
        SizedBox(
          height: 50,
          child: OutlinedButton.icon(
            onPressed: detail.externalUrl == null ? null : _openInSpotify,
            icon: const Icon(Icons.open_in_new_rounded),
            label: const Text('Abrir en Spotify'),
            style: OutlinedButton.styleFrom(
              foregroundColor: const Color(0xFF0D5C5E),
              side: const BorderSide(color: Color(0xFFBDE9E1)),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(15),
              ),
            ),
          ),
        ),
        if (_normalizedType == 'album' && detail.tracks.isNotEmpty) ...[
          const SizedBox(height: 28),
          const Text(
            'Canciones del álbum',
            style: TextStyle(
              color: Color(0xFF183D3B),
              fontSize: 17,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(height: 6),
          for (final track in detail.tracks)
            MusicTrackTile(track: track, onTap: () => _openNestedDetail(track)),
        ],
      ],
    );
  }

  Widget _buildArtwork(MusicDetail detail) {
    final image = detail.artworkUrl;
    return ClipRRect(
      borderRadius: BorderRadius.circular(22),
      child: SizedBox(
        width: 220,
        height: 220,
        child: image == null
            ? _artworkFallback()
            : Image.network(
                image,
                fit: BoxFit.cover,
                errorBuilder: (_, _, _) => _artworkFallback(),
              ),
      ),
    );
  }

  Widget _artworkFallback() {
    final icon = switch (_normalizedType) {
      'artist' => Icons.person_rounded,
      'album' => Icons.album_rounded,
      _ => Icons.music_note_rounded,
    };
    return ColoredBox(
      color: const Color(0xFFE8F7F3),
      child: Icon(icon, size: 94, color: const Color(0xFF0D7771)),
    );
  }

  void _openNestedDetail(MusicTrack track) {
    Navigator.push(
      context,
      MaterialPageRoute<void>(
        builder: (_) => MusicDetailScreen(
          type: 'track',
          spotifyId: track.id,
          apiClient: _apiClient,
        ),
      ),
    );
  }
}

class MusicTrackDetailScreen extends StatelessWidget {
  final MusicTrack track;
  final ApiClient? apiClient;

  const MusicTrackDetailScreen({
    super.key,
    required this.track,
    this.apiClient,
  });

  @override
  Widget build(BuildContext context) {
    return MusicDetailScreen(
      type: track.type,
      spotifyId: track.id,
      apiClient: apiClient,
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
    return Container(
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
    );
  }
}
